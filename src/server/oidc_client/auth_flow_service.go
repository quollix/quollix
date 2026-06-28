package oidc_client

import (
	"net/http"
	"server/configs"
	"server/tools"
	"strings"
	"time"

	u "github.com/quollix/common/utils"
)

const (
	MissingOidcAuthorizationCodeError = "OIDC authorization code is missing"
	InvalidOidcLoginStateError        = "OIDC sign-in state is invalid"
)

const (
	oidcSignInStateCookieName = "quollix-oidc-sign-in-state"
	oidcLoginStateTtl         = 10 * time.Minute
)

type OidcAuthFlowService interface {
	StartLogin(providerId int) (*OidcLoginStart, error)
	FinishLogin(state string, stateCookieValue string, code string) (*http.Cookie, error)
}

type OidcAuthFlowServiceImpl struct {
	ProviderRepo   OidcAuthProviderRepository
	ProviderClient OidcProviderClient
	LoginService   LoginService
	StateCache     OidcLoginStateCache
	ConfigsService configs.ConfigsService
	AuthHelper     u.AuthHelper
	OsWrapper      u.OsWrapper
}

type OidcLoginStart struct {
	RedirectUrl string
	StateCookie *http.Cookie
}

func (s *OidcAuthFlowServiceImpl) StartLogin(providerId int) (*OidcLoginStart, error) {
	provider, err := s.ProviderRepo.GetProviderById(providerId)
	if err != nil {
		return nil, err
	}
	state, err := s.AuthHelper.GenerateSecret()
	if err != nil {
		return nil, err
	}
	now := s.OsWrapper.Now()
	expiresAt := now.Add(oidcLoginStateTtl)
	s.StateCache.StoreLoginState(OidcLoginState{
		State:      state,
		ProviderId: providerId,
		ExpiresAt:  expiresAt,
	})

	callbackUrl, err := s.callbackUrl()
	if err != nil {
		return nil, err
	}
	redirectUrl, err := s.ProviderClient.GetAuthorizationUrl(provider, callbackUrl, state)
	if err != nil {
		return nil, err
	}
	return &OidcLoginStart{
		RedirectUrl: redirectUrl,
		StateCookie: newOidcLoginStateCookie(state, expiresAt),
	}, nil
}

func (s *OidcAuthFlowServiceImpl) FinishLogin(state string, stateCookieValue string, code string) (*http.Cookie, error) {
	state = strings.TrimSpace(state)
	code = strings.TrimSpace(code)
	if code == "" {
		return nil, u.Logger.NewError(MissingOidcAuthorizationCodeError)
	}
	if state == "" || state != stateCookieValue {
		return nil, u.Logger.NewError(InvalidOidcLoginStateError)
	}
	loginState, found := s.StateCache.ConsumeLoginState(state, s.OsWrapper.Now())
	if !found {
		return nil, u.Logger.NewError(InvalidOidcLoginStateError)
	}
	provider, err := s.ProviderRepo.GetProviderById(loginState.ProviderId)
	if err != nil {
		return nil, err
	}
	callbackUrl, err := s.callbackUrl()
	if err != nil {
		return nil, err
	}
	claims, err := s.ProviderClient.ExchangeCodeForClaims(provider, callbackUrl, code)
	if err != nil {
		return nil, err
	}
	return s.LoginService.LoginWithClaims(loginState.ProviderId, claims)
}

func (s *OidcAuthFlowServiceImpl) callbackUrl() (string, error) {
	host, err := s.ConfigsService.GetBaseDomain()
	if err != nil {
		return "", err
	}
	return "https://quollix." + host + tools.Paths.BackendOidcCallback, nil
}

func newOidcLoginStateCookie(state string, expiresAt time.Time) *http.Cookie {
	return &http.Cookie{
		Name:     oidcSignInStateCookieName,
		Value:    state,
		Path:     tools.Paths.BackendOidcCallback,
		Expires:  expiresAt,
		MaxAge:   int(oidcLoginStateTtl.Seconds()),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}
}

func clearOidcLoginStateCookie() *http.Cookie {
	return &http.Cookie{
		Name:     oidcSignInStateCookieName,
		Value:    "",
		Path:     tools.Paths.BackendOidcCallback,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}
}
