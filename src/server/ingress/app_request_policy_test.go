package ingress

import (
	"testing"

	"github.com/quollix/common/assert"
	"github.com/quollix/deepstack"
)

func TestIsRequestAddressedToAnApp(t *testing.T) {
	service := AppRequestPolicyImpl{}
	assert.False(t, service.IsRequestAddressedToAnApp("quollix.somedomain.org", ""))
	assert.False(t, service.IsRequestAddressedToAnApp("somedomain.org", ""))
	assert.False(t, service.IsRequestAddressedToAnApp("some-app.somedomain.org", ""))
	assert.False(t, service.IsRequestAddressedToAnApp("some-app.somedomain.org", "127.0.0.1"))
	assert.False(t, service.IsRequestAddressedToAnApp("127.0.0.1", "127.0.0.1"))
	assert.False(t, service.IsRequestAddressedToAnApp("127.0.0.1", "somedomain.org"))
	assert.False(t, service.IsRequestAddressedToAnApp("quollix.somedomain.org", "somedomain.org"))

	assert.True(t, service.IsRequestAddressedToAnApp("some-app.somedomain.org", "somedomain.org"))

}

func TestIsOriginAllowed(t *testing.T) {
	tests := []struct {
		name        string
		requestHost string
		origin      string
		serverHost  string
		errorMsg    string
	}{
		// actually they are redirected to quollix
		{"NoHostSetAppRequest", "app.example.com", "app.example.com", "", ""},
		{"OtherHostSetAppRequest", "app.example.com", "app.example.com", "other.com", ""},

		{"PlatformOriginToAppDenied", "app.example.com", "example.com", "example.com", crossRequestsToAppsOnlyFromBrandAppOriginErrorMessage},
		{"ValidBrandAppOrigin", "app.example.com", "quollix.example.com", "example.com", ""},
		{"InvalidAppOrigin", "app.example.com", "other.com", "example.com", crossRequestsToAppsOnlyFromBrandAppOriginErrorMessage},

		{"BrandAppHostValid", "quollix.example.com", "quollix.example.com", "example.com", ""},
		{"BrandAppHostWithCrossAppOrigin", "quollix.example.com", "example.com", "example.com", CrossRequestsToBrandAppNotAllowedErrorMessage},
		{"BrandAppHostEmptyServerAndOrigin", "quollix.example.com", "", "", ""},
		{"BrandAppHostEmptyOrigin", "quollix.example.com", "", "example.com", ""},
		{"BrandAppHostInvalidOrigin", "quollix.example.com", "somewhere.else", "example.com", CrossRequestsToBrandAppNotAllowedErrorMessage},

		{"PlatformValid", "example.com", "example.com", "example.com", ""},
		{"PlatformNoServerNoOrigin", "example.com", "", "", ""},
		{"PlatformNoOriginServerSet", "example.com", "", "example.com", ""},
		{"PlatformWithBrandAppOrigin", "example.com", "quollix.example.com", "example.com", CrossRequestsToBrandAppNotAllowedErrorMessage},
		{"PlatformWithRandomOrigin", "example.com", "somewhere.else", "example.com", CrossRequestsToBrandAppNotAllowedErrorMessage},
	}

	service := AppRequestPolicyImpl{}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidateRequestOrigin(tt.requestHost, tt.origin, tt.serverHost)
			if tt.errorMsg == "" {
				assert.Nil(t, err)
			} else {
				deepStackError, ok := err.(*deepstack.DeepStackError)
				assert.True(t, ok)
				assert.Equal(t, tt.errorMsg, deepStackError.Message)
			}
		})
	}
}

func TestIsCrossRequest(t *testing.T) {
	service := AppRequestPolicyImpl{}
	assert.False(t, service.isCrossOriginRequest("localhost", "localhost"))
	assert.False(t, service.isCrossOriginRequest("localhost", ""))
	assert.True(t, service.isCrossOriginRequest("localhost", "localhost2"))
}
