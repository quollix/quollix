package frontend_pages

import (
	"fmt"
	"strings"

	"github.com/quollix/common/assert"
	"github.com/quollix/common/browsertest"
	"github.com/quollix/common/quollix/api"
	u "github.com/quollix/common/utils"
)

type AppsWithSecretsPage struct {
	Frame *FrameType
}

type AppSecretPage struct {
	Frame *FrameType
}

type AppSecretAppEntry struct {
	Maintainer string
	AppName    string
	Version    string
}

type AppSecretEntry struct {
	Name               string
	ValueDisplayedInUi string
}

func (a *AppsWithSecretsPage) ListApps() []AppSecretAppEntry {
	rows, err := a.Frame.Page.Elements("tr.app-with-secrets-row")
	assert.Nil(a.Frame.T, err)

	apps := make([]AppSecretAppEntry, 0, len(rows))
	for _, row := range rows {
		apps = append(apps, a.readAppEntry(row))
	}
	return apps
}

func (a *AppsWithSecretsPage) AssertAppCount(expected int) *AppsWithSecretsPage {
	assert.Equal(a.Frame.T, expected, len(a.ListApps()))
	return a
}

func (a *AppsWithSecretsPage) AssertAppPresent(maintainer, appName, version string) *AppsWithSecretsPage {
	app := a.GetRequiredApp(appName)
	assert.Equal(a.Frame.T, maintainer, app.Maintainer)
	assert.Equal(a.Frame.T, appName, app.AppName)
	assert.Equal(a.Frame.T, version, app.Version)
	return a
}

func (a *AppsWithSecretsPage) GetRequiredApp(appName string) *AppSecretAppEntry {
	for _, app := range a.ListApps() {
		if app.AppName == appName {
			appCopy := app
			return &appCopy
		}
	}
	assert.Nil(a.Frame.T, u.Logger.NewError("app secrets row not found", "app_name", appName))
	return nil
}

func (a *AppsWithSecretsPage) OpenApp(appName string) *AppSecretPage {
	row := a.findRowByAppName(appName)
	showButton, err := row.Element(".app-secret-show-button")
	assert.Nil(a.Frame.T, err)
	a.Frame.Browser.DoAndWaitDOMContentLoaded(func() {
		showButton.MustClick()
	})
	a.Frame.Assert.PagePath(api.Paths.FrontendAppSecret)
	return a.Frame.Pages.AppSecretPage
}

func (a *AppsWithSecretsPage) readAppEntry(row *browsertest.Element) AppSecretAppEntry {
	maintainerCell, err := row.Element(".app-secret-maintainer-cell")
	assert.Nil(a.Frame.T, err)
	appNameCell, err := row.Element(".app-secret-app-name-cell")
	assert.Nil(a.Frame.T, err)
	versionCell, err := row.Element(".app-secret-version-cell")
	assert.Nil(a.Frame.T, err)

	return AppSecretAppEntry{
		Maintainer: strings.TrimSpace(maintainerCell.MustText()),
		AppName:    strings.TrimSpace(appNameCell.MustText()),
		Version:    strings.TrimSpace(versionCell.MustText()),
	}
}

func (a *AppsWithSecretsPage) findRowByAppName(appName string) *browsertest.Element {
	var foundRow *browsertest.Element
	err := u.Eventually(func() error {
		rows, err := a.Frame.Page.Elements("tr.app-with-secrets-row")
		assert.Nil(a.Frame.T, err)
		for _, row := range rows {
			appNameAttr, err := row.Attribute("data-app-name")
			assert.Nil(a.Frame.T, err)
			if appNameAttr != nil && strings.TrimSpace(*appNameAttr) == appName {
				foundRow = row
				return nil
			}
		}
		return fmt.Errorf("app secrets row not found: %s", appName)
	})
	assert.Nil(a.Frame.T, err)
	return foundRow
}

func (a *AppSecretPage) AssertMetadata(maintainer, appName, version string) *AppSecretPage {
	assert.Equal(a.Frame.T, "Maintainer: "+maintainer, strings.TrimSpace(a.Frame.Controls.GetRequiredElementEventually("#app-secret-maintainer").MustText()))
	assert.Equal(a.Frame.T, "App Name: "+appName, strings.TrimSpace(a.Frame.Controls.GetRequiredElementEventually("#app-secret-app-name").MustText()))
	assert.Equal(a.Frame.T, "Version: "+version, strings.TrimSpace(a.Frame.Controls.GetRequiredElementEventually("#app-secret-version").MustText()))
	return a
}

func (a *AppSecretPage) ListSecrets() []AppSecretEntry {
	rows, err := a.Frame.Page.Elements("tr.app-secret-row")
	assert.Nil(a.Frame.T, err)

	secrets := make([]AppSecretEntry, 0, len(rows))
	for _, row := range rows {
		secrets = append(secrets, a.readSecretEntry(row))
	}
	return secrets
}

func (a *AppSecretPage) AssertSecretCount(expected int) *AppSecretPage {
	assert.Equal(a.Frame.T, expected, len(a.ListSecrets()))
	return a
}

func (a *AppSecretPage) AssertSecretMasked(name string) *AppSecretPage {
	actual := a.GetRequiredSecret(name).ValueDisplayedInUi
	assert.Equal(a.Frame.T, "****************", actual)
	return a
}

func (a *AppSecretPage) GetRequiredSecret(name string) *AppSecretEntry {
	for _, secret := range a.ListSecrets() {
		if secret.Name == name {
			secretCopy := secret
			return &secretCopy
		}
	}
	assert.Nil(a.Frame.T, u.Logger.NewError("app secret row not found", "secret_name", name))
	return nil
}

func (a *AppSecretPage) RegenerateSecret(name string) *AppSecretPage {
	row := a.findRowBySecretName(name)
	regenerateButton, err := row.Element(".app-secret-regenerate-button")
	assert.Nil(a.Frame.T, err)
	regenerateButton.MustClick()
	a.Frame.Browser.ConfirmDialog()
	a.Frame.Assert.SnackbarVisibleWithTextEventually("Secret regenerated successfully.")
	return a
}

func (a *AppSecretPage) readSecretEntry(row *browsertest.Element) AppSecretEntry {
	nameCell, err := row.Element(".app-secret-name-cell")
	assert.Nil(a.Frame.T, err)
	valueCell, err := row.Element(".app-secret-value-cell span.mono")
	assert.Nil(a.Frame.T, err)

	return AppSecretEntry{
		Name:               strings.TrimSpace(nameCell.MustText()),
		ValueDisplayedInUi: strings.TrimSpace(valueCell.MustText()),
	}
}

func (a *AppSecretPage) findRowBySecretName(name string) *browsertest.Element {
	var foundRow *browsertest.Element
	err := u.Eventually(func() error {
		rows, err := a.Frame.Page.Elements("tr.app-secret-row")
		assert.Nil(a.Frame.T, err)
		for _, row := range rows {
			secretNameAttr, err := row.Attribute("data-secret-name")
			assert.Nil(a.Frame.T, err)
			if secretNameAttr != nil && strings.TrimSpace(*secretNameAttr) == name {
				foundRow = row
				return nil
			}
		}
		return fmt.Errorf("app secret row not found: %s", name)
	})
	assert.Nil(a.Frame.T, err)
	return foundRow
}
