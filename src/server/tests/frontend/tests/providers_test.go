//go:build frontend

package frontend

import (
	"testing"

	"server/tests/component"
	"server/tests/frontend_pages"
	test_tools "server/tests/test_tools"

	"github.com/quollix/common/assert"
	"github.com/quollix/common/quollix/api"
)

func TestProvidersPageProviderCrud(t *testing.T) {
	frame := frontend_pages.Setup(t)
	defer frame.Client.Test.ResetTestState()

	page := frame.Pages.OpenProvidersPage()

	provider := test_tools.GetSampleOidcAuthProvider()
	provider.IssuerDomainPath = "quollix.localhost"
	page.
		FillCreateProvider(provider).
		TestCreateProviderDiscoveryAndAssertSuccess().
		CreateProviderAndAssertSuccess().
		AssertClientSecretVisibility(provider.Name, false).
		ToggleClientSecretVisibility(provider.Name).
		AssertClientSecretVisibility(provider.Name, true).
		ToggleClientSecretVisibility(provider.Name).
		AssertClientSecretVisibility(provider.Name, false)
	createdProvider := component.GetOnlyOidcAuthProvider(t, frame.Client)
	assert.Equal(t, createdProvider, page.GetRequiredProvider(provider.Name))
	assertOidcAuthProvider(t, *provider, createdProvider)

	updatedProvider := test_tools.GetUpdatedSampleOidcAuthProvider()
	updatedProvider.Id = createdProvider.Id
	page.UpdateProvider(updatedProvider)
	actualUpdatedProvider := component.GetOnlyOidcAuthProvider(t, frame.Client)
	assertOidcAuthProvider(t, *updatedProvider, actualUpdatedProvider)

	page.DeleteProvider(actualUpdatedProvider.Id)
	providers, err := frame.Client.OidcProviders.List()
	assert.Nil(t, err)
	assert.Equal(t, 0, len(providers))
}

func assertOidcAuthProvider(t *testing.T, expected api.OidcAuthProviderDto, actual api.OidcAuthProviderDto) {
	assert.Equal(t, expected.Name, actual.Name)
	assert.Equal(t, expected.IssuerDomainPath, actual.IssuerDomainPath)
	assert.Equal(t, expected.ClientId, actual.ClientId)
	assert.Equal(t, expected.ClientSecret, actual.ClientSecret)
}
