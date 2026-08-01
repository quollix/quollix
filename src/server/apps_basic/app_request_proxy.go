package apps_basic

import (
	"net/http"
	"net/url"
	"server/tools"
	"server/users"
	"time"

	api "github.com/quollix/common/quollix/api"
	u "github.com/quollix/common/utils"
)

type AppRequestProxy struct {
	AppRequestResolver     AppRequestResolver
	AppSessionService      AppSessionService
	AppRequestParser       AppRequestParser
	AppReverseProxyFactory AppReverseProxyFactory
}

var (
	expectedSecretDoesNotExistErrors = u.MapOf(users.SecretDoesNotExistError)
)

func (a *AppRequestProxy) ProxyRequestToTheAppsDockerContainer(w http.ResponseWriter, r *http.Request) {
	u.Logger.Debug("Proxying request")
	resolvedAppRequest, err := a.AppRequestResolver.ResolveAppRequest(r)
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}
	if !resolvedAppRequest.AppExists {
		a.writeAppUnavailablePage(w, resolvedAppRequest.BaseDomain)
		return
	}
	app := resolvedAppRequest.App

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

	authorizationStatus, err := a.AppSessionService.AuthorizeAppRequest(r, app)
	if err != nil {
		u.WriteResponseError(w, nil, err, "path", r.URL.String())
		return
	}
	switch authorizationStatus {
	case AppRequestAuthorized:
	case AppRequestMissingSession:
		a.redirectToBrandAppOpen(w, r, resolvedAppRequest.BaseDomain, app.AppName)
		return
	case AppRequestAccessDenied:
		a.writeAppUnavailablePage(w, resolvedAppRequest.BaseDomain)
		return
	default:
		u.WriteResponseErrorAlways(w, u.Logger.NewError("unknown app request authorization status"))
		return
	}

	considerAllowingLongLivedConnection(w, app)

	proxy := a.AppReverseProxyFactory.CreateProxyRequest(r, *app)
	proxy.ServeHTTP(w, r)
}

func (a *AppRequestProxy) writeAppUnavailablePage(w http.ResponseWriter, baseDomain string) {
	if err := tools.WriteAppUnavailablePage(w, baseDomain); err != nil {
		u.Logger.Error(err)
	}
}

func (a *AppRequestProxy) redirectToBrandAppOpen(w http.ResponseWriter, r *http.Request, baseDomain, appName string) {
	openURL := url.URL{
		Scheme: "https",
		Host:   tools.BrandAppDomainPrefix + baseDomain,
		Path:   api.Paths.FrontendAppOpen,
	}
	query := openURL.Query()
	query.Set("app", appName)
	query.Set("path", r.URL.RequestURI())
	openURL.RawQuery = query.Encode()

	http.Redirect(w, r, openURL.String(), http.StatusFound)
}

// Jitsi keeps XMPP WebSocket connections open for the whole call. The server deadlines are useful for normal requests, but would close those calls.
func considerAllowingLongLivedConnection(w http.ResponseWriter, app *AppRequestData) {
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
