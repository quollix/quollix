//go:build component

package component

import (
	"fmt"
	"server/apps_advanced"
	"server/apps_basic"
	"server/tools"
	"strings"
	"testing"
	"time"

	"github.com/quollix/common/assert"
	u "github.com/quollix/common/utils"
	"github.com/quollix/common/validation"
	"github.com/quollix/deepstack"
)

func TestPruningApp(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()
	assert.Equal(t, 1, len(client.Apps.ListInstalled()))
	installedSampleApp, err := client.Apps.InstallSample("2.0")
	assert.Nil(t, err)

	assert.Equal(t, 2, len(client.Apps.ListInstalled()))
	assert.Nil(t, client.Apps.Delete(installedSampleApp.AppId))
	assert.Equal(t, 1, len(client.Apps.ListInstalled()))
}

func TestSampleAppReceivesConfiguredEnvValues(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()
	_, err := client.Apps.InstallSample("2.0")
	assert.Nil(t, err)

	app := client.Apps.GetInstalledSample()
	assert.Nil(t, client.Apps.Start(app.AppId))
	appClient := GetAppClient(t, client)

	serverURL, err := appClient.Content.ReadSampleAppEnvValue("SERVER_URL")
	assert.Nil(t, err)
	assert.Equal(t, "https://sampleapp.localhost", serverURL)

	clientId, err := appClient.Content.ReadSampleAppEnvValue("OIDC_CLIENT_ID")
	assert.Nil(t, err)
	assert.Equal(t, app.ClientId, clientId)

	clientSecret, err := appClient.Content.ReadSampleAppEnvValue("OIDC_CLIENT_SECRET")
	assert.Nil(t, err)
	assert.Equal(t, app.ClientSecret, clientSecret)

	ianaTimezone, err := appClient.Content.ReadSampleAppEnvValue(tools.ComposeEnvVars.IanaTimeZone)
	assert.Nil(t, err)
	assert.Equal(t, "Europe/London", ianaTimezone)
}

func TestStartingAppAlreadyRunningIsPossible(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()
	app, err := client.Apps.InstallSample("2.0")
	assert.Nil(t, err)

	assert.Nil(t, client.Apps.Start(app.AppId))
	assert.Nil(t, client.Apps.Start(app.AppId))
}

func TestStartingAndStoppingApps(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()
	_, err := client.Apps.InstallSample("2.0")
	assert.Nil(t, err)

	sampleApp := client.Apps.GetInstalledSample()

	assert.False(t, sampleApp.IsRunning)
	assert.Nil(t, client.Apps.Start(sampleApp.AppId))
	sampleApp = client.Apps.GetInstalledSample()
	assert.True(t, sampleApp.IsRunning)
}

func TestStopAppNotExisting(t *testing.T) {
	client := GetClientAndLogin(t)
	notExistingId := "123"
	err := client.Apps.Stop(notExistingId)
	assert.NotNil(t, err)
	u.AssertDeepStackErrorFromRequest(t, err, u.OperationFailedError)
}

func TestProhibitedOperationsOnPostgresApp(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	installedApps := client.Apps.ListInstalled()
	postgresApp := installedApps[0]

	err := client.Apps.Start(postgresApp.AppId)
	assert.NotNil(t, err)
	u.AssertDeepStackErrorFromRequest(t, err, apps_basic.OperationNotAllowedOnOfficialDatabaseAppError)

	err = client.Apps.Update(postgresApp.AppId)
	assert.NotNil(t, err)
	u.AssertDeepStackErrorFromRequest(t, err, apps_basic.OperationNotAllowedOnSystemAppError)

	err = client.Apps.Delete(postgresApp.AppId)
	assert.NotNil(t, err)
	u.AssertDeepStackErrorFromRequest(t, err, apps_basic.OperationNotAllowedOnOfficialDatabaseAppError)
}

