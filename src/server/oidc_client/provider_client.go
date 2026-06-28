package oidc_client

import (
	"context"
	"crypto/tls"
	"net/http"
	"time"

	"server/tools"

	"github.com/coreos/go-oidc/v3/oidc"
	u "github.com/quollix/common/utils"
	"golang.org/x/oauth2"
)

const (
	MissingIdTokenError        = "OIDC provider did not return an id token" // #nosec G101 (CWE-798): Potential hardcoded credentials
	oidcProviderRequestTimeout = 3 * time.Second
)

type OidcProviderClient interface {
	GetAuthorizationUrl(provider *OidcAuthProviderDto, redirectUrl string, state string) (string, error)
	ExchangeCodeForClaims(provider *OidcAuthProviderDto, redirectUrl string, code string) (OidcLoginClaims, error)
	TestDiscovery(issuerDomainPath string) error
}

type OidcProviderClientImpl struct {
	httpClient *http.Client
}

func NewOidcProviderClient(config *tools.GlobalConfig) OidcProviderClient {
	httpClient := http.DefaultClient
	if config.AllowInsecureOidcProviderTls {
		transport := http.DefaultTransport.(*http.Transport).Clone()
		transport.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true, // #nosec G402: dev profile must connect to local OIDC providers with self-signed certificates
		}
		httpClient = &http.Client{
			Transport: transport,
		}
	}
	return &OidcProviderClientImpl{
		httpClient: httpClient,
	}
}

func (c *OidcProviderClientImpl) GetAuthorizationUrl(provider *OidcAuthProviderDto, redirectUrl string, state string) (string, error) {
	ctx, cancel := c.requestContext()
	defer cancel()

	oidcProvider, err := oidc.NewProvider(ctx, issuerDomainPathWithHttps(provider.IssuerDomainPath))
	if err != nil {
		return "", err
	}
	return c.oauth2Config(provider, oidcProvider, redirectUrl).AuthCodeURL(state), nil
}

func (c *OidcProviderClientImpl) ExchangeCodeForClaims(provider *OidcAuthProviderDto, redirectUrl string, code string) (OidcLoginClaims, error) {
	ctx, cancel := c.requestContext()
	defer cancel()

	oidcProvider, err := oidc.NewProvider(ctx, issuerDomainPathWithHttps(provider.IssuerDomainPath))
	if err != nil {
		return OidcLoginClaims{}, err
	}

	oauth2Config := c.oauth2Config(provider, oidcProvider, redirectUrl)
	token, err := oauth2Config.Exchange(ctx, code)
	if err != nil {
		return OidcLoginClaims{}, err
	}

	rawIdToken, ok := token.Extra("id_token").(string)
	if !ok {
		return OidcLoginClaims{}, u.Logger.NewError(MissingIdTokenError)
	}
	idToken, err := oidcProvider.Verifier(&oidc.Config{ClientID: provider.ClientId}).Verify(ctx, rawIdToken)
	if err != nil {
		return OidcLoginClaims{}, err
	}

	var claims OidcLoginClaims
	if err = idToken.Claims(&claims); err != nil {
		return OidcLoginClaims{}, err
	}
	return claims, nil
}

func (c *OidcProviderClientImpl) TestDiscovery(issuerDomainPath string) error {
	ctx, cancel := c.requestContext()
	defer cancel()

	_, err := oidc.NewProvider(ctx, issuerDomainPathWithHttps(issuerDomainPath))
	return err
}

func (c *OidcProviderClientImpl) oauth2Config(provider *OidcAuthProviderDto, oidcProvider *oidc.Provider, redirectUrl string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     provider.ClientId,
		ClientSecret: provider.ClientSecret,
		Endpoint:     oidcProvider.Endpoint(),
		RedirectURL:  redirectUrl,
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}
}

func (c *OidcProviderClientImpl) requestContext() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(context.Background(), oidcProviderRequestTimeout)
	return oidc.ClientContext(ctx, c.httpClient), cancel
}

func issuerDomainPathWithHttps(issuerDomainPath string) string {
	return "https://" + issuerDomainPath
}
