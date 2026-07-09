package oidc_client

import (
	"net/http"
	"testing"

	"server/tools"

	"github.com/quollix/common/assert"
)

func TestNewOidcProviderClient_ProdUsesDefaultHttpClient(t *testing.T) {
	client := NewOidcProviderClient(&tools.GlobalConfig{
		AllowInsecureOidcProviderTls: false,
	}).(*OidcProviderClientImpl)

	assert.Equal(t, http.DefaultClient, client.httpClient)
}

func TestNewOidcProviderClient_DevAllowsSelfSignedCertificates(t *testing.T) {
	client := NewOidcProviderClient(&tools.GlobalConfig{
		AllowInsecureOidcProviderTls: true,
	}).(*OidcProviderClientImpl)

	transport := client.httpClient.Transport.(*http.Transport)
	assert.NotEqual(t, http.DefaultTransport.(*http.Transport).TLSClientConfig, transport.TLSClientConfig)
	assert.NotEqual(t, nil, transport.TLSClientConfig)
	assert.Equal(t, true, transport.TLSClientConfig.InsecureSkipVerify)
}
