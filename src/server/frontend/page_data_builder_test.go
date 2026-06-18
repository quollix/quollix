package frontend

import (
	"server/app_store"
	"server/apps_basic"
	"server/backup_server"
	"server/configs"
	"server/maintenance/retention"
	"server/tools"
	"server/users"
	"testing"
	"time"

	"github.com/quollix/common/assert"
	"github.com/quollix/common/store"
	u "github.com/quollix/common/utils"
)

type frontendBuilderTestObjects struct {
	Builder                    *FrontendPageDataBuilderImpl
	AppService                 *apps_basic.AppServiceMock
	UserRepo                   *users.UserRepositoryMock
	ConfigsRepo                *configs.ConfigsRepositoryMock
	SshRepositoryConfigService *backup_server.SshRepositoryServiceMock
	AppStoreClient             *app_store.AppStoreClientLeanMock
	SshRepo                    *backup_server.SshRepositoryMock
	MaintenanceRepo            *configs.MaintenanceRepositoryMock
	RetentionRepo              *retention.RetentionPolicyRepositoryMock
	OsWrapper                  *tools.CommonOsWrapperMock
	TimezoneProvider           *tools.TimezoneProviderMock
	EmailRepo                  *configs.EmailRepositoryMock
}

func getTestObjects(t *testing.T) frontendBuilderTestObjects {
	appService := apps_basic.NewAppServiceMock(t)
	userRepo := users.NewUserRepositoryMock(t)
	configsRepo := configs.NewConfigsRepositoryMock(t)
	sshRepositoryConfigService := backup_server.NewSshRepositoryServiceMock(t)
	appStoreClient := app_store.NewAppStoreClientLeanMock(t)
	sshRepo := backup_server.NewSshRepositoryMock(t)
	maintenanceRepo := configs.NewMaintenanceRepositoryMock(t)
	retentionRepo := retention.NewRetentionPolicyRepositoryMock(t)
	osWrapper := tools.NewCommonOsWrapperMock(t)
	timezoneProvider := tools.NewTimezoneProviderMock(t)
	emailRepo := configs.NewEmailRepositoryMock(t)

	builder := &FrontendPageDataBuilderImpl{
		AppService:                 appService,
		UserRepo:                   userRepo,
		ConfigsRepo:                configsRepo,
		SshRepositoryConfigService: sshRepositoryConfigService,
		AppStoreClient:             appStoreClient,
		SshRepository:              sshRepo,
		MaintenanceRepo:            maintenanceRepo,
		RetentionPolicyRepo:        retentionRepo,
		OsWrapper:                  osWrapper,
		TimezoneProvider:           timezoneProvider,
		EmailRepository:            emailRepo,
		GlobalConfig:               tools.NewGlobalConfigFromEnv(),
	}

	return frontendBuilderTestObjects{
		Builder:                    builder,
		AppService:                 appService,
		UserRepo:                   userRepo,
		ConfigsRepo:                configsRepo,
		SshRepositoryConfigService: sshRepositoryConfigService,
		AppStoreClient:             appStoreClient,
		SshRepo:                    sshRepo,
		MaintenanceRepo:            maintenanceRepo,
		RetentionRepo:              retentionRepo,
		OsWrapper:                  osWrapper,
		TimezoneProvider:           timezoneProvider,
		EmailRepo:                  emailRepo,
	}
}

func TestBuildInstalledAppsPage_SortsByMaintainerThenAppName(t *testing.T) {
	testObjects := getTestObjects(t)
	now := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)

	testObjects.AppService.EXPECT().
		ListAppsForRole(123, tools.UserLevel).
		Return([]apps_basic.AppDto{
			{Maintainer: "b-maintainer", AppName: "a-app", VersionCreationTimestamp: now.Add(-3 * time.Hour)},
			{Maintainer: "a-maintainer", AppName: "z-app", VersionCreationTimestamp: now.Add(-2 * time.Hour)},
			{Maintainer: "a-maintainer", AppName: "a-app", VersionCreationTimestamp: now.Add(-1 * time.Hour)},
		}, nil)
	testObjects.SshRepo.EXPECT().IsRemoteBackupEnabled().Return(false, nil)
	testObjects.OsWrapper.EXPECT().Now().Return(now)

	pageContent, err := testObjects.Builder.BuildInstalledAppsPage(123, tools.UserLevel)
	assert.Nil(t, err)

	assert.Equal(t, "a-maintainer", pageContent.Apps[0].Maintainer)
	assert.Equal(t, "a-app", pageContent.Apps[0].AppName)
	assert.Equal(t, "1h ago", pageContent.Apps[0].VersionCreationTimestampFormatted)
	assert.Equal(t, "2025-12-31 23:00:00", pageContent.Apps[0].VersionCreationTimestampTooltip)

	assert.Equal(t, "a-maintainer", pageContent.Apps[1].Maintainer)
	assert.Equal(t, "z-app", pageContent.Apps[1].AppName)

	assert.Equal(t, "b-maintainer", pageContent.Apps[2].Maintainer)
	assert.Equal(t, "a-app", pageContent.Apps[2].AppName)

	assert.False(t, pageContent.IsBackupEnabled)
}

