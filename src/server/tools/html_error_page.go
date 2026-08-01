package tools

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/quollix/common/quollix/api"
	u "github.com/quollix/common/utils"
)

const PageCouldNotBeLoadedTitle = "Page could not be loaded"
const AppUnavailableTitle = "App unavailable"
const AppUnavailableMessage = "This app does not exist or you do not have access rights. If this is a mistake, please contact your administrator."
const AppUnavailableInstalledAppsLinkText = "Go to installed apps"

var pageCouldNotBeLoadedContent = u.RenderDefaultPage(fmt.Sprintf(`
<h3>%s</h3>
<p>
	The page you tried to open a not existing page or the provided parameters are invalid.
</p>
<p>
	<a href="/">Go to dashboard</a>
</p>
`, PageCouldNotBeLoadedTitle))

const appUnavailableContentTemplate = `
<h3>%s</h3>
<p>
	%s
</p>
<p>
	<a href="%s">%s</a>
</p>
`

func WritePageCouldNotBeLoaded(w http.ResponseWriter, statusCode int) error {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(statusCode)
	_, err := w.Write([]byte(pageCouldNotBeLoadedContent))
	return err
}

func WriteAppUnavailablePage(w http.ResponseWriter, baseDomain string) error {
	installedAppsURL := url.URL{
		Scheme: "https",
		Host:   BrandAppDomainPrefix + baseDomain,
		Path:   api.Paths.FrontendInstalledApps,
	}
	content := u.RenderDefaultPage(fmt.Sprintf(appUnavailableContentTemplate, AppUnavailableTitle, AppUnavailableMessage, installedAppsURL.String(), AppUnavailableInstalledAppsLinkText))
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusForbidden)
	_, err := w.Write([]byte(content))
	return err
}
