package di

import (
	"crypto/ed25519"
	"log/slog"

	"server/app_store"
	"server/apps_basic"
	certificates2 "server/certificates"
	"server/configs"
	"server/email"
	"server/tools"

	"github.com/quollix/common/store"
	u "github.com/quollix/common/utils"
	"github.com/quollix/common/validation"
	"github.com/quollix/deepstack"
)

func NewGlobalConfig() *tools.GlobalConfig {
	return tools.NewGlobalConfigFromEnv()
}

func GetLogger(config *tools.GlobalConfig) deepstack.DeepStackLogger {
	if config.UseDevelopmentLogger {
		return deepstack.NewDeepStackLogger(deepstack.NewRawConsoleHandler(slog.LevelInfo))
	}
	return deepstack.NewDeepStackLogger(deepstack.NewJsonConsoleHandler(slog.LevelInfo))
}

func NewAppStoreClient(
	config *tools.GlobalConfig,
	directoryProvider tools.DirectoryProvider,
	validator validation.VersionValidator,
	generator apps_basic.ClientCredentialsGenerator,
	authHelper u.AuthHelper,
	appRepository apps_basic.AppRepository,
	appService apps_basic.AppService,
	appServiceHelper apps_basic.AppServiceHelper,
	versionSigningService store.VersionSigningService,
) app_store.AppStoreClientLean {
	if config.UseLocalAppStoreClient {
		return &app_store.AppStoreClientMock{
			DirectoryProvider:          directoryProvider,
			Config:                     config,
			VersionValidator:           validator,
			ClientCredentialsGenerator: generator,
			AuthHelper:                 authHelper,
			AppRepository:              appRepository,
			AppService:                 appService,
			AppServiceHelper:           appServiceHelper,
			VersionSigningService:      versionSigningService,
		}
	}

	appStoreClient := app_store.AppStoreClientImpl{
		AppStoreClientImpl: store.AppStoreClientImpl{
			Parent: u.ComponentClient{
				RootUrl: "https://store.quollix.org",
			},
		},
	}
	return &appStoreClient
}

func NewVersionValidator() validation.VersionValidator {
	return validation.NewVersionValidator(false)
}

func NewOfficialMaintainerPublicKey(config *tools.GlobalConfig) (ed25519.PublicKey, error) {
	var publicKeyBytes []byte
	if config.UseLocalTestingAuthorizedKey {
		publicKeyBytes = u.LocalTestingPublicKeyOpenSSHBytes
	} else {
		publicKeyBytes = []byte(store.AppStoreOfficialMaintainerPublicKeyOpenSSH)
	}
	return u.DecodeAuthorizedEd25519PublicKey(publicKeyBytes)
}

func NewEmailClient(config *tools.GlobalConfig) u.EmailClient {
	if config.UseStrictEmailClientStub {
		return &email.StrictEmailClientStub{}
	}
	return &u.EmailClientImpl{}
}

func NewWildcardCertificateService(
	config *tools.GlobalConfig,
	configsRepo *configs.ConfigsRepositoryImpl,
	configsService configs.ConfigsService,
	certificatePersister *certificates2.CertificatePersisterImpl,
	networkWaiter certificates2.NetworkWaiter,
	acmeClient certificates2.AcmeClient,
	certificateService certificates2.CertificateService,
	operationMonitor certificates2.OperationMonitor,
) certificates2.WildcardCertificateService {
	if config.UseMockWildcardCertificateService {
		return &certificates2.WildcardCertificateServiceMock{
			OperationMonitor: operationMonitor,
		}
	}
	return &certificates2.WildcardCertificateServiceImpl{
		ConfigsRepository:    configsRepo,
		ConfigsService:       configsService,
		CertificatePersister: certificatePersister,
		NetworkWaiter:        networkWaiter,
		AcmeClient:           acmeClient,
		CertificateService:   certificateService,
		OperationMonitor:     operationMonitor,
	}
}

func NewDatabaseSnapshotRepo(
	globalConfig *tools.GlobalConfig,
	dbProvider *tools.DatabaseConnectorImpl,
	databaseUtils u.DatabaseUtils,
) u.DatabaseSnapshotRepository {
	return u.NewDatabaseSnapshotRepository(globalConfig.DatabaseHostName, dbProvider, databaseUtils)
}
