package frontend_pages

import (
	"fmt"
	"strconv"
	"strings"

	"server/oidc_provider"
	"server/tools"

	"github.com/go-rod/rod"
	"github.com/quollix/common/assert"
)

type ClientsPage struct {
	Frame *FrameType
}

func (c *ClientsPage) FillCreateClient(client *oidc_provider.OidcRelyingPartyDto) *ClientsPage {
	c.Frame.Controls.SetInputValue("#oidc-client-name-input", client.Name)
	c.Frame.Controls.SetInputValue("#oidc-client-domain-input", client.Domain)
	return c
}

func (c *ClientsPage) CreateClientAndAssertSuccess() *ClientsPage {
	c.Frame.Controls.GetRequiredElement("#oidc-client-create-button").MustClick()
	c.Frame.Assert.SnackbarVisibleWithTextEventually("OIDC client created.")
	return c
}

func (c *ClientsPage) UpdateClient(client *oidc_provider.OidcRelyingPartyDto) *ClientsPage {
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

func (c *ClientsPage) ToggleClientSecretVisibility(name string) *ClientsPage {
	row := c.findRowByClientName(name)
	GetRequiredElementInRow(c.Frame.T, row, ".oidc-client-client-secret-toggle").MustClick()
	return c
}

func (c *ClientsPage) AssertClientSecretVisibility(name string, visible bool) *ClientsPage {
	expectedType := "password"
	if visible {
		expectedType = "text"
	}

	err := tools.Eventually(func() error {
		row := c.findRowByClientName(name)
		actualType := getInputTypeInRow(c.Frame.T, row, ".oidc-client-client-secret-value")
		if actualType != expectedType {
			return fmt.Errorf("unexpected client secret input type: %q", actualType)
		}
		return nil
	})
	assert.Nil(c.Frame.T, err)
	return c
}

func (c *ClientsPage) GetRequiredClient(name string) oidc_provider.OidcRelyingPartyDto {
	var client oidc_provider.OidcRelyingPartyDto
	err := tools.Eventually(func() error {
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

func (c *ClientsPage) readClientEntry(row *rod.Element) oidc_provider.OidcRelyingPartyDto {
	clientRecordId, err := row.Attribute("data-client-record-id")
	assert.Nil(c.Frame.T, err)
	assert.NotNil(c.Frame.T, clientRecordId)

	id, err := strconv.Atoi(strings.TrimSpace(*clientRecordId))
	assert.Nil(c.Frame.T, err)

	return oidc_provider.OidcRelyingPartyDto{
		Id:           id,
		Name:         getInputValueInRow(c.Frame.T, row, ".oidc-client-name-edit"),
		Domain:       getInputValueInRow(c.Frame.T, row, ".oidc-client-domain-edit"),
		ClientId:     strings.TrimSpace(GetRequiredElementInRow(c.Frame.T, row, ".oidc-client-client-id-value").MustText()),
		ClientSecret: getInputValueInRow(c.Frame.T, row, ".oidc-client-client-secret-value"),
	}
}

func (c *ClientsPage) findRowByClientName(name string) *rod.Element {
	var foundRow *rod.Element
	err := tools.Eventually(func() error {
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

func (c *ClientsPage) findRowByClientId(clientId int) *rod.Element {
	var foundRow *rod.Element
	expectedClientId := strconv.Itoa(clientId)
	err := tools.Eventually(func() error {
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
