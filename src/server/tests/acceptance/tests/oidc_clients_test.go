//go:build acceptance

package acceptance

import (
	"testing"

	"server/tests/component"
	"server/tests/frontend_pages"
	test_tools "server/tests/test_tools"
	"server/tools"

	"github.com/quollix/common/assert"
	u "github.com/quollix/common/utils"
)

const (
	oidcSecretMask = "****************"
)

func TestOidcClientsPage(t *testing.T) {
	frame := frontend_pages.Setup(t)
	defer frame.Client.Test.ResetTestState()

	_, err := component.InstallSample(t, frame.Client, "2.0")
	assert.Nil(t, err)

	page := frame.Pages.OpenOidcClientsPage()

	apps := page.ListApps()
	assert.Equal(t, 1, len(apps))

	sampleApp := page.GetRequiredApp("sampleapp")
	assert.Equal(t, "samplemaintainer", sampleApp.Maintainer)
	assert.Equal(t, "sampleapp", sampleApp.AppName)
	assert.Equal(t, test_tools.OidcClientIDLength, len(sampleApp.ClientIDDisplayedInUi))
	assert.Equal(t, oidcSecretMask, sampleApp.ClientSecretDisplayedInUi)

	oldSampleAppFromBackend := component.GetInstalledSample(t, frame.Client)
	oldClientID := oldSampleAppFromBackend.ClientId
	oldClientSecret := oldSampleAppFromBackend.ClientSecret
	test_tools.AssertGeneratedOidcCredentials(t, oldClientID, oldClientSecret)
	assert.Equal(t, oldClientID, sampleApp.ClientIDDisplayedInUi)

	page.RegenerateCredentials("sampleapp")

	err = tools.Eventually(func() error {
		newSampleAppFromBackend := component.GetInstalledSample(t, frame.Client)
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
	assert.Equal(t, test_tools.OidcClientIDLength, len(sampleApp.ClientIDDisplayedInUi))
	assert.Equal(t, oidcSecretMask, sampleApp.ClientSecretDisplayedInUi)

	newSampleAppFromBackend := component.GetInstalledSample(t, frame.Client)
	newClientID := newSampleAppFromBackend.ClientId
	newClientSecret := newSampleAppFromBackend.ClientSecret
	test_tools.AssertGeneratedOidcCredentials(t, newClientID, newClientSecret)
	assert.Equal(t, newClientID, sampleApp.ClientIDDisplayedInUi)

	test_tools.AssertOidcCredentialsChanged(t, oldClientID, oldClientSecret, newClientID, newClientSecret)
}
