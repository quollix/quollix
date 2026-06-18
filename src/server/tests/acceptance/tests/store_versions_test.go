//go:build acceptance

package acceptance

import (
	"server/tests/acceptance/pages"
	"server/tools"
	"testing"
	"time"

	"github.com/quollix/common/assert"
)

func TestStoreVersionsPage(t *testing.T) {
	frame := pages.Setup(t)
	defer frame.Client.Test.ResetTestState()

	expectedVersions := map[string]time.Time{
		"0.0": tools.SampleAppVersion0CreationTimestamp,
		"1.0": tools.SampleAppVersion1CreationTimestamp,
		"1.5": tools.SampleAppCreationTimestamp,
		"2.0": tools.SampleAppVersion2CreationTimestamp,
	}

	versionsPage := frame.GoToStorePage().
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
		WaitUntilAppVersionInstalled(tools.SampleApp, "1.0").
		Frame.AssertSnackbarVisibleWithTextEventually("Installation successful")

	sampleApp := frame.Client.Apps.GetInstalledSample()
	assert.Equal(t, "1.0", sampleApp.VersionName)
	assert.Nil(t, frame.Client.Apps.Delete(sampleApp.AppId))

	versionsPage.
		SetVersionFilter("0.0").
		AssertVisibleVersionNames([]string{"0.0"}).
		InstallFilteredVersion().
		WaitUntilAppVersionInstalled(tools.SampleApp, "0.0")
	frame.AssertSnackbarVisibleWithTextEventually("Installation successful")

	sampleApp = frame.Client.Apps.GetInstalledSample()
	assert.Equal(t, "0.0", sampleApp.VersionName)
}
