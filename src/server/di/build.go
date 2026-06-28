package di

import (
	"io/fs"

	"server/app_store"
	"server/apps_advanced"
	"server/apps_basic"
	"server/backup_server"
	"server/backups"
	certificates2 "server/certificates"
	"server/configs"
	"server/email"
	"server/frontend"
	"server/frontend/assets"
	frontendpages "server/frontend/pages"
	"server/frontend/renderer"
	"server/groups"
	"server/ingress"
	"server/maintenance"
	"server/maintenance/retention"
	"server/oidc_client"
	"server/oidc_provider"
	"server/setup"
	"server/system_config_migrations"
	"server/terminal"
	"server/tools"
	"server/users"

	"github.com/go-chi/chi/v5"
	"github.com/google/wire"
	"github.com/moby/moby/client"
	"github.com/quollix/common/store"
	u "github.com/quollix/common/utils"
	"github.com/quollix/common/validation"
	"github.com/quollix/deepstack"
)

type Dependencies struct {
	MaintenanceAgent            maintenance.MaintenanceAgent
	ResticDockerImageService    backup_server.ResticDockerImageService
	SystemConfigMigrator        system_config_migrations.SystemConfigMigrator
	DatabaseSnapshotRepository  u.DatabaseSnapshotRepository
	HandlerRegisterer           *setup.HandlerRegisterer
	AppManager                  apps_basic.AppService
	ServerListener              *ingress.ServerListener
	DatabaseConnector           tools.DatabaseConnector
	CliToolInstallationVerifier setup.CliToolInstallationVerifier
	DirectoryProvider           tools.DirectoryProvider
	AppStoreClientLean          app_store.AppStoreClientLean
	Config                      *tools.GlobalConfig
	TemplateService             renderer.TemplateService
	OidcService                 oidc_provider.OidcService
	ConfigRepo                  configs.ConfigsRepository
	TimezoneProvider            tools.TimezoneProvider
	ShutdownHandler             *ingress.ShutdownHandlerImpl
	AcmeClient                  certificates2.AcmeClient
	GroupHandler                *groups.GroupHandler
	EmailHandler                *email.EmailHandler
	TerminalHandler             *terminal.TerminalHandlerImpl
	Logger                      deepstack.DeepStackLogger

	Repos *Repositories
}

type Repositories struct {
	UserRepo             *users.UserRepositoryImpl
	SessionRepo          *users.SessionRepositoryImpl
	AppsRepo             *apps_basic.AppRepositoryImpl
	ConfigsRepo          *configs.ConfigsRepositoryImpl
	SshRepo              *backup_server.SshRepositoryImpl
	EmailRepo            *configs.EmailRepoImpl
	RetentionPolicyRepo  *retention.RetentionPolicyRepositoryImpl
	MaintenanceRepo      *configs.MaintenanceRepositoryImpl
	GroupsRepo           *groups.GroupRepositoryImpl
	OidcAuthProviderRepo *oidc_client.OidcAuthProviderRepositoryImpl
	UserAuthMethodRepo   *oidc_client.UserAuthMethodRepositoryImpl
	OidcRelyingPartyRepo *oidc_provider.OidcRelyingPartyRepositoryImpl
}

