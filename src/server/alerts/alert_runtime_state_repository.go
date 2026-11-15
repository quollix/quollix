package alerts

import (
	"time"

	"server/configs"

	u "github.com/quollix/common/utils"
)

type AlertStatus string

const (
	AlertStatusHealthy AlertStatus = "healthy"
	AlertStatusPending AlertStatus = "pending"
	AlertStatusFiring  AlertStatus = "firing"
)

type AlertRuntimeState struct {
	Status       AlertStatus
	PendingSince time.Time
}

type RuntimeState struct {
	CPUUsage          AlertRuntimeState
	MemoryUsage       AlertRuntimeState
	DiskUsage         AlertRuntimeState
	CertificateExpiry AlertRuntimeState
	QuollixErrors     AlertRuntimeState
}

type StateRepository interface {
	Read() (*RuntimeState, error)
	Save(*RuntimeState) error
}

type StateRepositoryImpl struct {
	ConfigsRepository configs.ConfigsRepository
}

func DefaultRuntimeState() *RuntimeState {
	healthy := AlertRuntimeState{Status: AlertStatusHealthy}
	return &RuntimeState{
		CPUUsage:          healthy,
		MemoryUsage:       healthy,
		DiskUsage:         healthy,
		CertificateExpiry: healthy,
		QuollixErrors:     healthy,
	}
}

func (r *StateRepositoryImpl) Read() (*RuntimeState, error) {
	state := &RuntimeState{}
	for _, field := range alertStateFields(state) {
		encodedStatus, err := r.ConfigsRepository.GetConfig(field.statusKey)
		if err != nil {
			return nil, err
		}
		field.state.Status = AlertStatus(encodedStatus)
		if !isValidAlertStatus(field.state.Status) || (field.state.Status == AlertStatusPending && field.pendingSinceKey == "") {
			return nil, u.Logger.NewError("invalid alert runtime state", "config_key", field.statusKey, "value", encodedStatus)
		}
		if field.pendingSinceKey == "" {
			continue
		}
		encodedPendingSince, err := r.ConfigsRepository.GetConfig(field.pendingSinceKey)
		if err != nil {
			return nil, err
		}
		if encodedPendingSince != "" {
			field.state.PendingSince, err = time.Parse(time.RFC3339Nano, encodedPendingSince)
			if err != nil {
				return nil, u.Logger.NewError(err.Error(), "config_key", field.pendingSinceKey)
			}
		}
		if (field.state.Status == AlertStatusPending) != !field.state.PendingSince.IsZero() {
			return nil, u.Logger.NewError("invalid alert pending time", "config_key", field.pendingSinceKey)
		}
	}
	return state, nil
}

func (r *StateRepositoryImpl) Save(state *RuntimeState) error {
	for _, field := range alertStateFields(state) {
		if !isValidAlertStatus(field.state.Status) || (field.state.Status == AlertStatusPending && field.pendingSinceKey == "") {
			return u.Logger.NewError("invalid alert runtime state", "value", field.state.Status)
		}
		if field.pendingSinceKey != "" && (field.state.Status == AlertStatusPending) != !field.state.PendingSince.IsZero() {
			return u.Logger.NewError("invalid alert pending time", "config_key", field.pendingSinceKey)
		}
		if err := r.ConfigsRepository.SetConfig(field.statusKey, string(field.state.Status)); err != nil {
			return err
		}
		if field.pendingSinceKey == "" {
			continue
		}
		pendingSince := ""
		if !field.state.PendingSince.IsZero() {
			pendingSince = field.state.PendingSince.UTC().Format(time.RFC3339Nano)
		}
		if err := r.ConfigsRepository.SetConfig(field.pendingSinceKey, pendingSince); err != nil {
			return err
		}
	}
	return nil
}

type alertStateField struct {
	statusKey       string
	pendingSinceKey string
	state           *AlertRuntimeState
}

func alertStateFields(state *RuntimeState) []alertStateField {
	return []alertStateField{
		{configs.ConfigKeys.AlertCPUUsageState, configs.ConfigKeys.AlertCPUUsagePendingSince, &state.CPUUsage},
		{configs.ConfigKeys.AlertMemoryUsageState, configs.ConfigKeys.AlertMemoryUsagePendingSince, &state.MemoryUsage},
		{configs.ConfigKeys.AlertDiskUsageState, configs.ConfigKeys.AlertDiskUsagePendingSince, &state.DiskUsage},
		{statusKey: configs.ConfigKeys.AlertCertificateExpiryState, state: &state.CertificateExpiry},
		{statusKey: configs.ConfigKeys.AlertQuollixErrorsState, state: &state.QuollixErrors},
	}
}

func isValidAlertStatus(status AlertStatus) bool {
	return status == AlertStatusHealthy || status == AlertStatusPending || status == AlertStatusFiring
}
