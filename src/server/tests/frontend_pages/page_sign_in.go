package frontend_pages

import "github.com/quollix/common/quollix/api"

type SignInPage struct {
	Frame *FrameType
}

func (l *SignInPage) SignInAsAdmin() *InstalledAppsPage {
	l.Frame.Page.MustElement("#sign-in-tab").MustClick()
	l.Frame.Assert.PagePath(api.Paths.FrontendSignIn)
	l.Frame.Page.MustElement("#username-input").MustInput("admin")
	l.Frame.Page.MustElement("#password-input").MustInput("password")
	l.Frame.Browser.DoAndWaitDOMContentLoaded(func() {
		l.Frame.Page.MustElement("#sign-in-button").MustClick()
	})
	l.Frame.Assert.PagePath(api.Paths.FrontendInstalledApps)
	l.Frame.Session.syncClientCookieFromBrowser()
	return l.Frame.Pages.InstalledAppsPage
}
