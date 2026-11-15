package system_config_migrations

import (
	"server/configs"
	"server/email"
	"server/users"

	u "github.com/quollix/common/utils"
)

func (s *SystemConfigMigrationsProviderImpl) createBroadcastEmailTemplate() error {
	u.Logger.Info(systemConfigMigrationLogMessage, "step", 6, "action", "create broadcast email template")
	if err := email.ValidateBroadcastEmailTemplate(users.DefaultBroadcastEmailTemplate); err != nil {
		return u.Logger.NewError("default broadcast email template is invalid", "error", err.Error())
	}
	return s.ConfigsRepo.SetConfig(configs.ConfigKeys.BroadcastEmailTemplate, users.DefaultBroadcastEmailTemplate)
}
