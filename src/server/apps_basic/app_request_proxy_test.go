package apps_basic

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"server/tools"
	"strings"
	"testing"

	"github.com/quollix/common/assert"
	api "github.com/quollix/common/quollix/api"
	u "github.com/quollix/common/utils"
)

type appRequestProxyTestDependencies struct {
	appRequestResolver *AppRequestResolverMock
	appSessionService  *AppSessionServiceMock
	appRequestParser   *AppRequestParserMock
	proxy              *AppRequestProxy
}

func newAppRequestProxyTestDependencies(t *testing.T) *appRequestProxyTestDependencies {
	appRequestResolver := NewAppRequestResolverMock(t)
	appSessionService := NewAppSessionServiceMock(t)
	appRequestParser := NewAppRequestParserMock(t)
	return &appRequestProxyTestDependencies{
		appRequestResolver: appRequestResolver,
		appSessionService:  appSessionService,
		appRequestParser:   appRequestParser,
		proxy: &AppRequestProxy{
			AppRequestResolver: appRequestResolver,
			AppSessionService:  appSessionService,
			AppRequestParser:   appRequestParser,
		},
	}
}

func getSampleAppRequestData() *AppRequestData {
	return &AppRequestData{
		Maintainer: "maintainer",
		AppName:    "sample-app",
	}
}

func getResolvedSampleAppRequest() *ResolvedAppRequest {
	return &ResolvedAppRequest{
		BaseDomain: "example.com",
		AppName:    "sample-app",
		App:        getSampleAppRequestData(),
		AppExists:  true,
	}
}

func getResolvedUnknownAppRequest() *ResolvedAppRequest {
	return &ResolvedAppRequest{
		BaseDomain: "example.com",
		AppName:    "unknown-app",
		AppExists:  false,
	}
}

func TestAppRequestProxy_ExchangeSecretSetsCookieAndRedirectsWithoutSecret(t *testing.T) {
	appSessionService := NewAppSessionServiceMock(t)
	proxy := &AppRequestProxy{
		AppSessionService: appSessionService,
	}
	app := &AppRequestData{
		Maintainer: "maintainer",
		AppName:    "sample-app",
	}
	cookie := &http.Cookie{
		Name:     api.BrandAppAuthCookieName,
		Value:    "app-cookie-value",
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}
	appSessionService.EXPECT().
		CreateAppSessionCookieFromSecret("secret-value", app).
		Return(cookie, nil)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/path?secret=secret-value", nil)
	err := proxy.exchangeSecretAgainstAuthenticationCookieAndInstructBrowserToRepeatThatRequest(response, request, "secret-value", app)
	assert.Nil(t, err)

	result := response.Result()
	defer u.Close(result.Body)
	assert.Equal(t, http.StatusFound, result.StatusCode)
	cookies := result.Cookies()
	assert.Equal(t, 1, len(cookies))
	assert.Equal(t, cookie.Value, cookies[0].Value)
	assert.Equal(t, "/path", result.Header.Get("Location"))
}

func TestAppRequestProxy_MissingAppCookieRedirectsToBrandAppOpen(t *testing.T) {
	deps := newAppRequestProxyTestDependencies(t)
	request := httptest.NewRequest(http.MethodGet, "https://sample-app.example.com/docs/view?id=123", nil)
	response := httptest.NewRecorder()
	resolvedAppRequest := getResolvedSampleAppRequest()
	deps.appRequestResolver.EXPECT().ResolveAppRequest(request).Return(resolvedAppRequest, nil)
	deps.appRequestParser.EXPECT().GetQuerySecret(request).Return("", false, true)
	deps.appSessionService.EXPECT().AuthorizeAppRequest(request, resolvedAppRequest.App).Return(AppRequestMissingSession, nil)

	deps.proxy.ProxyRequestToTheAppsDockerContainer(response, request)

	result := response.Result()
	defer u.Close(result.Body)
	assert.Equal(t, http.StatusFound, result.StatusCode)
	location, err := url.Parse(result.Header.Get("Location"))
	assert.Nil(t, err)
	assert.Equal(t, "https", location.Scheme)
	assert.Equal(t, tools.BrandAppDomainPrefix+"example.com", location.Host)
	assert.Equal(t, api.Paths.FrontendAppOpen, location.Path)
	assert.Equal(t, "sample-app", location.Query().Get("app"))
	assert.Equal(t, "/docs/view?id=123", location.Query().Get("path"))
}

func TestAppRequestProxy_AccessDeniedReturnsAppUnavailablePage(t *testing.T) {
	deps := newAppRequestProxyTestDependencies(t)
	request := httptest.NewRequest(http.MethodGet, "https://sample-app.example.com/docs/view", nil)
	response := httptest.NewRecorder()
	resolvedAppRequest := getResolvedSampleAppRequest()
	deps.appRequestResolver.EXPECT().ResolveAppRequest(request).Return(resolvedAppRequest, nil)
	deps.appRequestParser.EXPECT().GetQuerySecret(request).Return("", false, true)
	deps.appSessionService.EXPECT().AuthorizeAppRequest(request, resolvedAppRequest.App).Return(AppRequestAccessDenied, nil)

	deps.proxy.ProxyRequestToTheAppsDockerContainer(response, request)

	result := response.Result()
	defer u.Close(result.Body)
	assert.Equal(t, http.StatusForbidden, result.StatusCode)
	assert.Equal(t, "text/html; charset=utf-8", result.Header.Get("Content-Type"))
	assert.True(t, strings.Contains(response.Body.String(), tools.AppUnavailableTitle))
	assert.True(t, strings.Contains(response.Body.String(), "https://"+tools.BrandAppDomainPrefix+"example.com"+api.Paths.FrontendInstalledApps))
}

func TestAppRequestProxy_UnknownAppReturnsAppUnavailablePage(t *testing.T) {
	deps := newAppRequestProxyTestDependencies(t)
	request := httptest.NewRequest(http.MethodGet, "https://unknown-app.example.com/docs/view", nil)
	response := httptest.NewRecorder()
	deps.appRequestResolver.EXPECT().ResolveAppRequest(request).Return(getResolvedUnknownAppRequest(), nil)

	deps.proxy.ProxyRequestToTheAppsDockerContainer(response, request)

	result := response.Result()
	defer u.Close(result.Body)
	assert.Equal(t, http.StatusForbidden, result.StatusCode)
	assert.Equal(t, "text/html; charset=utf-8", result.Header.Get("Content-Type"))
	assert.True(t, strings.Contains(response.Body.String(), tools.AppUnavailableTitle))
	assert.True(t, strings.Contains(response.Body.String(), tools.AppUnavailableMessage))
	assert.True(t, strings.Contains(response.Body.String(), "https://"+tools.BrandAppDomainPrefix+"example.com"+api.Paths.FrontendInstalledApps))
}
