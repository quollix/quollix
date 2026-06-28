//go:build component

package component

import (
	"fmt"
	"server/apps_advanced"
	"server/apps_basic"
	"server/tests/api_client"
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
	assert.Equal(t, 1, len(ListInstalledApps(t, client)))
	installedSampleApp, err := InstallSample(t, client, "2.0")
	assert.Nil(t, err)

	assert.Equal(t, 2, len(ListInstalledApps(t, client)))
	assert.Nil(t, client.Apps.Delete(installedSampleApp.AppId))
	assert.Equal(t, 1, len(ListInstalledApps(t, client)))
}

func TestSampleAppReceivesConfiguredEnvValues(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()
	_, err := InstallSample(t, client, "2.0")
	assert.Nil(t, err)

	app := GetInstalledSample(t, client)
	assert.Nil(t, client.Apps.Start(app.AppId))
	appClient := GetAppClient(t, client)

	serverURL, err := ReadSampleAppEnvValue(appClient, "SERVER_URL")
	assert.Nil(t, err)
	assert.Equal(t, "https://sampleapp.localhost", serverURL)

	clientId, err := ReadSampleAppEnvValue(appClient, "OIDC_CLIENT_ID")
	assert.Nil(t, err)
	assert.Equal(t, app.ClientId, clientId)

	clientSecret, err := ReadSampleAppEnvValue(appClient, "OIDC_CLIENT_SECRET")
	assert.Nil(t, err)
	assert.Equal(t, app.ClientSecret, clientSecret)

	ianaTimezone, err := ReadSampleAppEnvValue(appClient, tools.ComposeEnvVars.IanaTimeZone)
	assert.Nil(t, err)
	assert.Equal(t, "Europe/London", ianaTimezone)
}

func TestStartingAppAlreadyRunningIsPossible(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()
	app, err := InstallSample(t, client, "2.0")
	assert.Nil(t, err)

	assert.Nil(t, client.Apps.Start(app.AppId))
	assert.Nil(t, client.Apps.Start(app.AppId))
}

func TestStartingAndStoppingApps(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()
	_, err := InstallSample(t, client, "2.0")
	assert.Nil(t, err)

	sampleApp := GetInstalledSample(t, client)

	assert.False(t, sampleApp.IsRunning)
	assert.Nil(t, client.Apps.Start(sampleApp.AppId))
	sampleApp = GetInstalledSample(t, client)
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

	installedApps := ListInstalledApps(t, client)
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

	apps := ListInstalledApps(t, client)
	assert.Equal(t, 1, len(apps))
	app, err := InstallSample(t, client, "2.0")
	assert.Nil(t, err)
	apps = ListInstalledApps(t, client)
	assert.Equal(t, 2, len(apps))
	assert.Nil(t, client.Apps.Start(app.AppId))

	ExpectDockerObject(t, Network, true)
	ExpectDockerObject(t, Volume, true)
	ExpectDockerObject(t, Container, true)

	assert.Nil(t, client.Apps.Delete(app.AppId))
	apps = ListInstalledApps(t, client)
	assert.Equal(t, 1, len(apps))

	ExpectDockerObject(t, Network, false)
	ExpectDockerObject(t, Volume, false)
	ExpectDockerObject(t, Container, false)
}

func TestCantChangeAccessPolicyOfDatabaseApp(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	installedApps := ListInstalledApps(t, client)
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

	sampleApp, err := InstallSample(t, adminClient, "2.0")
	assert.Nil(t, err)
	assert.Nil(t, adminClient.Apps.Start(sampleApp.AppId))
	InviteUserAndSetPassword(t, adminClient, SampleUsername, "userpassword", SampleUserEmail)

	userClient := api_client.NewQuollixClient()
	assert.Nil(t, userClient.Auth.SignIn(SampleUsername, "userpassword"))
	anonymousClient := api_client.NewQuollixClient()

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

			adminApps := ListInstalledApps(t, adminClient)
			assert.Equal(t, testCase.adminVisibleAppCount, len(adminApps))

			userApps := ListInstalledApps(t, userClient)
			assert.Equal(t, testCase.userVisibleAppCount, len(userApps))

			anonymousApps := ListInstalledApps(t, anonymousClient)
			assert.Equal(t, testCase.anonymousVisibleAppCount, len(anonymousApps))
		})
	}
}

