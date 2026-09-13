//go:build component

package component

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"server/apps_advanced"
	"server/apps_basic"
	"server/tools"

	"github.com/quollix/common/quollix/api"
	"github.com/quollix/common/quollix/api_client"

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
	_, err := InstallAndStartSample(t, client, "2.0")
	assert.Nil(t, err)

	app := GetInstalledSample(t, client)
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

	appSecret, err := ReadSampleAppEnvValue(appClient, tools.ComposeEnvVars.AppSecret)
	assert.Nil(t, err)
	assert.Equal(t, app.AppSecret, appSecret)

	ianaTimezone, err := ReadSampleAppEnvValue(appClient, tools.ComposeEnvVars.IanaTimeZone)
	assert.Nil(t, err)
	assert.Equal(t, "Europe/London", ianaTimezone)

	sharedSecret, err := ReadSampleAppEnvValue(appClient, "SECRET_SAMPLE_SHARED")
	assert.Nil(t, err)
	assert.Equal(t, 64, len(sharedSecret))

	versionSecret, err := ReadSampleAppEnvValue(appClient, "SECRET_SAMPLE_VERSION_TWO")
	assert.Nil(t, err)
	assert.Equal(t, 64, len(versionSecret))
	assert.True(t, sharedSecret != versionSecret)

	migratedPassword, err := ReadSampleAppEnvValue(appClient, "SAMPLE_LEGACY_PASSWORD")
	assert.Nil(t, err)
	assert.Equal(t, 64, len(migratedPassword))
	assert.True(t, migratedPassword != "password")
}

func TestReinstallingSampleAppRenewsSecrets(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	firstInstall, err := InstallAndStartSample(t, client, "2.0")
	assert.Nil(t, err)
	firstAppClient := GetAppClient(t, client)
	firstSharedSecret := ReadRequiredSampleAppEnvValue(t, firstAppClient, "SECRET_SAMPLE_SHARED")
	firstVersionSecret := ReadRequiredSampleAppEnvValue(t, firstAppClient, "SECRET_SAMPLE_VERSION_TWO")

	assert.Nil(t, client.Apps.Delete(firstInstall.AppId))

	_, err = InstallAndStartSample(t, client, "2.0")
	assert.Nil(t, err)
	secondAppClient := GetAppClient(t, client)
	secondSharedSecret := ReadRequiredSampleAppEnvValue(t, secondAppClient, "SECRET_SAMPLE_SHARED")
	secondVersionSecret := ReadRequiredSampleAppEnvValue(t, secondAppClient, "SECRET_SAMPLE_VERSION_TWO")

	assert.True(t, firstSharedSecret != secondSharedSecret)
	assert.True(t, firstVersionSecret != secondVersionSecret)
}

func TestUpdatingSampleAppPreservesExistingSecretsAndGeneratesNewSecrets(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	appBeforeUpdate, err := InstallAndStartSample(t, client, "1.0")
	assert.Nil(t, err)
	appClientBeforeUpdate := GetAppClient(t, client)
	sharedSecretBeforeUpdate := ReadRequiredSampleAppEnvValue(t, appClientBeforeUpdate, "SECRET_SAMPLE_SHARED")
	versionOneSecret := ReadRequiredSampleAppEnvValue(t, appClientBeforeUpdate, "SECRET_SAMPLE_VERSION_ONE")
	legacyPasswordBeforeUpdate, err := ReadSampleAppEnvValue(appClientBeforeUpdate, "SAMPLE_LEGACY_PASSWORD")
	assert.Nil(t, err)
	assert.Equal(t, "password", legacyPasswordBeforeUpdate)

	assert.Nil(t, client.Apps.Update(appBeforeUpdate.AppId))

	appClientAfterUpdate := GetAppClient(t, client)
	sharedSecretAfterUpdate := ReadRequiredSampleAppEnvValue(t, appClientAfterUpdate, "SECRET_SAMPLE_SHARED")
	versionTwoSecret := ReadRequiredSampleAppEnvValue(t, appClientAfterUpdate, "SECRET_SAMPLE_VERSION_TWO")
	migratedPasswordAfterUpdate, err := ReadSampleAppEnvValue(appClientAfterUpdate, "SAMPLE_LEGACY_PASSWORD")
	assert.Nil(t, err)

	assert.Equal(t, sharedSecretBeforeUpdate, sharedSecretAfterUpdate)
	assert.Equal(t, "password", migratedPasswordAfterUpdate)
	assert.Equal(t, 64, len(versionTwoSecret))
	assert.True(t, versionTwoSecret != versionOneSecret)
	assert.True(t, versionTwoSecret != sharedSecretAfterUpdate)
}

