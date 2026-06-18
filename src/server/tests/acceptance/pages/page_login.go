//go:build acceptance

package pages

import (
	"server/tools"
)

type LoginPage struct {
	Frame *FrameType
}

func (l *LoginPage) LoginAsAdmin() *InstalledAppsPage {
	l.Frame.page.MustElement("#login-tab").MustClick()
	l.Frame.AssertPagePath(tools.Paths.FrontendLogin)
	l.Frame.page.MustElement("#username-input").MustInput("admin")
	l.Frame.page.MustElement("#password-input").MustInput("password")
	l.Frame.DoAndWaitDOMContentLoaded(func() {
		l.Frame.page.MustElement("#login-button").MustClick()
	})
	l.Frame.syncClientCookieFromBrowser()
	l.Frame.AssertPagePath(tools.Paths.FrontendInstalledApps)
	return l.Frame.InstalledAppsPage
}
