package apps_basic

import (
	"net/http"
	"net/url"
	"server/tools"
	"testing"

	"github.com/quollix/common/assert"
	u "github.com/quollix/common/utils"
)

func TestAppRequestParser_GetHostFromRequestHost(t *testing.T) {
	parser := &AppRequestParserImpl{}

	assert.Equal(t, "example.com:8080", parser.GetHostFromRequestHost("example.com:8080"))
	assert.Equal(t, "localhost", parser.GetHostFromRequestHost("localhost:80"))
	assert.Equal(t, "localhost", parser.GetHostFromRequestHost("localhost:443"))
	assert.Equal(t, "example.com", parser.GetHostFromRequestHost("example.com"))
	assert.Equal(t, "", parser.GetHostFromRequestHost(""))
}

func TestAppRequestParser_GetQuerySecret(t *testing.T) {
	parser := &AppRequestParserImpl{}
	sampleSecret := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	tests := []struct {
		name        string
		queryParams url.Values
		expected    string
		isPresent   bool
		isValid     bool
	}{
		{"secret present", url.Values{tools.BrandAppQuerySecretName: {sampleSecret}}, sampleSecret, true, true},
		{"secret missing", url.Values{}, "", false, true},
		{"secret empty", url.Values{tools.BrandAppQuerySecretName: {""}}, "", false, true},
		{"invalid secret value", url.Values{tools.BrandAppQuerySecretName: {"invalid"}}, "", false, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := &http.Request{URL: &url.URL{RawQuery: tc.queryParams.Encode()}}
			secret, ok, isValid := parser.GetQuerySecret(r)
			assert.Equal(t, tc.expected, secret)
			assert.Equal(t, tc.isPresent, ok)
			assert.Equal(t, tc.isValid, isValid)
		})
	}
}

func TestAppRequestParser_GetAppNameFromRequestHost(t *testing.T) {
	parser := &AppRequestParserImpl{}

	appName, err := parser.GetAppNameFromRequestHost("gitea.localhost", "localhost")
	assert.Nil(t, err)
	assert.Equal(t, "gitea", appName)

	_, err = parser.GetAppNameFromRequestHost("gitea.localhost", "localhost2")
	assert.NotNil(t, err)
	assert.Equal(t, HostDoesNotEndWithDatabaseHostError, u.ExtractError(err))
}
