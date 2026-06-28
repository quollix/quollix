//go:build acceptance

package acceptance

import (
	"testing"

	"server/tests/component"
	"server/tests/frontend_pages"
	test_tools "server/tests/test_tools"

	"github.com/quollix/common/assert"
)

func TestClientsPageClientCrud(t *testing.T) {
	frame := frontend_pages.Setup(t)
	defer frame.Client.Test.ResetTestState()

	page := frame.Pages.OpenClientsPage()

	client := test_tools.GetSampleOidcRelyingParty()
	page.
		FillCreateClient(client).
		CreateClientAndAssertSuccess().
		AssertClientSecretVisibility(client.Name, false).
		ToggleClientSecretVisibility(client.Name).
		AssertClientSecretVisibility(client.Name, true).
		ToggleClientSecretVisibility(client.Name).
		AssertClientSecretVisibility(client.Name, false)
	createdClient := component.GetOnlyOidcRelyingParty(t, frame.Client)
	assert.Equal(t, createdClient, page.GetRequiredClient(client.Name))
	assert.Equal(t, client.Name, createdClient.Name)
	assert.Equal(t, client.Domain, createdClient.Domain)
	test_tools.AssertGeneratedOidcCredentials(t, createdClient.ClientId, createdClient.ClientSecret)

	updatedClient := test_tools.GetUpdatedSampleOidcRelyingParty()
	updatedClient.Id = createdClient.Id
	page.UpdateClient(updatedClient)
	actualUpdatedClient := component.GetOnlyOidcRelyingParty(t, frame.Client)
	assert.Equal(t, updatedClient.Id, actualUpdatedClient.Id)
	assert.Equal(t, updatedClient.Name, actualUpdatedClient.Name)
	assert.Equal(t, updatedClient.Domain, actualUpdatedClient.Domain)
	assert.Equal(t, createdClient.ClientId, actualUpdatedClient.ClientId)
	assert.Equal(t, createdClient.ClientSecret, actualUpdatedClient.ClientSecret)

	page.RegenerateCredentials(actualUpdatedClient.Id)
	regeneratedClient := component.GetOnlyOidcRelyingParty(t, frame.Client)
	assert.Equal(t, actualUpdatedClient.Id, regeneratedClient.Id)
	assert.Equal(t, actualUpdatedClient.Name, regeneratedClient.Name)
	assert.Equal(t, actualUpdatedClient.Domain, regeneratedClient.Domain)
	test_tools.AssertGeneratedOidcCredentials(t, regeneratedClient.ClientId, regeneratedClient.ClientSecret)
	test_tools.AssertOidcCredentialsChanged(t, actualUpdatedClient.ClientId, actualUpdatedClient.ClientSecret, regeneratedClient.ClientId, regeneratedClient.ClientSecret)

	page.DeleteClient(regeneratedClient.Id)
	clients, err := frame.Client.OidcClients.List()
	assert.Nil(t, err)
	assert.Equal(t, 0, len(clients))
}
