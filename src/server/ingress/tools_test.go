package ingress

import (
	"testing"

	"github.com/quollix/common/assert"
)

func TestGetHostFromOriginHeaderValue(t *testing.T) {
	certificateTools := &CertificateToolsImpl{}

	tests := []struct {
		name         string
		input        string
		expectedHost string
		expectErr    bool
	}{
		{"empty", "", "", false},
		{"null", "null", "", false},
		{"valid", "https://example.com", "example.com", false},
		{"valid", "http://example.com", "example.com", false},
		{"invalid", "example.com", "", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			host, err := certificateTools.GetHostFromOriginHeaderValue(tc.input)
			if tc.expectErr {
				assert.NotNil(t, err)
				assert.Equal(t, "", host)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tc.expectedHost, host)
			}
		})
	}
}
