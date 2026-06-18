//go:build acceptance

package acceptance

import (
	"server/tests/acceptance/pages"
	"testing"

	"github.com/quollix/common/assert"
)

func TestMaintenancePage(t *testing.T) {
	frame := pages.Setup(t)
	defer frame.Client.Test.ResetTestState()

	_, err := frame.Client.Apps.InstallSample("2.0")
	assert.Nil(t, err)

	page := frame.OpenMaintenancePage()

	apps := page.ListApps()
	assert.Equal(t, 2, len(apps))

	postgres := page.GetRequiredApp("postgres")
	assert.Equal(t, "quollix", postgres.Maintainer)
	assert.Equal(t, "postgres", postgres.AppName)
	assert.False(t, postgres.AutomaticUpdatesCheckboxPresent)
	assert.True(t, postgres.AutomaticBackupsCheckboxPresent)
	assert.True(t, postgres.AutomaticBackupsEnabled)

	sampleApp := page.GetRequiredApp("sampleapp")
	assert.Equal(t, "samplemaintainer", sampleApp.Maintainer)
	assert.Equal(t, "sampleapp", sampleApp.AppName)
	assert.True(t, sampleApp.AutomaticUpdatesCheckboxPresent)
	assert.True(t, sampleApp.AutomaticUpdatesEnabled)
	assert.True(t, sampleApp.AutomaticBackupsCheckboxPresent)
	assert.True(t, sampleApp.AutomaticBackupsEnabled)

	assertSampleMaintenanceStateInUi := func(expectedUpdatesEnabled, expectedBackupsEnabled bool) {
		frame.ReloadPage()
		sampleApp = page.GetRequiredApp("sampleapp")
		assert.Equal(t, expectedUpdatesEnabled, sampleApp.AutomaticUpdatesEnabled)
		assert.Equal(t, expectedBackupsEnabled, sampleApp.AutomaticBackupsEnabled)
	}

	page.SetAutomaticUpdatesEnabled("sampleapp", false)
	page.WaitForSampleMaintenanceState(false, true)
	assertSampleMaintenanceStateInUi(false, true)

	page.SetAutomaticBackupsEnabled("sampleapp", false)
	page.WaitForSampleMaintenanceState(false, false)
	assertSampleMaintenanceStateInUi(false, false)

	page.SetAutomaticUpdatesEnabled("sampleapp", true)
	frame.ConfirmDialog()
	page.WaitForSampleMaintenanceState(true, false)
	assertSampleMaintenanceStateInUi(true, false)

	page.SetAutomaticBackupsEnabled("sampleapp", true)
	page.WaitForSampleMaintenanceState(true, true)
	assertSampleMaintenanceStateInUi(true, true)
}
