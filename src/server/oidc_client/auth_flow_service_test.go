package oidc_client

import (
	"net/http"
	"server/configs"
	"server/tools"
	"testing"
	"time"

	"github.com/quollix/common/assert"
	u "github.com/quollix/common/utils"
	"github.com/stretchr/testify/mock"
)

var oidcAuthFlowNow = time.Date(2026, time.June, 21, 11, 0, 0, 0, time.UTC)

func TestOidcAuthFlowServiceImpl_StartLoginStoresStateAndReturnsRedirect(t *testing.T) {
	testObjects := newOidcAuthFlowTestObjects(t)
	provider := &OidcAuthProviderDto{Id: 10, IssuerDomainPath: "issuer.example", ClientId: "client-id"}
	var storedState OidcLoginState
	testObjects.ProviderRepo.EXPECT().GetProviderById(10).Return(provider, nil)
	testObjects.AuthHelper.EXPECT().GenerateSecret().Return("state-secret", nil)
	testObjects.OsWrapper.EXPECT().Now().Return(oidcAuthFlowNow)
	testObjects.StateCache.EXPECT().StoreLoginState(mock.AnythingOfType("oidc_client.OidcLoginState")).Run(func(state OidcLoginState) {
		storedState = state
	})
	testObjects.ConfigsService.EXPECT().GetBaseDomain().Return("example.com", nil)
	testObjects.ProviderClient.EXPECT().GetAuthorizationUrl(provider, "https://quollix.example.com/api/auth/oidc/callback", "state-secret").Return("https://issuer.example/auth", nil)

	start, err := testObjects.Service.StartLogin(10)

	assert.Nil(t, err)
	assert.Equal(t, "https://issuer.example/auth", start.RedirectUrl)
	assert.Equal(t, oidcSignInStateCookieName, start.StateCookie.Name)
	assert.Equal(t, "state-secret", start.StateCookie.Value)
	assert.Equal(t, tools.Paths.BackendOidcCallback, start.StateCookie.Path)
	assert.True(t, start.StateCookie.HttpOnly)
	assert.True(t, start.StateCookie.Secure)
	assert.Equal(t, "state-secret", storedState.State)
	assert.Equal(t, 10, storedState.ProviderId)
	assert.Equal(t, oidcAuthFlowNow.Add(oidcLoginStateTtl), storedState.ExpiresAt)
}

func TestOidcAuthFlowServiceImpl_FinishLoginExchangesCodeAndCreatesSession(t *testing.T) {
	testObjects := newOidcAuthFlowTestObjects(t)
	provider := &OidcAuthProviderDto{Id: 10, IssuerDomainPath: "issuer.example", ClientId: "client-id"}
	claims := OidcLoginClaims{
		Subject:           "subject",
		PreferredUsername: "tom",
	}
	testObjects.OsWrapper.EXPECT().Now().Return(oidcAuthFlowNow)
	testObjects.StateCache.EXPECT().ConsumeLoginState("state-secret", oidcAuthFlowNow).Return(OidcLoginState{
		State:      "state-secret",
		ProviderId: 10,
		ExpiresAt:  oidcAuthFlowNow.Add(oidcLoginStateTtl),
	}, true)
	testObjects.ProviderRepo.EXPECT().GetProviderById(10).Return(provider, nil)
	testObjects.ConfigsService.EXPECT().GetBaseDomain().Return("example.com", nil)
	testObjects.ProviderClient.EXPECT().ExchangeCodeForClaims(provider, "https://quollix.example.com/api/auth/oidc/callback", "code").Return(claims, nil)
	testObjects.LoginService.EXPECT().LoginWithClaims(10, claims).Return(&http.Cookie{Name: "session", Value: "cookie"}, nil)

	cookie, err := testObjects.Service.FinishLogin("state-secret", "state-secret", "code")

	assert.Nil(t, err)
	assert.Equal(t, "session", cookie.Name)
	assert.Equal(t, "cookie", cookie.Value)
}

