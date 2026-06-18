package users

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"server/tools"
	"testing"
	"time"

	"github.com/quollix/common/assert"
	u "github.com/quollix/common/utils"
)

const requestAuthCookieValue = "plain-cookie-value"

type userServiceTestObjects struct {
	Service     *UserServiceImpl
	UserRepo    *UserRepositoryMock
	SessionRepo *SessionRepositoryMock
	OsWrapper   *tools.CommonOsWrapperMock
}

func getUserServiceTestObjects(t *testing.T) *userServiceTestObjects {
	userRepo := NewUserRepositoryMock(t)
	sessionRepo := NewSessionRepositoryMock(t)
	osWrapper := tools.NewCommonOsWrapperMock(t)
	authHelper := &u.AuthHelperImpl{}
	service := &UserServiceImpl{
		UserRepo: userRepo,
		SessionService: &SessionServiceImpl{
			SessionRepo: sessionRepo,
			AuthHelper:  authHelper,
		},
		AuthHelper: authHelper,
		OsWrapper:  osWrapper,
	}
	return &userServiceTestObjects{
		Service:     service,
		UserRepo:    userRepo,
		SessionRepo: sessionRepo,
		OsWrapper:   osWrapper,
	}
}

func newRequestWithAuthCookie() *http.Request {
	request := httptest.NewRequest("GET", "/", nil)
	request.AddCookie(&http.Cookie{
		Name:  tools.BrandAppAuthCookieName,
		Value: requestAuthCookieValue,
	})
	return request
}

func TestGetUserIdAndRoleFromQuollixRequest_WithoutAuthCookieReturnsAnonymousUser(t *testing.T) {
	testObjects := getUserServiceTestObjects(t)
	request := httptest.NewRequest("GET", "/", nil)

	userId, accessLevel, err := testObjects.Service.GetUserIdAndRoleFromQuollixRequest(request)
	assert.Nil(t, err)
	assert.Equal(t, AnonymousUserId, userId)
	assert.Equal(t, tools.AnonymousLevel, accessLevel)
}

func TestGetUserIdAndRoleFromQuollixRequest_WhenCookieDoesNotResolveReturnsAnonymousUser(t *testing.T) {
	testObjects := getUserServiceTestObjects(t)
	request := newRequestWithAuthCookie()
	hashedCookie := testObjects.Service.AuthHelper.GetSHA256Hash(requestAuthCookieValue)
	testObjects.SessionRepo.EXPECT().
		GetAuthenticatedSession(hashedCookie, QuollixSessionAudience()).
		Return(nil, errors.New("cookie not found"))

	userId, accessLevel, err := testObjects.Service.GetUserIdAndRoleFromQuollixRequest(request)
	assert.Nil(t, err)
	assert.Equal(t, AnonymousUserId, userId)
	assert.Equal(t, tools.AnonymousLevel, accessLevel)
}

func TestGetUserIdAndRoleFromQuollixRequest_AdminCookieReturnsAdminUser(t *testing.T) {
	testObjects := getUserServiceTestObjects(t)
	request := newRequestWithAuthCookie()
	hashedCookie := testObjects.Service.AuthHelper.GetSHA256Hash(requestAuthCookieValue)
	testObjects.SessionRepo.EXPECT().
		GetAuthenticatedSession(hashedCookie, QuollixSessionAudience()).
		Return(getAuthenticatedSession(42, true, time.Now().Add(time.Hour)), nil)

	userId, accessLevel, err := testObjects.Service.GetUserIdAndRoleFromQuollixRequest(request)
	assert.Nil(t, err)
	assert.Equal(t, 42, userId)
	assert.Equal(t, tools.AdminLevel, accessLevel)
}

func TestGetUserIdAndRoleFromQuollixRequest_RegularCookieReturnsRegularUser(t *testing.T) {
	testObjects := getUserServiceTestObjects(t)
	request := newRequestWithAuthCookie()
	hashedCookie := testObjects.Service.AuthHelper.GetSHA256Hash(requestAuthCookieValue)
	testObjects.SessionRepo.EXPECT().
		GetAuthenticatedSession(hashedCookie, QuollixSessionAudience()).
		Return(getAuthenticatedSession(7, false, time.Now().Add(time.Hour)), nil)

	userId, accessLevel, err := testObjects.Service.GetUserIdAndRoleFromQuollixRequest(request)
	assert.Nil(t, err)
	assert.Equal(t, 7, userId)
	assert.Equal(t, tools.UserLevel, accessLevel)
}

