package system_config_migrations

import (
	"server/alerts"

	"github.com/quollix/common/utils"
)

func (s *SystemConfigMigrationsProviderImpl) createAlertConfiguration() error {
	utils.Logger.Info(systemConfigMigrationLogMessage, "step", 5, "action", "create alert settings and runtime state")
	if err := s.AlertSettingsRepository.Write(alerts.DefaultAlertSettings()); err != nil {
		return err
	}
	return s.AlertStateRepository.Save(alerts.DefaultRuntimeState())
}
