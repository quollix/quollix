package frontend_pages

import (
	"strings"

	"github.com/go-rod/rod"
	"github.com/quollix/common/assert"
	u "github.com/quollix/common/utils"
)

type OidcClientsPage struct {
	Frame *FrameType
}

type OidcClientEntry struct {
	Maintainer                string
	AppName                   string
	ClientIDDisplayedInUi     string
	ClientSecretDisplayedInUi string
}

func (o *OidcClientsPage) ListApps() []OidcClientEntry {
	rows := o.Frame.Page.MustElements("tr.oidc-client-row")
	out := make([]OidcClientEntry, 0, len(rows))

	for _, row := range rows {
		entry := o.readAppEntry(row)
		out = append(out, *entry)
	}
	return out
}

func (o *OidcClientsPage) GetRequiredApp(appName string) *OidcClientEntry {
	apps := o.ListApps()
	for _, app := range apps {
		if app.AppName == appName {
			appCopy := app
			return &appCopy
		}
	}
	assert.True(o.Frame.T, false)
	return nil
}

func (o *OidcClientsPage) RegenerateCredentials(appName string) *OidcClientsPage {
	row := o.findRowByAppName(appName)
	regenButton, err := row.Element("button.regen-btn")
	assert.Nil(o.Frame.T, err)
	regenButton.MustClick()
	o.Frame.Browser.ConfirmDialog()
	o.Frame.Assert.SnackbarVisibleWithTextEventually("Credentials regenerated successfully.")
	return o
}

func (o *OidcClientsPage) readAppEntry(row *rod.Element) *OidcClientEntry {
	maintainerCell, err := row.Element(".oidc-maintainer-cell")
	assert.Nil(o.Frame.T, err)
	maintainer, err := maintainerCell.Text()
	assert.Nil(o.Frame.T, err)

	appNameCell, err := row.Element(".oidc-app-name-cell")
	assert.Nil(o.Frame.T, err)
	appName, err := appNameCell.Text()
	assert.Nil(o.Frame.T, err)

	clientIDDisplayedInUi, err := row.Element(".client-id-cell span.mono")
	assert.Nil(o.Frame.T, err)
	clientIDDisplayedInUiText, err := clientIDDisplayedInUi.Text()
	assert.Nil(o.Frame.T, err)

	clientSecretDisplayedInUi, err := row.Element(".client-secret-cell span.mono")
	assert.Nil(o.Frame.T, err)
	clientSecretDisplayedInUiText, err := clientSecretDisplayedInUi.Text()
	assert.Nil(o.Frame.T, err)

	return &OidcClientEntry{
		Maintainer:                strings.TrimSpace(maintainer),
		AppName:                   strings.TrimSpace(appName),
		ClientIDDisplayedInUi:     strings.TrimSpace(clientIDDisplayedInUiText),
		ClientSecretDisplayedInUi: strings.TrimSpace(clientSecretDisplayedInUiText),
	}
}

func (o *OidcClientsPage) findRowByAppName(appName string) *rod.Element {
	rows, err := o.Frame.Page.Elements("tr.oidc-client-row")
	assert.Nil(o.Frame.T, err)
	for _, row := range rows {
		appNameAttr, err := row.Attribute("data-app-name")
		if err == nil && appNameAttr != nil && strings.TrimSpace(*appNameAttr) == appName {
			return row
		}
	}
	assert.Nil(o.Frame.T, u.Logger.NewError("oidc row not found", "app_name", appName))
	return nil
}
