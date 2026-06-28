package oidc_client

import (
	"testing"

	"github.com/quollix/common/assert"
	u "github.com/quollix/common/utils"
)

func TestOidcAuthProviderServiceImpl_CreateProvider_WhenFieldIsEmptyReturnsError(t *testing.T) {
	testObjects := newAuthProviderServiceTestObjects(t)

	err := testObjects.Service.CreateProvider(&OidcAuthProviderDto{
		Name:             "Corporate-SSO",
		IssuerDomainPath: "auth.example.com/realms/main",
		ClientId:         "client-id",
		ClientSecret:     "",
	})

	assert.Equal(t, OidcAuthProviderRequiredFieldMissingError, u.ExtractError(err))
}

func TestOidcAuthProviderServiceImpl_UpdateProvider_WhenNameAndIssuerBelongToSameProviderUpdatesProvider(t *testing.T) {
	testObjects := newAuthProviderServiceTestObjects(t)
	provider := getSampleServiceOidcAuthProvider()
	provider.Id = 7
	testObjects.ProviderRepo.EXPECT().GetProviderByName(provider.Name).Return(provider, true, nil)
	testObjects.ProviderRepo.EXPECT().GetProviderByIssuerDomainPath(provider.IssuerDomainPath).Return(provider, true, nil)
	testObjects.ProviderRepo.EXPECT().UpdateProvider(provider).Return(nil)

	err := testObjects.Service.UpdateProvider(provider)

	assert.Nil(t, err)
}

type authProviderServiceTestObjects struct {
	Service      *OidcAuthProviderServiceImpl
	ProviderRepo *OidcAuthProviderRepositoryMock
}

func newAuthProviderServiceTestObjects(t *testing.T) authProviderServiceTestObjects {
	providerRepo := NewOidcAuthProviderRepositoryMock(t)
	return authProviderServiceTestObjects{
		Service:      &OidcAuthProviderServiceImpl{ProviderRepo: providerRepo},
		ProviderRepo: providerRepo,
	}
}

func getSampleServiceOidcAuthProvider() *OidcAuthProviderDto {
	return &OidcAuthProviderDto{
		Name:             "Corporate-SSO",
		IssuerDomainPath: "auth.example.com/realms/main",
		ClientId:         "client-id",
		ClientSecret:     "client-secret",
	}
}
