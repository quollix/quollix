package users

import (
	"net/http"
	"net/url"

	"github.com/quollix/common/quollix/api"
	u "github.com/quollix/common/utils"
)

var expectedAuthProtectionErrors = u.MapOf(
	CookieNotFoundError,
	UnauthorizedError,
	"Invalid input. The content of the field Cookie must be exactly 64 characters long. Allowed symbols are: a-f0-9.",
)

func (r *RouteRegisterer) writeAuthFailure(w http.ResponseWriter, req *http.Request, path string, err error) {
	isFrontendRoute := IsFrontendRequest(path)
	u.Logger.Debug("printing component addressed", "is_frontend_request", isFrontendRoute, "path", req.URL.Path)
	if isFrontendRoute {
		http.Redirect(w, req, BuildSignInURLForRequest(req), http.StatusFound) // #nosec G710 (CWE-601): Open redirect; redirect target is the local sign-in path, request URI is encoded into next and validated before use.
		return
	}
	u.WriteResponseError(w, expectedAuthProtectionErrors, err)
}

func BuildSignInURLForRequest(req *http.Request) string {
	signInURL := api.Paths.FrontendSignIn
	if requestURI := req.URL.RequestURI(); requestURI != "" {
		query := url.Values{}
		query.Set("next", requestURI)
		signInURL += "?" + query.Encode()
	}
	return signInURL
}