func TestDeletingAppSecretOnlyAllowsUnusedSecrets(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	appBeforeUpdate, err := InstallSample(t, client, "1.0")
	assert.Nil(t, err)
	versionOneSecretBeforeDeletion, err := client.Apps.GetInstalledAppSecret(tools.SampleApp, "SECRET_SAMPLE_VERSION_ONE")
	assert.Nil(t, err)
	assert.True(t, versionOneSecretBeforeDeletion != "")

	assert.Nil(t, client.Apps.Update(appBeforeUpdate.AppId))
	versionOneSecretAfterUpdate, err := client.Apps.GetInstalledAppSecret(tools.SampleApp, "SECRET_SAMPLE_VERSION_ONE")
	assert.Nil(t, err)
	assert.True(t, versionOneSecretAfterUpdate != "")
	versionTwoSecretAfterUpdate, err := client.Apps.GetInstalledAppSecret(tools.SampleApp, "SECRET_SAMPLE_VERSION_TWO")
	assert.Nil(t, err)
	assert.True(t, versionTwoSecretAfterUpdate != "")

	assert.Nil(t, client.Apps.DeleteSecret(appBeforeUpdate.AppId, "SECRET_SAMPLE_VERSION_ONE"))
	appAfterUnusedSecretDeletion := GetInstalledSample(t, client)
	_, oldSecretExists := appAfterUnusedSecretDeletion.Secrets["SECRET_SAMPLE_VERSION_ONE"]
	assert.False(t, oldSecretExists)

	err = client.Apps.DeleteSecret(appBeforeUpdate.AppId, "SECRET_SAMPLE_VERSION_TWO")
	u.AssertDeepStackErrorFromRequest(t, err, apps_basic.AppSecretInUseError)
	versionTwoSecretAfterUsedSecretDeletionAttempt, err := client.Apps.GetInstalledAppSecret(tools.SampleApp, "SECRET_SAMPLE_VERSION_TWO")
	assert.Nil(t, err)
	assert.True(t, versionTwoSecretAfterUsedSecretDeletionAttempt != "")
}

func TestRegeneratingSampleAppSecretChangesStoredSecret(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	installedApp, err := InstallAndStartSample(t, client, "2.0")
	assert.Nil(t, err)
	storedSecretBeforeRegeneration, err := client.Apps.GetInstalledAppSecret(tools.SampleApp, "SECRET_SAMPLE_SHARED")
	assert.Nil(t, err)
	assert.Equal(t, 64, len(storedSecretBeforeRegeneration))

	appClient := GetAppClient(t, client)
	runningSecretBeforeRegeneration := ReadRequiredSampleAppEnvValue(t, appClient, "SECRET_SAMPLE_SHARED")
	assert.Equal(t, storedSecretBeforeRegeneration, runningSecretBeforeRegeneration)

	assert.Nil(t, client.Apps.RegenerateSecret(installedApp.AppId, "SECRET_SAMPLE_SHARED"))

	storedSecretAfterRegeneration, err := client.Apps.GetInstalledAppSecret(tools.SampleApp, "SECRET_SAMPLE_SHARED")
	assert.Nil(t, err)
	assert.Equal(t, 64, len(storedSecretAfterRegeneration))
	assert.True(t, storedSecretAfterRegeneration != storedSecretBeforeRegeneration)

	runningSecretAfterRegeneration := ReadRequiredSampleAppEnvValue(t, appClient, "SECRET_SAMPLE_SHARED")
	assert.Equal(t, runningSecretBeforeRegeneration, runningSecretAfterRegeneration)

	assert.Nil(t, client.Apps.Stop(installedApp.AppId))
	assert.Nil(t, client.Apps.Start(installedApp.AppId))

	appClientAfterRestart := GetAppClient(t, client)
	runningSecretAfterRestart := ReadRequiredSampleAppEnvValue(t, appClientAfterRestart, "SECRET_SAMPLE_SHARED")
	assert.Equal(t, storedSecretAfterRegeneration, runningSecretAfterRestart)
}

