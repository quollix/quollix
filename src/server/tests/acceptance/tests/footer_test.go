//go:build acceptance

package acceptance

import (
	"server/tests/frontend_pages"
	"server/tools"
	"strings"
	"testing"

	"github.com/quollix/common/assert"
)

func TestFooterShowsApplicationVersion(t *testing.T) {
	frame := frontend_pages.Setup(t)
	defer frame.Client.Test.ResetTestState()

	frame.Pages.GoToInstalledAppsPage()

	footerText := strings.TrimSpace(frame.Page.MustElement("#app-footer").MustText())
	assert.Equal(t, "Quollix "+tools.ApplicationVersion, footerText)
}
