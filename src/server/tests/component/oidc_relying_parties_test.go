//go:build component

package component

import (
	"net/http"
	"testing"

	"server/oidc_provider"
	test_tools "server/tests/test_tools"

	"github.com/quollix/common/assert"
	u "github.com/quollix/common/utils"
)

func TestOidcRelyingParty_Crud(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	oidcClient := test_tools.GetSampleOidcRelyingParty()
	assert.Nil(t, client.OidcClients.Create(oidcClient))

	createdClient := GetOnlyOidcRelyingParty(t, client)
	assertOidcRelyingPartyConfiguration(t, *oidcClient, createdClient)
	test_tools.AssertGeneratedOidcCredentials(t, createdClient.ClientId, createdClient.ClientSecret)

	updatedClient := *test_tools.GetUpdatedSampleOidcRelyingParty()
	updatedClient.Id = createdClient.Id
	assert.Nil(t, client.OidcClients.Update(&updatedClient))

	actualUpdatedClient := GetOnlyOidcRelyingParty(t, client)
	updatedClient.ClientId = createdClient.ClientId
	updatedClient.ClientSecret = createdClient.ClientSecret
	assert.Equal(t, updatedClient, actualUpdatedClient)

	assert.Nil(t, client.OidcClients.Regenerate(actualUpdatedClient.Id))
	regeneratedClient := GetOnlyOidcRelyingParty(t, client)
	assertOidcRelyingPartyConfiguration(t, updatedClient, regeneratedClient)
	test_tools.AssertOidcCredentialsChanged(t, actualUpdatedClient.ClientId, actualUpdatedClient.ClientSecret, regeneratedClient.ClientId, regeneratedClient.ClientSecret)

	assert.Nil(t, client.OidcClients.Delete(updatedClient.Id))
	clients, err := client.OidcClients.List()
	assert.Nil(t, err)
	assert.Equal(t, 0, len(clients))
}

func TestOidcRelyingParty_CreateClientWithSchemeInDomainReturnsError(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	oidcClient := test_tools.GetSampleOidcRelyingParty()
	oidcClient.Domain = "http://client.example.com"

	err := client.OidcClients.Create(oidcClient)

	assert.NotNil(t, err)
}

func TestOidcRelyingParty_CreateClientWithPathInDomainReturnsError(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	oidcClient := test_tools.GetSampleOidcRelyingParty()
	oidcClient.Domain = "client.example.com/callback"

	err := client.OidcClients.Create(oidcClient)

	assert.NotNil(t, err)
}

func TestOidcRelyingParty_CreateClientWithQueryInDomainReturnsError(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	oidcClient := test_tools.GetSampleOidcRelyingParty()
	oidcClient.Domain = "client.example.com?tenant=main"

	err := client.OidcClients.Create(oidcClient)

	assert.NotNil(t, err)
}

func TestOidcRelyingParty_CreateClientWithDuplicateNameReturnsError(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	oidcClient := test_tools.GetSampleOidcRelyingParty()
	assert.Nil(t, client.OidcClients.Create(oidcClient))

	duplicateClient := test_tools.GetUpdatedSampleOidcRelyingParty()
	duplicateClient.Name = oidcClient.Name
	err := client.OidcClients.Create(duplicateClient)

	u.AssertDeepStackErrorFromRequest(t, err, oidc_provider.OidcRelyingPartyNameAlreadyExistsError)
}