func TestUpdatingSampleAppSecretChangesStoredSecret(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	installedApp, err := InstallAndStartSample(t, client, "2.0")
	assert.Nil(t, err)
	updatedSecret := "manually-updated-secret"

	assert.Nil(t, client.Apps.UpdateSecret(installedApp.AppId, "SECRET_SAMPLE_SHARED", updatedSecret))

	storedSecretAfterUpdate, err := client.Apps.GetInstalledAppSecret(tools.SampleApp, "SECRET_SAMPLE_SHARED")
	assert.Nil(t, err)
	assert.Equal(t, updatedSecret, storedSecretAfterUpdate)
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

	assert.True(t, sampleApp.IsRunning)
	assert.Nil(t, client.Apps.Stop(sampleApp.AppId))
	sampleApp = GetInstalledSample(t, client)
	assert.False(t, sampleApp.IsRunning)
	assert.Nil(t, client.Apps.Start(sampleApp.AppId))
	sampleApp = GetInstalledSample(t, client)
	assert.True(t, sampleApp.IsRunning)
	assert.Nil(t, client.Apps.Stop(sampleApp.AppId))
	sampleApp = GetInstalledSample(t, client)
	assert.False(t, sampleApp.IsRunning)
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
	err := client.Apps.SetAccessPolicy(databaseApp.AppId, api.Policies.PublicAccessPolicy)
	assert.NotNil(t, err)
	u.AssertDeepStackErrorFromRequest(t, err, apps_basic.OperationNotAllowedOnOfficialDatabaseAppError)
}

func TestAccessPolicy_NonGroupRestricted(t *testing.T) {
	adminClient := GetClientAndLogin(t)
	RunAccessPoliciesTest(t, adminClient, []AccessPolicyTestCase{
		{
			AccessPolicy:              api.Policies.AdminOnlyAccessPolicy,
			ShouldAdminHaveAccess:     true,
			ShouldUserHaveAccess:      false,
			ShouldAnonymousHaveAccess: false,
		},
		{
			AccessPolicy:              api.Policies.PublicAccessPolicy,
			ShouldAdminHaveAccess:     true,
			ShouldUserHaveAccess:      true,
			ShouldAnonymousHaveAccess: true,
		},
		{
			AccessPolicy:              api.Policies.AuthenticatedAccessPolicy,
			ShouldAdminHaveAccess:     true,
			ShouldUserHaveAccess:      true,
			ShouldAnonymousHaveAccess: false,
		},
	})
}

