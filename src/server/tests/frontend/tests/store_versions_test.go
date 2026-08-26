//go:build frontend

package frontend

import (
	"server/tests/component"
	"server/tests/frontend_pages"
	"server/tools"
	"testing"
	"time"

	"github.com/quollix/common/assert"
)

func TestStoreVersionsPage(t *testing.T) {
	frame := frontend_pages.Setup(t)
	defer frame.Client.Test.ResetTestState()

	expectedVersions := map[string]time.Time{
		"0.0": tools.SampleAppVersion0CreationTimestamp,
		"1.0": tools.SampleAppVersion1CreationTimestamp,
		"1.5": tools.SampleAppCreationTimestamp,
		"2.0": tools.SampleAppVersion2CreationTimestamp,
	}

	versionsPage := frame.Pages.GoToStorePage().
		EnableUnofficialSearchAndConfirm().
		SetMaintainerFilter("samplemaintainer").
		SetSearchAppName("sampleapp").
		Search().
		AssertSearchRowCount(1).
		OpenVersionsFromResult("samplemaintainer", "sampleapp")

	versionsPage.
		AssertVersionsPageHeader("samplemaintainer", "sampleapp").
		AssertVersionRowCount(4).
		AssertVersionNames([]string{"0.0", "1.0", "1.5", "2.0"}).
		AssertVersionsAndCreationDates(expectedVersions).
		InstallVersion("1.0").
		AssertVersionInstallButtonDisabled("0.0").
		AssertVersionInstallButtonDisabled("1.0").
		AssertVersionInstallButtonEnabled("1.5").
		AssertVersionInstallButtonEnabled("2.0").
		WaitUntilAppVersionInstalled(tools.SampleApp, "1.0").
			Frame.Assert.SnackbarVisibleWithTextEventually("Installation/update started")

	sampleApp := component.GetInstalledSample(t, frame.Client)
	assert.Equal(t, "1.0", sampleApp.VersionName)
	assert.True(t, sampleApp.IsRunning)

	versionsPage.
		InstallVersion("2.0").
		AssertVersionInstallButtonDisabled("0.0").
		AssertVersionInstallButtonDisabled("1.0").
		AssertVersionInstallButtonDisabled("1.5").
		AssertVersionInstallButtonDisabled("2.0").
		WaitUntilAppVersionInstalled(tools.SampleApp, "2.0")
	frame.Assert.SnackbarVisibleWithTextEventually("Installation/update started")

	sampleApp = component.GetInstalledSample(t, frame.Client)
	assert.Equal(t, "2.0", sampleApp.VersionName)
	assert.True(t, sampleApp.IsRunning)
}
