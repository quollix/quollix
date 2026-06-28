//go:build acceptance

package acceptance

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
		InstallFromResult("samplemaintainer", "sampleapp")
	frame.Assert.SnackbarVisibleWithTextEventually("Installation successful")

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
}