func TestBuildInstalledAppsPage_WhenBackupEnabled_SetsFlag(t *testing.T) {
	testObjects := getTestObjects(t)

	testObjects.AppService.EXPECT().ListAppsForRole(123, tools.UserLevel).Return([]apps_basic.AppDto{}, nil)
	testObjects.SshRepo.EXPECT().IsRemoteBackupEnabled().Return(true, nil)
	testObjects.OsWrapper.EXPECT().Now().Return(time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC))

	pageContent, err := testObjects.Builder.BuildInstalledAppsPage(123, tools.UserLevel)
	assert.Nil(t, err)
	assert.True(t, pageContent.IsBackupEnabled)
}

func TestBuildUsersPage_SortsAndBuildsSetPasswordLink(t *testing.T) {
	testObjects := getTestObjects(t)

	tokenExpiration := time.Date(2026, 1, 17, 10, 0, 0, 0, time.UTC)

	testObjects.UserRepo.EXPECT().
		ListUsers().
		Return([]tools.User{
			{
				Id:                             2,
				Username:                       "zara",
				Email:                          "zara@sample.com",
				IsAdmin:                        false,
				SetPasswordToken:               "",
				SetPasswordTokenExpirationDate: tools.DefaultTime,
				CreationDate:                   time.Date(2026, 1, 16, 8, 30, 0, 0, time.UTC),
			},
			{
				Id:                             1,
				Username:                       "alice",
				Email:                          "alice@sample.com",
				IsAdmin:                        true,
				SetPasswordToken:               "token-123",
				SetPasswordTokenExpirationDate: tokenExpiration,
				CreationDate:                   time.Date(2026, 1, 15, 9, 45, 0, 0, time.UTC),
			},
		}, nil)

	testObjects.ConfigsRepo.EXPECT().GetConfig(configs.ConfigKeys.ServerHost).Return("example.com", nil)
	testObjects.EmailRepo.EXPECT().ReadEmailConfig().Return(&u.EmailConfig{IsEnabled: true}, nil)

	pageContent, err := testObjects.Builder.BuildUsersPage()
	assert.Nil(t, err)
	assert.True(t, pageContent.IsEmailEnabled)

	assert.Equal(t, 1, pageContent.Users[0].Id)
	assert.Equal(t, "alice", pageContent.Users[0].Username)
	assert.Equal(t, "alice@sample.com", pageContent.Users[0].Email)
	assert.True(t, pageContent.Users[0].IsAdmin)
	assert.Equal(t, "https://quollix.example.com/set-password?token=token-123", pageContent.Users[0].SetPasswordLink)
	assert.Equal(t, "2026-01-17 10:00:00", pageContent.Users[0].SetPasswordTokenExpirationDate)
	assert.Equal(t, "2026-01-15 09:45:00", pageContent.Users[0].CreatedAt)

	assert.Equal(t, 2, pageContent.Users[1].Id)
	assert.Equal(t, "zara", pageContent.Users[1].Username)
	assert.Equal(t, "zara@sample.com", pageContent.Users[1].Email)
	assert.False(t, pageContent.Users[1].IsAdmin)
	assert.True(t, pageContent.Users[1].SetPasswordLink == "")
	assert.True(t, pageContent.Users[1].SetPasswordTokenExpirationDate == "")
	assert.Equal(t, "2026-01-16 08:30:00", pageContent.Users[1].CreatedAt)
}

func TestBuildBackedUpAppsPage_WhenBackupDisabled_ReturnsDisabledPage(t *testing.T) {
	testObjects := getTestObjects(t)

	testObjects.SshRepo.EXPECT().
		IsRemoteBackupEnabled().
		Return(false, nil)

	pageContent, err := testObjects.Builder.BuildBackedUpAppsPage()
	assert.Nil(t, err)
	assert.True(t, pageContent.IsBackupEnabled == false)
	assert.True(t, pageContent.Apps == nil)
}

