//go:build integration

package repository

import (
	"server/oidc_client"
	test_tools "server/tests/test_tools"
	"testing"

	"github.com/quollix/common/assert"
	u "github.com/quollix/common/utils"
)

func TestOidcAuthProviderRepository_CRUD(t *testing.T) {
	InitDeps()
	defer cleanupOidcAuthProviderRepoTest()

	provider := test_tools.GetSampleOidcAuthProvider()
	var err error
	provider.Id, err = OidcAuthProviderRepo.CreateProvider(provider)
	assert.Nil(t, err)

	actualProvider, err := OidcAuthProviderRepo.GetProviderById(provider.Id)
	assert.Nil(t, err)
	assertOidcAuthProviderEquality(t, provider, actualProvider)

	actualProvider, found, err := OidcAuthProviderRepo.GetProviderByIssuerDomainPath(provider.IssuerDomainPath)
	assert.Nil(t, err)
	assert.True(t, found)
	assertOidcAuthProviderEquality(t, provider, actualProvider)

	actualProvider, found, err = OidcAuthProviderRepo.GetProviderByName(provider.Name)
	assert.Nil(t, err)
	assert.True(t, found)
	assertOidcAuthProviderEquality(t, provider, actualProvider)

	providers, err := OidcAuthProviderRepo.ListProviders()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(providers))
	assertOidcAuthProviderEquality(t, provider, &providers[0])

	updatedProvider := test_tools.GetUpdatedSampleOidcAuthProvider()
	updatedProvider.Id = provider.Id
	provider = updatedProvider
	assert.Nil(t, OidcAuthProviderRepo.UpdateProvider(provider))
	actualProvider, err = OidcAuthProviderRepo.GetProviderById(provider.Id)
	assert.Nil(t, err)
	assertOidcAuthProviderEquality(t, provider, actualProvider)

	assert.Nil(t, OidcAuthProviderRepo.DeleteProvider(provider.Id))
	providers, err = OidcAuthProviderRepo.ListProviders()
	assert.Nil(t, err)
	assert.Equal(t, 0, len(providers))
}

func TestOidcAuthProviderRepository_DuplicateIssuerDomainPathReturnsError(t *testing.T) {
	InitDeps()
	defer cleanupOidcAuthProviderRepoTest()

	provider := test_tools.GetSampleOidcAuthProvider()
	_, err := OidcAuthProviderRepo.CreateProvider(provider)
	assert.Nil(t, err)

	provider.Name = "other-provider"
	_, err = OidcAuthProviderRepo.CreateProvider(provider)
	assert.NotNil(t, err)
}

func TestOidcAuthProviderRepository_DuplicateNameReturnsError(t *testing.T) {
	InitDeps()
	defer cleanupOidcAuthProviderRepoTest()

	provider := test_tools.GetSampleOidcAuthProvider()
	_, err := OidcAuthProviderRepo.CreateProvider(provider)
	assert.Nil(t, err)

	provider.IssuerDomainPath = "https://other-auth.example.com/realms/sample"
	_, err = OidcAuthProviderRepo.CreateProvider(provider)
	assert.NotNil(t, err)
}

func TestOidcAuthProviderRepository_UpdateProvider_WhenProviderDoesNotExistReturnsError(t *testing.T) {
	InitDeps()
	defer cleanupOidcAuthProviderRepoTest()

	provider := test_tools.GetSampleOidcAuthProvider()
	provider.Id = 987654

	err := OidcAuthProviderRepo.UpdateProvider(provider)

	assert.Equal(t, oidc_client.OidcAuthProviderNotFoundError, u.ExtractError(err))
}

func TestOidcAuthProviderRepository_GetProviderByName_WhenProviderDoesNotExistReturnsFalse(t *testing.T) {
	InitDeps()
	defer cleanupOidcAuthProviderRepoTest()

	provider, found, err := OidcAuthProviderRepo.GetProviderByName("missing-provider")

	assert.Nil(t, err)
	assert.False(t, found)
	assert.Nil(t, provider)
}

func TestOidcAuthProviderRepository_GetProviderByIssuerDomainPath_WhenProviderDoesNotExistReturnsFalse(t *testing.T) {
	InitDeps()
	defer cleanupOidcAuthProviderRepoTest()

	provider, found, err := OidcAuthProviderRepo.GetProviderByIssuerDomainPath("https://missing-auth.example.com/realms/main")

	assert.Nil(t, err)
	assert.False(t, found)
	assert.Nil(t, provider)
}

func cleanupOidcAuthProviderRepoTest() {
	OidcAuthProviderRepo.Wipe()
}

func assertOidcAuthProviderEquality(t *testing.T, expected, actual *oidc_client.OidcAuthProviderDto) {
	assert.Equal(t, expected.Id, actual.Id)
	assert.Equal(t, expected.Name, actual.Name)
	assert.Equal(t, expected.IssuerDomainPath, actual.IssuerDomainPath)
	assert.Equal(t, expected.ClientId, actual.ClientId)
	assert.Equal(t, expected.ClientSecret, actual.ClientSecret)
}
