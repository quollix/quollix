package frontend_pages

import (
	"net/http"

	"server/tools"

	"github.com/go-rod/rod/lib/proto"
	"github.com/quollix/common/assert"
)

type FrameSession struct {
	Frame *FrameType
}

func (s *FrameSession) SetBrowserAuthCookie(cookie *http.Cookie) {
	s.Frame.Page.MustSetCookies(&proto.NetworkCookieParam{
		Name:     cookie.Name,
		Value:    cookie.Value,
		URL:      s.Frame.BaseUrl,
		Path:     "/",
		Secure:   cookie.Secure,
		HTTPOnly: cookie.HttpOnly,
	})
}

func (s *FrameSession) ClearBrowserCookies() {
	s.Frame.Page.MustSetCookies()
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
	err := s.Frame.Client.Auth.SignIn(username, password)
	assert.Nil(s.Frame.T, err)
	s.syncBrowserCookieFromClient()
	s.Frame.Pages.Visit(tools.Paths.FrontendInstalledApps)
	return s.Frame
}

func (s *FrameSession) GetAuthCookie() *http.Cookie {
	cookies := s.Frame.Page.MustCookies(s.Frame.BaseUrl)
	for _, cookie := range cookies {
		if cookie.Name == tools.BrandAppAuthCookieName {
			return &http.Cookie{ // #nosec G124: acceptance tests reconstruct the browser cookie only for local test replay
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
