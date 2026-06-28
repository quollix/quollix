package users

import (
	"net/http"
	"net/http/httptest"
	"server/tools"
	"testing"
	"time"

	"github.com/quollix/common/assert"
	u "github.com/quollix/common/utils"
	"github.com/stretchr/testify/mock"
)

const authenticationServiceCookieValue = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

type authenticationServiceTestObjects struct {
	Service     *AuthenticationServiceImpl
	SessionRepo *SessionRepositoryMock
	AuthHelper  u.AuthHelper
}

func getAuthenticationServiceTestObjects(t *testing.T) *authenticationServiceTestObjects {
	sessionRepo := NewSessionRepositoryMock(t)
	authHelper := &u.AuthHelperImpl{}
	service := &AuthenticationServiceImpl{
		SessionService: &SessionServiceImpl{
			SessionRepo: sessionRepo,
			AuthHelper:  authHelper,
		},
	}
	return &authenticationServiceTestObjects{
		Service:     service,
		SessionRepo: sessionRepo,
		AuthHelper:  authHelper,
	}
}

func newAuthenticationServiceRequest() *http.Request {
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.AddCookie(&http.Cookie{
		Name:  tools.BrandAppAuthCookieName,
		Value: authenticationServiceCookieValue,
	})
	return request
}

func getAuthenticationServiceSession(id int, isAdmin bool, expirationTime time.Time) *AuthenticatedSession {
	return &AuthenticatedSession{
		User: tools.User{
			Id:        id,
			IsAdmin:   isAdmin,
			IsEnabled: true,
		},
		Session: UserSession{
			CookieExpirationDate: expirationTime,
		},
	}
}

func TestAuthenticationService_GetRequestWithAuthContextAddsAuthenticatedUserToContext(t *testing.T) {
	testObjects := getAuthenticationServiceTestObjects(t)
	request := newAuthenticationServiceRequest()
	hashedCookie := testObjects.AuthHelper.GetSHA256Hash(authenticationServiceCookieValue)
	testObjects.SessionRepo.EXPECT().
		GetAuthenticatedSession(hashedCookie, QuollixSessionAudience()).
		Return(getAuthenticationServiceSession(42, true, time.Now().UTC().Add(time.Hour)), nil)
	testObjects.SessionRepo.EXPECT().UpdateCookieExpirationDate(hashedCookie, QuollixSessionAudience(), mock.MatchedBy(func(cookieExpirationDate time.Time) bool {
		return cookieExpirationDate.After(time.Now().UTC())
	})).Return(nil)

	requestWithAuth, err := testObjects.Service.GetRequestWithAuthContext(httptest.NewRecorder(), request)
	assert.Nil(t, err)
	authenticatedUser := requestWithAuth.Context().Value(tools.AuthKey).(tools.User)
	assert.Equal(t, 42, authenticatedUser.Id)
	assert.True(t, authenticatedUser.IsAdmin)
	assert.True(t, authenticatedUser.CookieExpirationDate.After(time.Now().UTC()))
}

func TestAuthenticationService_GetRequestWithAuthContextReturnsErrorWhenCookieIsExpired(t *testing.T) {
	testObjects := getAuthenticationServiceTestObjects(t)
	request := newAuthenticationServiceRequest()
	hashedCookie := testObjects.AuthHelper.GetSHA256Hash(authenticationServiceCookieValue)
	testObjects.SessionRepo.EXPECT().
		GetAuthenticatedSession(hashedCookie, QuollixSessionAudience()).
		Return(getAuthenticationServiceSession(42, false, time.Now().UTC().Add(-time.Hour)), nil)

	requestWithAuth, err := testObjects.Service.GetRequestWithAuthContext(httptest.NewRecorder(), request)
	assert.Nil(t, requestWithAuth)
	assert.Equal(t, CookieExpiredError, u.ExtractError(err))
}
