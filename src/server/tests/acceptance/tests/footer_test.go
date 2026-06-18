//go:build acceptance

package acceptance

import (
	"server/tests/acceptance/pages"
	"server/tools"
	"strings"
	"testing"

	"github.com/quollix/common/assert"
)

func TestFooterShowsApplicationVersion(t *testing.T) {
	frame := pages.Setup(t)
	defer frame.Client.Test.ResetTestState()

	frame.GoToInstalledAppsPage()

	footerText := strings.TrimSpace(frame.Page().MustElement("#app-footer").MustText())
	assert.Equal(t, "Quollix "+tools.ApplicationVersion, footerText)
}
