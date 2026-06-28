package oidc_client

import (
	"net/http"
	"strconv"
	"testing"
	"time"

	"server/tools"
	"server/users"

	"github.com/quollix/common/assert"
	u "github.com/quollix/common/utils"
	"github.com/stretchr/testify/mock"
)

var loginFlowSampleTime = time.Date(2026, time.June, 20, 12, 0, 0, 0, time.UTC)

func TestLoginServiceImpl_LoginWithClaims_WhenAuthMethodExistsUpdatesTimestampAndCreatesSession(t *testing.T) {
	testObjects := newLoginFlowTestObjects(t)
	testObjects.expectTrustedProvider()
	method := &UserAuthMethod{
		Id:                 55,
		UserId:             42,
		OidcAuthProviderId: 1,
		ExternalSubject:    "external-subject",
	}
	testObjects.OsWrapper.EXPECT().Now().Return(loginFlowSampleTime)
	testObjects.UserAuthMethodRepo.EXPECT().GetUserAuthMethodByProviderAndSubject(1, "external-subject").Return(method, true, nil)
	testObjects.UserRepo.EXPECT().GetUserById(42).Return(&tools.User{Id: 42, IsEnabled: true}, nil)
	testObjects.UserAuthMethodRepo.EXPECT().UpdateLastOidcAuthenticatedAt(55, loginFlowSampleTime).Return(nil)
	testObjects.SessionService.EXPECT().GenerateAndSaveCookie(42, users.QuollixSessionAudience()).Return(testCookie(42), nil)

	cookie, err := testObjects.Service.LoginWithClaims(1, OidcLoginClaims{Subject: "external-subject"})

	assert.Nil(t, err)
	assert.Equal(t, "cookie-42", cookie.Value)
}

func TestLoginServiceImpl_LoginWithClaims_WhenLinkedUserIsDisabledReturnsError(t *testing.T) {
	testObjects := newLoginFlowTestObjects(t)
	testObjects.expectTrustedProvider()
	method := &UserAuthMethod{
		Id:                 55,
		UserId:             42,
		OidcAuthProviderId: 1,
		ExternalSubject:    "external-subject",
	}
	testObjects.OsWrapper.EXPECT().Now().Return(loginFlowSampleTime)
	testObjects.UserAuthMethodRepo.EXPECT().GetUserAuthMethodByProviderAndSubject(1, "external-subject").Return(method, true, nil)
	testObjects.UserRepo.EXPECT().GetUserById(42).Return(&tools.User{Id: 42, IsEnabled: false}, nil)

	_, err := testObjects.Service.LoginWithClaims(1, OidcLoginClaims{Subject: "external-subject"})

	assert.Equal(t, users.UserDisabledError, u.ExtractError(err))
}

func TestLoginServiceImpl_LoginWithClaims_TrimsSubjectBeforeLookup(t *testing.T) {
	testObjects := newLoginFlowTestObjects(t)
	testObjects.expectTrustedProvider()
	method := &UserAuthMethod{
		Id:                 55,
		UserId:             42,
		OidcAuthProviderId: 1,
		ExternalSubject:    "external-subject",
	}
	testObjects.OsWrapper.EXPECT().Now().Return(loginFlowSampleTime)
	testObjects.UserAuthMethodRepo.EXPECT().GetUserAuthMethodByProviderAndSubject(1, "external-subject").Return(method, true, nil)
	testObjects.UserRepo.EXPECT().GetUserById(42).Return(&tools.User{Id: 42, IsEnabled: true}, nil)
	testObjects.UserAuthMethodRepo.EXPECT().UpdateLastOidcAuthenticatedAt(55, loginFlowSampleTime).Return(nil)
	testObjects.SessionService.EXPECT().GenerateAndSaveCookie(42, users.QuollixSessionAudience()).Return(testCookie(42), nil)

	cookie, err := testObjects.Service.LoginWithClaims(1, OidcLoginClaims{Subject: " external-subject "})

	assert.Nil(t, err)
	assert.Equal(t, "cookie-42", cookie.Value)
}

