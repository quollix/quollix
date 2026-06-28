package apps_basic

import (
	"net/http"
	"server/tools"
	"server/users"
	"time"

	u "github.com/quollix/common/utils"
)

type AppSessionService interface {
	AuthorizeAppRequest(r *http.Request, app *AppRequestData) error
	CreateAppSessionCookieFromSecret(urlSecret string, app *AppRequestData) (*http.Cookie, error)
}

type AppSessionServiceImpl struct {
	UserService            users.UserService
	SessionService         users.SessionService
	SecretAndCookieStorage users.SecretAndCookieStorage
	Authorizer             Authorizer
}

func (a *AppSessionServiceImpl) AuthorizeAppRequest(r *http.Request, app *AppRequestData) error {
	if app.AccessPolicy == tools.Policies.PublicAccessPolicy {
		return nil
	}

	userId, role, err := a.UserService.GetUserIdAndRoleFromRequestForAudience(
		r,
		users.SessionAudience(app.Maintainer, app.AppName),
	)
	if err != nil {
		return err
	}

	return a.Authorizer.Authorize(app.AccessPolicy, role, userId, app.AppName)
}

func (a *AppSessionServiceImpl) CreateAppSessionCookieFromSecret(urlSecret string, app *AppRequestData) (*http.Cookie, error) {
	cookieValue, err := a.SecretAndCookieStorage.LoadCookieViaSecret(urlSecret)
	if err != nil {
		return nil, err
	}

	authenticatedSession, err := a.SessionService.GetAuthenticatedSession(cookieValue, users.QuollixSessionAudience())
	if err != nil {
		return nil, err
	}
	if authenticatedSession.Session.CookieExpirationDate.Before(time.Now().UTC()) {
		return nil, u.Logger.NewError(users.CookieExpiredError)
	}

	return a.SessionService.GenerateAndSaveCookie(
		authenticatedSession.User.Id,
		users.SessionAudience(app.Maintainer, app.AppName),
	)
}
