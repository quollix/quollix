package apps_basic

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"strings"

	"server/tools"

	api "github.com/quollix/common/quollix/api"
	u "github.com/quollix/common/utils"
)

type AppReverseProxyFactory interface {
	CreateProxyRequest(originalRequest *http.Request, app AppRequestData) *httputil.ReverseProxy
}

type AppReverseProxyFactoryImpl struct {
	Config *tools.GlobalConfig
}

func (a *AppReverseProxyFactoryImpl) CreateProxyRequest(originalRequest *http.Request, app AppRequestData) *httputil.ReverseProxy {
	upstream := fmt.Sprintf("%s_%s_%s:%s", app.Maintainer, app.AppName, app.AppName, app.Port)
	errorHandler := appProxyErrorHandler{appName: app.AppName, upstream: upstream}
	proxy := &httputil.ReverseProxy{
		ErrorHandler: errorHandler.ServeHTTPError,
		ErrorLog:     log.New(errorHandler, "", 0),
	}
	proxy.Rewrite = func(proxyRequest *httputil.ProxyRequest) {
		newProxyRequest := proxyRequest.Out
		newProxyRequest.Header.Set("X-Forwarded-Host", originalRequest.Host)
		newProxyRequest.Header.Set("X-Forwarded-Proto", a.Config.AppForwardedProto)
		newProxyRequest.Host = originalRequest.Host
		newProxyRequest.URL.Scheme = "http"
		newProxyRequest.URL.Host = upstream
		query := newProxyRequest.URL.Query()
		query.Del(api.BrandAppQuerySecretName)
		newProxyRequest.URL.RawQuery = query.Encode()
		newProxyRequest.Header.Del(api.BrandAppAuthCookieName)

		removeAuthCookie(newProxyRequest, originalRequest)
	}
	return proxy
}

type appProxyErrorHandler struct {
	appName  string
	upstream string
}

func (h appProxyErrorHandler) ServeHTTPError(w http.ResponseWriter, r *http.Request, err error) {
	if r.Context().Err() != nil {
		return
	}
	u.Logger.Error(err, "app", h.appName, "upstream", h.upstream)
	w.WriteHeader(http.StatusBadGateway)
}

func (h appProxyErrorHandler) Write(p []byte) (int, error) {
	u.Logger.Error(strings.TrimSpace(string(p)), "app", h.appName, "upstream", h.upstream)
	return len(p), nil
}

func removeAuthCookie(proxyRequest *http.Request, originalRequest *http.Request) {
	var newCookies []string
	for _, cookie := range originalRequest.Cookies() {
		if cookie.Name != api.BrandAppAuthCookieName {
			newCookies = append(newCookies, cookie.String())
		}
	}
	if len(newCookies) > 0 {
		proxyRequest.Header.Set("Cookie", strings.Join(newCookies, "; "))
	} else {
		proxyRequest.Header.Del("Cookie")
	}
}
