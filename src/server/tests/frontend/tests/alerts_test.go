//go:build frontend

package frontend

import (
	"server/tests/frontend_pages"
	"testing"

	"github.com/quollix/common/assert"
	"github.com/quollix/common/quollix/api"
)

func TestAlertsPage_ShowsDefaultSettings(t *testing.T) {
	frame := frontend_pages.Setup(t)
	defer frame.Client.Test.ResetTestState()

	page := frame.Pages.OpenAlertsPage()
	frame.Assert.PagePath(api.Paths.FrontendAlerts)

	alerts := page.ListAlerts()
	assert.Equal(t, 5, len(alerts))
	assert.Equal(t, frontend_pages.AlertEntry{Id: "cpu", Enabled: true, Status: "healthy"}, *page.GetRequiredAlert("cpu"))
	assert.Equal(t, frontend_pages.AlertEntry{Id: "memory", Enabled: true, Status: "healthy"}, *page.GetRequiredAlert("memory"))
	assert.Equal(t, frontend_pages.AlertEntry{Id: "disk", Enabled: true, Status: "healthy"}, *page.GetRequiredAlert("disk"))
	assert.Equal(t, frontend_pages.AlertEntry{Id: "certificate", Enabled: true, Status: "healthy"}, *page.GetRequiredAlert("certificate"))
	assert.Equal(t, frontend_pages.AlertEntry{Id: "quollix-errors", Enabled: true, Status: "healthy"}, *page.GetRequiredAlert("quollix-errors"))

	expectedEnabled := map[string]bool{
		"cpu":            false,
		"memory":         false,
		"disk":           false,
		"certificate":    false,
		"quollix-errors": false,
	}
	for alertId, enabled := range expectedEnabled {
		page.SetAlertEnabled(alertId, enabled)
	}
	page.SaveAndAssertSuccessSnackbar()
	frame.Browser.ReloadPage()
	for alertId, enabled := range expectedEnabled {
		assert.Equal(t, enabled, page.GetRequiredAlert(alertId).Enabled)
	}
}
