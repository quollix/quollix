//go:build component

package component

import (
	"strings"
	"testing"
	"time"

	"server/apps_basic"
	"server/backup_server"
	"server/system_config_migrations"
	"server/tests/api_client"
	"server/tools"

	"github.com/quollix/common/assert"
	u "github.com/quollix/common/utils"
)

const notExistingBackupId = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

func TestBackupLifecycle(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	_, err := InstallSample(t, client, "2.0")
	assert.Nil(t, err)
	configureBackupRepo(t, client)

	appBackups, err := client.Backups.ListByApp(tools.SampleMaintainer, tools.SampleApp)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(appBackups))

	assert.Nil(t, CreateSampleBackup(t, client))
	// we create two backups to check if backend can handle this, and backup operations are heavy so we combine this feature with whole lifecycle flow
	assert.Nil(t, CreateSampleBackup(t, client))

	appBackups, err = client.Backups.ListByApp(tools.SampleMaintainer, tools.SampleApp)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(appBackups))

	for _, backup := range appBackups {
		assert.Equal(t, tools.SampleMaintainer, backup.Maintainer)
		assert.Equal(t, tools.SampleApp, backup.AppName)
		assert.Equal(t, "2.0", backup.VersionName)
		assert.Equal(t, tools.ManualBackupDescription, backup.Description)
		assert.True(t, backup.BackupCreationTimestamp.Before(time.Now()))
		assert.True(t, backup.BackupCreationTimestamp.After(time.Now().Add(-1*time.Minute)))
		assert.Equal(t, backup.ApplicationVersion, tools.ApplicationVersion)
	}

	assert.Nil(t, client.Backups.Delete([]string{appBackups[0].BackupId, appBackups[1].BackupId}))

	appBackups, err = client.Backups.ListByApp(tools.SampleMaintainer, tools.SampleApp)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(appBackups))
}

func TestBackupOperationsRequireRepoConfigured(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	app, err := InstallSample(t, client, "2.0")
	assert.Nil(t, err)
	assert.Nil(t, client.Apps.Start(app.AppId))

	err = client.Backups.Create(app.AppId)
	u.AssertDeepStackErrorFromRequest(t, err, backup_server.ErrBackupRepoNotConfigured)

	err = client.Backups.Delete([]string{notExistingBackupId})
	u.AssertDeepStackErrorFromRequest(t, err, backup_server.ErrBackupRepoNotConfigured)

	err = client.Backups.Restore(notExistingBackupId)
	u.AssertDeepStackErrorFromRequest(t, err, backup_server.ErrBackupRepoNotConfigured)

	_, err = client.Backups.ListAppsInRepository()
	u.AssertDeepStackErrorFromRequest(t, err, backup_server.ErrBackupRepoNotConfigured)

	_, err = client.Backups.ListByApp(tools.SampleMaintainer, tools.SampleApp)
	u.AssertDeepStackErrorFromRequest(t, err, backup_server.ErrBackupRepoNotConfigured)
}

func TestAppListingInBackupRepo(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	_, err := InstallSample(t, client, "2.0")
	assert.Nil(t, err)

	configureBackupRepo(t, client)
	assertAppsNumbersInBackupRepo(t, client, 0)

	assert.Nil(t, CreateSampleBackup(t, client))
	assertAppsNumbersInBackupRepo(t, client, 1)
}

func TestFileBackupAndRestore(t *testing.T) {
	fixture := createSampleAppWithBackupAndDeleteIt(t, func(client *api_client.QuollixClient, app *apps_basic.AppDto) {
		appClient := GetAppClient(t, client)
		assert.Nil(t, StoreStringInSampleApp(appClient, "sample string"))

		appResponseString, err := ReadStringFromSampleApp(appClient)
		assert.Nil(t, err)
		assert.Equal(t, "sample string", appResponseString)
	})
	defer fixture.client.Test.ResetTestState()

	assert.Nil(t, fixture.client.Backups.Restore(fixture.backup.BackupId))

	appClient := GetAppClient(t, fixture.client)
	appResponseString, err := ReadStringFromSampleApp(appClient)
	assert.Nil(t, err)
	assert.Equal(t, "sample string", appResponseString)
}

