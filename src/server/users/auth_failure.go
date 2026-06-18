package users

import (
	"net/http"

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
		http.Redirect(w, req, "/login", http.StatusFound)
		return
	}
	u.WriteResponseError(w, expectedAuthProtectionErrors, err)
}
