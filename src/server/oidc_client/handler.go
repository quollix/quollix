package oidc_client

import (
	"net/http"
	"strconv"

	"server/tools"
	"server/users"

	u "github.com/quollix/common/utils"
	"github.com/quollix/common/validation"
)

const MissingOidcLoginStateCookieError = "OIDC sign-in state cookie is missing"

type OidcClientHandler struct {
	AuthFlowService     OidcAuthFlowService
	AuthProviderService OidcAuthProviderService
}

type OidcStartLoginResponse struct {
	RedirectUrl string `json:"redirect_url"`
}

func (h *OidcClientHandler) StartLogin(w http.ResponseWriter, r *http.Request) {
	providerIdString, ok := validation.ReadBody[tools.NumberString](w, r)
	if !ok {
		return
	}
	providerId, err := strconv.Atoi(providerIdString.Value)
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}
	start, err := h.AuthFlowService.StartLogin(providerId)
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}
	http.SetCookie(w, start.StateCookie)
	u.SendJsonResponse(w, OidcStartLoginResponse{RedirectUrl: start.RedirectUrl})
}

func (h *OidcClientHandler) Callback(w http.ResponseWriter, r *http.Request) {
	stateCookie, err := r.Cookie(oidcSignInStateCookieName)
	if err != nil {
		u.WriteResponseError(w, nil, u.Logger.NewError(MissingOidcLoginStateCookieError))
		return
	}
	sessionCookie, err := h.AuthFlowService.FinishLogin(
		r.URL.Query().Get("state"),
		stateCookie.Value,
		r.URL.Query().Get("code"),
	)
	if err != nil {
		http.SetCookie(w, clearOidcLoginStateCookie())
		u.WriteResponseError(w, expectedOidcLoginErrors, err)
		return
	}
	http.SetCookie(w, clearOidcLoginStateCookie())
	http.SetCookie(w, sessionCookie)
	http.Redirect(w, r, tools.Paths.FrontendIndex, http.StatusFound)
}

func (h *OidcClientHandler) CreateAuthProvider(w http.ResponseWriter, r *http.Request) {
	provider, ok := validation.ReadBody[OidcAuthProviderDto](w, r)
	if !ok {
		return
	}
	if err := h.AuthProviderService.CreateProvider(provider); err != nil {
		u.WriteResponseError(w, expectedOidcAuthProviderErrors, err)
		return
	}
}

func (h *OidcClientHandler) UpdateAuthProvider(w http.ResponseWriter, r *http.Request) {
	provider, ok := validation.ReadBody[OidcAuthProviderDto](w, r)
	if !ok {
		return
	}
	if err := h.AuthProviderService.UpdateProvider(provider); err != nil {
		u.WriteResponseError(w, expectedOidcAuthProviderErrors, err)
		return
	}
}

func (h *OidcClientHandler) ListAuthProviders(w http.ResponseWriter, _ *http.Request) {
	providers, err := h.AuthProviderService.ListProviders()
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}
	u.SendJsonResponse(w, providers)
}

func (h *OidcClientHandler) TestAuthProviderDiscovery(w http.ResponseWriter, r *http.Request) {
	request, ok := validation.ReadBody[OidcAuthProviderDiscoveryRequest](w, r)
	if !ok {
		return
	}
	if err := h.AuthProviderService.TestDiscovery(request.IssuerDomainPath); err != nil {
		u.WriteResponseErrorAlways(w, err)
		return
	}
}

func (h *OidcClientHandler) DeleteAuthProvider(w http.ResponseWriter, r *http.Request) {
	providerIdString, ok := validation.ReadBody[tools.NumberString](w, r)
	if !ok {
		return
	}
	providerId, err := strconv.Atoi(providerIdString.Value)
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}
	if err := h.AuthProviderService.DeleteProvider(providerId); err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}
}

var expectedOidcAuthProviderErrors = u.MapOf(
	OidcAuthProviderNameAlreadyExistsError,
	OidcAuthProviderIssuerDomainPathAlreadyExistsError,
	OidcAuthProviderRequiredFieldMissingError,
)

var expectedOidcLoginErrors = u.MapOf(
	OidcLoginEmailAlreadyExistsError,
	NoValidUsernameClaimError,
	MissingSubjectClaimError,
	MissingOidcAuthorizationCodeError,
	InvalidOidcLoginStateError,
	MissingOidcLoginStateCookieError,
	users.UserDisabledError,
)
