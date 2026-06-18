//go:build component

package component

import (
	"crypto/tls"
	"net/http"
	"server/tools"
	"strings"
	"testing"

	"github.com/quollix/common/assert"
	u "github.com/quollix/common/utils"
)

func TestEmailFrontendEndpointIsAvailable(t *testing.T) {
	client := GetQuollixClient(t)
	assert.Nil(t, client.Auth.Login("admin", "password"))
	defer client.Test.ResetTestState()

	req, err := http.NewRequest(http.MethodGet, client.Parent.RootUrl+tools.Paths.FrontendEmail, nil)
	assert.Nil(t, err)
	req.AddCookie(client.Parent.Cookie)

	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	resp, err := httpClient.Do(req)
	assert.Nil(t, err)
	defer u.Close(resp.Body)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.True(t, strings.Contains(resp.Header.Get("Content-Type"), "text/html"))
}