func TestBuildSettingsPage(t *testing.T) {
	testObjects := getTestObjects(t)

	expectedRemoteRepository := &tools.BackupServerConfigs{}
	expectedMaintenanceConfig := &configs.MaintenanceConfig{
		IanaTimezone:               "Europe/Berlin",
		MaintenanceWindowStartHour: 7,
		NextMaintenanceAt:          time.Date(2026, time.January, 15, 12, 30, 0, 0, time.UTC),
	}
	expectedRetentionPolicy := &retention.RetentionPolicy{}
	expectedIanaTimezones := []string{"UTC", "Europe/Berlin"}

	testObjects.SshRepositoryConfigService.EXPECT().GetRemoteBackupRepository().Return(expectedRemoteRepository, nil)
	testObjects.MaintenanceRepo.EXPECT().GetMaintenanceConfig().Return(expectedMaintenanceConfig, nil)
	testObjects.RetentionRepo.EXPECT().GetRetentionPolicy().Return(expectedRetentionPolicy, nil)
	testObjects.TimezoneProvider.EXPECT().ListIanaTimezones().Return(expectedIanaTimezones)

	pageContent, err := testObjects.Builder.BuildSettingsPage()
	assert.Nil(t, err)
	assert.Equal(t, expectedRemoteRepository, pageContent.BackupServer)
	assert.Equal(t, expectedMaintenanceConfig, pageContent.MaintenanceConfig)
	assert.Equal(t, expectedRetentionPolicy, pageContent.RetentionPolicy)
	assert.Equal(t, expectedIanaTimezones, pageContent.IanaTimezoneOptions)
	assert.Equal(t, "Thu, 15 Jan 2026 13:30", pageContent.NextMaintenanceAt)

	assert.Equal(t, 24, len(pageContent.MaintenanceWindowOptions))
	assert.Equal(t, 0, pageContent.MaintenanceWindowOptions[0].Value)
	assert.Equal(t, "00:00-01:00", pageContent.MaintenanceWindowOptions[0].Label)
	assert.Equal(t, 1, pageContent.MaintenanceWindowOptions[1].Value)
	assert.Equal(t, "01:00-02:00", pageContent.MaintenanceWindowOptions[1].Label)
	assert.Equal(t, 23, pageContent.MaintenanceWindowOptions[23].Value)
	assert.Equal(t, "23:00-00:00", pageContent.MaintenanceWindowOptions[23].Label)
}

func TestBuildVersionsPage_SortsByCreationTimestampDesc(t *testing.T) {
	testObjects := getTestObjects(t)

	olderTimestamp := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)
	newerTimestamp := time.Date(2026, 1, 2, 10, 0, 0, 0, time.UTC)

	testObjects.AppStoreClient.EXPECT().
		ListVersions("maintainer", "app").
		Return([]store.LeanVersionDto{
			{Name: "older", CreationTimestamp: olderTimestamp},
			{Name: "newer", CreationTimestamp: newerTimestamp},
		}, nil)

	pageContent, err := testObjects.Builder.BuildVersionsPage("maintainer", "app")
	assert.Nil(t, err)

	assert.Equal(t, "maintainer", pageContent.Maintainer)
	assert.Equal(t, "app", pageContent.App)

	assert.Equal(t, "newer", pageContent.Versions[0].Name)
	assert.Equal(t, "older", pageContent.Versions[1].Name)
}

func TestBuildOidcClientsPage_FiltersOfficialDatabaseApp_AndSortsByMaintainerThenAppName(t *testing.T) {
	testObjects := getTestObjects(t)

	testObjects.AppService.EXPECT().
		ListAppsForAdmin().
		Return([]apps_basic.AppDto{
			{Maintainer: "b-maintainer", AppName: "a-app"},
			{Maintainer: "a-maintainer", AppName: u.OfficialDatabaseAppName},
			{Maintainer: "a-maintainer", AppName: "z-app"},
			{Maintainer: "a-maintainer", AppName: "a-app"},
		}, nil)

	pageContent, err := testObjects.Builder.BuildOidcClientsPage()
	assert.Nil(t, err)

	assert.Equal(t, 3, len(pageContent.Apps))
	assert.Equal(t, "a-maintainer", pageContent.Apps[0].Maintainer)
	assert.Equal(t, "a-app", pageContent.Apps[0].AppName)
	assert.Equal(t, "a-maintainer", pageContent.Apps[1].Maintainer)
	assert.Equal(t, "z-app", pageContent.Apps[1].AppName)
	assert.Equal(t, "b-maintainer", pageContent.Apps[2].Maintainer)
	assert.Equal(t, "a-app", pageContent.Apps[2].AppName)
}

