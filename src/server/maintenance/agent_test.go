package maintenance

import (
	"server/backup_server"
	"server/configs"
	"server/oidc_client"
	"server/oidc_provider"
	"server/users"

	"server/tools"
	"testing"
	"time"

	"github.com/quollix/common/assert"
	"github.com/stretchr/testify/mock"
)

var sampleNowUtc = time.Date(2026, 2, 5, 12, 0, 0, 0, time.UTC)

type maintenanceAgentTestObjects struct {
	Agent               *MaintenanceAgentImpl
	Repository          *configs.MaintenanceRepositoryMock
	Helper              *AgentHelperMock
	OsWrapper           *tools.CommonOsWrapperMock
	CommandRunner       *tools.CommandRunnerMock
	OidcProviderCache   *oidc_provider.OidcCacheMock
	OidcLoginStateCache *oidc_client.OidcLoginStateCacheMock
	SessionRepo         *users.SessionRepositoryMock
	ResticImage         *backup_server.ResticDockerImageServiceMock
}

func newMaintenanceAgentTestObjects(t *testing.T) maintenanceAgentTestObjects {
	repository := configs.NewMaintenanceRepositoryMock(t)
	helper := NewAgentHelperMock(t)
	osWrapper := tools.NewCommonOsWrapperMock(t)
	commandRunner := tools.NewCommandRunnerMock(t)
	oidcProviderCache := oidc_provider.NewOidcCacheMock(t)
	oidcLoginStateCache := oidc_client.NewOidcLoginStateCacheMock(t)
	sessionRepo := users.NewSessionRepositoryMock(t)
	resticImage := backup_server.NewResticDockerImageServiceMock(t)

	agent := &MaintenanceAgentImpl{
		Repository:          repository,
		ServiceHelper:       helper,
		OsWrapper:           osWrapper,
		CommandRunner:       commandRunner,
		OidcProviderCache:   oidcProviderCache,
		OidcLoginStateCache: oidcLoginStateCache,
		SessionRepo:         sessionRepo,
		ResticImage:         resticImage,
		GlobalConfig:        tools.NewGlobalConfigFromEnv(),
	}
	return maintenanceAgentTestObjects{
		Agent:               agent,
		Repository:          repository,
		Helper:              helper,
		OsWrapper:           osWrapper,
		CommandRunner:       commandRunner,
		OidcProviderCache:   oidcProviderCache,
		OidcLoginStateCache: oidcLoginStateCache,
		SessionRepo:         sessionRepo,
		ResticImage:         resticImage,
	}
}

func assertMaintenanceAgentExpectations(t *testing.T, testObjects maintenanceAgentTestObjects) {
	testObjects.Repository.AssertExpectations(t)
	testObjects.Helper.AssertExpectations(t)
	testObjects.OsWrapper.AssertExpectations(t)
	testObjects.CommandRunner.AssertExpectations(t)
	testObjects.OidcProviderCache.AssertExpectations(t)
	testObjects.OidcLoginStateCache.AssertExpectations(t)
	testObjects.SessionRepo.AssertExpectations(t)
	testObjects.ResticImage.AssertExpectations(t)
}

