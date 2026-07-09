//go:build acceptance

package acceptance

import (
	"server/tests/frontend_pages"
	"server/tools"
	"testing"

	"github.com/quollix/common/assert"
)

func TestSignInAndSignOut(t *testing.T) {
	frame := frontend_pages.Setup(t)
	defer frame.Client.Test.ResetTestState()
	frame.Session.SignOutViaClient()

	frame.Pages.VisitSignInPage().SignInAsAdmin().Frame.Assert.PagePath(tools.Paths.FrontendInstalledApps)
	frame.Pages.GoToUsersPage().Frame.Assert.PagePath(tools.Paths.FrontendUsers)
	assert.NotNil(t, frame.Client.Parent.Cookie)

	cookie := frame.Session.GetAuthCookie()
	assert.NotNil(t, cookie)
	assert.Equal(t, frame.Client.Parent.Cookie.Value, cookie.Value)
	assert.Nil(t, checkAuthWithCookie(t, cookie))
	frame.Session.SignOut().Assert.PagePath(tools.Paths.FrontendInstalledApps)
	assert.Nil(t, frame.Client.Parent.Cookie)
}

func TestClientSignInAndSignOutSyncsBrowserCookie(t *testing.T) {
	frame := frontend_pages.Setup(t)
	defer frame.Client.Test.ResetTestState()

	assert.NotNil(t, frame.Client.Parent.Cookie)

	cookie := frame.Session.GetAuthCookie()
	assert.Equal(t, frame.Client.Parent.Cookie.Value, cookie.Value)
	assert.Nil(t, checkAuthWithCookie(t, cookie))

	frame.Session.SignOutViaClient()
	assert.Nil(t, frame.Client.Parent.Cookie)
	assert.Nil(t, frame.Session.GetAuthCookie())
}

func TestChangingPasswordRequiresMatchingNewPasswords(t *testing.T) {
	frame := frontend_pages.Setup(t)
	defer frame.Client.Test.ResetTestState()

	frame.Pages.GoToAccountPage().
		AssertChangePasswordFormState().
		EnterChangePassword(tools.DefaultAdminPassword, "new-admin-password", "different-admin-password").
		SaveChangePassword()

	frame.Assert.SnackbarVisibleWithTextEventually("Passwords do not match")
	assert.Nil(t, frame.Client.Auth.SignIn(tools.DefaultAdminName, tools.DefaultAdminPassword))
}