func TestOidcRelyingParty_OidcFlowWithGenericClient(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	oidcClient := test_tools.GetSampleOidcRelyingParty()
	assert.Nil(t, client.OidcClients.Create(oidcClient))
	createdClient := GetOnlyOidcRelyingParty(t, client)

	ctx := NewOidcTestClient(t)
	ctx.SignInAsAdmin()
	redirectURI := "https://" + createdClient.Domain + "/callback"
	authRes, verifier := ctx.AuthorizeWithPKCERedirectURI(createdClient.ClientId, redirectURI)
	tokens := ctx.ExchangeCodeForTokensWithRedirectURI(authRes.Code, verifier, redirectURI, createdClient.ClientId, createdClient.ClientSecret, ClientAuthMethodBasic)
	publicKey := FetchPublicKeyFromJWKS(t, ctx)
	claims := VerifyIDToken(t, tokens.IDToken, publicKey, ctx.Config.Issuer, createdClient.ClientId)
	assert.Equal(t, adminUsername, claims.PreferredUsername)

	userinfo := ctx.FetchUserinfo(tokens.AccessToken)
	assert.Equal(t, adminUsername, userinfo.PreferredUsername)
}

func TestOidcRelyingParty_OidcFlowAllowsAnyCallbackPathOnConfiguredOrigin(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	oidcClient := test_tools.GetSampleOidcRelyingParty()
	assert.Nil(t, client.OidcClients.Create(oidcClient))
	createdClient := GetOnlyOidcRelyingParty(t, client)

	ctx := NewOidcTestClient(t)
	ctx.SignInAsAdmin()
	redirectURI := "https://" + createdClient.Domain + "/custom/oidc/callback"
	authRes, verifier := ctx.AuthorizeWithPKCERedirectURI(createdClient.ClientId, redirectURI)
	tokens := ctx.ExchangeCodeForTokensWithRedirectURI(authRes.Code, verifier, redirectURI, createdClient.ClientId, createdClient.ClientSecret, ClientAuthMethodBasic)
	assert.NotEqual(t, "", tokens.AccessToken)
}

func TestOidcRelyingParty_OidcFlowWithWrongClientSecretReturnsError(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	oidcClient := test_tools.GetSampleOidcRelyingParty()
	assert.Nil(t, client.OidcClients.Create(oidcClient))
	createdClient := GetOnlyOidcRelyingParty(t, client)

	ctx := NewOidcTestClient(t)
	ctx.SignInAsAdmin()
	redirectURI := "https://" + createdClient.Domain + "/callback"
	authRes, verifier := ctx.AuthorizeWithPKCERedirectURI(createdClient.ClientId, redirectURI)
	resp := ctx.exchangeCodeForTokensRaw(authRes.Code, verifier, redirectURI, createdClient.ClientId, "wrong-secret", ClientAuthMethodBasic)
	defer u.Close(resp.Body)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestOidcRelyingParty_OidcFlowWithWrongRedirectDomainReturnsError(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	oidcClient := test_tools.GetSampleOidcRelyingParty()
	assert.Nil(t, client.OidcClients.Create(oidcClient))
	createdClient := GetOnlyOidcRelyingParty(t, client)

	ctx := NewOidcTestClient(t)
	ctx.SignInAsAdmin()
	resp, _ := ctx.authorizeWithPKCERaw(createdClient.ClientId, "https://evil.example.com/callback")
	defer u.Close(resp.Body)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestOidcRelyingParty_OidcFlowWithDeletedClientReturnsError(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	oidcClient := test_tools.GetSampleOidcRelyingParty()
	assert.Nil(t, client.OidcClients.Create(oidcClient))
	createdClient := GetOnlyOidcRelyingParty(t, client)
	assert.Nil(t, client.OidcClients.Delete(createdClient.Id))

	ctx := NewOidcTestClient(t)
	ctx.SignInAsAdmin()
	resp, _ := ctx.authorizeWithPKCERaw(createdClient.ClientId, "https://"+createdClient.Domain+"/callback")
	defer u.Close(resp.Body)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func assertOidcRelyingPartyConfiguration(t *testing.T, expected oidc_provider.OidcRelyingPartyDto, actual oidc_provider.OidcRelyingPartyDto) {
	assert.Equal(t, expected.Name, actual.Name)
	assert.Equal(t, expected.Domain, actual.Domain)
}