var SharedSet = wire.NewSet(
	NewAppStoreClient,
	NewTrustedAuthorizedKey,
	NewRouter,
	NewGlobalConfig,
	GetLogger,
	NewWildcardCertificateService,
	validation.NewVersionValidator,
	oidc_provider.NewOidcCache,
	oidc_client.NewOidcLoginStateCache,
	oidc_client.NewOidcProviderClient,
	NewEmailClient,
	NewDockerClient,
	NewDatabaseSnapshotRepo,
	wire.InterfaceValue(new(fs.FS), tools.FrontendResourceFilesystem),
	wire.Value(map[string][]byte{}),

	wire.Struct(new(apps_basic.OperationRegistryImpl)),
	wire.Struct(new(ingress.AppRequestPolicyImpl), "*"),
	wire.Struct(new(setup.CliToolInstallationVerifier), "*"),
	wire.Struct(new(tools.DockerServiceImpl), "*"),
	wire.Struct(new(tools.TimezoneProviderImpl)),
	tools.NewDirectoryProvider,
	tools.NewDatabaseConnector,
	wire.Struct(new(users.UserRepositoryImpl), "*"),
	wire.Struct(new(users.SessionRepositoryImpl), "*"),
	wire.Struct(new(users.SessionServiceImpl), "*"),
	wire.Struct(new(apps_basic.AppRequestParserImpl)),
	wire.Struct(new(apps_basic.AppReverseProxyFactoryImpl), "*"),
	wire.Struct(new(apps_basic.AppSessionServiceImpl), "*"),
	wire.Struct(new(apps_basic.AppRequestProxy), "*"),
	wire.Struct(new(users.SecretAndCookieStorageImpl), "*"),
	wire.Struct(new(users.UserServiceImpl), "*"),
	wire.Struct(new(certificates2.CertificateCacheImpl)),
	wire.Struct(new(ingress.ServerListener), "*"),
	wire.Struct(new(app_store.AppStoreHandler), "*"),
	wire.Struct(new(app_store.VersionVerifierImpl), "*"),
	wire.Struct(new(configs.ConfigsRepositoryImpl), "*"),
	wire.Struct(new(configs.ConfigsServiceImpl), "*"),
	wire.Struct(new(configs.OidcEmailExposureServiceImpl), "*"),
	wire.Struct(new(configs.SettingsHandler), "*"),
	wire.Struct(new(apps_basic.AppRepositoryImpl), "*"),
	wire.Struct(new(users.RouteRegisterer), "*"),
	wire.Struct(new(backup_server.SshClientImpl), "*"),
	wire.Struct(new(backup_server.ResticDockerImageServiceImpl), "*"),
	wire.Struct(new(backup_server.SshHandler), "*"),
	wire.Struct(new(backups.BackupServiceImpl), "*"),
	wire.Struct(new(backups.BackupHandler), "*"),
	wire.Struct(new(setup.TestStateResetHandler), "*"),
	wire.Struct(new(apps_basic.AppServiceImpl), "*"),
	wire.Struct(new(apps_basic.AppsHandler), "*"),
	wire.Struct(new(users.AuthenticationServiceImpl), "*"),
	wire.Struct(new(system_config_migrations.SystemConfigMigrationsProviderImpl), "*"),
	wire.Struct(new(system_config_migrations.SystemConfigMigratorImpl), "*"),
	wire.Struct(new(app_store.AppStoreServiceImpl), "*"),
	wire.Struct(new(users.UserHandler), "*"),
	wire.Struct(new(certificates2.CertificateHandler), "*"),
	wire.Struct(new(ingress.SetupHandler), "*"),
	wire.Struct(new(setup.HandlerRegisterer), "*"),
	wire.Struct(new(tools.CommandRunnerImpl), "*"),
	wire.Struct(new(tools.ResticContainerExecutorImpl), "*"),
	wire.Struct(new(frontend.TemplateHandlerImpl), "*"),
	wire.Struct(new(frontend.BackedUpAppsLoaderHandler), "*"),
	wire.Struct(new(frontend.BackupsPageLoaderHandler), "*"),
	wire.Struct(new(frontendpages.PageRendererImpl), "*"),
	wire.Struct(new(backup_server.SshRepositoryServiceImpl), "*"),
	wire.Struct(new(backup_server.SshRepositoryImpl), "*"),
	wire.Struct(new(oidc_provider.OidcHandler), "*"),
	wire.Struct(new(backup_server.ResticServiceImpl), "*"),
	wire.Struct(new(backups.SnapshotPackagerImpl), "*"),
	wire.Struct(new(backups.MetaCodecImpl), "*"),
	wire.Struct(new(apps_basic.ComposeExtractorImpl), "*"),
	wire.Struct(new(configs.EmailRepoImpl), "*"),
	wire.Struct(new(certificates2.CertificatePersisterImpl), "*"),
	wire.Struct(new(backup_server.ResticSnapshotsParserImpl), "*"),
	wire.Struct(new(u.OsWrapperImpl)),
	wire.Struct(new(backups.BackupCreationAssemblerImpl), "*"),
	wire.Struct(new(backups.BackupQueryServiceImpl), "*"),
	wire.Struct(new(backups.RestoreFinalizerImpl), "*"),
	wire.Struct(new(u.AuthHelperImpl)),
	wire.Struct(new(u.BytesSignerImpl)),
	wire.Struct(new(apps_basic.ClientCredentialsGeneratorImpl), "*"),
	wire.Struct(new(apps_basic.AuthorizerImpl), "*"),
	wire.Struct(new(apps_basic.AppServiceHelperImpl), "*"),
	wire.Struct(new(apps_basic.AppDetectorImpl), "*"),
	wire.Struct(new(assets.AssetStoreImpl), "*"),
	wire.Struct(new(assets.AssetTagBuilderImpl), "*"),
	wire.Struct(new(renderer.TemplateEngineImpl), "*"),
	wire.Struct(new(renderer.TemplateServiceImpl), "*"),
	wire.Struct(new(frontend.FrontendPageDataBuilderImpl), "*"),
	wire.Struct(new(oidc_provider.IdTokenIssuerImpl), "*"),
	wire.Struct(new(oidc_provider.OidcServiceImpl), "*"),
	wire.Struct(new(oidc_provider.TokenIssuerImpl), "*"),
	wire.Struct(new(oidc_provider.OidcRelyingPartyResolverImpl), "*"),
	wire.Struct(new(oidc_provider.OidcClientServiceImpl), "*"),
	wire.Struct(new(oidc_provider.OidcRelyingPartyRepositoryImpl), "*"),
	wire.Struct(new(oidc_provider.OidcRelyingPartyServiceImpl), "*"),
	wire.Struct(new(oidc_provider.OidcRelyingPartyHandler), "*"),
	wire.Struct(new(oidc_provider.ClockImpl), "*"),
	wire.Struct(new(oidc_client.OidcAuthProviderRepositoryImpl), "*"),
	wire.Struct(new(oidc_client.UserAuthMethodRepositoryImpl), "*"),
	wire.Struct(new(oidc_client.OidcUserResolverImpl), "*"),
	wire.Struct(new(oidc_client.LoginServiceImpl), "*"),
	wire.Struct(new(oidc_client.OidcAuthFlowServiceImpl), "*"),
	wire.Struct(new(oidc_client.OidcAuthProviderServiceImpl), "*"),
	wire.Struct(new(oidc_client.OidcClientHandler), "*"),
	wire.Struct(new(store.VersionSigningCodecImpl), "*"),
	wire.Struct(new(store.VersionSigningServiceImpl), "*"),
	wire.Struct(new(retention.RetentionPolicyRepositoryImpl), "*"),
	wire.Struct(new(maintenance.AgentHelperImpl), "*"),
	wire.Struct(new(maintenance.RandomProviderImpl), "*"),
	wire.Struct(new(configs.MaintenanceRepositoryImpl), "*"),
	wire.Struct(new(maintenance.MaintenanceAgentImpl), "*"),
	wire.Struct(new(retention.BackupDeletionFinderImpl), "*"),
	wire.Struct(new(retention.BackupRetentionSelectorImpl), "*"),
	wire.Struct(new(maintenance.MaintenanceConfigsHandler), "*"),
	wire.Struct(new(ingress.ShutdownHandlerImpl), "*"),
	wire.Struct(new(maintenance.MaintenanceServiceImpl), "*"),
	wire.Struct(new(ingress.AuthServiceImpl), "*"),
	wire.Struct(new(ingress.CertificateToolsImpl), "*"),
	wire.Struct(new(certificates2.OperationMonitorImpl)),
	wire.Struct(new(email.EmailHandler), "*"),
	wire.Struct(new(email.EmailServiceImpl), "*"),
	wire.Struct(new(email.UserEmailServiceImpl), "*"),
	wire.Struct(new(groups.GroupHandler), "*"),
	wire.Struct(new(groups.GroupRepositoryImpl), "*"),
	wire.Struct(new(terminal.DockerTerminalClientImpl), "*"),
	wire.Struct(new(terminal.DockerTerminalServiceImpl), "*"),
	wire.Struct(new(terminal.TerminalHandlerImpl), "*"),
	wire.Struct(new(certificates2.AcmeClientImpl), "*"),
	wire.Struct(new(certificates2.CertificateServiceImpl), "*"),
	wire.Struct(new(certificates2.NetworkWaiterImpl), "*"),
	wire.Struct(new(u.DatabaseUtilsImpl)),
	wire.Struct(new(apps_basic.DatabaseIndependentRuntimeImpl), "*"),
	wire.Struct(new(apps_basic.VersionFileNameEncoderImpl), "*"),
	wire.Struct(new(apps_advanced.AppsServiceAdvancedImpl), "*"),
	wire.Struct(new(apps_advanced.AppsAdvancedHandler), "*"),
	wire.Bind(new(apps_basic.ComposeExtractor), new(*apps_basic.ComposeExtractorImpl)),
	wire.Bind(new(apps_advanced.AppsServiceAdvanced), new(*apps_advanced.AppsServiceAdvancedImpl)),
	wire.Bind(new(app_store.AppStoreService), new(*app_store.AppStoreServiceImpl)),
	wire.Bind(new(app_store.VersionVerifier), new(*app_store.VersionVerifierImpl)),
	wire.Bind(new(apps_basic.VersionFileNameEncoder), new(*apps_basic.VersionFileNameEncoderImpl)),
	wire.Bind(new(tools.DatabaseConnector), new(*tools.DatabaseConnectorImpl)),
	wire.Bind(new(tools.TimezoneProvider), new(*tools.TimezoneProviderImpl)),
	wire.Bind(new(apps_basic.DatabaseIndependentRuntime), new(*apps_basic.DatabaseIndependentRuntimeImpl)),
	wire.Bind(new(u.DatabaseUtils), new(*u.DatabaseUtilsImpl)),
	wire.Bind(new(certificates2.AcmeClient), new(*certificates2.AcmeClientImpl)),
	wire.Bind(new(certificates2.OperationMonitor), new(*certificates2.OperationMonitorImpl)),
	wire.Bind(new(certificates2.NetworkWaiter), new(*certificates2.NetworkWaiterImpl)),
	wire.Bind(new(certificates2.CertificateService), new(*certificates2.CertificateServiceImpl)),
	wire.Bind(new(ingress.CertificateTools), new(*ingress.CertificateToolsImpl)),
	wire.Bind(new(ingress.AppRequestPolicy), new(*ingress.AppRequestPolicyImpl)),
	wire.Bind(new(ingress.AuthService), new(*ingress.AuthServiceImpl)),
	wire.Bind(new(certificates2.CertificateCache), new(*certificates2.CertificateCacheImpl)),
	wire.Bind(new(maintenance.MaintenanceService), new(*maintenance.MaintenanceServiceImpl)),
	wire.Bind(new(retention.BackupRetentionSelector), new(*retention.BackupRetentionSelectorImpl)),
	wire.Bind(new(retention.BackupDeletionFinder), new(*retention.BackupDeletionFinderImpl)),
	wire.Bind(new(maintenance.MaintenanceAgent), new(*maintenance.MaintenanceAgentImpl)),
	wire.Bind(new(configs.MaintenanceRepository), new(*configs.MaintenanceRepositoryImpl)),
	wire.Bind(new(maintenance.RandomProvider), new(*maintenance.RandomProviderImpl)),
	wire.Bind(new(maintenance.AgentHelper), new(*maintenance.AgentHelperImpl)),
	wire.Bind(new(retention.RetentionPolicyRepository), new(*retention.RetentionPolicyRepositoryImpl)),
	wire.Bind(new(certificates2.CertificatePersister), new(*certificates2.CertificatePersisterImpl)),
	wire.Bind(new(apps_basic.ClientCredentialsGenerator), new(*apps_basic.ClientCredentialsGeneratorImpl)),
	wire.Bind(new(frontendpages.PageRenderer), new(*frontendpages.PageRendererImpl)),
	wire.Bind(new(oidc_provider.Clock), new(*oidc_provider.ClockImpl)),
	wire.Bind(new(oidc_provider.OidcRelyingPartyResolver), new(*oidc_provider.OidcRelyingPartyResolverImpl)),
	wire.Bind(new(oidc_provider.OidcClientService), new(*oidc_provider.OidcClientServiceImpl)),
	wire.Bind(new(oidc_provider.OidcRelyingPartyRepository), new(*oidc_provider.OidcRelyingPartyRepositoryImpl)),
	wire.Bind(new(oidc_provider.OidcRelyingPartyService), new(*oidc_provider.OidcRelyingPartyServiceImpl)),
	wire.Bind(new(oidc_provider.TokenIssuer), new(*oidc_provider.TokenIssuerImpl)),
	wire.Bind(new(oidc_provider.OidcService), new(*oidc_provider.OidcServiceImpl)),
	wire.Bind(new(frontend.FrontendPageDataBuilder), new(*frontend.FrontendPageDataBuilderImpl)),
	wire.Bind(new(renderer.TemplateEngine), new(*renderer.TemplateEngineImpl)),
	wire.Bind(new(renderer.TemplateService), new(*renderer.TemplateServiceImpl)),
	wire.Bind(new(assets.AssetTagBuilder), new(*assets.AssetTagBuilderImpl)),
	wire.Bind(new(apps_basic.AppDetector), new(*apps_basic.AppDetectorImpl)),
	wire.Bind(new(apps_basic.OperationRegistry), new(*apps_basic.OperationRegistryImpl)),
	wire.Bind(new(apps_basic.AppServiceHelper), new(*apps_basic.AppServiceHelperImpl)),
	wire.Bind(new(apps_basic.Authorizer), new(*apps_basic.AuthorizerImpl)),
	wire.Bind(new(u.AuthHelper), new(*u.AuthHelperImpl)),
	wire.Bind(new(u.BytesSigner), new(*u.BytesSignerImpl)),
	wire.Bind(new(backup_server.ResticSnapshotsParser), new(*backup_server.ResticSnapshotsParserImpl)),
	wire.Bind(new(backups.RestoreFinalizer), new(*backups.RestoreFinalizerImpl)),
	wire.Bind(new(backups.BackupQueryService), new(*backups.BackupQueryServiceImpl)),
	wire.Bind(new(backups.BackupCreationAssembler), new(*backups.BackupCreationAssemblerImpl)),
	wire.Bind(new(backups.SnapshotPackager), new(*backups.SnapshotPackagerImpl)),
	wire.Bind(new(u.OsWrapper), new(*u.OsWrapperImpl)),
	wire.Bind(new(backup_server.ResticService), new(*backup_server.ResticServiceImpl)),
	wire.Bind(new(configs.ConfigsRepository), new(*configs.ConfigsRepositoryImpl)),
	wire.Bind(new(configs.ConfigsService), new(*configs.ConfigsServiceImpl)),
	wire.Bind(new(configs.OidcEmailExposureService), new(*configs.OidcEmailExposureServiceImpl)),
	wire.Bind(new(configs.EmailRepository), new(*configs.EmailRepoImpl)),
	wire.Bind(new(store.VersionSigningCodec), new(*store.VersionSigningCodecImpl)),
	wire.Bind(new(store.VersionSigningService), new(*store.VersionSigningServiceImpl)),
	wire.Bind(new(tools.DockerService), new(*tools.DockerServiceImpl)),
	wire.Bind(new(backups.BackupService), new(*backups.BackupServiceImpl)),
	wire.Bind(new(backup_server.ResticDockerImageService), new(*backup_server.ResticDockerImageServiceImpl)),
	wire.Bind(new(backup_server.SshClient), new(*backup_server.SshClientImpl)),
	wire.Bind(new(apps_basic.AppService), new(*apps_basic.AppServiceImpl)),
	wire.Bind(new(users.UserRepository), new(*users.UserRepositoryImpl)),
	wire.Bind(new(users.SessionRepository), new(*users.SessionRepositoryImpl)),
	wire.Bind(new(users.SessionService), new(*users.SessionServiceImpl)),
	wire.Bind(new(apps_basic.AppRepository), new(*apps_basic.AppRepositoryImpl)),
	wire.Bind(new(apps_basic.AppRequestParser), new(*apps_basic.AppRequestParserImpl)),
	wire.Bind(new(apps_basic.AppReverseProxyFactory), new(*apps_basic.AppReverseProxyFactoryImpl)),
	wire.Bind(new(apps_basic.AppSessionService), new(*apps_basic.AppSessionServiceImpl)),
	wire.Bind(new(email.EmailService), new(*email.EmailServiceImpl)),
	wire.Bind(new(email.UserEmailService), new(*email.UserEmailServiceImpl)),
	wire.Bind(new(groups.GroupRepository), new(*groups.GroupRepositoryImpl)),
	wire.Bind(new(terminal.DockerTerminalClient), new(*terminal.DockerTerminalClientImpl)),
	wire.Bind(new(terminal.DockerTerminalService), new(*terminal.DockerTerminalServiceImpl)),
	wire.Bind(new(system_config_migrations.SystemConfigMigrationsProvider), new(*system_config_migrations.SystemConfigMigrationsProviderImpl)),
	wire.Bind(new(system_config_migrations.SystemConfigMigrator), new(*system_config_migrations.SystemConfigMigratorImpl)),
	wire.Bind(new(tools.CommandRunner), new(*tools.CommandRunnerImpl)),
	wire.Bind(new(tools.ResticContainerExecutor), new(*tools.ResticContainerExecutorImpl)),
	wire.Bind(new(backup_server.SshRepositoryService), new(*backup_server.SshRepositoryServiceImpl)),
	wire.Bind(new(backup_server.SshRepository), new(*backup_server.SshRepositoryImpl)),
	wire.Bind(new(users.UserService), new(*users.UserServiceImpl)),
	wire.Bind(new(users.AuthenticationService), new(*users.AuthenticationServiceImpl)),
	wire.Bind(new(users.SecretAndCookieStorage), new(*users.SecretAndCookieStorageImpl)),
	wire.Bind(new(oidc_provider.IdTokenIssuer), new(*oidc_provider.IdTokenIssuerImpl)),
	wire.Bind(new(oidc_client.OidcAuthProviderRepository), new(*oidc_client.OidcAuthProviderRepositoryImpl)),
	wire.Bind(new(oidc_client.UserAuthMethodRepository), new(*oidc_client.UserAuthMethodRepositoryImpl)),
	wire.Bind(new(oidc_client.OidcUserResolver), new(*oidc_client.OidcUserResolverImpl)),
	wire.Bind(new(oidc_client.LoginService), new(*oidc_client.LoginServiceImpl)),
	wire.Bind(new(oidc_client.OidcAuthFlowService), new(*oidc_client.OidcAuthFlowServiceImpl)),
	wire.Bind(new(oidc_client.OidcAuthProviderService), new(*oidc_client.OidcAuthProviderServiceImpl)),
	wire.Bind(new(assets.AssetStore), new(*assets.AssetStoreImpl)),
)

var DependenciesSet = wire.NewSet(
	wire.Struct(new(Repositories), "*"),
	wire.Struct(new(Dependencies), "*"),
)

func NewRouter() chi.Router {
	return chi.NewRouter()
}

func NewDockerClient() (*client.Client, error) {
	dockerClient, err := client.New(client.FromEnv)
	if err != nil {
		return nil, err
	}
	return dockerClient, nil
}
