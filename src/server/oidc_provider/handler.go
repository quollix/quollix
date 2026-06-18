package oidc_provider

import (
	"encoding/json"
	"net/http"
	"net/url"
	"server/users"

	u "github.com/quollix/common/utils"
)

type OidcHandler struct {
	Service    OidcService
	UserRepo   users.UserRepository
	AuthHelper u.AuthHelper
}

func (h *OidcHandler) HandleAuthorize(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	user, err := users.GetAuthFromContext(r)
	if err != nil {
		// Preserve the full original /authorize request (including state, nonce, PKCE, etc.) so we can resume the exact same OIDC flow after successful login.
		nextUrl := r.URL.RequestURI()
		http.Redirect(w, r, "/login?next="+url.QueryEscape(nextUrl), http.StatusFound)
		return
	}
	in := AuthorizeInput{
		ResponseType:        q.Get("response_type"),
		ClientID:            q.Get("client_id"),
		RedirectURI:         q.Get("redirect_uri"),
		State:               q.Get("state"),
		Nonce:               q.Get("nonce"),
		CodeChallenge:       q.Get("code_challenge"),
		CodeChallengeMethod: q.Get("code_challenge_method"),
		UserID:              user.Id,
	}

	redirectURI, err := h.Service.Authorize(in)
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}
	http.Redirect(w, r, redirectURI, http.StatusFound)
}

func (h *OidcHandler) HandleDiscovery(w http.ResponseWriter, r *http.Request) {
	cfg, err := h.Service.Discovery()
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}
	writeJSON(w, cfg)
}

func (h *OidcHandler) HandleJWKS(w http.ResponseWriter, r *http.Request) {
	resp := h.Service.GetJwks()
	writeJSON(w, resp)
}

func (h *OidcHandler) HandleToken(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}

	// client credentials can be provided via basic auth ("client_secret_basic") or form parameters ("client_secret_post")
	clientId, clientSecret, ok := r.BasicAuth()
	if !ok {
		clientId = r.Form.Get("client_id")
		clientSecret = r.Form.Get("client_secret")
		if clientId == "" || clientSecret == "" {
			http.Error(w, "missing client authentication", http.StatusBadRequest)
			return
		}
	}

	grantType := r.Form.Get("grant_type")

	switch grantType {
	case "authorization_code":
		in := AuthCodeGrantInput{
			Code:         r.Form.Get("code"),
			RedirectURI:  r.Form.Get("redirect_uri"),
			CodeVerifier: r.Form.Get("code_verifier"),
			ClientID:     clientId,
			ClientSecret: clientSecret,
		}
		res, err := h.Service.TokenWithAuthCode(in)
		if err != nil {
			u.WriteResponseError(w, nil, err)
			return
		}
		writeJSON(w, res)

	case "refresh_token":
		in := RefreshTokenGrantInput{
			RefreshToken: r.Form.Get("refresh_token"),
			ClientID:     clientId,
			ClientSecret: clientSecret,
		}

		res, err := h.Service.TokenWithRefresh(in)
		if err != nil {
			u.WriteResponseError(w, nil, err)
			return
		}
		writeJSON(w, res)

	default:
		http.Error(w, "unsupported grant_type", http.StatusBadRequest)
	}
}

func (h *OidcHandler) HandleUserinfo(w http.ResponseWriter, r *http.Request) {
	token, ok := getBearerToken(r)
	if !ok {
		http.Error(w, "invalid authorization header", http.StatusUnauthorized)
		return
	}

	resp, err := h.Service.Userinfo(token)
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}

	writeJSON(w, resp)
}

func trimLeadingZeros(b []byte) []byte {
	i := 0
	for i < len(b) && b[i] == 0 {
		i++
	}
	return b[i:]
}

func writeJSON(w http.ResponseWriter, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(payload)
	if err != nil {
		u.Logger.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func getBearerToken(r *http.Request) (string, bool) {
	const prefix = "Bearer "
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", false
	}
	if len(authHeader) <= len(prefix) || authHeader[:len(prefix)] != prefix {
		return "", false
	}
	return authHeader[len(prefix):], true
}
