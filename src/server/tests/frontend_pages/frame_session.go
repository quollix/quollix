package frontend_pages

import (
	"net/http"
	"time"

	"server/tools"

	"github.com/quollix/common/assert"
	"github.com/quollix/common/browsertest"
	u "github.com/quollix/common/utils"
)

type FrameSession struct {
	Frame *FrameType
}

func (s *FrameSession) SetBrowserAuthCookie(cookie *http.Cookie) {
	assert.Nil(s.Frame.T, s.Frame.Quollix.UseAuthCookie(cookie))
}

func (s *FrameSession) ClearBrowserCookies() {
	assert.Nil(s.Frame.T, s.Frame.Quollix.ClearSession())
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
	err := u.EventuallyWithTimeout(backupOperationTimeout, 1*time.Second, func() error {
		err := s.Frame.Quollix.SyncedLoginWithClient(username, password)
		if browsertest.IsNetworkChangedError(err) {
			return err
		}
		assert.Nil(s.Frame.T, err)
		return nil
	})
	assert.Nil(s.Frame.T, err)
	return s.Frame
}

func (s *FrameSession) GetAuthCookie() *http.Cookie {
	return s.Frame.Quollix.Client.Parent.Cookie
}