func TestOidcAuthFlowServiceImpl_FinishLoginTrimsStateAndCode(t *testing.T) {
	testObjects := newOidcAuthFlowTestObjects(t)
	provider := &OidcAuthProviderDto{Id: 10, IssuerDomainPath: "issuer.example", ClientId: "client-id"}
	claims := OidcLoginClaims{
		Subject:           "subject",
		PreferredUsername: "tom",
	}
	testObjects.OsWrapper.EXPECT().Now().Return(oidcAuthFlowNow)
	testObjects.StateCache.EXPECT().ConsumeLoginState("state-secret", oidcAuthFlowNow).Return(OidcLoginState{
		State:      "state-secret",
		ProviderId: 10,
		ExpiresAt:  oidcAuthFlowNow.Add(oidcLoginStateTtl),
	}, true)
	testObjects.ProviderRepo.EXPECT().GetProviderById(10).Return(provider, nil)
	testObjects.ConfigsService.EXPECT().GetBaseDomain().Return("example.com", nil)
	testObjects.ProviderClient.EXPECT().ExchangeCodeForClaims(provider, "https://quollix.example.com/api/auth/oidc/callback", "code").Return(claims, nil)
	testObjects.LoginService.EXPECT().LoginWithClaims(10, claims).Return(&http.Cookie{Name: "session", Value: "cookie"}, nil)

	cookie, err := testObjects.Service.FinishLogin(" state-secret ", "state-secret", " code ")

	assert.Nil(t, err)
	assert.Equal(t, "session", cookie.Name)
	assert.Equal(t, "cookie", cookie.Value)
}

func TestOidcAuthFlowServiceImpl_FinishLoginReturnsErrorForMissingCode(t *testing.T) {
	testObjects := newOidcAuthFlowTestObjects(t)

	_, err := testObjects.Service.FinishLogin("state-secret", "state-secret", "")

	assert.Equal(t, MissingOidcAuthorizationCodeError, u.ExtractError(err))
}

func TestOidcAuthFlowServiceImpl_FinishLoginReturnsErrorForMissingState(t *testing.T) {
	testObjects := newOidcAuthFlowTestObjects(t)

	_, err := testObjects.Service.FinishLogin("", "", "code")

	assert.Equal(t, InvalidOidcLoginStateError, u.ExtractError(err))
}

func TestOidcAuthFlowServiceImpl_FinishLoginReturnsErrorForStateMismatch(t *testing.T) {
	testObjects := newOidcAuthFlowTestObjects(t)

	_, err := testObjects.Service.FinishLogin("state-secret", "different-state", "code")

	assert.Equal(t, InvalidOidcLoginStateError, u.ExtractError(err))
}

func TestOidcAuthFlowServiceImpl_FinishLoginReturnsErrorForUnknownState(t *testing.T) {
	testObjects := newOidcAuthFlowTestObjects(t)
	testObjects.OsWrapper.EXPECT().Now().Return(oidcAuthFlowNow)
	testObjects.StateCache.EXPECT().ConsumeLoginState("state-secret", oidcAuthFlowNow).Return(OidcLoginState{}, false)

	_, err := testObjects.Service.FinishLogin("state-secret", "state-secret", "code")

	assert.Equal(t, InvalidOidcLoginStateError, u.ExtractError(err))
}

type oidcAuthFlowTestObjects struct {
	Service        *OidcAuthFlowServiceImpl
	ProviderRepo   *OidcAuthProviderRepositoryMock
	ProviderClient *OidcProviderClientMock
	LoginService   *LoginServiceMock
	StateCache     *OidcLoginStateCacheMock
	ConfigsService *configs.ConfigsServiceMock
	AuthHelper     *tools.AuthHelperMock
	OsWrapper      *tools.CommonOsWrapperMock
}

func newOidcAuthFlowTestObjects(t *testing.T) oidcAuthFlowTestObjects {
	providerRepo := NewOidcAuthProviderRepositoryMock(t)
	providerClient := NewOidcProviderClientMock(t)
	loginService := NewLoginServiceMock(t)
	stateCache := NewOidcLoginStateCacheMock(t)
	configsService := configs.NewConfigsServiceMock(t)
	authHelper := tools.NewAuthHelperMock(t)
	osWrapper := tools.NewCommonOsWrapperMock(t)
	return oidcAuthFlowTestObjects{
		Service: &OidcAuthFlowServiceImpl{
			ProviderRepo:   providerRepo,
			ProviderClient: providerClient,
			LoginService:   loginService,
			StateCache:     stateCache,
			ConfigsService: configsService,
			AuthHelper:     authHelper,
			OsWrapper:      osWrapper,
		},
		ProviderRepo:   providerRepo,
		ProviderClient: providerClient,
		LoginService:   loginService,
		StateCache:     stateCache,
		ConfigsService: configsService,
		AuthHelper:     authHelper,
		OsWrapper:      osWrapper,
	}
}
