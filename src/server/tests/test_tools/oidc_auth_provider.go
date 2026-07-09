package test_tools

import "server/oidc_client"

func GetSampleOidcAuthProvider() *oidc_client.OidcAuthProviderDto {
	return &oidc_client.OidcAuthProviderDto{
		Name:             "Corporate-SSO",
		IssuerDomainPath: "auth.example.com/realms/main",
		ClientId:         "client-id",
		ClientSecret:     "client-secret",
	}
}

func GetUpdatedSampleOidcAuthProvider() *oidc_client.OidcAuthProviderDto {
	return &oidc_client.OidcAuthProviderDto{
		Name:             "Updated-SSO",
		IssuerDomainPath: "updated-auth.example.com/realms/main",
		ClientId:         "updated-client-id",
		ClientSecret:     "updated-client-secret",
	}
}
