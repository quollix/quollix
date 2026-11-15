package system_config_migrations

import (
	"server/configs"

	u "github.com/quollix/common/utils"
)

func (s *SystemConfigMigrationsProviderImpl) renameServerHostConfigToBaseDomain() error {
	u.Logger.Info(systemConfigMigrationLogMessage, "step", 2, "action", "rename server host to base domain")
	legacyServerHost, err := s.ConfigsRepo.GetConfig(legacyServerHostConfigKey)
	if err != nil {
		return err
	}

	if err := s.ConfigsRepo.SetConfig(configs.ConfigKeys.BaseDomain, legacyServerHost); err != nil {
		return err
	}
	return s.ConfigsRepo.DeleteConfig(legacyServerHostConfigKey)
}