func TestGetUserIdAndRoleFromRequestForAudience_UsesProvidedAudience(t *testing.T) {
	testObjects := getUserServiceTestObjects(t)
	request := newRequestWithAuthCookie()
	hashedCookie := testObjects.Service.AuthHelper.GetSHA256Hash(requestAuthCookieValue)
	audience := SessionAudience("sample-maintainer", "sample-app")
	testObjects.SessionRepo.EXPECT().
		GetAuthenticatedSession(hashedCookie, audience).
		Return(getAuthenticatedSession(7, false, time.Now().Add(time.Hour)), nil)

	userId, accessLevel, err := testObjects.Service.GetUserIdAndRoleFromRequestForAudience(request, audience)
	assert.Nil(t, err)
	assert.Equal(t, 7, userId)
	assert.Equal(t, tools.UserLevel, accessLevel)
}

func TestGetUserIdAndRoleFromQuollixRequest_ExpiredSessionReturnsAnonymousUser(t *testing.T) {
	testObjects := getUserServiceTestObjects(t)
	request := newRequestWithAuthCookie()
	hashedCookie := testObjects.Service.AuthHelper.GetSHA256Hash(requestAuthCookieValue)
	testObjects.SessionRepo.EXPECT().
		GetAuthenticatedSession(hashedCookie, QuollixSessionAudience()).
		Return(getAuthenticatedSession(7, false, time.Now().Add(-time.Hour)), nil)

	userId, accessLevel, err := testObjects.Service.GetUserIdAndRoleFromQuollixRequest(request)
	assert.Nil(t, err)
	assert.Equal(t, AnonymousUserId, userId)
	assert.Equal(t, tools.AnonymousLevel, accessLevel)
}

func getAuthenticatedSession(userId int, isAdmin bool, cookieExpirationDate time.Time) *AuthenticatedSession {
	return &AuthenticatedSession{
		User: tools.User{
			Id:      userId,
			IsAdmin: isAdmin,
		},
		Session: UserSession{
			CookieExpirationDate: cookieExpirationDate,
		},
	}
}

func TestAcceptNewPasswordViaToken_ValidTokenUpdatesPasswordAndClearsInvitationState(t *testing.T) {
	testObjects := getUserServiceTestObjects(t)
	user := &tools.User{
		Id:                             42,
		SetPasswordToken:               "token-123",
		SetPasswordTokenExpirationDate: time.Now().Add(time.Hour),
	}
	testObjects.UserRepo.EXPECT().GetUserByToken("token-123").Return(user, nil)
	testObjects.UserRepo.EXPECT().UpdateUser(user).Run(func(info *tools.User) {
		assert.True(t, testObjects.Service.AuthHelper.DoesMatchSaltedHash("new-password", info.HashedPassword))
		assert.Equal(t, "", info.SetPasswordToken)
		assert.Equal(t, tools.DefaultTime, info.SetPasswordTokenExpirationDate)
	}).Return(nil)

	err := testObjects.Service.AcceptNewPasswordViaToken("new-password", "token-123")
	assert.Nil(t, err)
}

func TestAcceptNewPasswordViaToken_ExpiredTokenReturnsError(t *testing.T) {
	testObjects := getUserServiceTestObjects(t)
	user := &tools.User{
		Id:                             42,
		SetPasswordToken:               "token-123",
		SetPasswordTokenExpirationDate: time.Now().Add(-time.Hour),
	}
	testObjects.UserRepo.EXPECT().GetUserByToken("token-123").Return(user, nil)

	err := testObjects.Service.AcceptNewPasswordViaToken("new-password", "token-123")
	assert.Equal(t, TokenExpiredError, u.ExtractError(err))
}

func TestAcceptNewPasswordViaToken_UnknownTokenReturnsError(t *testing.T) {
	testObjects := getUserServiceTestObjects(t)
	testObjects.UserRepo.EXPECT().GetUserByToken("missing-token").Return(nil, errors.New(UserNotFoundError))

	err := testObjects.Service.AcceptNewPasswordViaToken("new-password", "missing-token")
	assert.Equal(t, UserNotFoundError, u.ExtractError(err))
}
