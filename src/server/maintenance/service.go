package maintenance

import (
	"server/configs"
	"server/tools"

	u "github.com/quollix/common/utils"
)

type MaintenanceService interface {
	SaveMaintenanceConfig(timezone string, hour int) error
}

type MaintenanceServiceImpl struct {
	MaintenanceRepository configs.MaintenanceRepository
	OsWrapper             u.OsWrapper
	TimezoneProvider      tools.TimezoneProvider
	AgentHelper           AgentHelper
}

func (m *MaintenanceServiceImpl) SaveMaintenanceConfig(timezone string, hour int) error {
	if hour < 0 || hour > 23 {
		return u.Logger.NewError(InvalidMaintenanceWindowStartHourErrorMessage)
	}
	if !m.TimezoneProvider.IsIanaTimezoneValid(timezone) {
		return u.Logger.NewError(InvalidIanaTimezoneErrorMessage)
	}

	databaseMaintenanceConfig, err := m.MaintenanceRepository.GetMaintenanceConfig()
	if err != nil {
		return err
	}

	newMaintenanceTime, err := m.AgentHelper.CalculateNextMaintenanceAtUtc(m.OsWrapper.Now(), timezone, hour)
	if err != nil {
		return err
	}

	databaseMaintenanceConfig.NextMaintenanceAt = *newMaintenanceTime
	databaseMaintenanceConfig.IanaTimezone = timezone
	databaseMaintenanceConfig.MaintenanceWindowStartHour = hour

	return m.MaintenanceRepository.SetMaintenanceConfig(databaseMaintenanceConfig)
}
