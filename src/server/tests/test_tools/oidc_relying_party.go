package test_tools

import (
	"testing"

	"github.com/quollix/common/assert"
	"github.com/quollix/common/quollix/api"
)

const (
	OidcClientIDLength     = 16
	OidcClientSecretLength = 64
)

func GetSampleOidcRelyingParty() *api.OidcRelyingPartyDto {
	return &api.OidcRelyingPartyDto{
		Name:         "Corporate-Client",
		Domain:       "client.example.com",
		ClientId:     "client-id",
		ClientSecret: "client-secret",
	}
}

func GetUpdatedSampleOidcRelyingParty() *api.OidcRelyingPartyDto {
	return &api.OidcRelyingPartyDto{
		Name:         "Updated-Client",
		Domain:       "updated-client.example.com",
		ClientId:     "updated-client-id",
		ClientSecret: "updated-client-secret",
	}
}

func AssertGeneratedOidcCredentials(t *testing.T, clientID string, clientSecret string) {
	assert.Equal(t, OidcClientIDLength, len(clientID))
	assert.Equal(t, OidcClientSecretLength, len(clientSecret))
}

func AssertOidcCredentialsChanged(t *testing.T, previousClientID string, previousClientSecret string, currentClientID string, currentClientSecret string) {
	assert.NotEqual(t, previousClientID, currentClientID)
	assert.NotEqual(t, previousClientSecret, currentClientSecret)
}
