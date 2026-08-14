//go:build oidc

package oidc

import (
	"testing"

	"server/tests/frontend_pages"
	"server/tools"
	"server/users"

	"github.com/quollix/common/assert"
	"github.com/quollix/common/quollix/api"
	"github.com/quollix/common/quollix/api_client"
	"github.com/quollix/common/quollix/test_environment"
	u "github.com/quollix/common/utils"
)

func TestOidcLoginFlowBetweenTwoQuollixInstances(t *testing.T) {
	clients := SetupAndGetClients(t)
	defer clients.Reset()

	frame := frontend_pages.NewBrowserFrame(t, test_environment.OidcClientBaseUrl, nil)
	signInViaOidcInBrowser(frame)
	openAccountPage(frame)
	frame.Assert.PageContainsEventually("Name: " + test_environment.OidcProviderAdminUsername)
	frame.Assert.PageContainsEventually("Email: " + test_environment.OidcProviderAdminUsername + "@example.invalid")
}

func TestOidcAccountPage_SetLocalPasswordShowsChangePasswordForm(t *testing.T) {
	clients := SetupAndGetClients(t)
	defer clients.Reset()
	oidcUserClient := signInViaOidcHttpClient(t, clients)

	frame := frontend_pages.NewBrowserFrame(t, test_environment.OidcClientBaseUrl, oidcUserClient)
	frame.Session.SetBrowserAuthCookie(oidcUserClient.Parent.Cookie)
	frame.Pages.GoToAccountPage().AssertSetPasswordFormState()

	frame.Pages.AccountPage.
		EnterSetPassword(test_environment.OidcProviderAdminPassword, "different-password").
		SaveSetPassword()
	frame.Assert.SnackbarVisibleWithTextEventually("Passwords do not match")
	frame.Pages.AccountPage.AssertSetPasswordFormState()
	err := api_client.NewQuollixClientForRootUrl(test_environment.OidcClientBaseUrl).Auth.SignIn(test_environment.OidcProviderAdminUsername, test_environment.OidcProviderAdminPassword)
	u.AssertDeepStackErrorFromRequest(t, err, users.IncorrectLoginCredentialsError)

	frame.Pages.AccountPage.
		EnterSetPassword(test_environment.OidcProviderAdminPassword, test_environment.OidcProviderAdminPassword).
		SaveSetPasswordAndWaitForReload().
		AssertChangePasswordFormState()

	passwordLoginClient := api_client.NewQuollixClientForRootUrl(test_environment.OidcClientBaseUrl)
	assert.Nil(t, passwordLoginClient.Auth.SignIn(test_environment.OidcProviderAdminUsername, test_environment.OidcProviderAdminPassword))
}

func signInViaOidcInBrowser(frame *frontend_pages.FrameType) {
	frame.Page.MustNavigate(frame.BaseUrl + api.Paths.FrontendSignIn)
	frame.Page.MustElementMatchingText(".oidc-provider-button", test_environment.OidcProviderName).MustClick()
	frame.Assert.HostEventually("quollix." + test_environment.OidcProviderHost)

	loginViaBrowser(frame, test_environment.OidcProviderAdminUsername, tools.DefaultAdminPassword)
	frame.Assert.HostEventually("quollix." + test_environment.OidcClientHost)
}

func openAccountPage(frame *frontend_pages.FrameType) {
	frame.Browser.WaitForElement("#sidebar-user-link")
	frame.Browser.DoAndWaitDOMContentLoaded(func() {
		frame.Page.MustElement("#sidebar-user-link").MustClick()
	})
	frame.Assert.PathEventually(api.Paths.FrontendAccount)
}

func loginViaBrowser(frame *frontend_pages.FrameType, username string, password string) {
	frame.Browser.WaitForElement("#username-input")
	frame.Page.MustElement("#username-input").MustInput(username)
	frame.Page.MustElement("#password-input").MustInput(password)
	frame.Browser.DoAndWaitDOMContentLoaded(func() {
		frame.Page.MustElement("#sign-in-button").MustClick()
	})
}
