//go:build component

package component

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"server/oidc_provider"
	"server/tools"
	"strconv"
	"strings"
	"testing"

	"github.com/quollix/common/assert"
	u "github.com/quollix/common/utils"
)

func TestOIDC_HappyPath_AuthorizationCodeFlowWithUserinfo(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()
	expectedSub := strconv.Itoa(client.Users.GetByUsername(tools.DefaultAdminName).Id)

	app, err := client.Apps.InstallSample("2.0")
	assert.Nil(t, err)

	ctx := NewTestContext(t)
	ctx.LoginAsAdmin()
	authRes, verifier := ctx.AuthorizeWithPKCE("openid", app.ClientId)
	tokens := ctx.ExchangeCodeForTokens(authRes.Code, verifier, app.ClientId, app.ClientSecret, ClientAuthMethodBasic)
	pubKey := FetchPublicKeyFromJWKS(t, ctx)
	claims := VerifyIDToken(t, tokens.IDToken, pubKey, ctx.Config.Issuer, app.ClientId)
	assert.Equal(t, expectedSub, claims.Sub)
	assert.Equal(t, "admin", claims.Role)
	assert.Equal(t, authRes.Nonce, claims.Nonce)

	assert.Equal(t, adminUsername, claims.PreferredUsername)
	assert.Equal(t, tools.DefaultAdminEmail, claims.Email)
	assert.Equal(t, adminUsername, claims.Name)

	uinfo := ctx.FetchUserinfo(tokens.AccessToken)
	assert.Equal(t, expectedSub, uinfo.Sub)
	assert.Equal(t, "admin", uinfo.Role)

	assert.Equal(t, adminUsername, uinfo.PreferredUsername)
	assert.Equal(t, tools.DefaultAdminEmail, uinfo.Email)
	assert.Equal(t, adminUsername, uinfo.Name)
}

func TestOIDC_Discovery(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	ctx := NewTestContext(t)
	assert.Equal(t, baseURL, ctx.Config.Issuer)
	assert.Equal(t, baseURL+tools.Paths.BackendOidcAuthorize, ctx.Config.AuthorizationEndpoint)
	assert.Equal(t, baseURL+tools.Paths.BackendOidcToken, ctx.Config.TokenEndpoint)
	assert.Equal(t, baseURL+tools.Paths.BackendOidcJwks, ctx.Config.JwksURI)
	assert.Equal(t, []string{"code"}, ctx.Config.ResponseTypesSupported)
	assert.Equal(t, []string{"public"}, ctx.Config.SubjectTypesSupported)
	assert.Equal(t, []string{"RS256"}, ctx.Config.IDTokenSigningAlgValuesSupported)
	assert.Equal(t, []string{"openid", "profile", "email", "groups", "offline_access"}, ctx.Config.ScopesSupported)
	assert.Equal(t, []string{"client_secret_basic", "client_secret_post"}, ctx.Config.TokenEndpointAuthMethodsSupported)
	assert.Equal(t, []string{"S256"}, ctx.Config.CodeChallengeMethodsSupported)
}

func TestOIDC_JWKS(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	ctx := NewTestContext(t)
	assert.Equal(t, baseURL, ctx.Config.Issuer)
	assert.Equal(t, baseURL+tools.Paths.BackendOidcJwks, ctx.Config.JwksURI)

	req, err := http.NewRequest("GET", ctx.Config.JwksURI, nil)
	assert.Nil(t, err)
	resp := ctx.do(req)
	defer u.Close(resp.Body)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

	var jwks = oidc_provider.JWKS{}
	assert.Nil(t, json.NewDecoder(resp.Body).Decode(&jwks))

	assert.Equal(t, 1, len(jwks.Keys))
	k := jwks.Keys[0]
	assert.Equal(t, "RSA", k.Kty)
	assert.Equal(t, "RS256", k.Alg)
	assert.Equal(t, "sig", k.Use)
	assert.Equal(t, "key-1", k.Kid)

	nb, err := base64.RawURLEncoding.DecodeString(k.N)
	assert.Nil(t, err)
	eb, err := base64.RawURLEncoding.DecodeString(k.E)
	assert.Nil(t, err)

	n := new(big.Int).SetBytes(nb)
	e := new(big.Int).SetBytes(eb)

	assert.True(t, n.Sign() > 0)
	assert.True(t, e.Sign() > 0)
}

func TestOIDC_AuthorizeRedirectsToLoginWhenUnauthenticated(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	app, err := client.Apps.InstallSample("2.0")
	assert.Nil(t, err)

	ctx := NewTestContext(t)

	params := url.Values{}
	params.Set("response_type", "code")
	params.Set("client_id", app.ClientId)
	params.Set("redirect_uri", testRedirectURI)
	params.Set("scope", "openid")
	params.Set("state", "state-unauth")
	params.Set("nonce", "nonce-unauth")

	_, challenge := pkceValues(t)
	params.Set("code_challenge", challenge)
	params.Set("code_challenge_method", "S256")

	resp := ctx.doAuthorize(params)
	defer u.Close(resp.Body)

	assert.Equal(t, http.StatusFound, resp.StatusCode)
	loc, err := resp.Location()
	assert.Nil(t, err)
	assert.Equal(t, "https", loc.Scheme)
	assert.Equal(t, "quollix.localhost", loc.Host)
	assert.Equal(t, "/login", loc.Path)

	expectedNext := tools.Paths.BackendOidcAuthorize + "?" + params.Encode()
	assert.Equal(t, expectedNext, loc.Query().Get("next"))
}

