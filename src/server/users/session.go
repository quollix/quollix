package users

import (
	"net/http"
	"server/tools"
	"strings"
	"time"

	u "github.com/quollix/common/utils"
)

const (
	QuollixSessionMaintainer = "quollix"
	QuollixSessionAppName    = "quollix"
	sessionAudienceSeparator = "|"
)

type UserSession struct {
	SessionId            int
	UserId               int
	Audience             string
	HashedCookieValue    string
	CookieExpirationDate time.Time
}

type AuthenticatedSession struct {
	User    tools.User
	Session UserSession
}

type SessionService interface {
	GenerateAndSaveCookie(userId int, audience string) (*http.Cookie, error)
	GetAuthenticatedSession(cookieValue, audience string) (*AuthenticatedSession, error)
	UpdateCookieExpirationDate(cookieValue, audience string) (time.Time, error)
	DeleteSessionsByUserId(userId int) error
}

type SessionServiceImpl struct {
	SessionRepo SessionRepository
	AuthHelper  u.AuthHelper
}

func QuollixSessionAudience() string {
	return SessionAudience(QuollixSessionMaintainer, QuollixSessionAppName)
}

func SessionAudience(maintainer, appName string) string {
	return strings.Join([]string{maintainer, appName}, sessionAudienceSeparator)
}

func (s *SessionServiceImpl) GenerateAndSaveCookie(userId int, audience string) (*http.Cookie, error) {
	cookie, err := s.AuthHelper.GenerateCookie()
	if err != nil {
		return nil, u.Logger.NewError(err.Error())
	}
	cookie.Secure = true
	// This line is important. Using a strict cookie policy here would break the feature that when you click the 'Open' button in the GUI, you are redirected to the app web interface. Browsers treat that new tab to the app subdomain as a cross-site navigation, so the exchanged app cookie must be Lax to be sent on the follow-up request.
	cookie.SameSite = http.SameSiteLaxMode
	cookie.Name = tools.BrandAppAuthCookieName

	session := &UserSession{
		UserId:               userId,
		Audience:             audience,
		HashedCookieValue:    s.AuthHelper.GetSHA256Hash(cookie.Value),
		CookieExpirationDate: time.Now().UTC().Add(tools.CookieExpirationTime),
	}
	if _, err = s.SessionRepo.CreateSession(session); err != nil {
		return nil, err
	}
	return cookie, nil
}

func (s *SessionServiceImpl) GetAuthenticatedSession(cookieValue, audience string) (*AuthenticatedSession, error) {
	hashedCookie := s.AuthHelper.GetSHA256Hash(cookieValue)
	return s.SessionRepo.GetAuthenticatedSession(hashedCookie, audience)
}

func (s *SessionServiceImpl) UpdateCookieExpirationDate(cookieValue, audience string) (time.Time, error) {
	hashedCookie := s.AuthHelper.GetSHA256Hash(cookieValue)
	cookieExpirationDate := time.Now().UTC().Add(tools.CookieExpirationTime)
	return cookieExpirationDate, s.SessionRepo.UpdateCookieExpirationDate(
		hashedCookie,
		audience,
		cookieExpirationDate,
	)
}

func (s *SessionServiceImpl) DeleteSessionsByUserId(userId int) error {
	return s.SessionRepo.DeleteSessionsByUserId(userId)
}
