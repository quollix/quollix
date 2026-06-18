//go:build acceptance

package acceptance

import (
	"server/tests/acceptance/pages"
	"server/tools"
	"testing"

	"github.com/quollix/common/assert"
)

func TestLoginAndLogout(t *testing.T) {
	frame := pages.Setup(t)
	defer frame.Client.Test.ResetTestState()
	frame.LogoutViaClient()

	frame.VisitLoginPage().LoginAsAdmin().Frame.AssertPagePath(tools.Paths.FrontendInstalledApps)
	frame.GoToUsersPage().Frame.AssertPagePath(tools.Paths.FrontendUsers)
	assert.NotNil(t, frame.Client.Parent.Cookie)

	cookie := frame.GetAuthCookie()
	assert.NotNil(t, cookie)
	assert.Equal(t, frame.Client.Parent.Cookie.Value, cookie.Value)
	assert.Nil(t, checkAuthWithCookie(t, cookie))
	frame.Logout().AssertPagePath(tools.Paths.FrontendInstalledApps)
	assert.Nil(t, frame.Client.Parent.Cookie)
}

func TestClientLoginAndLogoutSyncsBrowserCookie(t *testing.T) {
	frame := pages.Setup(t)
	defer frame.Client.Test.ResetTestState()

	assert.NotNil(t, frame.Client.Parent.Cookie)

	cookie := frame.GetAuthCookie()
	assert.Equal(t, frame.Client.Parent.Cookie.Value, cookie.Value)
	assert.Nil(t, checkAuthWithCookie(t, cookie))

	frame.LogoutViaClient()
	assert.Nil(t, frame.Client.Parent.Cookie)
	assert.Nil(t, frame.GetAuthCookie())
}
