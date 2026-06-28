package maintenance

import (
	"net/http"
	"server/configs"
	"server/maintenance/retention"

	u "github.com/quollix/common/utils"
	"github.com/quollix/common/validation"
)

var (
	InvalidMaintenanceWindowStartHourErrorMessage = "Maintenance window start hour must be between 0 and 23"
	InvalidIanaTimezoneErrorMessage               = "Invalid time zone"
	expectedMaintenanceConfigSaveErrors           = u.MapOf(InvalidIanaTimezoneErrorMessage, InvalidMaintenanceWindowStartHourErrorMessage)

	NegativeRetentionValuesErrors     = "Retention policy values must be non-negative"
	expectedRetentionPolicySaveErrors = u.MapOf(NegativeRetentionValuesErrors)
)

type MaintenanceConfigsHandler struct {
	MaintenanceRepository     configs.MaintenanceRepository
	RetentionPolicyRepository retention.RetentionPolicyRepository
	MaintenanceAgent          MaintenanceAgent
	MaintenanceService        MaintenanceService
}

type MaintenanceConfigDto struct {
	IanaTimezone               string `json:"iana_timezone" validate:"ignore"` // "ignore" because we will do custom validation for this field in the handler
	MaintenanceWindowStartHour int    `json:"maintenance_window_start_hour"`
}

func (s *MaintenanceConfigsHandler) SaveMaintenanceConfigs(w http.ResponseWriter, r *http.Request) {
	config, ok := validation.ReadBody[MaintenanceConfigDto](w, r)
	if !ok {
		return
	}

	err := s.MaintenanceService.SaveMaintenanceConfig(config.IanaTimezone, config.MaintenanceWindowStartHour)
	if err != nil {
		u.WriteResponseError(w, expectedMaintenanceConfigSaveErrors, err)
		return
	}
}

// ReadMaintenanceConfigs only used for component tests
func (s *MaintenanceConfigsHandler) ReadMaintenanceConfigs(w http.ResponseWriter, r *http.Request) {
	databaseMaintenanceConfig, err := s.MaintenanceRepository.GetMaintenanceConfig()
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}
	u.SendJsonResponse(w, databaseMaintenanceConfig)
}

// ReadRetentionPolicyHandler only used for component tests
func (s *MaintenanceConfigsHandler) ReadRetentionPolicyHandler(w http.ResponseWriter, r *http.Request) {
	policy, err := s.RetentionPolicyRepository.GetRetentionPolicy()
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}
	u.SendJsonResponse(w, policy)
}

func (s *MaintenanceConfigsHandler) SaveRetentionPolicyHandler(w http.ResponseWriter, r *http.Request) {
	policy, ok := validation.ReadBody[retention.RetentionPolicy](w, r)
	if !ok {
		return
	}

	if policy.KeepPreUpdate < 0 || policy.KeepDaily < 0 || policy.KeepWeekly < 0 || policy.KeepMonthly < 0 || policy.KeepYearly < 0 {
		u.WriteResponseError(w, expectedRetentionPolicySaveErrors, u.Logger.NewError(NegativeRetentionValuesErrors))
		return
	}

	err := s.RetentionPolicyRepository.SetRetentionPolicy(policy)
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}
}

func (s *MaintenanceConfigsHandler) RunMaintenanceJobHandler(w http.ResponseWriter, r *http.Request) {
	s.MaintenanceAgent.RunMaintenanceJob()
}
