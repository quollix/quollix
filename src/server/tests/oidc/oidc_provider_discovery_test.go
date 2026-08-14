//go:build oidc

package oidc

import (
	"testing"

	"github.com/quollix/common/assert"
	"github.com/quollix/common/quollix/test_environment"
)

func TestOidcAuthProviderDiscoveryHealthcheckBetweenTwoQuollixInstances(t *testing.T) {
	clients := SetupAndGetClients(t)
	defer clients.Reset()

	err := clients.ClientAdmin.OidcProviders.TestDiscovery(test_environment.OidcProviderDomain)

	assert.Nil(t, err)
}
