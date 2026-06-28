package oidc

import (
	"fmt"
	"testing"

	"server/oidc_client"
	"server/oidc_provider"
	"server/tests/api_client"
	"server/tools"

	"github.com/quollix/common/assert"
)

const (
	ProviderDomain  = "quollix.oidc-provider.localhost"
	ClientDomain    = "quollix.oidc-client.localhost"
	ProviderBaseUrl = "https://" + ProviderDomain
	ClientBaseUrl   = "https://" + ClientDomain
	ProviderHost    = "oidc-provider.localhost"
	ClientHost      = "oidc-client.localhost"

	ProviderAdminUsername      = "provider-admin"
	ProviderAdminLocalPassword = "localuserpassword"
	OidcClientName             = "Two-Instance-Client"
	OidcProviderName           = "Provider-Quollix"
)

type TwoInstanceClients struct {
	ProviderAdmin *api_client.QuollixClient
	ClientAdmin   *api_client.QuollixClient
}

func SetupAndGetClients(t *testing.T) *TwoInstanceClients {
	clients := &TwoInstanceClients{
		ProviderAdmin: NewProviderClient(),
		ClientAdmin:   NewClientClient(),
	}
	assert.Nil(t, ConfigureTwoInstanceEnvironment(clients.ProviderAdmin, clients.ClientAdmin))
	return clients
}

func (c *TwoInstanceClients) Reset(t *testing.T) {
	assert.Nil(t, ResetTwoInstanceEnvironment(c.ProviderAdmin, c.ClientAdmin))
}

func NewProviderClient() *api_client.QuollixClient {
	return api_client.NewQuollixClientForRootUrl(ProviderBaseUrl)
}

func NewClientClient() *api_client.QuollixClient {
	return api_client.NewQuollixClientForRootUrl(ClientBaseUrl)
}

func ResetTwoInstanceEnvironment(providerAdmin *api_client.QuollixClient, clientAdmin *api_client.QuollixClient) error {
	if err := providerAdmin.Test.ResetTestState(); err != nil {
		return fmt.Errorf("reset OIDC provider server: %w", err)
	}
	if err := clientAdmin.Test.ResetTestState(); err != nil {
		return fmt.Errorf("reset OIDC client server: %w", err)
	}
	return nil
}

func ConfigureTwoInstanceEnvironment(providerAdmin *api_client.QuollixClient, clientAdmin *api_client.QuollixClient) error {
	if err := loginAdmin(providerAdmin); err != nil {
		return fmt.Errorf("sign in to OIDC provider server: %w", err)
	}
	if err := loginAdmin(clientAdmin); err != nil {
		return fmt.Errorf("sign in to OIDC client server: %w", err)
	}
	if err := renameProviderAdmin(providerAdmin); err != nil {
		return err
	}
	if err := providerAdmin.Settings.SetBaseDomainValue(ProviderHost); err != nil {
		return fmt.Errorf("set OIDC provider base domain: %w", err)
	}
	if err := clientAdmin.Settings.SetBaseDomainValue(ClientHost); err != nil {
		return fmt.Errorf("set OIDC client base domain: %w", err)
	}

	relyingParty, err := createRelyingPartyInProviderInstance(providerAdmin)
	if err != nil {
		return err
	}
	if err := createExternalProviderInClientInstance(clientAdmin, relyingParty); err != nil {
		return err
	}
	return nil
}

func loginAdmin(client *api_client.QuollixClient) error {
	return client.Auth.SignIn(tools.DefaultAdminName, tools.DefaultAdminPassword)
}

func renameProviderAdmin(providerAdmin *api_client.QuollixClient) error {
	admin, exists, err := providerAdmin.Users.GetByUsername(tools.DefaultAdminName)
	if err != nil {
		return fmt.Errorf("find provider admin user: %w", err)
	}
	if !exists {
		return fmt.Errorf("find provider admin user: admin user does not exist")
	}
	if err := providerAdmin.Users.ChangeUsername(admin.Id, ProviderAdminUsername); err != nil {
		return fmt.Errorf("rename provider admin user: %w", err)
	}
	return nil
}

func createRelyingPartyInProviderInstance(providerAdmin *api_client.QuollixClient) (oidc_provider.OidcRelyingPartyDto, error) {
	relyingParty := &oidc_provider.OidcRelyingPartyDto{
		Name:   OidcClientName,
		Domain: ClientDomain,
	}
	if err := providerAdmin.OidcClients.Create(relyingParty); err != nil {
		return oidc_provider.OidcRelyingPartyDto{}, fmt.Errorf("create relying party in OIDC provider server: %w", err)
	}
	return getRelyingPartyByName(providerAdmin, OidcClientName)
}

func getRelyingPartyByName(providerAdmin *api_client.QuollixClient, name string) (oidc_provider.OidcRelyingPartyDto, error) {
	clients, err := providerAdmin.OidcClients.List()
	if err != nil {
		return oidc_provider.OidcRelyingPartyDto{}, fmt.Errorf("list OIDC relying parties: %w", err)
	}
	for _, relyingParty := range clients {
		if relyingParty.Name == name {
			return relyingParty, nil
		}
	}
	return oidc_provider.OidcRelyingPartyDto{}, fmt.Errorf("find created OIDC relying party: created relying party was not found")
}

func createExternalProviderInClientInstance(clientAdmin *api_client.QuollixClient, relyingParty oidc_provider.OidcRelyingPartyDto) error {
	provider := &oidc_client.OidcAuthProviderDto{
		Name:             OidcProviderName,
		IssuerDomainPath: ProviderDomain,
		ClientId:         relyingParty.ClientId,
		ClientSecret:     relyingParty.ClientSecret,
	}
	if err := clientAdmin.OidcProviders.Create(provider); err != nil {
		return fmt.Errorf("create OIDC provider in OIDC client server: %w", err)
	}
	return nil
}
