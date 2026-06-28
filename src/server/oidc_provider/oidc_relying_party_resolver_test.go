package oidc_provider

import (
	"testing"

	"server/apps_basic"
	"server/configs"

	"github.com/quollix/common/assert"
	u "github.com/quollix/common/utils"
)

const (
	oidcResolverClientId     = "client-id"
	oidcResolverClientSecret = "client-secret"
)

type oidcRelyingPartyResolverTestObjects struct {
	Resolver       *OidcRelyingPartyResolverImpl
	AppRepo        *apps_basic.AppRepositoryMock
	ClientRepo     *OidcRelyingPartyRepositoryMock
	ConfigsService *configs.ConfigsServiceMock
}

func newOidcRelyingPartyResolverTestObjects(t *testing.T) oidcRelyingPartyResolverTestObjects {
	appRepo := apps_basic.NewAppRepositoryMock(t)
	clientRepo := NewOidcRelyingPartyRepositoryMock(t)
	configsService := configs.NewConfigsServiceMock(t)
	resolver := &OidcRelyingPartyResolverImpl{
		AppRepository:    appRepo,
		ClientRepository: clientRepo,
		ConfigsService:   configsService,
	}

	return oidcRelyingPartyResolverTestObjects{
		Resolver:       resolver,
		AppRepo:        appRepo,
		ClientRepo:     clientRepo,
		ConfigsService: configsService,
	}
}

func expectAppClient(testObjects oidcRelyingPartyResolverTestObjects) {
	testObjects.AppRepo.EXPECT().GetAppByClientId(oidcResolverClientId).Return(&apps_basic.RepoApp{
		AppName:      "sampleapp",
		ClientSecret: oidcResolverClientSecret,
	}, true, nil)
}

func expectNoAppClient(testObjects oidcRelyingPartyResolverTestObjects) {
	testObjects.AppRepo.EXPECT().GetAppByClientId(oidcResolverClientId).Return(nil, false, nil)
}

func expectGenericClient(testObjects oidcRelyingPartyResolverTestObjects) {
	testObjects.ClientRepo.EXPECT().GetClientByClientId(oidcResolverClientId).Return(&OidcRelyingPartyDto{
		Domain:       "client.example.com",
		ClientSecret: oidcResolverClientSecret,
	}, true, nil)
}

func TestOidcRelyingPartyResolver_CheckClientIdAndRedirectUri_AppHappyPath(t *testing.T) {
	testObjects := newOidcRelyingPartyResolverTestObjects(t)
	expectAppClient(testObjects)
	testObjects.ConfigsService.EXPECT().GetBaseDomain().Return("localhost", nil)

	err := testObjects.Resolver.CheckClientIdAndRedirectUri(oidcResolverClientId, "https://sampleapp.localhost/callback")
	assert.Nil(t, err)
}

func TestOidcRelyingPartyResolver_CheckClientIdAndRedirectUri_AppWrongHostReturnsError(t *testing.T) {
	testObjects := newOidcRelyingPartyResolverTestObjects(t)
	expectAppClient(testObjects)
	testObjects.ConfigsService.EXPECT().GetBaseDomain().Return("localhost", nil)

	err := testObjects.Resolver.CheckClientIdAndRedirectUri(oidcResolverClientId, "https://evil.localhost/callback")
	assert.Equal(t, "wrong redirect_uri host", u.ExtractError(err))
}

func TestOidcRelyingPartyResolver_CheckClientIdAndRedirectUri_AppWrongSchemeReturnsError(t *testing.T) {
	testObjects := newOidcRelyingPartyResolverTestObjects(t)
	expectAppClient(testObjects)
	testObjects.ConfigsService.EXPECT().GetBaseDomain().Return("localhost", nil)

	err := testObjects.Resolver.CheckClientIdAndRedirectUri(oidcResolverClientId, "http://sampleapp.localhost/callback")
	assert.Equal(t, "wrong redirect_uri scheme", u.ExtractError(err))
}

func TestOidcRelyingPartyResolver_CheckClientIdAndRedirectUri_AppWrongPortReturnsError(t *testing.T) {
	testObjects := newOidcRelyingPartyResolverTestObjects(t)
	expectAppClient(testObjects)
	testObjects.ConfigsService.EXPECT().GetBaseDomain().Return("localhost", nil)

	err := testObjects.Resolver.CheckClientIdAndRedirectUri(oidcResolverClientId, "https://sampleapp.localhost:8443/callback")
	assert.Equal(t, "wrong redirect_uri port", u.ExtractError(err))
}

func TestOidcRelyingPartyResolver_CheckClientIdAndRedirectUri_GenericClientHappyPath(t *testing.T) {
	testObjects := newOidcRelyingPartyResolverTestObjects(t)
	expectNoAppClient(testObjects)
	expectGenericClient(testObjects)

	err := testObjects.Resolver.CheckClientIdAndRedirectUri(oidcResolverClientId, "https://client.example.com/oidc/callback")
	assert.Nil(t, err)
}

