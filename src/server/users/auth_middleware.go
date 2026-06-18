package users

import (
	"net/http"
	"server/tools"
	"strings"

	u "github.com/quollix/common/utils"
)

var AuthNotFoundInContextError = "auth not found in context, but protected handlers must always have an auth context"
var UnauthorizedError = "Unauthorized"

func (r *RouteRegisterer) withOptionalAuthContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		requestWithAuthContext, err := r.AuthService.GetRequestWithAuthContext(w, req)
		if err == nil {
			next.ServeHTTP(w, requestWithAuthContext)
			return
		}
		next.ServeHTTP(w, req)
	})
}

func (r *RouteRegisterer) requireAuthContext(path string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		requestWithAuthContext, err := r.AuthService.GetRequestWithAuthContext(w, req)
		if err != nil {
			r.writeAuthFailure(w, req, path, err)
			return
		}
		next.ServeHTTP(w, requestWithAuthContext)
	})
}

func (r *RouteRegisterer) requireAccessLevel(level tools.UserAccessLevel, path string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		auth, err := GetAuthFromContext(req)
		if err != nil {
			r.writeAuthFailure(w, req, path, err)
			return
		}
		if !HasAccess(level, *auth) {
			r.writeAuthFailure(w, req, path, u.Logger.NewError(UnauthorizedError))
			return
		}
		next.ServeHTTP(w, req)
	})
}

func HasAccess(level tools.UserAccessLevel, auth tools.User) bool {
	switch level {
	case tools.AnonymousLevel:
		return true
	case tools.UserLevel:
		return true
	default:
		return auth.IsAdmin
	}
}

func IsFrontendRequest(path string) bool {
	isBackendRequest := strings.HasPrefix(path, tools.Paths.BackendApi)
	isWebResourceRequestedFromBackend := strings.HasPrefix(path, tools.FrontendResourcesPathWithSlash)
	isFrontendRequest := !isBackendRequest && !isWebResourceRequestedFromBackend
	return isFrontendRequest
}

func GetAuthFromContext(r *http.Request) (*tools.User, error) {
	if r.Context() == nil {
		return nil, u.Logger.NewError("request context is nil")
	}

	val := r.Context().Value(tools.AuthKey)
	if val == nil {
		return nil, u.Logger.NewError(AuthNotFoundInContextError)
	}

	user, ok := val.(tools.User)
	if !ok {
		return nil, u.Logger.NewError("auth context value is of invalid type")
	}

	return &user, nil
}
