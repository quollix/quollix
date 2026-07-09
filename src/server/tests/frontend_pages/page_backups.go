package frontend_pages

import (
	"server/tools"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/quollix/common/assert"
	u "github.com/quollix/common/utils"
)

type BackupsPage struct {
	Frame *FrameType
}

type BackupRow struct {
	VersionName                   string
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

func (b *BackupsPage) AssertLoadingBackupsVisible() *BackupsPage {
	loadingMessage, err := b.Frame.Page.Element("#backups-loading-message")
	assert.Nil(b.Frame.T, err)
	text, err := loadingMessage.Text()
	assert.Nil(b.Frame.T, err)
	assert.Equal(b.Frame.T, "Loading backups...", strings.TrimSpace(text))
	return b
}

func (b *BackupsPage) ClickBack() *BackedUpAppsPage {
	backButton, err := b.Frame.Page.Element("#backups-page-back-button")
	assert.Nil(b.Frame.T, err)
	b.Frame.Browser.DoAndWaitDOMContentLoaded(func() {
		backButton.MustClick()
	})
	b.Frame.Assert.PagePath(tools.Paths.FrontendBackedUpApps)
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
	err := tools.EventuallyWithTimeout(backupOperationTimeout, 50*time.Millisecond, func() error {
		apps, err := b.Frame.Client.Apps.ListInstalled()
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
	b.Frame.Assert.AppOperationStartedAndFinished().Browser.ReloadPage()
	assert.Equal(b.Frame.T, 0, len(b.ListBackups()))
	return b
}

func (b *BackupsPage) waitForSingleBackupRow() *rod.Element {
	var row *rod.Element
	err := tools.EventuallyWithTimeout(backupOperationTimeout, 50*time.Millisecond, func() error {
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
	err := tools.Eventually(func() error {
		isLoading, _, err := b.Frame.Page.Has("#backups-loading-message")
		if err != nil {
			return err
		}
		if isLoading {
			return u.Logger.NewError("backups page is still loading")
		}
		return nil
	})
	assert.Nil(b.Frame.T, err)
}
