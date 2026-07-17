package users

import (
	"server/tools"
	"testing"

	"github.com/quollix/common/assert"
	"github.com/quollix/common/quollix/api"
)

func TestHasAccess(t *testing.T) {
	tests := []struct {
		name     string
		level    tools.UserAccessLevel
		user     api.User
		expected bool
	}{
		{"anonymous user", tools.AnonymousLevel, api.User{}, true},
		{"anonymous admin", tools.AnonymousLevel, api.User{IsAdmin: true}, true},
		{"user regular", tools.UserLevel, api.User{IsAdmin: false}, true},
		{"user admin", tools.UserLevel, api.User{IsAdmin: true}, true},
		{"admin regular", tools.AdminLevel, api.User{IsAdmin: false}, false},
		{"admin admin", tools.AdminLevel, api.User{IsAdmin: true}, true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, HasAccess(tt.level, tt.user), tt.expected)
		})
	}
}

func TestIsFrontendRequest(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{name: "frontend root", path: "/", expected: true},
		{name: "frontend page", path: "/store", expected: true},
		{name: "frontend nested page", path: "/users/edit", expected: true},
		{name: "backend api root", path: api.Paths.BackendApi, expected: false},
		{name: "backend api endpoint", path: api.Paths.BackendSecret, expected: false},
		{name: "backend api nested endpoint", path: api.Paths.BackendApps + "/123", expected: false},
		{name: "frontend static resource root", path: tools.FrontendResourcesPathWithSlash, expected: false},
		{name: "frontend static resource file", path: tools.FrontendResourcesPathWithSlash + "pages/sign-in/sign-in.js", expected: false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, IsFrontendRequest(tt.path), tt.expected)
		})
	}
}
