//go:build component

package component

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"server/oidc_provider"
	"server/tests/api_client"
	"server/tools"

	"github.com/quollix/common/assert"
	u "github.com/quollix/common/utils"
)

const (
	baseURL         = "https://quollix.localhost"
	testRedirectURI = "https://sampleapp.localhost/callback"
	adminUsername   = "admin"
	adminPassword   = "password"
)

type OidcTestClient struct {
	T       *testing.T
	BaseURL string
	Config  oidc_provider.DiscoveryConfig
	Client  *http.Client
	Quollix *api_client.QuollixClient
}

func NewOidcTestClient(t *testing.T) *OidcTestClient {
	return NewOidcTestClientForBaseURL(t, baseURL)
}

func NewOidcTestClientForBaseURL(t *testing.T, baseURL string) *OidcTestClient {
	quollixClient := api_client.NewQuollixClientForRootUrl(baseURL)
	browser := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, // #nosec G402 (CWE-295): TLS InsecureSkipVerify set true.
			},
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	req, err := http.NewRequest("GET", baseURL+tools.Paths.BackendOidcWellKnown, nil)
	assert.Nil(t, err)
	resp, err := browser.Do(req)
	assert.Nil(t, err)
	defer u.Close(resp.Body)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var cfg oidc_provider.DiscoveryConfig
	assert.Nil(t, json.NewDecoder(resp.Body).Decode(&cfg))
	assert.Equal(t, baseURL, cfg.Issuer)

	return &OidcTestClient{
		T:       t,
		BaseURL: baseURL,
		Config:  cfg,
		Client:  browser,
		Quollix: quollixClient,
	}
}

func FetchPublicKeyFromJWKS(t *testing.T, ctx *OidcTestClient) *rsa.PublicKey {
	req, err := http.NewRequest("GET", ctx.Config.JwksURI, nil)
	assert.Nil(t, err)
	resp, err := ctx.Client.Do(req)
	assert.Nil(t, err)
	defer u.Close(resp.Body)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	jwks := oidc_provider.JWKS{}
	assert.Nil(t, json.NewDecoder(resp.Body).Decode(&jwks))
	assert.True(t, len(jwks.Keys) > 0)

	k := jwks.Keys[0]
	nb, err := base64.RawURLEncoding.DecodeString(k.N)
	assert.Nil(t, err)
	eb, err := base64.RawURLEncoding.DecodeString(k.E)
	assert.Nil(t, err)

	n := new(big.Int).SetBytes(nb)
	e := new(big.Int).SetBytes(eb)

	return &rsa.PublicKey{
		N: n,
		E: int(e.Int64()),
	}
}

func VerifyIDToken(t *testing.T, token string, pub *rsa.PublicKey, expectedIss, expectedAud string) oidc_provider.IDTokenClaims {
	parts := strings.Split(token, ".")
	assert.Equal(t, 3, len(parts))
	hb, err := base64.RawURLEncoding.DecodeString(parts[0])
	assert.Nil(t, err)
	cb, err := base64.RawURLEncoding.DecodeString(parts[1])
	assert.Nil(t, err)
	sb, err := base64.RawURLEncoding.DecodeString(parts[2])
	assert.Nil(t, err)

	var header map[string]any
	assert.Nil(t, json.Unmarshal(hb, &header))
	assert.Equal(t, "RS256", header["alg"])
	var claims oidc_provider.IDTokenClaims
	assert.Nil(t, json.Unmarshal(cb, &claims))

	unsigned := parts[0] + "." + parts[1]
	hash := sha256.Sum256([]byte(unsigned))
	assert.Nil(t, rsa.VerifyPKCS1v15(pub, crypto.SHA256, hash[:], sb))

	assert.Equal(t, expectedIss, claims.Iss)
	assert.Equal(t, expectedAud, claims.Aud)
	assert.True(t, claims.Exp > time.Now().Unix())
	assert.NotEqual(t, 0, claims.Exp)
	assert.True(t, claims.Iat <= time.Now().Unix())
	return claims
}

func (c *OidcTestClient) SignInAsAdmin() {
	c.SignIn(adminUsername, adminPassword)
}

func (c *OidcTestClient) SignIn(username, password string) {
	assert.Nil(c.T, c.Quollix.Auth.SignIn(username, password))
	assert.NotNil(c.T, c.Quollix.Parent.Cookie)
}

func (c *OidcTestClient) AuthorizeWithPKCE(clientId string) (oidc_provider.AuthCodeResult, string) {
	return c.AuthorizeWithPKCERedirectURI(clientId, testRedirectURI)
}

func (c *OidcTestClient) AuthorizeWithPKCERedirectURI(clientId string, redirectURI string) (oidc_provider.AuthCodeResult, string) {
	resp, verifier := c.authorizeWithPKCERaw(clientId, redirectURI)
	defer u.Close(resp.Body)
	assert.Equal(c.T, http.StatusFound, resp.StatusCode)

	loc, err := resp.Location()
	assert.Nil(c.T, err)
	expectedRedirectURI, err := url.Parse(redirectURI)
	assert.Nil(c.T, err)
	assert.Equal(c.T, expectedRedirectURI.Scheme, loc.Scheme)
	assert.Equal(c.T, expectedRedirectURI.Host, loc.Host)
	assert.Equal(c.T, expectedRedirectURI.Path, loc.Path)

	code := loc.Query().Get("code")
	assert.NotEqual(c.T, "", code)
	return oidc_provider.AuthCodeResult{
		Code:  code,
		State: "state-default",
		Nonce: "nonce-default",
	}, verifier
}

