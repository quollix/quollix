package oidc_provider

import (
	"crypto/subtle"
	"net/url"

	"server/apps_basic"
	"server/configs"

	u "github.com/quollix/common/utils"
)

type OidcRelyingPartyResolver interface {
	CheckClientIdAndRedirectUri(clientId, redirectUri string) error
	AuthenticateClient(clientId, clientSecret string) error
}

type OidcRelyingPartyResolverImpl struct {
	AppRepository    apps_basic.AppRepository
	ClientRepository OidcRelyingPartyRepository
	ConfigsService   configs.ConfigsService
}

func (r *OidcRelyingPartyResolverImpl) CheckClientIdAndRedirectUri(clientId string, redirectUri string) error {
	app, exists, err := r.AppRepository.GetAppByClientId(clientId)
	if err != nil {
		return err
	}
	if exists {
		return r.checkAppRedirectUri(app, redirectUri)
	}

	client, exists, err := r.ClientRepository.GetClientByClientId(clientId)
	if err != nil {
		return err
	}
	if !exists {
		return u.Logger.NewError(OidcRelyingPartyNotFoundError)
	}
	return r.checkGenericClientRedirectUri(client, redirectUri)
}

func (r *OidcRelyingPartyResolverImpl) AuthenticateClient(clientId string, clientSecret string) error {
	app, exists, err := r.AppRepository.GetAppByClientId(clientId)
	if err != nil {
		return err
	}
	if exists {
		return authenticateConfidentialClient(app.ClientSecret, clientSecret)
	}

	client, exists, err := r.ClientRepository.GetClientByClientId(clientId)
	if err != nil {
		return err
	}
	if !exists {
		return u.Logger.NewError(OidcRelyingPartyNotFoundError)
	}
	return authenticateConfidentialClient(client.ClientSecret, clientSecret)
}

func (r *OidcRelyingPartyResolverImpl) checkAppRedirectUri(app *apps_basic.RepoApp, redirectUri string) error {
	host, err := r.ConfigsService.GetBaseDomain()
	if err != nil {
		return err
	}
	redirectUrl, err := parseRedirectUri(redirectUri)
	if err != nil {
		return err
	}

	expectedHostname := app.AppName + "." + host
	if redirectUrl.Scheme != "https" {
		return u.Logger.NewError("wrong redirect_uri scheme")
	}
	if redirectUrl.Hostname() != expectedHostname {
		return u.Logger.NewError("wrong redirect_uri host")
	}
	if effectiveHttpsPort(redirectUrl) != "443" {
		return u.Logger.NewError("wrong redirect_uri port")
	}

	return nil
}

func (r *OidcRelyingPartyResolverImpl) checkGenericClientRedirectUri(client *OidcRelyingPartyDto, redirectUri string) error {
	redirectUrl, err := parseRedirectUri(redirectUri)
	if err != nil {
		return err
	}

	if redirectUrl.Scheme != "https" {
		return u.Logger.NewError("wrong redirect_uri scheme")
	}
	if redirectUrl.Hostname() != client.Domain {
		return u.Logger.NewError("wrong redirect_uri host")
	}
	if effectiveHttpsPort(redirectUrl) != "443" {
		return u.Logger.NewError("wrong redirect_uri port")
	}

	return nil
}

func parseRedirectUri(redirectUri string) (*url.URL, error) {
	redirectUrl, err := url.Parse(redirectUri)
	if err != nil {
		return nil, u.Logger.NewError("invalid redirect_uri format")
	}
	return redirectUrl, nil
}

func effectiveHttpsPort(parsedUrl *url.URL) string {
	if parsedUrl.Port() == "" {
		return "443"
	}
	return parsedUrl.Port()
}

func authenticateConfidentialClient(storedSecret string, requestSecret string) error {
	if storedSecret == "" {
		return u.Logger.NewError("confidential client missing stored client secret")
	}
	if requestSecret == "" {
		return u.Logger.NewError("confidential client missing client secret")
	}
	if subtle.ConstantTimeCompare([]byte(storedSecret), []byte(requestSecret)) != 1 {
		return u.Logger.NewError("wrong client secret")
	}
	return nil
}