func TestDeletingApp(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	apps := client.Apps.ListInstalled()
	assert.Equal(t, 1, len(apps))
	app, err := client.Apps.InstallSample("2.0")
	assert.Nil(t, err)
	apps = client.Apps.ListInstalled()
	assert.Equal(t, 2, len(apps))
	assert.Nil(t, client.Apps.Start(app.AppId))

	ExpectDockerObject(t, Network, true)
	ExpectDockerObject(t, Volume, true)
	ExpectDockerObject(t, Container, true)

	assert.Nil(t, client.Apps.Delete(app.AppId))
	apps = client.Apps.ListInstalled()
	assert.Equal(t, 1, len(apps))

	ExpectDockerObject(t, Network, false)
	ExpectDockerObject(t, Volume, false)
	ExpectDockerObject(t, Container, false)
}

func TestCantChangeAccessPolicyOfDatabaseApp(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	installedApps := client.Apps.ListInstalled()
	assert.Equal(t, 1, len(installedApps))
	databaseApp := installedApps[0]
	err := client.Apps.SetAccessPolicy(databaseApp.AppId, tools.Policies.PublicAccessPolicy)
	assert.NotNil(t, err)
	u.AssertDeepStackErrorFromRequest(t, err, apps_basic.OperationNotAllowedOnOfficialDatabaseAppError)
}

func TestAccessPolicy_NonGroupRestricted(t *testing.T) {
	adminClient := GetClientAndLogin(t)
	RunAccessPoliciesTest(t, adminClient, []AccessPolicyTestCase{
		{
			AccessPolicy:              tools.Policies.AdminOnlyAccessPolicy,
			ShouldAdminHaveAccess:     true,
			ShouldUserHaveAccess:      false,
			ShouldAnonymousHaveAccess: false,
		},
		{
			AccessPolicy:              tools.Policies.PublicAccessPolicy,
			ShouldAdminHaveAccess:     true,
			ShouldUserHaveAccess:      true,
			ShouldAnonymousHaveAccess: true,
		},
		{
			AccessPolicy:              tools.Policies.AuthenticatedAccessPolicy,
			ShouldAdminHaveAccess:     true,
			ShouldUserHaveAccess:      true,
			ShouldAnonymousHaveAccess: false,
		},
	})
}

func TestInstalledAppListing_ByAccessPolicy(t *testing.T) {
	adminClient := GetClientAndLogin(t)
	defer adminClient.Test.ResetTestState()

	sampleApp, err := adminClient.Apps.InstallSample("2.0")
	assert.Nil(t, err)
	adminClient.Apps.Start(sampleApp.AppId)
	InviteUserAndSetPassword(adminClient, SampleUsername, "userpassword", SampleUserEmail)

	userClient := GetQuollixClient(t)
	assert.Nil(t, userClient.Auth.Login(SampleUsername, "userpassword"))
	anonymousClient := GetQuollixClient(t)

	type listingExpectation struct {
		policy                   string
		adminVisibleAppCount     int
		userVisibleAppCount      int
		anonymousVisibleAppCount int
	}

	testCases := []listingExpectation{
		{
			policy:                   tools.Policies.AdminOnlyAccessPolicy,
			adminVisibleAppCount:     2,
			userVisibleAppCount:      0,
			anonymousVisibleAppCount: 0,
		},
		{
			policy:                   tools.Policies.AuthenticatedAccessPolicy,
			adminVisibleAppCount:     2,
			userVisibleAppCount:      1,
			anonymousVisibleAppCount: 0,
		},
		{
			policy:                   tools.Policies.PublicAccessPolicy,
			adminVisibleAppCount:     2,
			userVisibleAppCount:      1,
			anonymousVisibleAppCount: 1,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.policy, func(t *testing.T) {
			assert.Nil(t, adminClient.Apps.SetAccessPolicy(sampleApp.AppId, testCase.policy))

			adminApps := adminClient.Apps.ListInstalled()
			assert.Equal(t, testCase.adminVisibleAppCount, len(adminApps))

			userApps := userClient.Apps.ListInstalled()
			assert.Equal(t, testCase.userVisibleAppCount, len(userApps))

			anonymousApps := anonymousClient.Apps.ListInstalled()
			assert.Equal(t, testCase.anonymousVisibleAppCount, len(anonymousApps))
		})
	}
}

