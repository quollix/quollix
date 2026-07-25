package frontend_pages

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/quollix/common/assert"
	"github.com/quollix/common/browsertest"
	"github.com/quollix/common/quollix/api"
	u "github.com/quollix/common/utils"
)

type ClientsPage struct {
	Frame *FrameType
}

func (c *ClientsPage) FillCreateClient(client *api.OidcRelyingPartyDto) *ClientsPage {
	c.Frame.Controls.SetInputValue("#oidc-client-name-input", client.Name)
	c.Frame.Controls.SetInputValue("#oidc-client-domain-input", client.Domain)
	return c
}

func (c *ClientsPage) CreateClientAndAssertSuccess() *ClientsPage {
	c.Frame.Controls.GetRequiredElement("#oidc-client-create-button").MustClick()
	c.Frame.Assert.SnackbarVisibleWithTextEventually("OIDC client created.")
	return c
}

func (c *ClientsPage) UpdateClient(client *api.OidcRelyingPartyDto) *ClientsPage {
	row := c.findRowByClientId(client.Id)
	setInputValueInRow(c.Frame.T, row, ".oidc-client-name-edit", client.Name)
	setInputValueInRow(c.Frame.T, row, ".oidc-client-domain-edit", client.Domain)
	GetRequiredElementInRow(c.Frame.T, row, ".oidc-client-save-button").MustClick()
	c.Frame.Assert.SnackbarVisibleWithTextEventually("OIDC client saved.")
	return c
}

func (c *ClientsPage) RegenerateCredentials(clientId int) *ClientsPage {
	row := c.findRowByClientId(clientId)
	GetRequiredElementInRow(c.Frame.T, row, ".oidc-client-regenerate-button").MustClick()
	c.Frame.Browser.ConfirmDialog()
	c.Frame.Assert.SnackbarVisibleWithTextEventually("Credentials regenerated successfully.")
	return c
}

func (c *ClientsPage) DeleteClient(clientId int) *ClientsPage {
	row := c.findRowByClientId(clientId)
	GetRequiredElementInRow(c.Frame.T, row, ".oidc-client-delete-button").MustClick()
	c.Frame.Browser.ConfirmDialog()
	c.Frame.Assert.SnackbarVisibleWithTextEventually("OIDC client deleted.")
	return c
}

func (c *ClientsPage) AssertClientSecretMasked(name string) *ClientsPage {
	err := u.Eventually(func() error {
		row := c.findRowByClientName(name)
		actualText := strings.TrimSpace(GetRequiredElementInRow(c.Frame.T, row, ".oidc-client-client-secret-value").MustText())
		if actualText != "****************" {
			return fmt.Errorf("unexpected client secret display value: %q", actualText)
		}
		return nil
	})
	assert.Nil(c.Frame.T, err)
	return c
}

func (c *ClientsPage) GetRequiredClient(name string) api.OidcRelyingPartyDto {
	var client api.OidcRelyingPartyDto
	err := u.Eventually(func() error {
		rows := c.Frame.Page.MustElements("tr.oidc-relying-party-row")
		for _, row := range rows {
			entry := c.readClientEntry(row)
			if entry.Name == name {
				client = entry
				return nil
			}
		}
		return fmt.Errorf("client not found: %s", name)
	})
	assert.Nil(c.Frame.T, err)
	return client
}

func (c *ClientsPage) readClientEntry(row *browsertest.Element) api.OidcRelyingPartyDto {
	clientRecordId, err := row.Attribute("data-client-record-id")
	assert.Nil(c.Frame.T, err)
	assert.NotNil(c.Frame.T, clientRecordId)

	id, err := strconv.Atoi(strings.TrimSpace(*clientRecordId))
	assert.Nil(c.Frame.T, err)

	return api.OidcRelyingPartyDto{
		Id:       id,
		Name:     getInputValueInRow(c.Frame.T, row, ".oidc-client-name-edit"),
		Domain:   getInputValueInRow(c.Frame.T, row, ".oidc-client-domain-edit"),
		ClientId: strings.TrimSpace(GetRequiredElementInRow(c.Frame.T, row, ".oidc-client-client-id-value").MustText()),
	}
}

func (c *ClientsPage) findRowByClientName(name string) *browsertest.Element {
	var foundRow *browsertest.Element
	err := u.Eventually(func() error {
		rows := c.Frame.Page.MustElements("tr.oidc-relying-party-row")
		for _, row := range rows {
			if getInputValueInRow(c.Frame.T, row, ".oidc-client-name-edit") == name {
				foundRow = row
				return nil
			}
		}
		return fmt.Errorf("client row not found: %s", name)
	})
	assert.Nil(c.Frame.T, err)
	return foundRow
}

func (c *ClientsPage) findRowByClientId(clientId int) *browsertest.Element {
	var foundRow *browsertest.Element
	expectedClientId := strconv.Itoa(clientId)
	err := u.Eventually(func() error {
		rows := c.Frame.Page.MustElements("tr.oidc-relying-party-row")
		for _, row := range rows {
			clientRecordId, err := row.Attribute("data-client-record-id")
			assert.Nil(c.Frame.T, err)
			if clientRecordId != nil && *clientRecordId == expectedClientId {
				foundRow = row
				return nil
			}
		}
		return fmt.Errorf("client row not found: %s", expectedClientId)
	})
	assert.Nil(c.Frame.T, err)
	return foundRow
}
