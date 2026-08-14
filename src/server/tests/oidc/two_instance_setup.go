package oidc

import (
	"testing"

	"github.com/quollix/common/quollix/test_environment"

	"github.com/quollix/common/assert"
)

func SetupAndGetClients(t *testing.T) *test_environment.OidcTwoInstanceClients {
	clients := test_environment.NewOidcTwoInstanceClients()
	assert.Nil(t, clients.Configure())
	return clients
}
