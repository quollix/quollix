package frontend_pages

import (
	"server/tools"
	"strings"

	"github.com/go-rod/rod"
	"github.com/quollix/common/assert"
	u "github.com/quollix/common/utils"
)

type BackedUpAppsPage struct {
	Frame *FrameType
}

type BackedUpAppEntry struct {
	Maintainer string
	AppName    string
}

func (b *BackedUpAppsPage) AssertNoBackedUpAppsMessageVisible() *BackedUpAppsPage {
	b.waitUntilBackedUpAppsLoaded()
	err := b.assertParagraphTextVisible("No apps with backups were found yet.")
	assert.Nil(b.Frame.T, err)
	return b
}

func (b *BackedUpAppsPage) AssertBackupsDisabledMessageVisible() *BackedUpAppsPage {
	b.waitUntilBackedUpAppsLoaded()
	err := b.assertParagraphTextVisible("Backups are currently disabled. You can enable them in the settings page.")
	assert.Nil(b.Frame.T, err)
	return b
}

func (b *BackedUpAppsPage) AssertLoadingBackedUpAppsVisible() *BackedUpAppsPage {
	loadingMessage, err := b.Frame.Page.Element("#backed-up-apps-loading-message")
	assert.Nil(b.Frame.T, err)
	text, err := loadingMessage.Text()
	assert.Nil(b.Frame.T, err)
	assert.Equal(b.Frame.T, "Loading backed up apps...", strings.TrimSpace(text))
	return b
}

func (b *BackedUpAppsPage) assertParagraphTextVisible(expectedText string) error {
	return tools.Eventually(func() error {
		paragraphs, err := b.Frame.Page.Elements("p")
		if err != nil {
			return err
		}
		for _, paragraph := range paragraphs {
			text, textErr := paragraph.Text()
			if textErr != nil {
				return textErr
			}
			if strings.TrimSpace(text) == expectedText {
				return nil
			}
		}
		return u.Logger.NewError("paragraph text not found", "expected_text", expectedText)
	})
}

func (b *BackedUpAppsPage) ListBackedUpApps() []BackedUpAppEntry {
	b.waitUntilBackedUpAppsLoaded()

	rows, err := b.Frame.Page.Elements("tr.backed-up-app-row")
	assert.Nil(b.Frame.T, err)

	out := make([]BackedUpAppEntry, 0, len(rows))
	for _, row := range rows {
		maintainerCell, err := row.Element(".backed-up-app-maintainer-cell")
		assert.Nil(b.Frame.T, err)
		maintainer, err := maintainerCell.Text()
		assert.Nil(b.Frame.T, err)

		appNameCell, err := row.Element(".backed-up-app-name-cell")
		assert.Nil(b.Frame.T, err)
		appName, err := appNameCell.Text()
		assert.Nil(b.Frame.T, err)

		listBackupsButton, err := row.Element("button.backed-up-app-list-backups-button")
		assert.Nil(b.Frame.T, err)
		assert.NotNil(b.Frame.T, listBackupsButton)

		out = append(out, BackedUpAppEntry{
			Maintainer: strings.TrimSpace(maintainer),
			AppName:    strings.TrimSpace(appName),
		})
	}
	return out
}

func (b *BackedUpAppsPage) GetRequiredBackedUpApp(maintainer, appName string) *BackedUpAppEntry {
	var found *BackedUpAppEntry
	err := tools.Eventually(func() error {
		apps := b.ListBackedUpApps()
		for _, app := range apps {
			if app.Maintainer == maintainer && app.AppName == appName {
				appCopy := app
				found = &appCopy
				return nil
			}
		}
		return u.Logger.NewError("backed up app not found", "maintainer", maintainer, "app_name", appName)
	})
	assert.Nil(b.Frame.T, err)
	return found
}

func (b *BackedUpAppsPage) waitUntilBackedUpAppsLoaded() {
	err := tools.Eventually(func() error {
		isLoading, _, err := b.Frame.Page.Has("#backed-up-apps-loading-message")
		if err != nil {
			return err
		}
		if isLoading {
			return u.Logger.NewError("backed up apps page is still loading")
		}
		return nil
	})
	assert.Nil(b.Frame.T, err)
}

func (b *BackedUpAppsPage) OpenListBackupsPage(maintainer, appName string) *BackupsPage {
	row := b.getRequiredBackedUpAppRow(maintainer, appName)
	listBackupsButton, err := row.Element("button.backed-up-app-list-backups-button")
	assert.Nil(b.Frame.T, err)
	b.Frame.Browser.DoAndWaitDOMContentLoaded(func() {
		listBackupsButton.MustClick()
	})
	b.Frame.Assert.PagePath(tools.Paths.FrontendListBackups)
	b.Frame.Pages.BackupsPage.AssertLoadingBackupsVisible()
	return b.Frame.Pages.BackupsPage
}

func (b *BackedUpAppsPage) getRequiredBackedUpAppRow(maintainer, appName string) *rod.Element {
	var found *rod.Element
	err := tools.Eventually(func() error {
		row, err := b.Frame.Page.Element("tr.backed-up-app-row")
		if err != nil {
			return err
		}
		found = row
		return nil
	})
	assert.Nil(b.Frame.T, err)

	rowMaintainer, err := found.Attribute("data-maintainer")
	assert.Nil(b.Frame.T, err)
	assert.NotNil(b.Frame.T, rowMaintainer)
	assert.Equal(b.Frame.T, maintainer, strings.TrimSpace(*rowMaintainer))

	rowAppName, err := found.Attribute("data-app-name")
	assert.Nil(b.Frame.T, err)
	assert.NotNil(b.Frame.T, rowAppName)
	assert.Equal(b.Frame.T, appName, strings.TrimSpace(*rowAppName))

	return found
}
