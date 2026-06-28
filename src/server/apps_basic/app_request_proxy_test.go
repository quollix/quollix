package apps_basic

import (
	"net/http"
	"net/http/httptest"
	"server/tools"
	"testing"

	"github.com/quollix/common/assert"
	u "github.com/quollix/common/utils"
)

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
		Name:     tools.BrandAppAuthCookieName,
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