func TestStartMaintenanceJob_HappyPath_RunsAndReschedules(t *testing.T) {
	testObjects := newMaintenanceAgentTestObjects(t)
	defer assertMaintenanceAgentExpectations(t, testObjects)

	oldNextMaintenanceAt := time.Date(2026, 2, 5, 11, 0, 0, 0, time.UTC)
	newNextMaintenanceAt := time.Date(2026, 2, 6, 2, 15, 10, 0, time.UTC)

	inputConfig := getSampleMaintenanceConfig(oldNextMaintenanceAt)
	expectedConfig := *inputConfig
	expectedConfig.NextMaintenanceAt = newNextMaintenanceAt

	testObjects.Repository.EXPECT().GetMaintenanceConfig().Return(inputConfig, nil).Once()
	testObjects.OsWrapper.EXPECT().Now().Return(sampleNowUtc).Once()
	testObjects.Helper.EXPECT().
		CalculateNextMaintenanceAtUtc(sampleNowUtc, "Europe/Berlin", 2).
		Return(&newNextMaintenanceAt, nil).
		Once()

	testObjects.Repository.EXPECT().
		SetMaintenanceConfig(mock.Anything).
		Run(func(configPassedToDependency *configs.MaintenanceConfig) {
			assert.Equal(t, expectedConfig, *configPassedToDependency)
		}).
		Return(nil).
		Once()

	testObjects.Helper.EXPECT().UpdateAllApps().Once()
	testObjects.ResticImage.EXPECT().UpdateResticDockerImage().Return(nil).Once()
	testObjects.Helper.EXPECT().CreateBackupsForAllAppsIfConfigured().Once()
	testObjects.Helper.EXPECT().RetentBackupsForAllAppsIfConfigured().Once()
	testObjects.OidcProviderCache.EXPECT().Cleanup()
	testObjects.OidcLoginStateCache.EXPECT().CleanupExpiredLoginStates()
	testObjects.SessionRepo.EXPECT().DeleteExpiredSessions().Return(nil).Once()
	testObjects.Agent.GlobalConfig.PruneDockerSystemDuringMaintenance = true
	testObjects.CommandRunner.EXPECT().RunCommand(dockerImagePruneCommand).Return(&tools.CommandOutput{}, nil).Once()

	testObjects.Agent.considerRunningMaintenanceJob()
}

func getSampleMaintenanceConfig(oldNextMaintenanceAt time.Time) *configs.MaintenanceConfig {
	config := &configs.MaintenanceConfig{
		MaintenanceWindowStartHour: 2,
		NextMaintenanceAt:          oldNextMaintenanceAt,
		IanaTimezone:               "Europe/Berlin",
	}
	return config
}

func TestStartMaintenanceJob_WhenNextMaintenanceIsInTheFuture_DoesNothing(t *testing.T) {
	testObjects := newMaintenanceAgentTestObjects(t)
	defer assertMaintenanceAgentExpectations(t, testObjects)

	futureNextMaintenanceAt := time.Date(2026, 2, 5, 12, 0, 1, 0, time.UTC)
	config := getSampleMaintenanceConfig(futureNextMaintenanceAt)

	testObjects.Repository.EXPECT().GetMaintenanceConfig().Return(config, nil).Once()
	testObjects.OsWrapper.EXPECT().Now().Return(sampleNowUtc).Once()

	testObjects.Agent.considerRunningMaintenanceJob()
}

func TestStartMaintenanceJob_WhenDockerPurgeDisabled_DoesNotRunDockerPrune(t *testing.T) {
	testObjects := newMaintenanceAgentTestObjects(t)
	defer assertMaintenanceAgentExpectations(t, testObjects)

	oldNextMaintenanceAt := time.Date(2026, 2, 5, 11, 0, 0, 0, time.UTC)
	newNextMaintenanceAt := time.Date(2026, 2, 6, 2, 15, 10, 0, time.UTC)

	inputConfig := getSampleMaintenanceConfig(oldNextMaintenanceAt)

	testObjects.Repository.EXPECT().GetMaintenanceConfig().Return(inputConfig, nil).Once()
	testObjects.OsWrapper.EXPECT().Now().Return(sampleNowUtc).Once()
	testObjects.Helper.EXPECT().
		CalculateNextMaintenanceAtUtc(sampleNowUtc, "Europe/Berlin", 2).
		Return(&newNextMaintenanceAt, nil).
		Once()

	testObjects.Repository.EXPECT().SetMaintenanceConfig(mock.Anything).Return(nil).Once()

	testObjects.Helper.EXPECT().UpdateAllApps().Once()
	testObjects.ResticImage.EXPECT().UpdateResticDockerImage().Return(nil).Once()
	testObjects.Helper.EXPECT().CreateBackupsForAllAppsIfConfigured().Once()
	testObjects.Helper.EXPECT().RetentBackupsForAllAppsIfConfigured().Once()
	testObjects.OidcProviderCache.EXPECT().Cleanup()
	testObjects.OidcLoginStateCache.EXPECT().CleanupExpiredLoginStates()
	testObjects.SessionRepo.EXPECT().DeleteExpiredSessions().Return(nil).Once()

	testObjects.Agent.GlobalConfig.PruneDockerSystemDuringMaintenance = false

	testObjects.Agent.considerRunningMaintenanceJob()
}
