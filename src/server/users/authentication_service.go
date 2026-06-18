package users

import (
	"context"
	"net/http"
	"server/tools"
	"time"

	u "github.com/quollix/common/utils"
	"github.com/quollix/common/validation"
)

const (
	UserNotFoundError   = "user not found"
	CookieNotFoundError = "cookie not found"
	CookieExpiredError  = "cookie expired"
)

type AuthenticationService interface {
	GetRequestWithAuthContext(w http.ResponseWriter, r *http.Request) (*http.Request, error)
}

type AuthenticationServiceImpl struct {
	SessionService SessionService
}

func (a *AuthenticationServiceImpl) GetRequestWithAuthContext(w http.ResponseWriter, r *http.Request) (*http.Request, error) {
	auth, err := a.getUserAuthentication(r)
	if err != nil {
		return nil, err
	}

	// This simply adds context information to the request object for further security processing in the backend. The context information is not used in any http request that is proxied to an application.
	ctx := context.WithValue(r.Context(), tools.AuthKey, *auth)
	r = r.WithContext(ctx)

	return r, nil
}

func (a *AuthenticationServiceImpl) getUserAuthentication(r *http.Request) (*tools.User, error) {
	cookie, err := r.Cookie(tools.BrandAppAuthCookieName)
	if err != nil {
		return nil, u.Logger.NewError(CookieNotFoundError)
	}

	err = validation.Validate("Cookie", validation.FieldSecret, cookie.Value)
	if err != nil {
		return nil, err
	}

	authenticatedSession, err := a.SessionService.GetAuthenticatedSession(cookie.Value, QuollixSessionAudience())
	if err != nil {
		return nil, u.Logger.NewError(CookieNotFoundError)
	}

	if authenticatedSession.Session.CookieExpirationDate.Before(time.Now().UTC()) {
		return nil, u.Logger.NewError(CookieExpiredError)
	}

	renewedExpirationDate, err := a.SessionService.UpdateCookieExpirationDate(cookie.Value, QuollixSessionAudience())
	if err != nil {
		return nil, err
	}

	authenticatedSession.User.CookieExpirationDate = renewedExpirationDate
	return &authenticatedSession.User, nil
}
