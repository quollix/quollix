package test_tools

import "github.com/quollix/common/quollix/api"

func GetSampleOidcAuthProvider() *api.OidcAuthProviderDto {
	return &api.OidcAuthProviderDto{
		Name:             "Corporate-SSO",
		IssuerDomainPath: "auth.example.com/realms/main",
		ClientId:         "client-id",
		ClientSecret:     "client-secret",
	}
}

func GetUpdatedSampleOidcAuthProvider() *api.OidcAuthProviderDto {
	return &api.OidcAuthProviderDto{
		Name:             "Updated-SSO",
		IssuerDomainPath: "updated-auth.example.com/realms/main",
		ClientId:         "updated-client-id",
		ClientSecret:     "updated-client-secret",
	}
}