func TestBuildBackupsPage_ReturnsLoadingShell(t *testing.T) {
	testObjects := getTestObjects(t)

	request := tools.MaintainerAndApp{Maintainer: "maintainer", AppName: "app"}

	pageContent, err := testObjects.Builder.BuildBackupsPage(request)
	assert.Nil(t, err)

	assert.Equal(t, "maintainer", pageContent.Maintainer)
	assert.Equal(t, "app", pageContent.AppName)
	assert.True(t, pageContent.IsLoading)
	assert.Equal(t, 0, len(pageContent.Backups))
}

func TestBuildOidcClientsPage_ReturnsAppsForAdmin(t *testing.T) {
	testObjects := getTestObjects(t)

	expectedAppDtos := []apps_basic.AppDto{
		{Maintainer: "m1", AppName: "a1"},
		{Maintainer: "m2", AppName: "a2"},
	}
	appDtosReturnedByMock := expectedAppDtos
	appDtosReturnedByMock = append(appDtosReturnedByMock, apps_basic.AppDto{Maintainer: u.OfficialMaintainer, AppName: u.OfficialDatabaseAppName})

	testObjects.AppService.EXPECT().
		ListAppsForAdmin().
		Return(appDtosReturnedByMock, nil)

	pageContent, err := testObjects.Builder.BuildOidcClientsPage()
	assert.Nil(t, err)
	assert.Equal(t, 2, len(pageContent.Apps))
	assert.Equal(t, expectedAppDtos, pageContent.Apps)
}

func TestBuildStorePage_WhenNotSearch_ReturnsEmptyAppsAndEchoesInputs(t *testing.T) {
	testObjects := getTestObjects(t)
	testObjects.Builder.GlobalConfig.ShowUnofficialAppsSearch = true

	pageContent, err := testObjects.Builder.BuildStorePage("maintainer", "app", false, false)
	assert.Nil(t, err)

	assert.Equal(t, "maintainer", pageContent.MaintainerSearchTerm)
	assert.Equal(t, "app", pageContent.AppSearchTerm)
	assert.False(t, pageContent.ShowUnofficialApps)
	assert.True(t, pageContent.ShowUnofficialToggle)
	assert.Equal(t, 0, len(pageContent.Apps))
}

func TestBuildStorePage_WhenSearch_UsesCorrectMaintainerForSearch(t *testing.T) {
	testCases := []struct {
		name                    string
		showUnofficialApps      bool
		expectedMaintainerQuery string
	}{
		{
			name:                    "showUnofficialApps=true uses empty maintainer for search",
			showUnofficialApps:      true,
			expectedMaintainerQuery: "",
		},
		{
			name:                    "showUnofficialApps=false uses provided maintainer for search",
			showUnofficialApps:      false,
			expectedMaintainerQuery: "maintainer",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testObjects := getTestObjects(t)
			testObjects.Builder.GlobalConfig.ShowUnofficialAppsSearch = true

			creationTimestamp := time.Date(2026, time.January, 2, 10, 0, 0, 0, time.UTC)
			appsToDisplay := []store.AppWithLatestVersion{
				{
					Maintainer:                     "m1",
					AppName:                        "a1",
					LatestVersionName:              "1.2.3",
					LatestVersionCreationTimestamp: creationTimestamp,
				},
			}

			testObjects.AppStoreClient.EXPECT().
				SearchForApps(testCase.expectedMaintainerQuery, "app", testCase.showUnofficialApps).
				Return(appsToDisplay, nil)

			pageContent, err := testObjects.Builder.BuildStorePage("maintainer", "app", testCase.showUnofficialApps, true)
			assert.Nil(t, err)

			assert.Equal(t, "maintainer", pageContent.MaintainerSearchTerm)
			assert.Equal(t, "app", pageContent.AppSearchTerm)
			assert.Equal(t, testCase.showUnofficialApps, pageContent.ShowUnofficialApps)
			assert.True(t, pageContent.ShowUnofficialToggle)
			assert.Equal(t, 1, len(pageContent.Apps))
			assert.Equal(t, "m1", pageContent.Apps[0].Maintainer)
			assert.Equal(t, "a1", pageContent.Apps[0].AppName)
			assert.Equal(t, "1.2.3", pageContent.Apps[0].LatestVersionName)
			assert.Equal(t, "2026-01-02 10:00:00", pageContent.Apps[0].LatestVersionCreationTimestamp)
		})
	}
}