func TestDatabaseBackupAndRestore(t *testing.T) {
	client := prepareSshRemoteServerSetup(t)
	defer client.Test.ResetTestState()

	InviteUserAndSetPassword(t, client, SampleUsername, SampleUserPassword, SampleUserEmail)

	installedApps := ListInstalledApps(t, client)
	assert.Equal(t, 1, len(installedApps))
	postgresApp := installedApps[0]

	assert.Nil(t, client.Backups.Create(postgresApp.AppId))

	testuser := GetRequiredUserByUsername(t, client, SampleUsername)
	assert.Equal(t, 2, len(ListUsers(t, client)))
	assert.Nil(t, client.Users.Delete(testuser.Id))
	assert.Equal(t, 1, len(ListUsers(t, client)))

	backup := getSingleBackup(t, client, u.OfficialMaintainer, u.OfficialDatabaseAppName)
	assert.Nil(t, client.Backups.Restore(backup.BackupId))

	assert.Equal(t, 2, len(ListUsers(t, client)))
	_ = assertPostgresDetails(t, client)
}

func TestRestoreAppMetaData(t *testing.T) {
	fixture := createSampleAppWithBackupAndDeleteIt(t, func(client *api_client.QuollixClient, app *apps_basic.AppDto) {
		app.AccessPolicy = tools.Policies.PublicAccessPolicy
		assert.Nil(t, client.Apps.SetAccessPolicy(app.AppId, tools.Policies.PublicAccessPolicy))
		assert.True(t, app.AutomaticUpdatesEnabled)
		assert.True(t, app.AutomaticBackupsEnabled)
		app.AutomaticUpdatesEnabled = false
		app.AutomaticBackupsEnabled = false
		assert.Nil(t, client.Apps.UpdateMaintenanceSettings(app.AppId, false, false))
	})
	defer fixture.client.Test.ResetTestState()

	assert.Nil(t, fixture.client.Backups.Restore(fixture.backup.BackupId))

	restoredApp := GetInstalledSample(t, fixture.client)
	assertAppState(t, fixture.originalApp, restoredApp)
}

func configureBackupRepo(t *testing.T, client *api_client.QuollixClient) {
	repo := backup_server.GetSampleRemoteRepo()

	knownHosts, err := client.Settings.GetKnownHosts(repo)
	assert.Nil(t, err)
	assert.True(t, strings.Contains(knownHosts, "["+repo.Host+"]:2222"))

	repo.SshKnownHosts = knownHosts

	assert.Nil(t, client.Settings.SaveSshConfigs(repo))
}

func prepareSshRemoteServerSetup(t *testing.T) *api_client.QuollixClient {
	client := GetClientAndLogin(t)
	configureBackupRepo(t, client)
	return client
}

func getSingleBackup(t *testing.T, client *api_client.QuollixClient, maintainer string, appName string) tools.BackupInfo {
	appBackups, err := client.Backups.ListByApp(maintainer, appName)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(appBackups))
	return appBackups[0]
}

func assertAppBackupsCount(t *testing.T, client *api_client.QuollixClient, maintainer string, appName string, expectedCount int) {
	appBackups, err := client.Backups.ListByApp(maintainer, appName)
	assert.Nil(t, err)
	assert.Equal(t, expectedCount, len(appBackups))
}

func assertAppsNumbersInBackupRepo(t *testing.T, client *api_client.QuollixClient, expectedAppNumberInBackupRepo int) {
	backupRepoApps, err := client.Backups.ListAppsInRepository()
	assert.Nil(t, err)
	assert.Equal(t, expectedAppNumberInBackupRepo, len(backupRepoApps))
}

func assertAppState(t *testing.T, expected *apps_basic.AppDto, actual *apps_basic.AppDto) {
	assert.Equal(t, expected.Maintainer, actual.Maintainer)
	assert.Equal(t, expected.AppName, actual.AppName)
	assert.Equal(t, expected.VersionName, actual.VersionName)
	assert.Equal(t, expected.VersionCreationTimestamp, actual.VersionCreationTimestamp)
	assert.Equal(t, expected.Port, actual.Port)
	assert.Equal(t, expected.AccessPolicy, actual.AccessPolicy)
	assert.Equal(t, expected.IsOfficialDatabaseApp, actual.IsOfficialDatabaseApp)
	assert.Equal(t, expected.VersionContent, actual.VersionContent)
	assert.Equal(t, expected.IsRunning, actual.IsRunning)
	assert.Equal(t, expected.ClientId, actual.ClientId)
	assert.Equal(t, expected.ClientSecret, actual.ClientSecret)
	assert.Equal(t, expected.AutomaticUpdatesEnabled, actual.AutomaticUpdatesEnabled)
	assert.Equal(t, expected.AutomaticBackupsEnabled, actual.AutomaticBackupsEnabled)
}

