package tools

import (
	"fmt"
	"net/http"

	u "github.com/quollix/common/utils"
)

const PageCouldNotBeLoadedTitle = "Page could not be loaded"

var pageCouldNotBeLoadedContent = u.RenderDefaultPage(fmt.Sprintf(`
<h3>%s</h3>
<p>
	The page you tried to open a not existing page or the provided parameters are invalid.
</p>
<p>
	<a href="/">Go to dashboard</a>
</p>
`, PageCouldNotBeLoadedTitle))

func WritePageCouldNotBeLoaded(w http.ResponseWriter, statusCode int) error {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(statusCode)
	_, err := w.Write([]byte(pageCouldNotBeLoadedContent))
	return err
}
