//go:build acceptance

package pages

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
	rows := o.Frame.page.MustElements("tr.oidc-client-row")
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
	assert.True(o.Frame.t, false)
	return nil
}

func (o *OidcClientsPage) RegenerateCredentials(appName string) *OidcClientsPage {
	row := o.findRowByAppName(appName)
	regenButton, err := row.Element("button.regen-btn")
	assert.Nil(o.Frame.t, err)
	regenButton.MustClick()
	o.Frame.ConfirmDialog()
	o.Frame.AssertSnackbarVisibleWithTextEventually("Credentials regenerated successfully.")
	return o
}

func (o *OidcClientsPage) readAppEntry(row *rod.Element) *OidcClientEntry {
	maintainerCell, err := row.Element(".oidc-maintainer-cell")
	assert.Nil(o.Frame.t, err)
	maintainer, err := maintainerCell.Text()
	assert.Nil(o.Frame.t, err)

	appNameCell, err := row.Element(".oidc-app-name-cell")
	assert.Nil(o.Frame.t, err)
	appName, err := appNameCell.Text()
	assert.Nil(o.Frame.t, err)

	clientIDDisplayedInUi, err := row.Element(".client-id-cell span.mono")
	assert.Nil(o.Frame.t, err)
	clientIDDisplayedInUiText, err := clientIDDisplayedInUi.Text()
	assert.Nil(o.Frame.t, err)

	clientSecretDisplayedInUi, err := row.Element(".client-secret-cell span.mono")
	assert.Nil(o.Frame.t, err)
	clientSecretDisplayedInUiText, err := clientSecretDisplayedInUi.Text()
	assert.Nil(o.Frame.t, err)

	return &OidcClientEntry{
		Maintainer:                strings.TrimSpace(maintainer),
		AppName:                   strings.TrimSpace(appName),
		ClientIDDisplayedInUi:     strings.TrimSpace(clientIDDisplayedInUiText),
		ClientSecretDisplayedInUi: strings.TrimSpace(clientSecretDisplayedInUiText),
	}
}

func (o *OidcClientsPage) findRowByAppName(appName string) *rod.Element {
	rows, err := o.Frame.page.Elements("tr.oidc-client-row")
	assert.Nil(o.Frame.t, err)
	for _, row := range rows {
		appNameAttr, err := row.Attribute("data-app-name")
		if err == nil && appNameAttr != nil && strings.TrimSpace(*appNameAttr) == appName {
			return row
		}
	}
	assert.Nil(o.Frame.t, u.Logger.NewError("oidc row not found", "app_name", appName))
	return nil
}
