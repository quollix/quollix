//go:build integration

package repository

import (
	"testing"

	"server/oidc_provider"
	test_tools "server/tests/test_tools"

	"github.com/quollix/common/assert"
	u "github.com/quollix/common/utils"
)

func TestOidcRelyingPartyRepository_CRUD(t *testing.T) {
	InitDeps()
	defer cleanupOidcRelyingPartyRepoTest()

	client := test_tools.GetSampleOidcRelyingParty()
	var err error
	client.Id, err = OidcRelyingPartyRepo.CreateClient(client)
	assert.Nil(t, err)

	actualClient, err := OidcRelyingPartyRepo.GetClientById(client.Id)
	assert.Nil(t, err)
	assertOidcRelyingPartyEquality(t, client, actualClient)

	actualClient, found, err := OidcRelyingPartyRepo.GetClientByClientId(client.ClientId)
	assert.Nil(t, err)
	assert.True(t, found)
	assertOidcRelyingPartyEquality(t, client, actualClient)

	actualClient, found, err = OidcRelyingPartyRepo.GetClientByName(client.Name)
	assert.Nil(t, err)
	assert.True(t, found)
	assertOidcRelyingPartyEquality(t, client, actualClient)

	clients, err := OidcRelyingPartyRepo.ListClients()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(clients))
	assertOidcRelyingPartyEquality(t, client, &clients[0])

	updatedClient := test_tools.GetUpdatedSampleOidcRelyingParty()
	updatedClient.Id = client.Id
	client = updatedClient
	assert.Nil(t, OidcRelyingPartyRepo.UpdateClient(client))
	actualClient, err = OidcRelyingPartyRepo.GetClientById(client.Id)
	assert.Nil(t, err)
	assertOidcRelyingPartyEquality(t, client, actualClient)

	assert.Nil(t, OidcRelyingPartyRepo.DeleteClient(client.Id))
	clients, err = OidcRelyingPartyRepo.ListClients()
	assert.Nil(t, err)
	assert.Equal(t, 0, len(clients))
}

func TestOidcRelyingPartyRepository_DuplicateNameReturnsError(t *testing.T) {
	InitDeps()
	defer cleanupOidcRelyingPartyRepoTest()

	client := test_tools.GetSampleOidcRelyingParty()
	_, err := OidcRelyingPartyRepo.CreateClient(client)
	assert.Nil(t, err)

	client.ClientId = "other-client-id"
	_, err = OidcRelyingPartyRepo.CreateClient(client)
	assert.NotNil(t, err)
}

func TestOidcRelyingPartyRepository_DuplicateClientIdReturnsError(t *testing.T) {
	InitDeps()
	defer cleanupOidcRelyingPartyRepoTest()

	client := test_tools.GetSampleOidcRelyingParty()
	_, err := OidcRelyingPartyRepo.CreateClient(client)
	assert.Nil(t, err)

	client.Name = "Other-Client"
	_, err = OidcRelyingPartyRepo.CreateClient(client)
	assert.NotNil(t, err)
}

func TestOidcRelyingPartyRepository_UpdateClient_WhenClientDoesNotExistReturnsError(t *testing.T) {
	InitDeps()
	defer cleanupOidcRelyingPartyRepoTest()

	client := test_tools.GetSampleOidcRelyingParty()
	client.Id = 987654

	err := OidcRelyingPartyRepo.UpdateClient(client)

	assert.Equal(t, oidc_provider.OidcRelyingPartyNotFoundError, u.ExtractError(err))
}

func TestOidcRelyingPartyRepository_GetClientByName_WhenClientDoesNotExistReturnsFalse(t *testing.T) {
	InitDeps()
	defer cleanupOidcRelyingPartyRepoTest()

	client, found, err := OidcRelyingPartyRepo.GetClientByName("missing-client")

	assert.Nil(t, err)
	assert.False(t, found)
	assert.Nil(t, client)
}

func TestOidcRelyingPartyRepository_GetClientByClientId_WhenClientDoesNotExistReturnsFalse(t *testing.T) {
	InitDeps()
	defer cleanupOidcRelyingPartyRepoTest()

	client, found, err := OidcRelyingPartyRepo.GetClientByClientId("missing-client-id")

	assert.Nil(t, err)
	assert.False(t, found)
	assert.Nil(t, client)
}

func cleanupOidcRelyingPartyRepoTest() {
	OidcRelyingPartyRepo.Wipe()
}

func assertOidcRelyingPartyEquality(t *testing.T, expected, actual *oidc_provider.OidcRelyingPartyDto) {
	assert.Equal(t, expected.Id, actual.Id)
	assert.Equal(t, expected.Name, actual.Name)
	assert.Equal(t, expected.Domain, actual.Domain)
	assert.Equal(t, expected.ClientId, actual.ClientId)
	assert.Equal(t, expected.ClientSecret, actual.ClientSecret)
}
