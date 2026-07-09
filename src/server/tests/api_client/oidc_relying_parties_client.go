package api_client

import (
	"encoding/json"
	"strconv"

	"server/oidc_provider"
	"server/tools"
)

type OidcRelyingPartiesClient struct {
	quollix *QuollixClient
}

func (c *OidcRelyingPartiesClient) Create(client *oidc_provider.OidcRelyingPartyDto) error {
	request := oidc_provider.OidcRelyingPartyRequest{
		Id:     "0",
		Name:   client.Name,
		Domain: client.Domain,
	}
	_, err := c.quollix.Parent.DoRequest(tools.Paths.BackendOidcRelyingPartiesCreate, request)
	return err
}

func (c *OidcRelyingPartiesClient) Update(client *oidc_provider.OidcRelyingPartyDto) error {
	request := oidc_provider.OidcRelyingPartyRequest{
		Id:     strconv.Itoa(client.Id),
		Name:   client.Name,
		Domain: client.Domain,
	}
	_, err := c.quollix.Parent.DoRequest(tools.Paths.BackendOidcRelyingPartiesUpdate, request)
	return err
}

func (c *OidcRelyingPartiesClient) Regenerate(clientId int) error {
	_, err := c.quollix.Parent.DoRequest(tools.Paths.BackendOidcRelyingPartiesRegenerate, tools.NumberString{Value: strconv.Itoa(clientId)})
	return err
}

func (c *OidcRelyingPartiesClient) List() ([]oidc_provider.OidcRelyingPartyDto, error) {
	body, err := c.quollix.Parent.DoRequest(tools.Paths.BackendOidcRelyingPartiesList, nil)
	if err != nil {
		return nil, err
	}

	var clients []oidc_provider.OidcRelyingPartyDto
	err = json.Unmarshal(body, &clients)
	if err != nil {
		return nil, err
	}
	return clients, nil
}

func (c *OidcRelyingPartiesClient) Delete(clientId int) error {
	_, err := c.quollix.Parent.DoRequest(tools.Paths.BackendOidcRelyingPartiesDelete, tools.NumberString{Value: strconv.Itoa(clientId)})
	return err
}
