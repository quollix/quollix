package frontend_pages

import (
	"strings"
	"time"

	"github.com/quollix/common/assert"
	"github.com/quollix/common/browsertest"
	"github.com/quollix/common/quollix/api"
	u "github.com/quollix/common/utils"
)

type BackupsPage struct {
	Frame *FrameType
}

func (b *BackupsPage) Reload() {
	info, err := b.Frame.Page.Info()
	assert.Nil(b.Frame.T, err)
	navigateBackupPage(b.Frame, info.URL, "#backups-container table, #backups-empty-message", func() error {
		return b.Frame.Page.Navigate(info.URL)
	})
}

type BackupRow struct {
	VersionName                   string
	VersionCreationDate           string
	Description                   string
	BackupCreationDate            string
	CreatedWithApplicationVersion string
}

func (b *BackupsPage) AssertMaintainerAndApp(maintainer, appName string) *BackupsPage {
	maintainerElement, err := b.Frame.Page.Element("#backups-page-maintainer")
	assert.Nil(b.Frame.T, err)
	maintainerText, err := maintainerElement.Text()
	assert.Nil(b.Frame.T, err)
	assert.Equal(b.Frame.T, maintainer, strings.TrimSpace(maintainerText))

	appNameElement, err := b.Frame.Page.Element("#backups-page-app-name")
	assert.Nil(b.Frame.T, err)
	appNameText, err := appNameElement.Text()
	assert.Nil(b.Frame.T, err)
	assert.Equal(b.Frame.T, appName, strings.TrimSpace(appNameText))
	return b
}

func (b *BackupsPage) ClickBack() *BackedUpAppsPage {
	backButton, err := b.Frame.Page.Element("#backups-page-back-button")
	assert.Nil(b.Frame.T, err)
	navigateBackupPage(b.Frame, b.Frame.BaseUrl+api.Paths.FrontendBackedUpApps, "#backed-up-apps-container table, #backed-up-apps-empty-message, #backups-disabled-message", backButton.Click)
	return b.Frame.Pages.BackedUpAppsPage
}

func (b *BackupsPage) ListBackups() []BackupRow {
	b.waitUntilBackupsLoaded()

	rows, err := b.Frame.Page.Elements("tr.backup-row")
	assert.Nil(b.Frame.T, err)

	out := make([]BackupRow, 0, len(rows))
	for _, row := range rows {
		versionCell, err := row.Element(".backup-version-name-cell")
		assert.Nil(b.Frame.T, err)
		version, err := versionCell.Text()
		assert.Nil(b.Frame.T, err)

		versionCreationDateCell, err := row.Element(".backup-version-creation-date-cell")
		assert.Nil(b.Frame.T, err)
		versionCreationDate, err := versionCreationDateCell.Text()
		assert.Nil(b.Frame.T, err)

		descriptionCell, err := row.Element(".backup-description-cell")
		assert.Nil(b.Frame.T, err)
		description, err := descriptionCell.Text()
		assert.Nil(b.Frame.T, err)

		creationDateCell, err := row.Element(".backup-creation-date-cell")
		assert.Nil(b.Frame.T, err)
		creationDate, err := creationDateCell.Text()
		assert.Nil(b.Frame.T, err)

		createdWithVersionCell, err := row.Element(".backup-created-with-app-version-cell")
		assert.Nil(b.Frame.T, err)
		createdWithVersion, err := createdWithVersionCell.Text()
		assert.Nil(b.Frame.T, err)

		restoreButton, err := row.Element("button.backup-restore-button")
		assert.Nil(b.Frame.T, err)
		assert.NotNil(b.Frame.T, restoreButton)

		deleteButton, err := row.Element("button.backup-delete-button")
		assert.Nil(b.Frame.T, err)
		assert.NotNil(b.Frame.T, deleteButton)

		out = append(out, BackupRow{
			VersionName:                   strings.TrimSpace(version),
			VersionCreationDate:           strings.TrimSpace(versionCreationDate),
			Description:                   strings.TrimSpace(description),
			BackupCreationDate:            strings.TrimSpace(creationDate),
			CreatedWithApplicationVersion: strings.TrimSpace(createdWithVersion),
		})
	}
	return out
}

func (b *BackupsPage) ClickRestoreFirstBackup() *BackupsPage {
	row := b.waitForSingleBackupRow()
	restoreButton, err := row.Element("button.backup-restore-button")
	assert.Nil(b.Frame.T, err)
	restoreButton.MustClick()
	b.Frame.Browser.ConfirmDialog()
	return b
}

func (b *BackupsPage) WaitUntilAppAbsent(appName string) *BackupsPage {
	err := u.EventuallyWithTimeout(backupOperationTimeout, 50*time.Millisecond, func() error {
		apps, err := b.Frame.Client.Apps.ListInstalledForAdmin()
		assert.Nil(b.Frame.T, err)
		for _, app := range apps {
			if app.AppName == appName {
				return u.Logger.NewError("app still present", "app_name", appName)
			}
		}
		return nil
	})
	assert.Nil(b.Frame.T, err)
	return b
}

func (b *BackupsPage) ClickDeleteFirstBackup() *BackupsPage {
	row := b.waitForSingleBackupRow()
	deleteButton, err := row.Element("button.backup-delete-button")
	assert.Nil(b.Frame.T, err)
	deleteButton.MustClick()
	b.Frame.Browser.ConfirmDialog()
	return b
}

func (b *BackupsPage) DeleteFirstBackupUntilRemoved() *BackupsPage {
	b.ClickDeleteFirstBackup()
	b.Frame.Assert.AppOperationStartedAndFinished()
	b.Reload()
	assert.Equal(b.Frame.T, 0, len(b.ListBackups()))
	return b
}

func (b *BackupsPage) waitForSingleBackupRow() *browsertest.Element {
	var row *browsertest.Element
	err := u.EventuallyWithTimeout(backupOperationTimeout, 50*time.Millisecond, func() error {
		rows, listErr := b.Frame.Page.Elements("tr.backup-row")
		if listErr != nil {
			return listErr
		}
		if len(rows) != 1 {
			return u.Logger.NewError("expected exactly one backup row", "count", len(rows))
		}
		row = rows[0]
		return nil
	})
	assert.Nil(b.Frame.T, err)
	return row
}

func (b *BackupsPage) waitUntilBackupsLoaded() {
	err := u.EventuallyWithTimeout(backupOperationTimeout, 50*time.Millisecond, func() error {
		loaded, _, err := b.Frame.Page.Has("#backups-container table, #backups-empty-message")
		if err != nil {
			return err
		}
		if !loaded {
			return u.Logger.NewError("backups page is still loading")
		}
		return nil
	})
	assert.Nil(b.Frame.T, err)
}
