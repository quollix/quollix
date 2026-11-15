package system_config_migrations

import (
	"server/configs"
	"server/tools"

	u "github.com/quollix/common/utils"
)

func (s *SystemConfigMigrationsProviderImpl) createBackupFolderPath() error {
	u.Logger.Info(systemConfigMigrationLogMessage, "step", 4, "action", "create backup folder path")
	isBackupServerEnabled, err := s.ConfigsRepo.GetConfig(configs.ConfigKeys.BackupEnabled)
	if err != nil {
		return err
	}

	path := ""
	if isBackupServerEnabled == "true" {
		path = tools.LegacyRemoteBackupFolderPath
	}
	return s.ConfigsRepo.SetConfig(configs.ConfigKeys.BackupFolderPath, path)
}
