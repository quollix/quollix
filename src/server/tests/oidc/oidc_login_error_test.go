//go:build oidc

package oidc

import (
	"net/http"
	"testing"

	"server/oidc_client"
	"server/users"

	"github.com/quollix/common/quollix/api_client"
	"github.com/quollix/common/quollix/test_environment"

	"github.com/quollix/common/assert"
	u "github.com/quollix/common/utils"
)

func TestOidcLogin_EmailAlreadyExistsReturnsVisibleError(t *testing.T) {
	clients := SetupAndGetClients(t)
	defer clients.Reset()

	assert.Nil(t, clients.ClientAdmin.Users.Invite(test_environment.OidcProviderAdminUsername, test_environment.OidcProviderAdminUsername+"@example.invalid"))

	oidcClientLogin := api_client.NewQuollixClientForRootUrl(test_environment.OidcClientBaseUrl)
	callbackResponse := sendOidcLoginCallbackRequest(t, clients.ProviderAdmin, oidcClientLogin, clients.ClientAdmin)
	defer u.Close(callbackResponse.Body)
	assert.Equal(t, http.StatusBadRequest, callbackResponse.StatusCode)
	assert.Equal(t, oidc_client.OidcLoginEmailAlreadyExistsError, readResponseBody(t, callbackResponse))
}

func TestOidcLogin_SettingLocalPasswordEnablesPasswordLogin(t *testing.T) {
	clients := SetupAndGetClients(t)
	defer clients.Reset()

	oidcUserClient := signInViaOidcHttpClient(t, clients)
	currentUser, err := oidcUserClient.Auth.GetCurrentUser()
	assert.Nil(t, err)
	assert.Equal(t, test_environment.OidcProviderAdminUsername, currentUser.Username)

	passwordLoginClient := api_client.NewQuollixClientForRootUrl(test_environment.OidcClientBaseUrl)
	err = passwordLoginClient.Auth.SignIn(test_environment.OidcProviderAdminUsername, test_environment.OidcProviderAdminPassword)
	u.AssertDeepStackErrorFromRequest(t, err, users.IncorrectLoginCredentialsError)

	assert.Nil(t, oidcUserClient.Users.SetOwnPassword(test_environment.OidcProviderAdminPassword))
	assert.Nil(t, passwordLoginClient.Auth.SignIn(test_environment.OidcProviderAdminUsername, test_environment.OidcProviderAdminPassword))
}
