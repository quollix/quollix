//go:build acceptance

package acceptance

import (
	"server/tests/acceptance/pages"
	"server/tools"
	"testing"

	"github.com/quollix/common/assert"
	u "github.com/quollix/common/utils"
)

const (
	oidcClientIDLength     = 16
	oidcClientSecretLength = 64
	oidcSecretMask         = "****************"
)

func TestOidcClientsPage(t *testing.T) {
	frame := pages.Setup(t)
	defer frame.Client.Test.ResetTestState()

	_, err := frame.Client.Apps.InstallSample("2.0")
	assert.Nil(t, err)

	page := frame.OpenOidcClientsPage()

	apps := page.ListApps()
	assert.Equal(t, 1, len(apps))

	sampleApp := page.GetRequiredApp("sampleapp")
	assert.Equal(t, "samplemaintainer", sampleApp.Maintainer)
	assert.Equal(t, "sampleapp", sampleApp.AppName)
	assert.Equal(t, oidcClientIDLength, len(sampleApp.ClientIDDisplayedInUi))
	assert.Equal(t, oidcSecretMask, sampleApp.ClientSecretDisplayedInUi)

	oldSampleAppFromBackend := frame.Client.Apps.GetInstalledSample()
	oldClientID := oldSampleAppFromBackend.ClientId
	oldClientSecret := oldSampleAppFromBackend.ClientSecret
	assert.Equal(t, oidcClientIDLength, len(oldClientID))
	assert.Equal(t, oidcClientSecretLength, len(oldClientSecret))
	assert.Equal(t, oldClientID, sampleApp.ClientIDDisplayedInUi)

	page.RegenerateCredentials("sampleapp")

	err = tools.Eventually(func() error {
		newSampleAppFromBackend := frame.Client.Apps.GetInstalledSample()
		if oldClientID == newSampleAppFromBackend.ClientId {
			return u.Logger.NewError("client id was not regenerated")
		}
		if oldClientSecret == newSampleAppFromBackend.ClientSecret {
			return u.Logger.NewError("client secret was not regenerated")
		}
		return nil
	})
	assert.Nil(t, err)

	sampleApp = page.GetRequiredApp("sampleapp")
	assert.Equal(t, "samplemaintainer", sampleApp.Maintainer)
	assert.Equal(t, "sampleapp", sampleApp.AppName)
	assert.Equal(t, oidcClientIDLength, len(sampleApp.ClientIDDisplayedInUi))
	assert.Equal(t, oidcSecretMask, sampleApp.ClientSecretDisplayedInUi)

	newSampleAppFromBackend := frame.Client.Apps.GetInstalledSample()
	newClientID := newSampleAppFromBackend.ClientId
	newClientSecret := newSampleAppFromBackend.ClientSecret
	assert.Equal(t, oidcClientIDLength, len(newClientID))
	assert.Equal(t, oidcClientSecretLength, len(newClientSecret))
	assert.Equal(t, newClientID, sampleApp.ClientIDDisplayedInUi)

	assert.NotEqual(t, oldClientID, newClientID)
	assert.NotEqual(t, oldClientSecret, newClientSecret)
}
