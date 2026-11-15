package alerts

import (
	"strconv"

	"server/configs"

	"github.com/quollix/common/quollix/api"
	u "github.com/quollix/common/utils"
)

type AlertSettingsRepository interface {
	Read() (*api.AlertSettings, error)
	Write(*api.AlertSettings) error
}

type AlertSettingsRepositoryImpl struct {
	ConfigsRepository configs.ConfigsRepository
}

func (r *AlertSettingsRepositoryImpl) Read() (*api.AlertSettings, error) {
	settings := &api.AlertSettings{}
	values := []struct {
		key    string
		target *bool
	}{
		{configs.ConfigKeys.AlertCPUUsageEnabled, &settings.CPUUsageEnabled},
		{configs.ConfigKeys.AlertMemoryUsageEnabled, &settings.MemoryUsageEnabled},
		{configs.ConfigKeys.AlertDiskUsageEnabled, &settings.DiskUsageEnabled},
		{configs.ConfigKeys.AlertCertificateExpiryEnabled, &settings.CertificateExpiryEnabled},
		{configs.ConfigKeys.AlertQuollixErrorsEnabled, &settings.QuollixErrorsEnabled},
	}
	for _, value := range values {
		encoded, err := r.ConfigsRepository.GetConfig(value.key)
		if err != nil {
			return nil, err
		}
		parsed, err := strconv.ParseBool(encoded)
		if err != nil {
			return nil, u.Logger.NewError(err.Error(), "config_key", value.key)
		}
		*value.target = parsed
	}
	integerValues := []struct {
		key    string
		target *int
	}{
		{configs.ConfigKeys.AlertHostUsageTriggerPercent, &settings.HostUsageTriggerPercent},
		{configs.ConfigKeys.AlertHostUsageRecoveryPercent, &settings.HostUsageRecoveryPercent},
		{configs.ConfigKeys.AlertHostUsagePendingMinutes, &settings.HostUsagePendingMinutes},
	}
	for _, value := range integerValues {
		encoded, err := r.ConfigsRepository.GetConfig(value.key)
		if err != nil {
			return nil, err
		}
		parsed, err := strconv.Atoi(encoded)
		if err != nil {
			return nil, u.Logger.NewError(err.Error(), "config_key", value.key)
		}
		*value.target = parsed
	}
	return settings, nil
}

func (r *AlertSettingsRepositoryImpl) Write(settings *api.AlertSettings) error {
	values := []struct {
		key   string
		value string
	}{
		{configs.ConfigKeys.AlertCPUUsageEnabled, strconv.FormatBool(settings.CPUUsageEnabled)},
		{configs.ConfigKeys.AlertMemoryUsageEnabled, strconv.FormatBool(settings.MemoryUsageEnabled)},
		{configs.ConfigKeys.AlertDiskUsageEnabled, strconv.FormatBool(settings.DiskUsageEnabled)},
		{configs.ConfigKeys.AlertCertificateExpiryEnabled, strconv.FormatBool(settings.CertificateExpiryEnabled)},
		{configs.ConfigKeys.AlertQuollixErrorsEnabled, strconv.FormatBool(settings.QuollixErrorsEnabled)},
		{configs.ConfigKeys.AlertHostUsageTriggerPercent, strconv.Itoa(settings.HostUsageTriggerPercent)},
		{configs.ConfigKeys.AlertHostUsageRecoveryPercent, strconv.Itoa(settings.HostUsageRecoveryPercent)},
		{configs.ConfigKeys.AlertHostUsagePendingMinutes, strconv.Itoa(settings.HostUsagePendingMinutes)},
	}
	for _, value := range values {
		if err := r.ConfigsRepository.SetConfig(value.key, value.value); err != nil {
			return err
		}
	}
	return nil
}