func TestLoginServiceImpl_LoginWithClaims_WhenAuthMethodIsMissingCreatesUserAuthMethodAndSession(t *testing.T) {
	testObjects := newLoginFlowTestObjects(t)
	testObjects.expectTrustedProvider()
	var createdUser *tools.User
	var createdMethod *UserAuthMethod
	testObjects.OsWrapper.EXPECT().Now().Return(loginFlowSampleTime)
	testObjects.UserAuthMethodRepo.EXPECT().
		GetUserAuthMethodByProviderAndSubject(1, "external-subject").
		Return(nil, false, nil)
	testObjects.UserResolver.EXPECT().ResolveUser(OidcLoginClaims{
		Subject:           "external-subject",
		Email:             "external@example.com",
		PreferredUsername: "external",
	}).Return("external", "external@example.com", nil)
	testObjects.UserRepo.EXPECT().CreateUser(mock.AnythingOfType("*tools.User")).RunAndReturn(func(user *tools.User) (int, error) {
		createdUser = user
		return 100, nil
	})
	testObjects.UserAuthMethodRepo.EXPECT().CreateUserAuthMethod(mock.AnythingOfType("*oidc_client.UserAuthMethod")).RunAndReturn(func(method *UserAuthMethod) (int, error) {
		createdMethod = method
		return 500, nil
	})
	testObjects.SessionService.EXPECT().GenerateAndSaveCookie(100, users.QuollixSessionAudience()).Return(testCookie(100), nil)

	cookie, err := testObjects.Service.LoginWithClaims(1, OidcLoginClaims{
		Subject:           "external-subject",
		Email:             "external@example.com",
		PreferredUsername: "external",
	})

	assert.Nil(t, err)
	assert.Equal(t, "cookie-100", cookie.Value)
	assert.Equal(t, "external", createdUser.Username)
	assert.Equal(t, "external@example.com", createdUser.Email)
	assert.Equal(t, "", createdUser.HashedPassword)
	assert.True(t, createdUser.IsEnabled)
	assert.Equal(t, loginFlowSampleTime, createdUser.CreationDate)
	assert.Equal(t, 100, createdMethod.UserId)
	assert.Equal(t, 1, createdMethod.OidcAuthProviderId)
	assert.Equal(t, "external-subject", createdMethod.ExternalSubject)
	assert.Equal(t, loginFlowSampleTime, createdMethod.LastOidcAuthenticatedAt)
}

func TestLoginServiceImpl_LoginWithClaims_WhenSubjectIsMissingReturnsError(t *testing.T) {
	testObjects := newLoginFlowTestObjects(t)

	_, err := testObjects.Service.LoginWithClaims(1, OidcLoginClaims{
		PreferredUsername: "external",
	})

	assert.Equal(t, MissingSubjectClaimError, u.ExtractError(err))
}

func TestLoginServiceImpl_LoginWithClaims_WhenSubjectContainsControlCharacterReturnsError(t *testing.T) {
	testObjects := newLoginFlowTestObjects(t)

	_, err := testObjects.Service.LoginWithClaims(1, OidcLoginClaims{
		Subject: "external\nsubject",
	})

	assert.NotNil(t, err)
}

func TestLoginServiceImpl_LoginWithClaims_WhenOptionalClaimContainsControlCharacterReturnsError(t *testing.T) {
	testObjects := newLoginFlowTestObjects(t)

	_, err := testObjects.Service.LoginWithClaims(1, OidcLoginClaims{
		Subject:           "external-subject",
		PreferredUsername: "external\nuser",
	})

	assert.NotNil(t, err)
}

type loginFlowTestObjects struct {
	Service            *LoginServiceImpl
	ProviderRepo       *OidcAuthProviderRepositoryMock
	UserAuthMethodRepo *UserAuthMethodRepositoryMock
	UserRepo           *users.UserRepositoryMock
	UserResolver       *OidcUserResolverMock
	SessionService     *users.SessionServiceMock
	OsWrapper          *tools.CommonOsWrapperMock
}

func newLoginFlowTestObjects(t *testing.T) loginFlowTestObjects {
	providerRepo := NewOidcAuthProviderRepositoryMock(t)
	userAuthMethodRepo := NewUserAuthMethodRepositoryMock(t)
	userRepo := users.NewUserRepositoryMock(t)
	userResolver := NewOidcUserResolverMock(t)
	sessionService := users.NewSessionServiceMock(t)
	osWrapper := tools.NewCommonOsWrapperMock(t)
	return loginFlowTestObjects{
		Service: &LoginServiceImpl{
			ProviderRepo:       providerRepo,
			UserAuthMethodRepo: userAuthMethodRepo,
			UserRepo:           userRepo,
			UserResolver:       userResolver,
			SessionService:     sessionService,
			OsWrapper:          osWrapper,
		},
		ProviderRepo:       providerRepo,
		UserAuthMethodRepo: userAuthMethodRepo,
		UserRepo:           userRepo,
		UserResolver:       userResolver,
		SessionService:     sessionService,
		OsWrapper:          osWrapper,
	}
}

func (o loginFlowTestObjects) expectTrustedProvider() {
	o.ProviderRepo.EXPECT().GetProviderById(1).Return(&OidcAuthProviderDto{Id: 1, Name: "provider"}, nil)
}

func testCookie(userId int) *http.Cookie {
	return &http.Cookie{Value: "cookie-" + strconv.Itoa(userId)}
}
