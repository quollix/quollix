//go:build frontend

package frontend

import (
	"testing"

	"server/tests/component"
	"server/tests/frontend_pages"
	"server/tools"

	"github.com/quollix/common/assert"
	"github.com/quollix/common/quollix/api"
)

const (
	sampleAppDirectURL        = "https://sampleapp.localhost"
	sampleAppHost             = "sampleapp.localhost"
	sampleAppExpectedBodyText = "this is version 2.0"
)

func TestAnonymousUserSignsInAndReturnsToAppPath(t *testing.T) {
	frame := frontend_pages.Setup(t)
	defer frame.Client.Test.ResetTestState()

	app, err := component.InstallAndStartSample(t, frame.Client, "2.0")
	assert.Nil(t, err)
	assert.Nil(t, frame.Client.Apps.SetAccessPolicy(app.AppId, api.Policies.AuthenticatedAccessPolicy))
	frame.Session.SignOutViaClient()

	frame.Pages.VisitURL(sampleAppDirectURL + "/custom-path")
	frame.Assert.HostEventually("quollix.localhost").Assert.PathEventually(api.Paths.FrontendSignIn)

	signInAsAdminOnCurrentPage(t, frame)
	assertSampleAppLoaded(t, frame, "/custom-path")
}

func TestAuthenticatedUserWithoutAppCookieOpensAppSilently(t *testing.T) {
	frame := frontend_pages.Setup(t)
	defer frame.Client.Test.ResetTestState()

	app, err := component.InstallAndStartSample(t, frame.Client, "2.0")
	assert.Nil(t, err)
	assert.Nil(t, frame.Client.Apps.SetAccessPolicy(app.AppId, api.Policies.AuthenticatedAccessPolicy))

	frame.Pages.VisitURL(sampleAppDirectURL + "/")
	assertSampleAppLoaded(t, frame, "/")
}

func TestUserWithoutAccessClicksInstalledAppsLink(t *testing.T) {
	frame := frontend_pages.Setup(t)
	defer frame.Client.Test.ResetTestState()

	app, err := component.InstallAndStartSample(t, frame.Client, "2.0")
	assert.Nil(t, err)
	assert.Nil(t, frame.Client.Apps.SetAccessPolicy(app.AppId, api.Policies.AdminOnlyAccessPolicy))
	component.InviteUserAndSetPassword(t, frame.Client, component.SampleUsername, component.SampleUserPassword, component.SampleUserEmail)
	frame.Session.SignInViaClient(component.SampleUsername, component.SampleUserPassword)

	frame.Pages.VisitURL(sampleAppDirectURL + "/")
	frame.Assert.HostEventually(sampleAppHost).Assert.PageContainsEventually(tools.AppUnavailableTitle)
	frame.Assert.PageContainsEventually(tools.AppUnavailableMessage)

	clickInstalledAppsLink(t, frame)
	frame.Assert.HostEventually("quollix.localhost").Assert.PathEventually(api.Paths.FrontendInstalledApps)
}

func clickInstalledAppsLink(t *testing.T, frame *frontend_pages.FrameType) {
	assert.Nil(t, frame.Page.DoAndWaitLoad(func() error {
		link, err := frame.Page.ElementMatchingText("a", tools.AppUnavailableInstalledAppsLinkText)
		if err != nil {
			return err
		}
		return link.Click()
	}))
}

func signInAsAdminOnCurrentPage(t *testing.T, frame *frontend_pages.FrameType) {
	signInTab, err := frame.Page.Element("#sign-in-tab")
	assert.Nil(t, err)
	assert.Nil(t, signInTab.Click())

	usernameInput, err := frame.Page.Element("#username-input")
	assert.Nil(t, err)
	assert.Nil(t, usernameInput.Input(tools.DefaultAdminName))

	passwordInput, err := frame.Page.Element("#password-input")
	assert.Nil(t, err)
	assert.Nil(t, passwordInput.Input(tools.DefaultAdminPassword))

	assert.Nil(t, frame.Page.DoAndWaitLoad(func() error {
		signInButton, err := frame.Page.Element("#sign-in-button")
		if err != nil {
			return err
		}
		return signInButton.Click()
	}))
}

func assertSampleAppLoaded(t *testing.T, frame *frontend_pages.FrameType, expectedPath string) {
	frame.Assert.HostEventually(sampleAppHost)
	frame.Assert.PathEventually(expectedPath)
	frame.Assert.PageContainsEventually(sampleAppExpectedBodyText)
}