func TestBuildStorePage_WhenSearchFails_ReturnsError(t *testing.T) {
	testObjects := getTestObjects(t)
	expectedErr := u.Logger.NewError("search failed")

	testObjects.AppStoreClient.EXPECT().
		SearchForApps("maintainer", "app", false).
		Return(nil, expectedErr)

	pageContent, err := testObjects.Builder.BuildStorePage("maintainer", "app", false, true)
	assert.Equal(t, expectedErr, err)
	assert.Nil(t, pageContent)
}

func TestBuildStorePage_EchoesProfileToggleForUnofficialAppsSearch(t *testing.T) {
	testObjects := getTestObjects(t)
	testObjects.Builder.GlobalConfig.ShowUnofficialAppsSearch = false

	pageContent, err := testObjects.Builder.BuildStorePage("maintainer", "app", false, false)
	assert.Nil(t, err)
	assert.False(t, pageContent.ShowUnofficialToggle)
}

func TestBuildSetPasswordPage_ReturnsUsername(t *testing.T) {
	testObjects := getTestObjects(t)

	testObjects.UserRepo.EXPECT().
		GetUserByToken("token-123").
		Return(&tools.User{Username: "alice"}, nil)

	pageContent, err := testObjects.Builder.BuildSetPasswordPage("token-123")
	assert.Nil(t, err)
	assert.Equal(t, "alice", pageContent.Username)
}

func TestBuildBackedUpAppsPage_WhenBackupEnabled_ReturnsLoadingShell(t *testing.T) {
	testObjects := getTestObjects(t)

	testObjects.SshRepo.EXPECT().IsRemoteBackupEnabled().Return(true, nil)

	pageContent, err := testObjects.Builder.BuildBackedUpAppsPage()
	assert.Nil(t, err)
	assert.True(t, pageContent.IsBackupEnabled)
	assert.True(t, pageContent.IsLoading)
	assert.Equal(t, 0, len(pageContent.Apps))
}

func TestBuildMaintenancePage_SortsByMaintainerThenAppName(t *testing.T) {
	testObjects := getTestObjects(t)

	testObjects.AppService.EXPECT().
		ListAppsForAdmin().
		Return([]apps_basic.AppDto{
			{Maintainer: "b-maintainer", AppName: "a-app"},
			{Maintainer: "a-maintainer", AppName: "z-app"},
			{Maintainer: "a-maintainer", AppName: "a-app"},
		}, nil)

	pageContent, err := testObjects.Builder.BuildMaintenancePage()
	assert.Nil(t, err)

	assert.Equal(t, 3, len(pageContent.Apps))

	assert.Equal(t, "a-maintainer", pageContent.Apps[0].Maintainer)
	assert.Equal(t, "a-app", pageContent.Apps[0].AppName)

	assert.Equal(t, "a-maintainer", pageContent.Apps[1].Maintainer)
	assert.Equal(t, "z-app", pageContent.Apps[1].AppName)

	assert.Equal(t, "b-maintainer", pageContent.Apps[2].Maintainer)
	assert.Equal(t, "a-app", pageContent.Apps[2].AppName)
}

func TestBuildUserEditPageData(t *testing.T) {
	testObjects := getTestObjects(t)
	expectedUser := &tools.User{Id: 42, Username: "alice", Email: "alice@example.com"}

	testObjects.UserRepo.EXPECT().GetUserById(42).Return(expectedUser, nil)

	pageContent, err := testObjects.Builder.BuildUserEditPageData("42")
	assert.Nil(t, err)
	assert.Equal(t, "42", pageContent.UserId)
	assert.Equal(t, expectedUser, pageContent.User)
}

func TestBuildUserEditPageData_WhenUserIdIsInvalid_ReturnsError(t *testing.T) {
	testObjects := getTestObjects(t)

	pageContent, err := testObjects.Builder.BuildUserEditPageData("not-a-number")
	assert.NotNil(t, err)
	assert.Nil(t, pageContent)
}

func TestBuildAccountPageData(t *testing.T) {
	testObjects := getTestObjects(t)

	adminPage := testObjects.Builder.BuildAccountPageData(&tools.User{
		Username: "admin-user",
		Email:    "admin@example.com",
		IsAdmin:  true,
	})
	assert.Equal(t, "admin-user", adminPage.Username)
	assert.Equal(t, "admin@example.com", adminPage.Email)
	assert.Equal(t, "admin", adminPage.Role)

	userPage := testObjects.Builder.BuildAccountPageData(&tools.User{
		Username: "normal-user",
		Email:    "user@example.com",
		IsAdmin:  false,
	})
	assert.Equal(t, "normal-user", userPage.Username)
	assert.Equal(t, "user@example.com", userPage.Email)
	assert.Equal(t, "user", userPage.Role)
}
