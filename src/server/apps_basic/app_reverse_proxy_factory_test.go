package apps_basic

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"testing"

	"server/tools"

	"github.com/quollix/common/assert"
)

func TestAppReverseProxyFactory_CreateProxyRequest(t *testing.T) {
	factory := newAppReverseProxyFactory(tools.AppForwardedProtoHttps)
	app := AppRequestData{
		Maintainer: "maintainer",
		AppName:    "container",
		Port:       "8080",
	}
	testUrl := fmt.Sprintf("https://originalhost/path?%s=value&other=ok", tools.BrandAppQuerySecretName)
	r := httptest.NewRequest(http.MethodGet, testUrl, nil)
	r.Host = "originalhost"
	r.Header.Set(tools.BrandAppAuthCookieName, "cookievalue")
	authCookie := &http.Cookie{
		Name:  tools.BrandAppAuthCookieName,
		Value: "cookievalue",
	}
	r.AddCookie(authCookie)

	proxy := factory.CreateProxyRequest(r, app)
	r2 := r.Clone(r.Context())
	proxy.Rewrite(&httputil.ProxyRequest{In: r, Out: r2})
	assert.Equal(t, "originalhost", r2.Host)
	assert.Equal(t, "http", r2.URL.Scheme)
	assert.Equal(t, "maintainer_container_container:8080", r2.URL.Host)
	assert.Equal(t, "/path", r2.URL.Path)
	assert.Equal(t, "other=ok", r2.URL.RawQuery)
	assert.Equal(t, "originalhost", r2.Header.Get("X-Forwarded-Host"))
	assert.Equal(t, "https", r2.Header.Get("X-Forwarded-Proto"))
	assert.Equal(t, "", r2.Header.Get(tools.BrandAppAuthCookieName))
	assert.Equal(t, 0, len(r2.Cookies()))
}

func TestAppReverseProxyFactory_CreateProxyRequestUsesConfiguredForwardedProto(t *testing.T) {
	factory := newAppReverseProxyFactory(tools.AppForwardedProtoHttp)
	app := AppRequestData{
		Maintainer: "maintainer",
		AppName:    "container",
		Port:       "8080",
	}
	r := httptest.NewRequest(http.MethodGet, "https://originalhost/path", nil)
	r.Host = "originalhost"

	proxy := factory.CreateProxyRequest(r, app)
	r2 := r.Clone(r.Context())
	proxy.Rewrite(&httputil.ProxyRequest{In: r, Out: r2})

	assert.Equal(t, tools.AppForwardedProtoHttp, r2.Header.Get("X-Forwarded-Proto"))
}

func newAppReverseProxyFactory(appForwardedProto string) *AppReverseProxyFactoryImpl {
	return &AppReverseProxyFactoryImpl{
		Config: &tools.GlobalConfig{
			AppForwardedProto: appForwardedProto,
		},
	}
}

func TestRemoveAuthCookie_RemovesAuthCookieAndKeepsOtherCookies(t *testing.T) {
	originalRequest := httptest.NewRequest(http.MethodGet, "/", nil)
	proxyRequest := originalRequest.Clone(originalRequest.Context())
	authCookie := &http.Cookie{
		Name:  tools.BrandAppAuthCookieName,
		Value: "cookievalue",
	}
	otherCookie := &http.Cookie{
		Name:  "other-cookie",
		Value: "other-value",
	}
	originalRequest.AddCookie(authCookie)
	originalRequest.AddCookie(otherCookie)
	proxyRequest.Header.Set("Cookie", originalRequest.Header.Get("Cookie"))

	removeAuthCookie(proxyRequest, originalRequest)

	cookies := proxyRequest.Cookies()
	assert.Equal(t, 1, len(cookies))
	assert.Equal(t, otherCookie.Name, cookies[0].Name)
	assert.Equal(t, otherCookie.Value, cookies[0].Value)
}

func TestRemoveAuthCookie_RemovesCookieHeaderWhenOnlyAuthCookiesExist(t *testing.T) {
	originalRequest := httptest.NewRequest(http.MethodGet, "/", nil)
	proxyRequest := originalRequest.Clone(originalRequest.Context())
	authCookie := &http.Cookie{
		Name:  tools.BrandAppAuthCookieName,
		Value: "cookievalue",
	}
	originalRequest.AddCookie(authCookie)
	proxyRequest.Header.Set("Cookie", originalRequest.Header.Get("Cookie"))

	removeAuthCookie(proxyRequest, originalRequest)

	assert.Equal(t, "", proxyRequest.Header.Get("Cookie"))
	assert.Equal(t, 0, len(proxyRequest.Cookies()))
}
