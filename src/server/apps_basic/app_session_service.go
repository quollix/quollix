package apps_basic

import (
	"net/http"
	"server/tools"
	"server/users"
	"time"

	api "github.com/quollix/common/quollix/api"
	u "github.com/quollix/common/utils"
)

type AppSessionService interface {
	AuthorizeAppRequest(r *http.Request, app *AppRequestData) (AppRequestAuthorizationStatus, error)
	CreateAppSessionCookieFromSecret(urlSecret string, app *AppRequestData) (*http.Cookie, error)
}

type AppRequestAuthorizationStatus int

const (
	AppRequestAuthorizationUnknown AppRequestAuthorizationStatus = iota
	AppRequestAuthorized
	AppRequestMissingSession
	AppRequestAccessDenied
)

type AppSessionServiceImpl struct {
	UserService            users.UserService
	SessionService         users.SessionService
	SecretAndCookieStorage users.SecretAndCookieStorage
	Authorizer             Authorizer
}

func (a *AppSessionServiceImpl) AuthorizeAppRequest(r *http.Request, app *AppRequestData) (AppRequestAuthorizationStatus, error) {
	if app.AccessPolicy == api.Policies.PublicAccessPolicy {
		return AppRequestAuthorized, nil
	}

	userId, role, err := a.UserService.GetUserIdAndRoleFromRequestForAudience(
		r,
		users.SessionAudience(app.Maintainer, app.AppName),
	)
	if err != nil {
		return AppRequestAuthorizationUnknown, err
	}
	if role == tools.AnonymousLevel {
		return AppRequestMissingSession, nil
	}

	err = a.Authorizer.Authorize(app.AccessPolicy, role, userId, app.AppName)
	if err == nil {
		return AppRequestAuthorized, nil
	}
	if u.ExtractError(err) == AccessDeniedError {
		return AppRequestAccessDenied, nil
	}
	return AppRequestAuthorizationUnknown, err
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
