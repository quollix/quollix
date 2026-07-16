//go:build frontend

package frontend

import (
	"fmt"
	"server/tests/component"
	"server/tests/frontend_pages"
	"server/tools"
	"testing"

	"github.com/quollix/common/assert"
)

func TestStorePage(t *testing.T) {
	frame := frontend_pages.Setup(t)
	defer frame.Client.Test.ResetTestState()

	frame.Pages.GoToStorePage().
		SetSearchAppName("sampleapp").
		Search().
		AssertNoSearchRows().
		EnableUnofficialSearchAndConfirm().
		SetMaintainerFilter("samplemaintainer").
		Search().
		AssertSearchRowCount(1).
		AssertSearchContainsResult("samplemaintainer", "sampleapp", "2.0").
		AssertSearchResultCreatedAt("samplemaintainer", "sampleapp", tools.SampleAppVersion2CreationTimestamp.Format(tools.PrettyFrontendTimeLayout)).
		AssertInstallButtonEnabled("samplemaintainer", "sampleapp").
		InstallFromResult("samplemaintainer", "sampleapp")
	frame.Assert.SnackbarVisibleWithTextEventually("Installation successful")
	frame.Pages.StorePage.AssertInstallButtonDisabledAsInstalled("samplemaintainer", "sampleapp")

	err := tools.Eventually(func() error {
		installedApps := component.ListInstalledApps(t, frame.Client)
		for _, app := range installedApps {
			if app.Maintainer == "samplemaintainer" && app.AppName == "sampleapp" && app.VersionName == "2.0" {
				return nil
			}
		}
		return fmt.Errorf("sampleapp 2.0 not installed yet")
	})
	assert.Nil(t, err)

	sampleApp := component.GetInstalledSample(t, frame.Client)
	assert.Nil(t, frame.Client.Apps.Delete(sampleApp.AppId))

	frame.Pages.GoToStorePage().
		EnableUnofficialSearchAndConfirm().
		SetMaintainerFilter("samplemaintainer").
		SetSearchAppName("sampleapp").
		Search().
		AssertSearchRowCount(1).
		AssertInstallButtonEnabled("samplemaintainer", "sampleapp")
}