func TestSetUnknownAccessPolicy(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()
	app, err := InstallSample(t, client, "2.0")
	assert.Nil(t, err)

	err = client.Apps.SetAccessPolicy(app.AppId, "non-existing-policy")
	assert.NotNil(t, err)
	u.AssertDeepStackErrorFromRequest(t, err, apps_basic.InvalidAccessPolicyError)

	app = GetInstalledSample(t, client)
	assert.Equal(t, tools.Policies.AdminOnlyAccessPolicy, app.AccessPolicy)
}

func TestAppOperation(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	operations, isOngoing, err := client.Apps.GetCurrentOperations()
	assert.Nil(t, err)
	assert.False(t, isOngoing)
	assert.Equal(t, []string{}, operations)

	app, err := InstallSample(t, client, "2.0")
	assert.Nil(t, err)
	configureBackupRepo(t, client)
	go func() {
		err := client.Backups.Create(app.AppId)
		if err != nil {
			t.Error(err)
		}
	}()

	time.Sleep(100 * time.Millisecond)
	operations, isOngoing, err = client.Apps.GetCurrentOperations()
	assert.Nil(t, err)
	assert.True(t, isOngoing)
	assert.Equal(t, []string{"backing up 'sampleapp'"}, operations)

	deadline := time.Now().Add(3 * time.Second)
	for {
		_, isOngoing, err = client.Apps.GetCurrentOperations()
		assert.Nil(t, err)
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

	sampleApp := GetInstalledSample(t, client)
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

	downloadedVersionFile, err := client.Apps.DownloadVersionFile(sampleApp.AppId)
	assert.Nil(t, err)
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
	configureBackupRepo(t, client)

	sampleAppOld, err := InstallSample(t, client, "1.0")
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

	sampleAppNew := GetInstalledSample(t, client)
	assert.Equal(t, tools.SampleAppVersion2Name, sampleAppNew.VersionName)
	assert.Equal(t, tools.SampleAppVersion2CreationTimestamp, sampleAppNew.VersionCreationTimestamp)
	assert.Equal(t, "3001", sampleAppNew.Port)
	assert.Equal(t, originalVersionFile.Content, sampleAppNew.VersionContent)
}

func TestUploadingAppAlreadyExistWithDifferentMaintainerIsRejected(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	_, err := InstallSample(t, client, "1.0")
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

	_, err := InstallSample(t, client, "1.0")
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

	app, err := InstallSample(t, client, "2.0")
	assert.Nil(t, err)
	err = client.Apps.Update(app.AppId)
	u.AssertDeepStackErrorFromRequest(t, err, apps_advanced.CantUpdateAppError)
}

func TestUpdatesAndPreUpdateBackupCreation(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	appBeforeUpdate, err := InstallSample(t, client, "1.0")
	assert.Nil(t, err)

	assert.Nil(t, client.Apps.Start(appBeforeUpdate.AppId))
	appClient := GetAppClient(t, client)
	assert.Nil(t, AssertSampleAppContent(appClient, "this is version 1.0"))
	assert.Nil(t, StoreStringInSampleApp(appClient, "persisted before update"))
	persistedContent, err := ReadStringFromSampleApp(appClient)
	assert.Nil(t, err)
	assert.Equal(t, "persisted before update", persistedContent)

	configureBackupRepo(t, client)
	appBackups, err := client.Backups.ListByApp(tools.SampleMaintainer, tools.SampleApp)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(appBackups))

	assert.Nil(t, client.Apps.Update(appBeforeUpdate.AppId))
	appAfterUpdate := GetInstalledSample(t, client)
	assert.Equal(t, "2.0", appAfterUpdate.VersionName)
	assert.Nil(t, AssertSampleAppContent(appClient, "this is version 2.0"))
	persistedContent, err = ReadStringFromSampleApp(appClient)
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
