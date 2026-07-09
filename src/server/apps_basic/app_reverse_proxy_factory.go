package apps_basic

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"strings"

	"server/tools"
)

type AppReverseProxyFactory interface {
	CreateProxyRequest(originalRequest *http.Request, app AppRequestData) *httputil.ReverseProxy
}

type AppReverseProxyFactoryImpl struct {
	Config *tools.GlobalConfig
}

func (a *AppReverseProxyFactoryImpl) CreateProxyRequest(originalRequest *http.Request, app AppRequestData) *httputil.ReverseProxy {
	proxy := &httputil.ReverseProxy{}
	proxy.Rewrite = func(proxyRequest *httputil.ProxyRequest) {
		newProxyRequest := proxyRequest.Out
		newProxyRequest.Header.Set("X-Forwarded-Host", originalRequest.Host)
		newProxyRequest.Header.Set("X-Forwarded-Proto", a.Config.AppForwardedProto)
		newProxyRequest.Host = originalRequest.Host
		newProxyRequest.URL.Scheme = "http"
		newProxyRequest.URL.Host = fmt.Sprintf("%s_%s_%s:%s", app.Maintainer, app.AppName, app.AppName, app.Port)
		query := newProxyRequest.URL.Query()
		query.Del(tools.BrandAppQuerySecretName)
		newProxyRequest.URL.RawQuery = query.Encode()
		newProxyRequest.Header.Del(tools.BrandAppAuthCookieName)

		removeAuthCookie(newProxyRequest, originalRequest)
	}
	return proxy
}

func removeAuthCookie(proxyRequest *http.Request, originalRequest *http.Request) {
	var newCookies []string
	for _, cookie := range originalRequest.Cookies() {
		if cookie.Name != tools.BrandAppAuthCookieName {
			newCookies = append(newCookies, cookie.String())
		}
	}
	if len(newCookies) > 0 {
		proxyRequest.Header.Set("Cookie", strings.Join(newCookies, "; "))
	} else {
		proxyRequest.Header.Del("Cookie")
	}
}