func TestUnknownAppDomainReturnsAppUnavailablePage(t *testing.T) {
	statusCode, _, body, err := requestSampleAppWithUrlNoRedirect("https://unknownapp.localhost/", nil)
	assert.Nil(t, err)

	assertAppUnavailablePage(t, statusCode, body)
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
		policy                      string
		adminVisibleAppCount        int
		userVisibleAppCount         int
		anonymousVisibleAppCount    int
		shouldUserSeeSampleApp      bool
		shouldAnonymousSeeSampleApp bool
	}

	testCases := []listingExpectation{
		{
			policy:                      api.Policies.AdminOnlyAccessPolicy,
			adminVisibleAppCount:        2,
			userVisibleAppCount:         0,
			anonymousVisibleAppCount:    0,
			shouldUserSeeSampleApp:      false,
			shouldAnonymousSeeSampleApp: false,
		},
		{
			policy:                      api.Policies.AuthenticatedAccessPolicy,
			adminVisibleAppCount:        2,
			userVisibleAppCount:         1,
			anonymousVisibleAppCount:    0,
			shouldUserSeeSampleApp:      true,
			shouldAnonymousSeeSampleApp: false,
		},
		{
			policy:                      api.Policies.PublicAccessPolicy,
			adminVisibleAppCount:        2,
			userVisibleAppCount:         1,
			anonymousVisibleAppCount:    1,
			shouldUserSeeSampleApp:      true,
			shouldAnonymousSeeSampleApp: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.policy, func(t *testing.T) {
			assert.Nil(t, adminClient.Apps.SetAccessPolicy(sampleApp.AppId, testCase.policy))

			adminApps := ListInstalledApps(t, adminClient)
			assert.Equal(t, testCase.adminVisibleAppCount, len(adminApps))
			adminSampleApp, adminSampleAppExists := findAppByName(adminApps, tools.SampleApp)
			assert.True(t, adminSampleAppExists)
			assertAppSensitiveDataVisibleToAdmin(t, adminSampleApp)

			userApps := ListInstalledAppsForNonAdmin(t, userClient)
			assert.Equal(t, testCase.userVisibleAppCount, len(userApps))
			userSampleApp, userSampleAppExists := findNonAdminAppByName(userApps, tools.SampleApp)
			assert.Equal(t, testCase.shouldUserSeeSampleApp, userSampleAppExists)
			if userSampleAppExists {
				assertNonAdminAppDtoContainsSafeAppFields(t, userSampleApp, testCase.policy == api.Policies.PublicAccessPolicy)
			}

			anonymousApps := ListInstalledAppsForNonAdmin(t, anonymousClient)
			assert.Equal(t, testCase.anonymousVisibleAppCount, len(anonymousApps))
			anonymousSampleApp, anonymousSampleAppExists := findNonAdminAppByName(anonymousApps, tools.SampleApp)
			assert.Equal(t, testCase.shouldAnonymousSeeSampleApp, anonymousSampleAppExists)
			if anonymousSampleAppExists {
				assertNonAdminAppDtoContainsSafeAppFields(t, anonymousSampleApp, true)
			}
		})
	}
}

func assertAppSensitiveDataVisibleToAdmin(t *testing.T, app api.AdminAppDto) {
	assert.Equal(t, 64, len(app.ClientSecret))
	assert.Equal(t, 64, len(app.AppSecret))
	assert.True(t, len(app.VersionContent) > 0)
	assert.True(t, len(app.Secrets) > 0)
	assert.Equal(t, 64, len(app.Secrets["SECRET_SAMPLE_SHARED"]))
}

func assertNonAdminAppDtoContainsSafeAppFields(t *testing.T, app api.NonAdminAppDto, expectedIsPublic bool) {
	assert.Equal(t, tools.SampleMaintainer, app.Maintainer)
	assert.Equal(t, tools.SampleApp, app.AppName)
	assert.Equal(t, expectedIsPublic, app.IsPublic)
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
	assert.Equal(t, api.Policies.AdminOnlyAccessPolicy, app.AccessPolicy)
}

func TestManualBackupAppearsInCurrentOperations(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	operations, isOngoing, err := client.Apps.GetCurrentOperations()
	assert.Nil(t, err)
	assert.False(t, isOngoing)
	assert.Equal(t, []string{}, operations)

	app, err := InstallAndStartSample(t, client, "2.0")
	assert.Nil(t, err)
	configureBackupRepo(t, client)

	backupDone := make(chan error, 1)
	go func() {
		backupDone <- client.Backups.Create(app.AppId)
	}()

	time.Sleep(100 * time.Millisecond)
	operations, isOngoing, err = client.Apps.GetCurrentOperations()
	assert.Nil(t, err)
	assert.True(t, isOngoing)
	assert.Equal(t, []string{"backing up 'sampleapp'"}, operations)

	select {
	case err = <-backupDone:
		assert.Nil(t, err)
	case <-time.After(10 * time.Second):
		t.Fatal("backup did not finish")
	}

	operations, isOngoing, err = client.Apps.GetCurrentOperations()
	assert.Nil(t, err)
	assert.False(t, isOngoing)
	assert.Equal(t, []string{}, operations)
}

