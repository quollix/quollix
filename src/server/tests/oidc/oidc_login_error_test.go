//go:build oidc

package oidc

import (
	"net/http"
	"testing"

	"server/oidc_client"
	"server/tests/api_client"
	"server/users"

	"github.com/quollix/common/assert"
	u "github.com/quollix/common/utils"
)

func TestOidcLogin_EmailAlreadyExistsReturnsVisibleError(t *testing.T) {
	clients := SetupAndGetClients(t)
	defer clients.Reset(t)

	assert.Nil(t, clients.ClientAdmin.Users.Invite(ProviderAdminUsername, ProviderAdminUsername+"@example.invalid"))

	oidcClientLogin := api_client.NewQuollixClientForRootUrl(ClientBaseUrl)
	callbackResponse := sendOidcLoginCallbackRequest(t, clients.ProviderAdmin, oidcClientLogin, clients.ClientAdmin)
	defer u.Close(callbackResponse.Body)
	assert.Equal(t, http.StatusBadRequest, callbackResponse.StatusCode)
	assert.Equal(t, oidc_client.OidcLoginEmailAlreadyExistsError, readResponseBody(t, callbackResponse))
}

func TestOidcLogin_SettingLocalPasswordEnablesPasswordLogin(t *testing.T) {
	clients := SetupAndGetClients(t)
	defer clients.Reset(t)

	oidcUserClient := signInViaOidcHttpClient(t, clients)
	currentUser, err := oidcUserClient.Auth.GetCurrentUser()
	assert.Nil(t, err)
	assert.Equal(t, ProviderAdminUsername, currentUser.Username)

	passwordLoginClient := NewClientClient()
	err = passwordLoginClient.Auth.SignIn(ProviderAdminUsername, ProviderAdminLocalPassword)
	u.AssertDeepStackErrorFromRequest(t, err, users.IncorrectLoginCredentialsError)

	assert.Nil(t, oidcUserClient.Users.SetOwnPassword(ProviderAdminLocalPassword))
	assert.Nil(t, passwordLoginClient.Auth.SignIn(ProviderAdminUsername, ProviderAdminLocalPassword))
}
