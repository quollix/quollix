package frontend_pages

import (
	"net/http"

	"server/tools"

	"github.com/quollix/common/assert"
	"github.com/quollix/common/quollix/api"
)

type FrameSession struct {
	Frame *FrameType
}

func (s *FrameSession) SetBrowserAuthCookie(cookie *http.Cookie) {
	assert.Nil(s.Frame.T, s.Frame.Page.SetCookie(cookie, s.Frame.BaseUrl))
}

func (s *FrameSession) ClearBrowserCookies() {
	assert.Nil(s.Frame.T, s.Frame.Page.ClearCookies())
}

func (s *FrameSession) SignOut() *FrameType {
	s.Frame.Browser.DoAndWaitDOMContentLoaded(func() {
		s.Frame.Page.MustElement("#sign-out-tab a").MustClick()
	})
	s.Frame.Client.Parent.Cookie = nil
	s.ClearBrowserCookies()
	return s.Frame
}

func (s *FrameSession) SignOutViaClient() *FrameType {
	err := s.Frame.Client.Auth.SignOut()
	assert.Nil(s.Frame.T, err)
	s.Frame.Client.Parent.Cookie = nil
	s.ClearBrowserCookies()
	return s.Frame
}

func (s *FrameSession) SignInAsAdminViaClient() *FrameType {
	return s.SignInViaClient(tools.DefaultAdminName, tools.DefaultAdminPassword)
}

func (s *FrameSession) SignInViaClient(username, password string) *FrameType {
	s.Frame.Pages.Visit(api.Paths.FrontendSignIn)
	err := s.Frame.Client.Auth.SignIn(username, password)
	assert.Nil(s.Frame.T, err)
	s.syncBrowserCookieFromClient()
	s.Frame.Pages.Visit(api.Paths.FrontendInstalledApps)
	return s.Frame
}

func (s *FrameSession) GetAuthCookie() *http.Cookie {
	cookies, err := s.Frame.Page.Cookies(s.Frame.BaseUrl)
	assert.Nil(s.Frame.T, err)
	for _, cookie := range cookies {
		if cookie.Name == api.BrandAppAuthCookieName {
			return &http.Cookie{ // #nosec G124: frontend tests reconstruct the browser cookie only for local test replay
				Name:     cookie.Name,
				Value:    cookie.Value,
				Path:     cookie.Path,
				Secure:   cookie.Secure,
				HttpOnly: cookie.HTTPOnly,
			}
		}
	}
	return nil
}

func (s *FrameSession) syncClientCookieFromBrowser() {
	s.Frame.Client.Parent.Cookie = s.GetAuthCookie()
}

func (s *FrameSession) syncBrowserCookieFromClient() {
	s.ClearBrowserCookies()
	cookie := s.Frame.Client.Parent.Cookie
	if cookie == nil {
		return
	}
	s.SetBrowserAuthCookie(cookie)
}
