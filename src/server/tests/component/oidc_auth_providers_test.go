//go:build component

package component

import (
	"testing"

	"server/oidc_client"
	test_tools "server/tests/test_tools"

	"github.com/quollix/common/assert"
	u "github.com/quollix/common/utils"
)

func TestOidcAuthProvider_Crud(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	provider := test_tools.GetSampleOidcAuthProvider()
	assert.Nil(t, client.OidcProviders.Create(provider))

	createdProvider := GetOnlyOidcAuthProvider(t, client)
	provider.Id = createdProvider.Id
	assert.Equal(t, *provider, createdProvider)

	updatedProvider := *test_tools.GetUpdatedSampleOidcAuthProvider()
	updatedProvider.Id = createdProvider.Id
	assert.Nil(t, client.OidcProviders.Update(&updatedProvider))

	actualUpdatedProvider := GetOnlyOidcAuthProvider(t, client)
	assert.Equal(t, updatedProvider, actualUpdatedProvider)

	assert.Nil(t, client.OidcProviders.Delete(updatedProvider.Id))
	providers, err := client.OidcProviders.List()
	assert.Nil(t, err)
	assert.Equal(t, 0, len(providers))
}

func TestOidcAuthProvider_CreateProviderWithSchemeInIssuerReturnsError(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	provider := test_tools.GetSampleOidcAuthProvider()
	provider.Name = "HTTP-Provider"
	provider.IssuerDomainPath = "http://auth.example.com/realms/main"

	err := client.OidcProviders.Create(provider)

	assert.NotNil(t, err)
}

func TestOidcAuthProvider_UpdateProviderWithSchemeInIssuerReturnsError(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	provider := test_tools.GetSampleOidcAuthProvider()
	assert.Nil(t, client.OidcProviders.Create(provider))
	createdProvider := GetOnlyOidcAuthProvider(t, client)
	createdProvider.IssuerDomainPath = "http://auth.example.com/realms/main"

	err := client.OidcProviders.Update(&createdProvider)

	assert.NotNil(t, err)
}

func TestOidcAuthProvider_CreateProviderWithDuplicateNameReturnsError(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	provider := test_tools.GetSampleOidcAuthProvider()
	assert.Nil(t, client.OidcProviders.Create(provider))

	duplicateProvider := test_tools.GetUpdatedSampleOidcAuthProvider()
	duplicateProvider.Name = provider.Name
	err := client.OidcProviders.Create(duplicateProvider)

	u.AssertDeepStackErrorFromRequest(t, err, oidc_client.OidcAuthProviderNameAlreadyExistsError)
}

func TestOidcAuthProvider_CreateProviderWithDuplicateIssuerDomainPathReturnsError(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	provider := test_tools.GetSampleOidcAuthProvider()
	assert.Nil(t, client.OidcProviders.Create(provider))

	duplicateProvider := test_tools.GetUpdatedSampleOidcAuthProvider()
	duplicateProvider.IssuerDomainPath = provider.IssuerDomainPath
	err := client.OidcProviders.Create(duplicateProvider)

	u.AssertDeepStackErrorFromRequest(t, err, oidc_client.OidcAuthProviderIssuerDomainPathAlreadyExistsError)
}
