//go:build acceptance

package acceptance

import (
	"server/tests/acceptance/pages"
	"server/tools"
	"testing"

	u "github.com/quollix/common/utils"
)

func TestTerminalPageFlow(t *testing.T) {
	frame := pages.Setup(t)
	defer frame.Client.Test.ResetTestState()

	frame.GoToTerminalPage().
		AssertAppPresent(u.OfficialMaintainer, u.OfficialDatabaseAppName).
		AssertAppPresent(u.OfficialMaintainer, u.OfficialBrandAppName).
		OpenServicesPage(u.OfficialMaintainer, u.OfficialBrandAppName).
		AssertSelection(u.OfficialMaintainer, u.OfficialBrandAppName).
		ClickBackAndAssertTerminalAppsPage().
		OpenServicesPage(u.OfficialMaintainer, u.OfficialBrandAppName).
		AssertSelection(u.OfficialMaintainer, u.OfficialBrandAppName).
		AssertServicePresent(tools.BrandAppService).
		OpenTerminal(tools.BrandAppService).
		AssertSelection(u.OfficialMaintainer, u.OfficialBrandAppName, tools.BrandAppService).
		ClickBackAndAssertTerminalServicesPage().
		AssertSelection(u.OfficialMaintainer, u.OfficialBrandAppName).
		OpenTerminal(tools.BrandAppService).
		AssertSelection(u.OfficialMaintainer, u.OfficialBrandAppName, tools.BrandAppService).
		AssertReady().
		RunCommand("echo test").
		AssertOutputContains(`/opt/server # echo test
test
/opt/server #`)
}