func TestUploadToAndDownloadFromApplication(t *testing.T) {
	sampleAppContent := getSampleAppContent()

	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	originalVersionFile := api.BinaryFile{
		FileName: getSampleFileNameForAppUpload(),
		Content:  sampleAppContent,
	}

	assert.Nil(t, client.Apps.UploadVersionFile(originalVersionFile))

	sampleApp := GetInstalledSample(t, client)
	assertUploadedSampleApp(t, sampleApp, tools.SampleAppVersion2Name)
	assert.Equal(t, tools.SampleAppVersion2CreationTimestamp, sampleApp.VersionCreationTimestamp)
	assert.Equal(t, "3001", sampleApp.Port)
	assert.Equal(t, 16, len(sampleApp.ClientId))
	assert.Equal(t, 64, len(sampleApp.ClientSecret))
	assert.Equal(t, 64, len(sampleApp.AppSecret))
	assert.Equal(t, api.Policies.AdminOnlyAccessPolicy, sampleApp.AccessPolicy)

	downloadedVersionFile, err := client.Apps.DownloadVersionFile(sampleApp.AppId)
	assert.Nil(t, err)
	assert.Equal(t, originalVersionFile.FileName, downloadedVersionFile.FileName)
	assert.Equal(t, originalVersionFile.Content, downloadedVersionFile.Content)
}

func TestUploadTestAppDefinitionToApplication(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	uploadStartedAt := time.Now().UTC()
	uploadDeadline := uploadStartedAt.Add(5 * time.Minute)
	versionFile := api.BinaryFile{
		FileName: tools.SampleApp + ".yml",
		Content:  getSampleAppContent(),
	}

	assert.Nil(t, client.Apps.UploadVersionFile(versionFile))

	sampleApp := GetInstalledSample(t, client)
	assertUploadedSampleApp(t, sampleApp, apps_basic.TestAppDefinitionVersion)
	assert.True(t, uploadStartedAt.Before(sampleApp.VersionCreationTimestamp) || uploadStartedAt.Equal(sampleApp.VersionCreationTimestamp))
	assert.True(t, uploadDeadline.After(sampleApp.VersionCreationTimestamp) || uploadDeadline.Equal(sampleApp.VersionCreationTimestamp))
}

func assertUploadedSampleApp(t *testing.T, sampleApp *api.AdminAppDto, expectedVersionName string) {
	assert.Equal(t, tools.SampleMaintainer, sampleApp.Maintainer)
	assert.Equal(t, tools.SampleApp, sampleApp.AppName)
	assert.Equal(t, expectedVersionName, sampleApp.VersionName)
	assert.True(t, sampleApp.IsRunning)
	assert.False(t, sampleApp.AutomaticUpdatesEnabled)
	assert.True(t, sampleApp.AutomaticBackupsEnabled)
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

	originalVersionFile := api.BinaryFile{
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

	originalVersionFile := api.BinaryFile{
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

	sampleAppOld, err := InstallAndStartSample(t, client, "1.0")
	assert.Nil(t, err)
	assert.Equal(t, tools.SampleAppVersion1Name, sampleAppOld.VersionName)
	assert.Equal(t, tools.SampleAppVersion1CreationTimestamp, sampleAppOld.VersionCreationTimestamp)
	assert.Equal(t, "3000", sampleAppOld.Port)
	assert.True(t, sampleAppOld.AutomaticUpdatesEnabled)
	assert.True(t, sampleAppOld.AutomaticBackupsEnabled)

	originalVersionFile := api.BinaryFile{
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
	assert.False(t, sampleAppNew.AutomaticUpdatesEnabled)
	assert.True(t, sampleAppNew.AutomaticBackupsEnabled)
}

func TestUploadingAppAlreadyExistWithDifferentMaintainerIsRejected(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	_, err := InstallSample(t, client, "1.0")
	assert.Nil(t, err)

	fileNameWithDifferentMaintainer := strings.ReplaceAll(getSampleFileNameForAppUpload(), tools.SampleMaintainer, tools.SampleMaintainer+"2")
	sampleAppContentWithDifferentMaintainer := strings.ReplaceAll(string(getSampleAppContent()), tools.SampleMaintainer, tools.SampleMaintainer+"2")

	originalVersionFile := api.BinaryFile{
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

	originalVersionFile := api.BinaryFile{
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

	originalVersionFile := api.BinaryFile{
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
