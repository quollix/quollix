package alerts

import (
	"github.com/quollix/common/quollix/api"
	u "github.com/quollix/common/utils"
)

const InvalidAlertSettingsError = "Host usage thresholds must be integers from 10 to 100, recovery must be lower than trigger, and pending time must be zero or greater"

type AlertSettingsService interface {
	Save(*api.AlertSettings) error
}

type AlertSettingsServiceImpl struct {
	Repository AlertSettingsRepository
}

func (s *AlertSettingsServiceImpl) Save(settings *api.AlertSettings) error {
	if settings.HostUsageTriggerPercent < minimumHostUsagePercent || settings.HostUsageTriggerPercent > maximumHostUsagePercent ||
		settings.HostUsageRecoveryPercent < minimumHostUsagePercent || settings.HostUsageRecoveryPercent > maximumHostUsagePercent ||
		settings.HostUsageRecoveryPercent >= settings.HostUsageTriggerPercent || settings.HostUsagePendingMinutes < minimumHostUsagePendingMinutes {
		return u.Logger.NewError(InvalidAlertSettingsError)
	}
	return s.Repository.Write(settings)
}
