//go:build acceptance

package acceptance

import (
	"server/backup_server"
	"server/tests/component"
	"server/tests/frontend_pages"
	"server/tools"
	"strings"
	"testing"
	"time"

	"github.com/quollix/common/assert"
)

func TestSampleAppBackupViaInstalledAppsPage(t *testing.T) {
	frame := frontend_pages.Setup(t)
	defer frame.Client.Test.ResetTestState()

	_, err := component.InstallSample(t, frame.Client, "2.0")
	assert.Nil(t, err)

	installedAppsPage := frame.Pages.OpenInstalledAppsPage()
	installedAppsPage.AssertNoOngoingAppOperation()
	installedAppsPage.AssertOperationOptionNotPresent(tools.SampleApp, "Backup")

	repo := backup_server.GetSampleRemoteRepo()
	repo.SshKnownHosts, err = frame.Client.Settings.GetKnownHosts(repo)
	assert.Nil(t, err)
	assert.Nil(t, frame.Client.Settings.SaveSshConfigs(repo))

	sampleBackupsBefore, err := frame.Client.Backups.ListByApp(tools.SampleMaintainer, tools.SampleApp)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(sampleBackupsBefore))

	frame.Browser.ReloadPage()
	installedAppsPage.AssertOperationOptionPresent(tools.SampleApp, "Backup")
	installedAppsPage.BackupAppViaOperations(tools.SampleApp)
	frame.Assert.AppOperationStartedAndFinished().Browser.ReloadPage()
	installedAppsPage.AssertNoOngoingAppOperation()

	sampleBackupsAfter, err := frame.Client.Backups.ListByApp(tools.SampleMaintainer, tools.SampleApp)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(sampleBackupsAfter))
}

func TestBackupsPageListRestoreAndDeleteFlow(t *testing.T) {
	frame := frontend_pages.Setup(t)
	defer frame.Client.Test.ResetTestState()

	backedUpAppsPage := frame.Pages.OpenBackedUpAppsPage()
	backedUpAppsPage.AssertBackupsDisabledMessageVisible()

	repo := backup_server.GetSampleRemoteRepo()
	var err error
	repo.SshKnownHosts, err = frame.Client.Settings.GetKnownHosts(repo)
	assert.Nil(t, err)
	assert.Nil(t, frame.Client.Settings.SaveSshConfigs(repo))

	frame.Browser.ReloadPage()
	backedUpAppsPage = frame.Pages.BackedUpAppsPage
	backedUpAppsPage.AssertLoadingBackedUpAppsVisible()
	backedUpAppsPage.AssertNoBackedUpAppsMessageVisible()

	_, err = component.InstallSample(t, frame.Client, "2.0")
	assert.Nil(t, err)
	sampleApp := component.GetInstalledSample(t, frame.Client)
	assert.Nil(t, frame.Client.Backups.Create(sampleApp.AppId))

	frame.Browser.ReloadPage()
	backedUpAppsPage = frame.Pages.BackedUpAppsPage
	backedUpAppsPage.AssertLoadingBackedUpAppsVisible()
	sampleBackedUpApp := backedUpAppsPage.GetRequiredBackedUpApp(tools.SampleMaintainer, tools.SampleApp)
	assert.Equal(t, tools.SampleMaintainer, sampleBackedUpApp.Maintainer)
	assert.Equal(t, tools.SampleApp, sampleBackedUpApp.AppName)

	backupsPage := backedUpAppsPage.OpenListBackupsPage(tools.SampleMaintainer, tools.SampleApp)
	backupsPage.AssertMaintainerAndApp(tools.SampleMaintainer, tools.SampleApp)
	backupsPage.ClickBack()

	backupsPage = frame.Pages.BackedUpAppsPage.OpenListBackupsPage(tools.SampleMaintainer, tools.SampleApp)
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
	frame.Browser.ReloadPage()

	backupsPage.ClickRestoreFirstBackup()
	backupsPage.Frame.Assert.AppOperationStartedAndFinished()

	restoredApp := component.GetInstalledSample(t, frame.Client)
	assert.Equal(t, tools.SampleApp, restoredApp.AppName)

	frame.Browser.ReloadPage()
	backupsPage.ClickDeleteFirstBackup()
	backupsPage.Frame.Assert.AppOperationStartedAndFinished()
	sampleBackupsAfterDelete, err := frame.Client.Backups.ListByApp(tools.SampleMaintainer, tools.SampleApp)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(sampleBackupsAfterDelete))
}
