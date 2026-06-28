package frontend_pages

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"server/oidc_client"
	"server/tools"

	"github.com/go-rod/rod"
	"github.com/quollix/common/assert"
)

type ProvidersPage struct {
	Frame *FrameType
}

func (e *ProvidersPage) FillCreateProvider(provider *oidc_client.OidcAuthProviderDto) *ProvidersPage {
	e.Frame.Controls.SetInputValue("#oidc-provider-name-input", provider.Name)
	e.Frame.Controls.SetInputValue("#oidc-provider-issuer-domain-path-input", provider.IssuerDomainPath)
	e.Frame.Controls.SetInputValue("#oidc-provider-client-id-input", provider.ClientId)
	e.Frame.Controls.SetInputValue("#oidc-provider-client-secret-input", provider.ClientSecret)
	return e
}

func (e *ProvidersPage) CreateProviderAndAssertSuccess() *ProvidersPage {
	e.Frame.Controls.GetRequiredElement("#oidc-provider-create-button").MustClick()
	e.Frame.Assert.SnackbarVisibleWithTextEventually("Provider created.")
	return e
}

func (e *ProvidersPage) TestCreateProviderDiscoveryAndAssertSuccess() *ProvidersPage {
	e.Frame.Controls.GetRequiredElement("#oidc-provider-test-discovery-button").MustClick()
	e.Frame.Assert.SnackbarVisibleWithTextEventually("Discovery endpoint is available and seems valid.")
	return e
}

func (e *ProvidersPage) UpdateProvider(provider *oidc_client.OidcAuthProviderDto) *ProvidersPage {
	row := e.findRowByProviderId(provider.Id)
	setInputValueInRow(e.Frame.T, row, ".oidc-provider-name-edit", provider.Name)
	setInputValueInRow(e.Frame.T, row, ".oidc-provider-issuer-domain-path-edit", provider.IssuerDomainPath)
	setInputValueInRow(e.Frame.T, row, ".oidc-provider-client-id-edit", provider.ClientId)
	setInputValueInRow(e.Frame.T, row, ".oidc-provider-client-secret-edit", provider.ClientSecret)
	GetRequiredElementInRow(e.Frame.T, row, ".oidc-provider-save-button").MustClick()
	e.Frame.Assert.SnackbarVisibleWithTextEventually("Provider saved.")
	return e
}

func (e *ProvidersPage) DeleteProvider(providerId int) *ProvidersPage {
	row := e.findRowByProviderId(providerId)
	GetRequiredElementInRow(e.Frame.T, row, ".oidc-provider-delete-button").MustClick()
	e.Frame.Browser.ConfirmDialog()
	e.Frame.Assert.SnackbarVisibleWithTextEventually("Provider deleted.")
	return e
}

func (e *ProvidersPage) ToggleClientSecretVisibility(name string) *ProvidersPage {
	row := e.findRowByProviderName(name)
	GetRequiredElementInRow(e.Frame.T, row, ".oidc-provider-client-secret-toggle").MustClick()
	return e
}

func (e *ProvidersPage) AssertClientSecretVisibility(name string, visible bool) *ProvidersPage {
	expectedType := "password"
	if visible {
		expectedType = "text"
	}

	err := tools.Eventually(func() error {
		row := e.findRowByProviderName(name)
		actualType := getInputTypeInRow(e.Frame.T, row, ".oidc-provider-client-secret-edit")
		if actualType != expectedType {
			return fmt.Errorf("unexpected client secret input type: %q", actualType)
		}
		return nil
	})
	assert.Nil(e.Frame.T, err)
	return e
}

func (e *ProvidersPage) GetRequiredProvider(name string) oidc_client.OidcAuthProviderDto {
	var provider oidc_client.OidcAuthProviderDto
	err := tools.Eventually(func() error {
		rows := e.Frame.Page.MustElements("tr.oidc-auth-provider-row")
		for _, row := range rows {
			entry := e.readProviderEntry(row)
			if entry.Name == name {
				provider = entry
				return nil
			}
		}
		return fmt.Errorf("provider not found: %s", name)
	})
	assert.Nil(e.Frame.T, err)
	return provider
}

func (e *ProvidersPage) readProviderEntry(row *rod.Element) oidc_client.OidcAuthProviderDto {
	providerId, err := row.Attribute("data-provider-id")
	assert.Nil(e.Frame.T, err)
	assert.NotNil(e.Frame.T, providerId)

	id, err := strconv.Atoi(strings.TrimSpace(*providerId))
	assert.Nil(e.Frame.T, err)

	return oidc_client.OidcAuthProviderDto{
		Id:               id,
		Name:             getInputValueInRow(e.Frame.T, row, ".oidc-provider-name-edit"),
		IssuerDomainPath: getInputValueInRow(e.Frame.T, row, ".oidc-provider-issuer-domain-path-edit"),
		ClientId:         getInputValueInRow(e.Frame.T, row, ".oidc-provider-client-id-edit"),
		ClientSecret:     getInputValueInRow(e.Frame.T, row, ".oidc-provider-client-secret-edit"),
	}
}

func (e *ProvidersPage) findRowByProviderName(name string) *rod.Element {
	var foundRow *rod.Element
	err := tools.Eventually(func() error {
		rows := e.Frame.Page.MustElements("tr.oidc-auth-provider-row")
		for _, row := range rows {
			if getInputValueInRow(e.Frame.T, row, ".oidc-provider-name-edit") == name {
				foundRow = row
				return nil
			}
		}
		return fmt.Errorf("provider row not found: %s", name)
	})
	assert.Nil(e.Frame.T, err)
	return foundRow
}

func (e *ProvidersPage) findRowByProviderId(providerId int) *rod.Element {
	var foundRow *rod.Element
	expectedProviderId := strconv.Itoa(providerId)
	err := tools.Eventually(func() error {
		rows := e.Frame.Page.MustElements("tr.oidc-auth-provider-row")
		for _, row := range rows {
			actualProviderId, err := row.Attribute("data-provider-id")
			assert.Nil(e.Frame.T, err)
			if actualProviderId != nil && *actualProviderId == expectedProviderId {
				foundRow = row
				return nil
			}
		}
		return fmt.Errorf("provider row not found: %s", expectedProviderId)
	})
	assert.Nil(e.Frame.T, err)
	return foundRow
}

func GetRequiredElementInRow(t *testing.T, row *rod.Element, selector string) *rod.Element {
	element, err := row.Element(selector)
	assert.Nil(t, err)
	return element
}

func setInputValueInRow(t *testing.T, row *rod.Element, selector, value string) {
	GetRequiredElementInRow(t, row, selector).MustSelectAllText().MustInput(value)
}

func getInputValueInRow(t *testing.T, row *rod.Element, selector string) string {
	value, err := GetRequiredElementInRow(t, row, selector).Property("value")
	assert.Nil(t, err)
	return value.String()
}

func getInputTypeInRow(t *testing.T, row *rod.Element, selector string) string {
	value, err := GetRequiredElementInRow(t, row, selector).Property("type")
	assert.Nil(t, err)
	return value.String()
}