func TestSetUnknownAccessPolicy(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()
	app, err := client.Apps.InstallSample("2.0")
	assert.Nil(t, err)

	err = client.Apps.SetAccessPolicy(app.AppId, "non-existing-policy")
	assert.NotNil(t, err)
	u.AssertDeepStackErrorFromRequest(t, err, apps_basic.InvalidAccessPolicyError)

	app = client.Apps.GetInstalledSample()
	assert.Equal(t, tools.Policies.AdminOnlyAccessPolicy, app.AccessPolicy)
}

func TestAppOperation(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	operations, isOngoing := client.Apps.GetCurrentOperations()
	assert.False(t, isOngoing)
	assert.Equal(t, []string{}, operations)

	app, err := client.Apps.InstallSample("2.0")
	assert.Nil(t, err)
	configureBackupRepo(client)
	go func() {
		err := client.Backups.Create(app.AppId)
		if err != nil {
			t.Error(err)
		}
	}()

	time.Sleep(100 * time.Millisecond)
	operations, isOngoing = client.Apps.GetCurrentOperations()
	assert.True(t, isOngoing)
	assert.Equal(t, []string{"backing up 'sampleapp'"}, operations)

	deadline := time.Now().Add(3 * time.Second)
	for {
		_, isOngoing = client.Apps.GetCurrentOperations()
		if !isOngoing {
			break
		}
		if time.Now().After(deadline) {
			t.Fail()
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func TestUploadToAndDownloadFromApplication(t *testing.T) {
	sampleAppContent := getSampleAppContent()

	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	originalVersionFile := tools.BinaryFile{
		FileName: getSampleFileNameForAppUpload(),
		Content:  sampleAppContent,
	}

	assert.Nil(t, client.Apps.UploadVersionFile(originalVersionFile))

	sampleApp := client.Apps.GetInstalledSample()
	assert.Equal(t, tools.SampleMaintainer, sampleApp.Maintainer)
	assert.Equal(t, tools.SampleApp, sampleApp.AppName)
	assert.Equal(t, tools.SampleAppVersion2Name, sampleApp.VersionName)
	assert.False(t, sampleApp.IsRunning)

	assert.Equal(t, tools.SampleAppVersion2CreationTimestamp, sampleApp.VersionCreationTimestamp)

	assert.Equal(t, "3001", sampleApp.Port)
	assert.Equal(t, 16, len(sampleApp.ClientId))
	assert.Equal(t, 64, len(sampleApp.ClientSecret))

	assert.True(t, sampleApp.AutomaticBackupsEnabled)
	assert.False(t, sampleApp.AutomaticUpdatesEnabled)

	assert.Equal(t, tools.Policies.AdminOnlyAccessPolicy, sampleApp.AccessPolicy)

	downloadedVersionFile := client.Apps.DownloadVersionFile(sampleApp.AppId)
	assert.Equal(t, originalVersionFile.FileName, downloadedVersionFile.FileName)
	assert.Equal(t, originalVersionFile.Content, downloadedVersionFile.Content)
}

func getSampleAppContent() []byte {
	return []byte(tools.SampleAppVersion2ComposeYAML)
}

func TestUploadWithBadContent(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	sampleAppContentString := string(getSampleAppContent())
	invalidContainerName := fmt.Sprintf("container_name: %s_%s_%s", tools.SampleMaintainer, tools.SampleApp, tools.SampleApp)
	replacementContainerName := fmt.Sprintf("container_name: %s2_%s_%s", tools.SampleMaintainer, tools.SampleApp, tools.SampleApp)
	badSampleAppContent := []byte(strings.Replace(sampleAppContentString, invalidContainerName, replacementContainerName, 1))

	originalVersionFile := tools.BinaryFile{
		FileName: getSampleFileNameForAppUpload(),
		Content:  badSampleAppContent,
	}
	err := client.Apps.UploadVersionFile(originalVersionFile)
	assert.NotNil(t, err)
	deepStackError, ok := err.(*deepstack.DeepStackError)
	assert.True(t, ok)
	errorMessageAny, ok := deepStackError.Context["response_body"]
	assert.True(t, ok)
	errorMessage, ok := errorMessageAny.(string)
	assert.True(t, ok)
	assert.True(t, strings.Contains(errorMessage, "service has invalid container_name"))
}

func getSampleFileNameForAppUpload() string {
	return fmt.Sprintf("%s_%s_%s_%s.yml", tools.SampleMaintainer, tools.SampleApp, tools.SampleAppVersion2Name, tools.SampleAppVersion2CreationTimestamp.Format(apps_basic.VersionFileUploadTimestampLayout))
}

func TestUploadWithBadFileNameFormat(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	originalVersionFile := tools.BinaryFile{
		FileName: getSampleFileNameForAppUpload() + "2",
		Content:  getSampleAppContent(),
	}
	err := client.Apps.UploadVersionFile(originalVersionFile)
	u.AssertDeepStackErrorFromRequest(t, err, "file must end with .yml")
}

func TestUploadingAppAlreadyExistingUpdatesIt(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()
	configureBackupRepo(client)

	sampleAppOld, err := client.Apps.InstallSample("1.0")
	assert.Nil(t, err)
	assert.Equal(t, tools.SampleAppVersion1Name, sampleAppOld.VersionName)
	assert.Equal(t, tools.SampleAppVersion1CreationTimestamp, sampleAppOld.VersionCreationTimestamp)
	assert.Equal(t, "3000", sampleAppOld.Port)

	originalVersionFile := tools.BinaryFile{
		FileName: getSampleFileNameForAppUpload(),
		Content:  getSampleAppContent(),
	}

	backupsBeforeUpload, err := client.Backups.ListByApp(sampleAppOld.Maintainer, sampleAppOld.AppName)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(backupsBeforeUpload))

	assert.Nil(t, client.Apps.UploadVersionFile(originalVersionFile))

	backupsAfterUpload, err := client.Backups.ListByApp(sampleAppOld.Maintainer, sampleAppOld.AppName)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(backupsAfterUpload))
	backup := backupsAfterUpload[0]
	assert.Equal(t, tools.PreUpdateBackupDescription, backup.Description)

	sampleAppNew := client.Apps.GetInstalledSample()
	assert.Equal(t, tools.SampleAppVersion2Name, sampleAppNew.VersionName)
	assert.Equal(t, tools.SampleAppVersion2CreationTimestamp, sampleAppNew.VersionCreationTimestamp)
	assert.Equal(t, "3001", sampleAppNew.Port)
	assert.Equal(t, originalVersionFile.Content, sampleAppNew.VersionContent)
}

func TestUploadingAppAlreadyExistWithDifferentMaintainerIsRejected(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	_, err := client.Apps.InstallSample("1.0")
	assert.Nil(t, err)

	fileNameWithDifferentMaintainer := strings.ReplaceAll(getSampleFileNameForAppUpload(), tools.SampleMaintainer, tools.SampleMaintainer+"2")
	sampleAppContentWithDifferentMaintainer := strings.ReplaceAll(string(getSampleAppContent()), tools.SampleMaintainer, tools.SampleMaintainer+"2")

	originalVersionFile := tools.BinaryFile{
		FileName: fileNameWithDifferentMaintainer,
		Content:  []byte(sampleAppContentWithDifferentMaintainer),
	}
	err = client.Apps.UploadVersionFile(originalVersionFile)
	u.AssertDeepStackErrorFromRequest(t, err, apps_advanced.AppFromAnotherMaintainerExistsAlreadyError)
}

func TestUploadingOlderVersionMakesUpdateFail(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	_, err := client.Apps.InstallSample("1.0")
	assert.Nil(t, err)

	sampleNameWithOlderVersion := getSampleFileNameForAppUpload()
	sampleNameWithOlderVersion = strings.ReplaceAll(sampleNameWithOlderVersion, "2021", "1999")

	originalVersionFile := tools.BinaryFile{
		FileName: sampleNameWithOlderVersion,
		Content:  getSampleAppContent(),
	}

	err = client.Apps.UploadVersionFile(originalVersionFile)
	u.AssertDeepStackErrorFromRequest(t, err, apps_advanced.CanNotUploadOlderAppVersionOverNewer)
}

func TestUploadingSystemAppFails(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	sampleSystemAppName := getSampleFileNameForAppUpload()
	sampleSystemAppName = strings.ReplaceAll(sampleSystemAppName, tools.SampleApp, u.OfficialDatabaseAppName)

	originalVersionFile := tools.BinaryFile{
		FileName: sampleSystemAppName,
		Content:  getSampleAppContent(),
	}

	err := client.Apps.UploadVersionFile(originalVersionFile)
	u.AssertDeepStackErrorFromRequest(t, err, validation.SystemAppNamesAreAlreadyReserved)
}

func TestErrorWhenUpdatingAndLatestVersionAlreadyInstalled(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	app, err := client.Apps.InstallSample("2.0")
	assert.Nil(t, err)
	err = client.Apps.Update(app.AppId)
	u.AssertDeepStackErrorFromRequest(t, err, apps_advanced.CantUpdateAppError)
}

func TestUpdatesAndPreUpdateBackupCreation(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	appBeforeUpdate, err := client.Apps.InstallSample("1.0")
	assert.Nil(t, err)

	assert.Nil(t, client.Apps.Start(appBeforeUpdate.AppId))
	appClient := GetAppClient(t, client)
	assert.Nil(t, appClient.Content.AssertContent("this is version 1.0"))
	assert.Nil(t, appClient.Content.StoreStringInSampleApp("persisted before update"))
	persistedContent, err := appClient.Content.ReadStringFromSampleApp()
	assert.Nil(t, err)
	assert.Equal(t, "persisted before update", persistedContent)

	configureBackupRepo(client)
	appBackups, err := client.Backups.ListByApp(tools.SampleMaintainer, tools.SampleApp)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(appBackups))

	assert.Nil(t, client.Apps.Update(appBeforeUpdate.AppId))
	appAfterUpdate := client.Apps.GetInstalledSample()
	assert.Equal(t, "2.0", appAfterUpdate.VersionName)
	assert.Nil(t, appClient.Content.AssertContent("this is version 2.0"))
	persistedContent, err = appClient.Content.ReadStringFromSampleApp()
	assert.Nil(t, err)
	assert.Equal(t, "persisted before update", persistedContent)

	appBackups, err = client.Backups.ListByApp(tools.SampleMaintainer, tools.SampleApp)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(appBackups))
	backup := appBackups[0]
	assert.Equal(t, "1.0", backup.VersionName)
	assert.Equal(t, tools.PreUpdateBackupDescription, backup.Description)
	assert.Equal(t, tools.SampleApp, backup.AppName)
	assert.Equal(t, tools.SampleMaintainer, backup.Maintainer)

	expectedAppState := *appBeforeUpdate
	expectedAppState.VersionName = "2.0"
	expectedAppState.VersionCreationTimestamp = tools.SampleAppVersion2CreationTimestamp
	expectedAppState.Port = "3001"
	expectedAppState.VersionContent = appAfterUpdate.VersionContent // not ideal, but if endpoint check worked previously, then the correct version content of docker-compose.yml was implicitly used to run the app
	expectedAppState.IsRunning = true
	assertAppState(t, &expectedAppState, appAfterUpdate)
}
