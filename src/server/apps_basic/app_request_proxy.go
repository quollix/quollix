package apps_basic

import (
	"net/http"
	"server/configs"
	"server/users"
	"time"

	u "github.com/quollix/common/utils"
)

type AppRequestProxy struct {
	ConfigsService         configs.ConfigsService
	AppRepo                AppRepository
	AppSessionService      AppSessionService
	AppRequestParser       AppRequestParser
	AppReverseProxyFactory AppReverseProxyFactory
}

var (
	expectedAccessDeniedErrors       = u.MapOf(AccessDeniedError)
	expectedSecretDoesNotExistErrors = u.MapOf(users.SecretDoesNotExistError)
)

func (a *AppRequestProxy) ProxyRequestToTheAppsDockerContainer(w http.ResponseWriter, r *http.Request) {
	u.Logger.Debug("Proxying request")
	requestHost := a.AppRequestParser.GetHostFromRequestHost(r.Host)
	baseDomain, err := a.ConfigsService.GetBaseDomain()
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}

	appName, err := a.AppRequestParser.GetAppNameFromRequestHost(requestHost, baseDomain)
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}

	app, err := a.AppRepo.GetAppRequestData(appName)
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}

	secret, isPresent, isValidValue := a.AppRequestParser.GetQuerySecret(r)
	if !isValidValue {
		u.WriteResponseErrorAlways(w, u.Logger.NewError("invalid input"))
		return
	}
	if isPresent {
		err = a.exchangeSecretAgainstAuthenticationCookieAndInstructBrowserToRepeatThatRequest(w, r, secret, app)
		if err != nil {
			u.Logger.Error(err)
		}
		return
	}

	if err = a.AppSessionService.AuthorizeAppRequest(r, app); err != nil {
		u.WriteResponseError(w, expectedAccessDeniedErrors, err, "path", r.URL.String())
		return
	}

	allowLongLivedConnection(w, app)

	proxy := a.AppReverseProxyFactory.CreateProxyRequest(r, *app)
	proxy.ServeHTTP(w, r)
}

// Jitsi keeps XMPP WebSocket connections open for the whole call. The server deadlines are useful for normal requests, but would close those calls.
func allowLongLivedConnection(w http.ResponseWriter, app *AppRequestData) {
	if app.AppName != "jitsi" {
		return
	}

	responseController := http.NewResponseController(w)
	if err := responseController.SetReadDeadline(time.Time{}); err != nil {
		u.Logger.Error(err, "app", app.AppName)
	}
	if err := responseController.SetWriteDeadline(time.Time{}); err != nil {
		u.Logger.Error(err, "app", app.AppName)
	}
}

func (a *AppRequestProxy) exchangeSecretAgainstAuthenticationCookieAndInstructBrowserToRepeatThatRequest(w http.ResponseWriter, r *http.Request, urlSecret string, app *AppRequestData) error {
	cookie, err := a.AppSessionService.CreateAppSessionCookieFromSecret(urlSecret, app)
	if err != nil {
		u.WriteResponseError(w, expectedSecretDoesNotExistErrors, err)
		return err
	}

	http.SetCookie(w, cookie)

	redirectURL := *r.URL
	redirectURL.RawQuery = ""
	http.Redirect(w, r, redirectURL.String(), http.StatusFound) // #nosec G710: redirect intentionally returns the browser to the same request URL without the exchanged secret
	return nil
}
