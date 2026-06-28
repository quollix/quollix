package system_config_migrations

import "server/configs"

func (s *SystemConfigMigrationsProviderImpl) renameServerHostConfigToBaseDomain() error {
	legacyServerHost, err := s.ConfigsRepo.GetConfig(legacyServerHostConfigKey)
	if err != nil {
		return err
	}

	if err := s.ConfigsRepo.SetConfig(configs.ConfigKeys.BaseDomain, legacyServerHost); err != nil {
		return err
	}
	return s.ConfigsRepo.DeleteConfig(legacyServerHostConfigKey)
}