func assertPostgresDetails(t *testing.T, client *api_client.QuollixClient) apps_basic.AppDto {
	installedApps := ListInstalledApps(t, client)

	var postgresApp apps_basic.AppDto
	for _, app := range installedApps {
		if app.AppName == u.OfficialDatabaseAppName {
			postgresApp = app
		}
	}

	assert.True(t, postgresApp.IsRunning)
	assert.Equal(t, u.OfficialMaintainer, postgresApp.Maintainer)
	assert.Equal(t, u.OfficialDatabaseAppName, postgresApp.AppName)
	assert.Equal(t, system_config_migrations.PostgresVersion, postgresApp.VersionName)
	assert.Equal(t, tools.Policies.AdminOnlyAccessPolicy, postgresApp.AccessPolicy)

	return postgresApp
}

type SampleAppBackupFixture struct {
	client      *api_client.QuollixClient
	app         *apps_basic.AppDto
	originalApp *apps_basic.AppDto
	backup      tools.BackupInfo
}

func createSampleAppWithBackupAndDeleteIt(t *testing.T, beforeBackup func(client *api_client.QuollixClient, app *apps_basic.AppDto)) SampleAppBackupFixture {
	client := prepareSshRemoteServerSetup(t)

	app, err := InstallSample(t, client, "2.0")
	assert.Nil(t, err)
	assert.Nil(t, client.Apps.Start(app.AppId))

	if beforeBackup != nil {
		beforeBackup(client, app)
	}

	originalApp := GetInstalledSample(t, client)

	assert.Nil(t, client.Backups.Create(originalApp.AppId))
	assert.Nil(t, client.Apps.Delete(originalApp.AppId))
	assertAppBackupsCount(t, client, tools.SampleMaintainer, tools.SampleApp, 1)

	backup := getSingleBackup(t, client, tools.SampleMaintainer, tools.SampleApp)

	return SampleAppBackupFixture{
		client:      client,
		app:         app,
		originalApp: originalApp,
		backup:      backup,
	}
}

func TestWrongEncryptionPassword_PurgeBackupServerAndRecreateWithNewPassword(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	repo := backup_server.GetSampleRemoteRepo()
	knownHosts, err := client.Settings.GetKnownHosts(repo)
	assert.Nil(t, err)
	repo.SshKnownHosts = knownHosts

	assert.Nil(t, client.Settings.SaveSshConfigs(repo))

	repo.EncryptionPassword = "wrongpassword"

	err = client.Settings.SaveSshConfigs(repo)
	u.AssertDeepStackErrorFromRequest(t, err, backup_server.WrongEncryptionPasswordErr)

	assert.Nil(t, client.Backups.PurgeServer(repo))
	assert.Nil(t, client.Settings.SaveSshConfigs(repo))
}

func TestDeletionOfTwoBackupsAtOnce(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	_, err := InstallSample(t, client, "2.0")
	assert.Nil(t, err)
	configureBackupRepo(t, client)

	assert.Nil(t, CreateSampleBackup(t, client))
	assert.Nil(t, CreateSampleBackup(t, client))

	appBackups, err := client.Backups.ListByApp(tools.SampleMaintainer, tools.SampleApp)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(appBackups))

	var backupIds []string
	for _, backup := range appBackups {
		backupIds = append(backupIds, backup.BackupId)
	}

	assert.Nil(t, client.Backups.Delete(backupIds))

	appBackups, err = client.Backups.ListByApp(tools.SampleMaintainer, tools.SampleApp)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(appBackups))
}

func TestCantUpdateAppWithoutVolumes(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	sampleApp, err := InstallSample(t, client, "0.0")
	assert.Nil(t, err)
	configureBackupRepo(t, client)

	err = client.Backups.Create(sampleApp.AppId)
	u.AssertDeepStackErrorFromRequest(t, err, backup_server.CantBackupAppWithoutVolumes)
}
