//go:build component

package component

import (
	"net/http"
	"testing"

	"server/tools"

	"github.com/quollix/common/assert"
	u "github.com/quollix/common/utils"
)

func TestDefaultDeploymentRedirectsHttpAndForwardsHttpsProto(t *testing.T) {
	httpClient := &http.Client{
		CheckRedirect: func(request *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	httpRequest, err := http.NewRequest(http.MethodGet, "http://localhost/some/path?value=1", nil)
	assert.Nil(t, err)
	httpResponse, err := httpClient.Do(httpRequest)
	assert.Nil(t, err)
	defer u.Close(httpResponse.Body)
	assert.Equal(t, http.StatusPermanentRedirect, httpResponse.StatusCode)
	expectedLocation := "https://localhost/some/path?value=1"
	actualLocation := httpResponse.Header.Get("Location")
	assert.Equal(t, expectedLocation, actualLocation)
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()
	InstallAndStartSample(t, client, "2.0")
	appClient := GetAppClient(t, client)
	proto, err := ReadSampleAppHeaderValue(appClient, sampleAppHttpsUrl, "X-Forwarded-Proto")
	assert.Nil(t, err)
	assert.Equal(t, tools.AppForwardedProtoHttps, proto)
}
