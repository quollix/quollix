//go:build acceptance

package acceptance

import (
	"server/backup_server"
	"server/tests/acceptance/pages"
	"server/tools"
	"strings"
	"testing"
	"time"

	"github.com/quollix/common/assert"
)

func TestSampleAppBackupViaInstalledAppsPage(t *testing.T) {
	frame := pages.Setup(t)
	defer frame.Client.Test.ResetTestState()

	_, err := frame.Client.Apps.InstallSample("2.0")
	assert.Nil(t, err)

	installedAppsPage := frame.OpenInstalledAppsPage()
	installedAppsPage.AssertNoOngoingAppOperation()
	installedAppsPage.AssertOperationOptionNotPresent(tools.SampleApp, "Backup")

	repo := backup_server.GetSampleRemoteRepo()
	repo.SshKnownHosts = frame.Client.Settings.GetKnownHosts(repo)
	assert.Nil(t, frame.Client.Settings.SaveSshConfigs(repo))

	sampleBackupsBefore, err := frame.Client.Backups.ListByApp(tools.SampleMaintainer, tools.SampleApp)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(sampleBackupsBefore))

	frame.ReloadPage()
	installedAppsPage.AssertOperationOptionPresent(tools.SampleApp, "Backup")
	installedAppsPage.BackupAppViaOperations(tools.SampleApp)
	frame.AssertAppOperationStartedAndFinished().ReloadPage()
	installedAppsPage.AssertNoOngoingAppOperation()

	sampleBackupsAfter, err := frame.Client.Backups.ListByApp(tools.SampleMaintainer, tools.SampleApp)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(sampleBackupsAfter))
}

func TestBackupsPageListRestoreAndDeleteFlow(t *testing.T) {
	frame := pages.Setup(t)
	defer frame.Client.Test.ResetTestState()

	backedUpAppsPage := frame.OpenBackedUpAppsPage()
	backedUpAppsPage.AssertBackupsDisabledMessageVisible()

	repo := backup_server.GetSampleRemoteRepo()
	repo.SshKnownHosts = frame.Client.Settings.GetKnownHosts(repo)
	assert.Nil(t, frame.Client.Settings.SaveSshConfigs(repo))

	frame.ReloadPage()
	backedUpAppsPage = frame.BackedUpAppsPage
	backedUpAppsPage.AssertLoadingBackedUpAppsVisible()
	backedUpAppsPage.AssertNoBackedUpAppsMessageVisible()

	_, err := frame.Client.Apps.InstallSample("2.0")
	assert.Nil(t, err)
	sampleApp := frame.Client.Apps.GetInstalledSample()
	assert.Nil(t, frame.Client.Backups.Create(sampleApp.AppId))

	frame.ReloadPage()
	backedUpAppsPage = frame.BackedUpAppsPage
	backedUpAppsPage.AssertLoadingBackedUpAppsVisible()
	sampleBackedUpApp := backedUpAppsPage.GetRequiredBackedUpApp(tools.SampleMaintainer, tools.SampleApp)
	assert.Equal(t, tools.SampleMaintainer, sampleBackedUpApp.Maintainer)
	assert.Equal(t, tools.SampleApp, sampleBackedUpApp.AppName)

	backupsPage := backedUpAppsPage.OpenListBackupsPage(tools.SampleMaintainer, tools.SampleApp)
	backupsPage.AssertMaintainerAndApp(tools.SampleMaintainer, tools.SampleApp)
	backupsPage.ClickBack()

	backupsPage = frame.BackedUpAppsPage.OpenListBackupsPage(tools.SampleMaintainer, tools.SampleApp)
	backupRows := backupsPage.ListBackups()
	assert.Equal(t, 1, len(backupRows))

	backupRow := backupRows[0]
	assert.True(t, strings.HasPrefix(backupRow.VersionName, "2.0"))
	assert.NotEqual(t, "", backupRow.BackupCreationDate)
	_, err = time.Parse(tools.PrettyFrontendTimeLayoutWithDay, backupRow.BackupCreationDate)
	assert.Nil(t, err)
	assert.Equal(t, tools.ApplicationVersion, backupRow.CreatedWithApplicationVersion)

	assert.Nil(t, frame.Client.Apps.Delete(sampleApp.AppId))
	backupsPage.WaitUntilAppAbsent(tools.SampleApp)
	frame.ReloadPage()

	backupsPage.ClickRestoreFirstBackup()
	backupsPage.Frame.AssertAppOperationStartedAndFinished()

	restoredApp := frame.Client.Apps.GetInstalledSample()
	assert.Equal(t, tools.SampleApp, restoredApp.AppName)

	frame.ReloadPage()
	backupsPage.ClickDeleteFirstBackup()
	backupsPage.Frame.AssertAppOperationStartedAndFinished()
	sampleBackupsAfterDelete, err := frame.Client.Backups.ListByApp(tools.SampleMaintainer, tools.SampleApp)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(sampleBackupsAfterDelete))
}
