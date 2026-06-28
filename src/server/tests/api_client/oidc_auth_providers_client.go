package api_client

import (
	"encoding/json"
	"server/oidc_client"
	"server/tools"
	"strconv"
)

type OidcAuthProvidersClient struct {
	quollix *QuollixClient
}

func (c *OidcAuthProvidersClient) Create(provider *oidc_client.OidcAuthProviderDto) error {
	_, err := c.quollix.Parent.DoRequest(tools.Paths.BackendOidcAuthProvidersCreate, provider)
	return err
}

func (c *OidcAuthProvidersClient) Update(provider *oidc_client.OidcAuthProviderDto) error {
	_, err := c.quollix.Parent.DoRequest(tools.Paths.BackendOidcAuthProvidersUpdate, provider)
	return err
}

func (c *OidcAuthProvidersClient) List() ([]oidc_client.OidcAuthProviderDto, error) {
	body, err := c.quollix.Parent.DoRequest(tools.Paths.BackendOidcAuthProvidersList, nil)
	if err != nil {
		return nil, err
	}

	var providers []oidc_client.OidcAuthProviderDto
	err = json.Unmarshal(body, &providers)
	if err != nil {
		return nil, err
	}
	return providers, nil
}

func (c *OidcAuthProvidersClient) Delete(providerId int) error {
	_, err := c.quollix.Parent.DoRequest(tools.Paths.BackendOidcAuthProvidersDelete, tools.NumberString{Value: strconv.Itoa(providerId)})
	return err
}

func (c *OidcAuthProvidersClient) TestDiscovery(issuerDomainPath string) error {
	_, err := c.quollix.Parent.DoRequest(tools.Paths.BackendOidcAuthProvidersTestDiscovery, oidc_client.OidcAuthProviderDiscoveryRequest{IssuerDomainPath: issuerDomainPath})
	return err
}

func (c *OidcAuthProvidersClient) StartLogin(providerId int) (oidc_client.OidcStartLoginResponse, error) {
	body, err := c.quollix.Parent.DoRequest(tools.Paths.BackendOidcSignIn, tools.NumberString{Value: strconv.Itoa(providerId)})
	if err != nil {
		return oidc_client.OidcStartLoginResponse{}, err
	}

	var response oidc_client.OidcStartLoginResponse
	err = json.Unmarshal(body, &response)
	return response, err
}
