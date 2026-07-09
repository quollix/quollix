//go:build behind_proxy

package behind_proxy

import (
	"testing"

	"server/tests/component"
	"server/tools"

	"github.com/quollix/common/assert"
)

func TestBehindProxyDeploymentAllowsHttpAndForwardsHttpProto(t *testing.T) {
	client := component.GetClientAndLogin(t)
	component.InstallAndStartSample(t, client, "2.0")
	secret, err := client.AppAccess.GetSecret()
	assert.Nil(t, err)
	assert.Nil(t, component.ExchangeAppAccessSecretForCookieWithUrl(client, secret, "http://sampleapp.localhost/"))
	proto, err := component.ReadSampleAppHeaderValue(client, "http://sampleapp.localhost", "X-Forwarded-Proto")
	assert.Nil(t, err)
	assert.Equal(t, tools.AppForwardedProtoHttp, proto)
}
