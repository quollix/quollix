package system_config_migrations

import (
	"server/apps_basic"
	"server/backup_server"
	"server/certificates"
	"server/configs"
	"server/maintenance"
	"server/maintenance/retention"
	"server/tools"
	"server/users"
	"time"

	u "github.com/quollix/common/utils"
)

var (
	PostgresVersion            = "17.5"
	SamplePostgresCreationDate = time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
)

const (
	DefaultInitialAdminName           = "administrator"
	GeneratedInitialAdminPasswordSize = 20
	InitialAdminPasswordEnvVar        = "INITIAL_ADMIN_PASSWORD"
	InitialAdminNameEnvVar            = "INITIAL_ADMIN_NAME"
	legacyServerHostConfigKey         = "server_host"
)

type SystemConfigMigrationsProviderImpl struct {
	UserRepo                   users.UserRepository
	AppService                 apps_basic.AppService
	AppRepo                    apps_basic.AppRepository
	DirectoryProvider          tools.DirectoryProvider
	AuthHelper                 u.AuthHelper
	ClientCredentialsGenerator apps_basic.ClientCredentialsGenerator
	ConfigsRepo                configs.ConfigsRepository
	OidcEmailService           configs.OidcEmailExposureService
	CertificatePersister       certificates.CertificatePersister
	EmailRepo                  configs.EmailRepository
	SshRepository              backup_server.SshRepository
	RetentionPolicyRepository  retention.RetentionPolicyRepository
	MaintenanceServiceHelper   maintenance.AgentHelper
	MaintenanceRepo            configs.MaintenanceRepository
	OsWrapper                  u.OsWrapper
	CertificateService         certificates.CertificateService
}

func (s *SystemConfigMigrationsProviderImpl) List() []func() error {
	// Append only. The stored system_config_version is the 1-based index in this slice.
	return []func() error{
		s.runInitialBootstrapMigration,
		s.renameServerHostConfigToBaseDomain,
	}
}
