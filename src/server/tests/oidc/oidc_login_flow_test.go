//go:build oidc

package oidc

import (
	"testing"

	"server/tests/frontend_pages"
	"server/tools"
	"server/users"

	"github.com/quollix/common/assert"
	u "github.com/quollix/common/utils"
)

func TestOidcLoginFlowBetweenTwoQuollixInstances(t *testing.T) {
	clients := SetupAndGetClients(t)
	defer clients.Reset(t)

	frame := frontend_pages.NewBrowserFrame(t, ClientBaseUrl, nil)
	signInViaOidcInBrowser(frame)
	openAccountPage(frame)
	frame.Assert.PageContainsEventually("Name: " + ProviderAdminUsername)
	frame.Assert.PageContainsEventually("Email: " + ProviderAdminUsername + "@example.invalid")
}

func TestOidcAccountPage_SetLocalPasswordShowsChangePasswordForm(t *testing.T) {
	clients := SetupAndGetClients(t)
	defer clients.Reset(t)
	oidcUserClient := signInViaOidcHttpClient(t, clients)

	frame := frontend_pages.NewBrowserFrame(t, ClientBaseUrl, oidcUserClient)
	frame.Session.SetBrowserAuthCookie(oidcUserClient.Parent.Cookie)
	frame.Pages.GoToAccountPage().AssertSetPasswordFormState()

	frame.Pages.AccountPage.
		EnterSetPassword(ProviderAdminLocalPassword, "different-password").
		SaveSetPassword()
	frame.Assert.SnackbarVisibleWithTextEventually("Passwords do not match")
	frame.Pages.AccountPage.AssertSetPasswordFormState()
	err := NewClientClient().Auth.SignIn(ProviderAdminUsername, ProviderAdminLocalPassword)
	u.AssertDeepStackErrorFromRequest(t, err, users.IncorrectLoginCredentialsError)

	frame.Pages.AccountPage.
		EnterSetPassword(ProviderAdminLocalPassword, ProviderAdminLocalPassword).
		SaveSetPasswordAndWaitForReload().
		AssertChangePasswordFormState()

	passwordLoginClient := NewClientClient()
	assert.Nil(t, passwordLoginClient.Auth.SignIn(ProviderAdminUsername, ProviderAdminLocalPassword))
}

func signInViaOidcInBrowser(frame *frontend_pages.FrameType) {
	frame.Page.MustNavigate(frame.BaseUrl + tools.Paths.FrontendSignIn).MustWaitLoad()
	frame.Page.MustElementR(".oidc-provider-button", OidcProviderName).MustClick()
	frame.Assert.HostEventually("quollix." + ProviderHost)

	loginViaBrowser(frame, ProviderAdminUsername, tools.DefaultAdminPassword)
	frame.Assert.HostEventually("quollix." + ClientHost)
}

func openAccountPage(frame *frontend_pages.FrameType) {
	frame.Browser.WaitForElement("#sidebar-user-link")
	frame.Browser.DoAndWaitDOMContentLoaded(func() {
		frame.Page.MustElement("#sidebar-user-link").MustClick()
	})
	frame.Assert.PathEventually(tools.Paths.FrontendAccount)
}

func loginViaBrowser(frame *frontend_pages.FrameType, username string, password string) {
	frame.Browser.WaitForElement("#username-input")
	frame.Page.MustElement("#username-input").MustInput(username)
	frame.Page.MustElement("#password-input").MustInput(password)
	frame.Browser.DoAndWaitDOMContentLoaded(func() {
		frame.Page.MustElement("#sign-in-button").MustClick()
	})
}
