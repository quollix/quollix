package apps_basic

import (
	"net/http"
	"net/http/httptest"
	"server/tools"
	"server/users"
	"testing"
	"time"

	"github.com/quollix/common/assert"
	u "github.com/quollix/common/utils"
	"github.com/stretchr/testify/mock"
)

func TestAppSessionService_AuthorizeAppRequestSkipsAuthForPublicApp(t *testing.T) {
	service := &AppSessionServiceImpl{}
	app := &AppRequestData{
		AppName:      "sample-app",
		AccessPolicy: tools.Policies.PublicAccessPolicy,
	}
	request := httptest.NewRequest(http.MethodGet, "https://sample-app.example.com", nil)

	err := service.AuthorizeAppRequest(request, app)
	assert.Nil(t, err)
}

func TestAppSessionService_AuthorizeAppRequestUsesAppAudience(t *testing.T) {
	sessionRepo := users.NewSessionRepositoryMock(t)
	authorizer := NewAuthorizerMock(t)
	authHelper := &u.AuthHelperImpl{}
	service := &AppSessionServiceImpl{
		UserService: &users.UserServiceImpl{
			SessionService: &users.SessionServiceImpl{
				SessionRepo: sessionRepo,
				AuthHelper:  authHelper,
			},
			AuthHelper: authHelper,
		},
		Authorizer: authorizer,
	}
	app := &AppRequestData{
		Maintainer:   "maintainer",
		AppName:      "sample-app",
		AccessPolicy: tools.Policies.AuthenticatedAccessPolicy,
	}
	request := httptest.NewRequest(http.MethodGet, "https://sample-app.example.com", nil)
	request.AddCookie(&http.Cookie{
		Name:  tools.BrandAppAuthCookieName,
		Value: "app-cookie-value",
	})
	hashedCookie := authHelper.GetSHA256Hash("app-cookie-value")
	sessionRepo.EXPECT().
		GetAuthenticatedSession(hashedCookie, users.SessionAudience(app.Maintainer, app.AppName)).
		Return(&users.AuthenticatedSession{
			User: tools.User{
				Id:        42,
				IsEnabled: true,
			},
			Session: users.UserSession{
				CookieExpirationDate: time.Now().Add(time.Hour),
			},
		}, nil)
	authorizer.EXPECT().
		Authorize(app.AccessPolicy, tools.UserLevel, 42, app.AppName).
		Return(nil)

	err := service.AuthorizeAppRequest(request, app)
	assert.Nil(t, err)
}

func TestAppSessionService_CreateAppSessionCookieFromSecretCreatesAppAudienceSessionCookie(t *testing.T) {
	sessionRepo := users.NewSessionRepositoryMock(t)
	authHelper := &u.AuthHelperImpl{}
	secretStorage := &users.SecretAndCookieStorageImpl{
		AuthHelper: authHelper,
	}
	service := &AppSessionServiceImpl{
		SecretAndCookieStorage: secretStorage,
		SessionService: &users.SessionServiceImpl{
			SessionRepo: sessionRepo,
			AuthHelper:  authHelper,
		},
	}
	app := &AppRequestData{
		Maintainer: "maintainer",
		AppName:    "sample-app",
	}
	quollixCookieValue := "quollix-cookie-value"
	secret, err := secretStorage.GenerateSecretForCookie(quollixCookieValue)
	assert.Nil(t, err)
	sessionRepo.EXPECT().
		GetAuthenticatedSession(authHelper.GetSHA256Hash(quollixCookieValue), users.QuollixSessionAudience()).
		Return(&users.AuthenticatedSession{
			User: tools.User{
				Id:        42,
				IsEnabled: true,
			},
			Session: users.UserSession{
				CookieExpirationDate: time.Now().Add(time.Hour),
			},
		}, nil)
	sessionRepo.EXPECT().
		CreateSession(mock.MatchedBy(func(session *users.UserSession) bool {
			return session.UserId == 42 &&
				session.Audience == users.SessionAudience(app.Maintainer, app.AppName) &&
				session.HashedCookieValue != "" &&
				session.CookieExpirationDate.After(time.Now().UTC())
		})).
		Return(7, nil)

	cookie, err := service.CreateAppSessionCookieFromSecret(secret, app)
	assert.Nil(t, err)
	assert.Equal(t, tools.BrandAppAuthCookieName, cookie.Name)
	assert.NotEqual(t, quollixCookieValue, cookie.Value)
	assert.True(t, cookie.Secure)
	assert.Equal(t, http.SameSiteLaxMode, cookie.SameSite)
}
