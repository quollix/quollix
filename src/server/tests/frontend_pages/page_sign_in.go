package frontend_pages

import (
	"server/tools"
)

type SignInPage struct {
	Frame *FrameType
}

func (l *SignInPage) SignInAsAdmin() *InstalledAppsPage {
	l.Frame.Page.MustElement("#sign-in-tab").MustClick()
	l.Frame.Assert.PagePath(tools.Paths.FrontendSignIn)
	l.Frame.Page.MustElement("#username-input").MustInput("admin")
	l.Frame.Page.MustElement("#password-input").MustInput("password")
	l.Frame.Browser.DoAndWaitDOMContentLoaded(func() {
		l.Frame.Page.MustElement("#sign-in-button").MustClick()
	})
	l.Frame.Session.syncClientCookieFromBrowser()
	l.Frame.Assert.PagePath(tools.Paths.FrontendInstalledApps)
	return l.Frame.Pages.InstalledAppsPage
}
