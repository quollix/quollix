//go:build oidc

package oidc

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"server/tests/api_client"
	"server/tools"

	"github.com/quollix/common/assert"
	u "github.com/quollix/common/utils"
)

var oidcHttpClient = &http.Client{
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // #nosec G402: OIDC tests intentionally connect to the local test certificate
		},
	},
	CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

func signInViaOidcHttpClient(t *testing.T, clients *TwoInstanceClients) *api_client.QuollixClient {
	oidcClientLogin := api_client.NewQuollixClientForRootUrl(ClientBaseUrl)
	callbackResponse := sendOidcLoginCallbackRequest(t, clients.ProviderAdmin, oidcClientLogin, clients.ClientAdmin)
	defer u.Close(callbackResponse.Body)

	assert.Equal(t, http.StatusFound, callbackResponse.StatusCode)
	authCookie, err := getAuthCookie(callbackResponse)
	assert.Nil(t, err)
	oidcClientLogin.Parent.Cookie = authCookie
	return oidcClientLogin
}

func sendOidcLoginCallbackRequest(t *testing.T, oidcProviderAdmin *api_client.QuollixClient, oidcClientLogin *api_client.QuollixClient, oidcClientAdmin *api_client.QuollixClient) *http.Response {
	externalProviders, err := oidcClientAdmin.OidcProviders.List()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(externalProviders))
	externalProvider := externalProviders[0]
	start, err := oidcClientLogin.OidcProviders.StartLogin(externalProvider.Id)
	assert.Nil(t, err)

	providerResponse := sendOidcGetRequest(t, oidcProviderAdmin, start.RedirectUrl)
	defer u.Close(providerResponse.Body)
	assert.Equal(t, http.StatusFound, providerResponse.StatusCode)
	callbackUrl := getRedirectLocation(t, providerResponse)
	assert.Equal(t, "quollix."+ClientHost, callbackUrl.Host)

	return sendOidcGetRequest(t, oidcClientLogin, callbackUrl.String())
}

func sendOidcGetRequest(t *testing.T, client *api_client.QuollixClient, targetUrl string) *http.Response {
	req, err := http.NewRequest(http.MethodGet, targetUrl, nil)
	assert.Nil(t, err)
	u.SetCookieHeaders(req, client.Parent)

	resp, err := oidcHttpClient.Do(req)
	assert.Nil(t, err)
	return resp
}

func getAuthCookie(resp *http.Response) (*http.Cookie, error) {
	for _, cookie := range resp.Cookies() {
		if cookie.Name == tools.BrandAppAuthCookieName {
			return cookie, nil
		}
	}
	return nil, fmt.Errorf("auth cookie was not set")
}

func getRedirectLocation(t *testing.T, resp *http.Response) *url.URL {
	location, err := resp.Location()
	assert.Nil(t, err)
	return location
}

func readResponseBody(t *testing.T, resp *http.Response) string {
	body, err := io.ReadAll(resp.Body)
	assert.Nil(t, err)
	return strings.TrimSpace(string(body))
}