func (c *OidcTestClient) authorizeWithPKCERaw(clientId string, redirectURI string) (*http.Response, string) {
	state := "state-default"
	nonce := "nonce-default"

	params := url.Values{}
	params.Set("response_type", "code")
	params.Set("client_id", clientId)
	params.Set("redirect_uri", redirectURI)
	params.Set("scope", "openid")
	params.Set("state", state)
	params.Set("nonce", nonce)

	verifier, challenge := pkceValues(c.T)
	params.Set("code_challenge", challenge)
	params.Set("code_challenge_method", "S256")

	resp := c.doAuthorize(params)
	return resp, verifier
}

func pkceValues(t *testing.T) (verifier string, challenge string) {
	var err error
	authHelper := u.AuthHelperImpl{}
	verifier, err = authHelper.GenerateSecret()
	assert.Nil(t, err)
	sum := sha256.Sum256([]byte(verifier))
	challenge = base64.RawURLEncoding.EncodeToString(sum[:])
	return
}

func (c *OidcTestClient) ExchangeCodeForTokens(code, verifier, clientId, clientSecret string, clientAuthMethod ClientAuthMethod) oidc_provider.TokenResponse {
	return c.ExchangeCodeForTokensWithRedirectURI(code, verifier, testRedirectURI, clientId, clientSecret, clientAuthMethod)
}

func (c *OidcTestClient) ExchangeCodeForTokensWithRedirectURI(code, verifier, redirectURI, clientId, clientSecret string, clientAuthMethod ClientAuthMethod) oidc_provider.TokenResponse {
	resp := c.exchangeCodeForTokensRaw(code, verifier, redirectURI, clientId, clientSecret, clientAuthMethod)
	defer u.Close(resp.Body)
	assert.Equal(c.T, http.StatusOK, resp.StatusCode)

	var tokenResponse oidc_provider.TokenResponse
	assert.Nil(c.T, json.NewDecoder(resp.Body).Decode(&tokenResponse))

	assert.NotEqual(c.T, "", tokenResponse.AccessToken)
	assert.NotEqual(c.T, "", tokenResponse.IDToken)
	assert.NotEqual(c.T, "", tokenResponse.RefreshToken)
	assert.Equal(c.T, "Bearer", tokenResponse.TokenType)
	assert.True(c.T, tokenResponse.ExpiresIn > 0)
	assert.True(c.T, tokenResponse.ExpiresIn <= 600)
	return tokenResponse
}

func (c *OidcTestClient) exchangeCodeForTokensRaw(code, verifier, redirectURI, clientId, clientSecret string, clientAuthMethod ClientAuthMethod) *http.Response {
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", code)
	form.Set("redirect_uri", redirectURI)
	form.Set("code_verifier", verifier)

	req, err := http.NewRequest("POST", c.Config.TokenEndpoint, nil)
	assert.Nil(c.T, err)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	applyClientAuthenticationToTokenRequest(req, form, clientId, clientSecret, clientAuthMethod)
	req.Body = io.NopCloser(strings.NewReader(form.Encode()))
	req.ContentLength = int64(len(form.Encode()))

	resp, err := c.Client.Do(req)
	assert.Nil(c.T, err)
	return resp
}

func (c *OidcTestClient) FetchUserinfo(accessToken string) oidc_provider.UserinfoResponse {
	req, err := http.NewRequest("GET", c.Config.UserinfoEndpoint, nil)
	assert.Nil(c.T, err)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := c.Client.Do(req)
	assert.Nil(c.T, err)
	defer u.Close(resp.Body)
	assert.Equal(c.T, http.StatusOK, resp.StatusCode)

	var uinfo oidc_provider.UserinfoResponse
	assert.Nil(c.T, json.NewDecoder(resp.Body).Decode(&uinfo))
	return uinfo
}

func (c *OidcTestClient) doAuthorize(params url.Values) *http.Response {
	authURL := c.Config.AuthorizationEndpoint + "?" + params.Encode()
	req, err := http.NewRequest(http.MethodGet, authURL, nil)
	assert.Nil(c.T, err)
	u.SetCookieHeaders(req, c.Quollix.Parent)
	resp, err := c.Client.Do(req)
	assert.Nil(c.T, err)
	return resp
}

func (c *OidcTestClient) do(req *http.Request) *http.Response {
	resp, err := c.Client.Do(req) // #nosec G704: OIDC tests intentionally issue requests to locally constructed test endpoints
	assert.Nil(c.T, err)
	return resp
}

type ClientAuthMethod string

const (
	ClientAuthMethodBasic ClientAuthMethod = "client_secret_basic"
	ClientAuthMethodPost  ClientAuthMethod = "client_secret_post"
)

func applyClientAuthenticationToTokenRequest(req *http.Request, form url.Values, clientId string, clientSecret string, clientAuthMethod ClientAuthMethod) {
	if clientAuthMethod == ClientAuthMethodBasic {
		authHeader := base64.StdEncoding.EncodeToString([]byte(clientId + ":" + clientSecret))
		req.Header.Set("Authorization", "Basic "+authHeader)
		return
	}

	form.Set("client_id", clientId)
	form.Set("client_secret", clientSecret)
}
