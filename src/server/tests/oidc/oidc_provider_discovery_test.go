//go:build oidc

package oidc

import (
	"testing"

	"github.com/quollix/common/assert"
)

func TestOidcAuthProviderDiscoveryHealthcheckBetweenTwoQuollixInstances(t *testing.T) {
	clients := SetupAndGetClients(t)
	defer clients.Reset(t)

	err := clients.ClientAdmin.OidcProviders.TestDiscovery(ProviderDomain)

	assert.Nil(t, err)
}