func TestOIDC_Userinfo_WithoutTokenCausesUnauthorizedCode(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	ctx := NewTestContext(t)
	req, err := http.NewRequest("GET", ctx.Config.UserinfoEndpoint, nil)
	assert.Nil(t, err)
	resp := ctx.do(req)
	defer u.Close(resp.Body)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestOIDC_RefreshTokenFlow(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()
	expectedSub := strconv.Itoa(client.Users.GetByUsername(tools.DefaultAdminName).Id)

	app, err := client.Apps.InstallSample("2.0")
	assert.Nil(t, err)

	ctx := NewTestContext(t)
	ctx.LoginAsAdmin()
	authRes, verifier := ctx.AuthorizeWithPKCE("openid", app.ClientId)
	initialTokens := ctx.ExchangeCodeForTokens(authRes.Code, verifier, app.ClientId, app.ClientSecret, ClientAuthMethodBasic)
	refreshedTokens := ctx.refreshTokens(initialTokens.RefreshToken, app.ClientId, app.ClientSecret, ClientAuthMethodBasic)
	// optional: ensure we got a new access token
	assert.NotEqual(t, initialTokens.AccessToken, refreshedTokens.AccessToken)
	pubKey := FetchPublicKeyFromJWKS(t, ctx)
	claims := VerifyIDToken(t, refreshedTokens.IDToken, pubKey, ctx.Config.Issuer, app.ClientId)
	assert.Equal(t, expectedSub, claims.Sub)
	assert.Equal(t, adminUsername, claims.PreferredUsername)
	assert.Equal(t, tools.DefaultAdminEmail, claims.Email)
	assert.Equal(t, adminUsername, claims.Name)

	refreshedAgainTokens := ctx.refreshTokens(refreshedTokens.RefreshToken, app.ClientId, app.ClientSecret, ClientAuthMethodBasic)
	assert.NotEqual(t, refreshedTokens.AccessToken, refreshedAgainTokens.AccessToken)
	assert.NotEqual(t, refreshedTokens.RefreshToken, refreshedAgainTokens.RefreshToken)

	respOld := ctx.refreshTokensRaw(refreshedTokens.RefreshToken, app.ClientId, app.ClientSecret, ClientAuthMethodBasic)
	defer u.Close(respOld.Body)
	assert.Equal(t, http.StatusBadRequest, respOld.StatusCode)

	resp := ctx.refreshTokensRaw(initialTokens.RefreshToken, app.ClientId, app.ClientSecret, ClientAuthMethodBasic)
	defer u.Close(resp.Body)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func runAuthorizationCodeFlowWithAuthMethod(t *testing.T, clientAuthMethod ClientAuthMethod) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()
	expectedSub := strconv.Itoa(client.Users.GetByUsername(tools.DefaultAdminName).Id)

	app, err := client.Apps.InstallSample("2.0")
	assert.Nil(t, err)

	ctx := NewTestContext(t)
	ctx.LoginAsAdmin()
	authRes, verifier := ctx.AuthorizeWithPKCE("openid", app.ClientId)
	tokens := ctx.ExchangeCodeForTokens(authRes.Code, verifier, app.ClientId, app.ClientSecret, clientAuthMethod)

	pubKey := FetchPublicKeyFromJWKS(t, ctx)
	claims := VerifyIDToken(t, tokens.IDToken, pubKey, ctx.Config.Issuer, app.ClientId)

	assert.Equal(t, expectedSub, claims.Sub)
	assert.Equal(t, authRes.Nonce, claims.Nonce)

	uinfo := ctx.FetchUserinfo(tokens.AccessToken)
	assert.Equal(t, expectedSub, uinfo.Sub)
	assert.Equal(t, adminUsername, uinfo.PreferredUsername)
	assert.Equal(t, tools.DefaultAdminEmail, uinfo.Email)
	assert.Equal(t, adminUsername, uinfo.Name)
}

func TestOIDC_HappyPath_AuthorizationCodeFlow_ClientSecretBasic(t *testing.T) {
	runAuthorizationCodeFlowWithAuthMethod(t, ClientAuthMethodBasic)
}

func TestOIDC_HappyPath_AuthorizationCodeFlow_ClientSecretPost(t *testing.T) {
	runAuthorizationCodeFlowWithAuthMethod(t, ClientAuthMethodPost)
}

func (c *TestContext) refreshTokens(refreshToken, clientId, clientSecret string, clientAuthMethod ClientAuthMethod) oidc_provider.TokenResponse {
	resp := c.refreshTokensRaw(refreshToken, clientId, clientSecret, clientAuthMethod)
	defer u.Close(resp.Body)
	assert.Equal(c.T, http.StatusOK, resp.StatusCode)

	var tokenResponse oidc_provider.TokenResponse
	assert.Nil(c.T, json.NewDecoder(resp.Body).Decode(&tokenResponse))

	assert.NotEqual(c.T, "", tokenResponse.AccessToken)
	assert.NotEqual(c.T, "", tokenResponse.IDToken)
	assert.NotEqual(c.T, "", tokenResponse.RefreshToken)
	return tokenResponse
}

func (c *TestContext) refreshTokensRaw(refreshToken, clientId, clientSecret string, clientAuthMethod ClientAuthMethod) *http.Response {
	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", refreshToken)

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
