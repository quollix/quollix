package users

import (
	"net/http"
	"server/tools"

	"github.com/go-chi/chi/v5"
)

type Route struct {
	Path        string
	HandlerFunc http.HandlerFunc
	AccessLevel tools.UserAccessLevel
}

type RouteRegisterer struct {
	AuthService AuthenticationService
	Router      chi.Router
}

func (r *RouteRegisterer) RegisterRoutes(routes []Route) {
	for _, route := range routes {
		r.Router.Handle(route.Path, r.buildProtectedHandler(route))
	}
}

func (r *RouteRegisterer) buildProtectedHandler(route Route) http.Handler {
	handler := http.Handler(route.HandlerFunc)

	if route.AccessLevel == tools.AnonymousLevel {
		handler = r.withOptionalAuthContext(handler)
	} else {
		handler = r.requireAccessLevel(route.AccessLevel, route.Path, handler)
		handler = r.requireAuthContext(route.Path, handler)
	}

	return handler
}