func TestOidcRelyingPartyResolver_CheckClientIdAndRedirectUri_GenericClientIgnoresCallbackPath(t *testing.T) {
	testObjects := newOidcRelyingPartyResolverTestObjects(t)
	expectNoAppClient(testObjects)
	expectGenericClient(testObjects)

	err := testObjects.Resolver.CheckClientIdAndRedirectUri(oidcResolverClientId, "https://client.example.com/other/path")
	assert.Nil(t, err)
}

func TestOidcRelyingPartyResolver_CheckClientIdAndRedirectUri_GenericClientWrongSchemeReturnsError(t *testing.T) {
	testObjects := newOidcRelyingPartyResolverTestObjects(t)
	expectNoAppClient(testObjects)
	expectGenericClient(testObjects)

	err := testObjects.Resolver.CheckClientIdAndRedirectUri(oidcResolverClientId, "http://client.example.com/oidc/callback")
	assert.Equal(t, "wrong redirect_uri scheme", u.ExtractError(err))
}

func TestOidcRelyingPartyResolver_CheckClientIdAndRedirectUri_GenericClientWrongHostReturnsError(t *testing.T) {
	testObjects := newOidcRelyingPartyResolverTestObjects(t)
	expectNoAppClient(testObjects)
	expectGenericClient(testObjects)

	err := testObjects.Resolver.CheckClientIdAndRedirectUri(oidcResolverClientId, "https://evil.example.com/oidc/callback")
	assert.Equal(t, "wrong redirect_uri host", u.ExtractError(err))
}

func TestOidcRelyingPartyResolver_CheckClientIdAndRedirectUri_GenericClientWrongPortReturnsError(t *testing.T) {
	testObjects := newOidcRelyingPartyResolverTestObjects(t)
	expectNoAppClient(testObjects)
	expectGenericClient(testObjects)

	err := testObjects.Resolver.CheckClientIdAndRedirectUri(oidcResolverClientId, "https://client.example.com:8443/oidc/callback")
	assert.Equal(t, "wrong redirect_uri port", u.ExtractError(err))
}

func TestOidcRelyingPartyResolver_CheckClientIdAndRedirectUri_GenericClientAllowsDefaultHttpsPort(t *testing.T) {
	testObjects := newOidcRelyingPartyResolverTestObjects(t)
	expectNoAppClient(testObjects)
	expectGenericClient(testObjects)

	err := testObjects.Resolver.CheckClientIdAndRedirectUri(oidcResolverClientId, "https://client.example.com:443/oidc/callback")
	assert.Nil(t, err)
}

func TestOidcRelyingPartyResolver_AuthenticateClient_AppSecretMatches(t *testing.T) {
	testObjects := newOidcRelyingPartyResolverTestObjects(t)
	expectAppClient(testObjects)

	err := testObjects.Resolver.AuthenticateClient(oidcResolverClientId, oidcResolverClientSecret)
	assert.Nil(t, err)
}

func TestOidcRelyingPartyResolver_AuthenticateClient_AppSecretMismatchReturnsError(t *testing.T) {
	testObjects := newOidcRelyingPartyResolverTestObjects(t)
	expectAppClient(testObjects)

	err := testObjects.Resolver.AuthenticateClient(oidcResolverClientId, "wrong-secret")
	assert.Equal(t, "wrong client secret", u.ExtractError(err))
}

func TestOidcRelyingPartyResolver_AuthenticateClient_GenericClientSecretMatches(t *testing.T) {
	testObjects := newOidcRelyingPartyResolverTestObjects(t)
	expectNoAppClient(testObjects)
	expectGenericClient(testObjects)

	err := testObjects.Resolver.AuthenticateClient(oidcResolverClientId, oidcResolverClientSecret)
	assert.Nil(t, err)
}

func TestOidcRelyingPartyResolver_AuthenticateClient_GenericClientSecretMismatchReturnsError(t *testing.T) {
	testObjects := newOidcRelyingPartyResolverTestObjects(t)
	expectNoAppClient(testObjects)
	expectGenericClient(testObjects)

	err := testObjects.Resolver.AuthenticateClient(oidcResolverClientId, "wrong-secret")
	assert.Equal(t, "wrong client secret", u.ExtractError(err))
}

func TestOidcRelyingPartyResolver_AuthenticateClient_GenericClientMissingRequestSecretReturnsError(t *testing.T) {
	testObjects := newOidcRelyingPartyResolverTestObjects(t)
	expectNoAppClient(testObjects)
	expectGenericClient(testObjects)

	err := testObjects.Resolver.AuthenticateClient(oidcResolverClientId, "")
	assert.Equal(t, "confidential client missing client secret", u.ExtractError(err))
}
