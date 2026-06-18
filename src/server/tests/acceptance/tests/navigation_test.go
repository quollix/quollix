//go:build acceptance

package acceptance

import (
	"server/tests/acceptance/pages"
	"server/tools"
	"testing"
)

const (
	sidebarGroupApps         = "sidebar-group-apps"
	sidebarGroupAccess       = "sidebar-group-access"
	sidebarGroupSystem       = "sidebar-group-system"
	sidebarLinkInstalledApps = "sidebar-link-installed-apps"
	sidebarLinkStore         = "sidebar-link-store"
	sidebarLinkMaintenance   = "sidebar-link-maintenance"
	sidebarLinkOidc          = "sidebar-link-oidc"
	sidebarLinkUsers         = "sidebar-link-users"
	sidebarLinkGroups        = "sidebar-link-groups"
	sidebarLinkBackups       = "sidebar-link-backups"
	sidebarLinkTerminal      = "sidebar-link-terminal"
	sidebarLinkEmail         = "sidebar-link-email"
	sidebarLinkFeedback      = "sidebar-link-feedback"
	sidebarLinkSettings      = "sidebar-link-settings"
	sidebarUserLink          = "sidebar-user-link"
)

func TestSidebarNavigationAsAdmin(t *testing.T) {
	frame := pages.Setup(t)
	defer frame.Client.Test.ResetTestState()

	frame.OpenInstalledAppsPage()
	frame.AssertPagePath(tools.Paths.FrontendInstalledApps)

	frame.ClickSidebarLink(sidebarGroupApps, sidebarLinkStore)
	frame.AssertPagePath(tools.Paths.FrontendStore)

	frame.ClickSidebarLink(sidebarGroupApps, sidebarLinkMaintenance)
	frame.AssertPagePath(tools.Paths.FrontendMaintenance)

	frame.ClickSidebarLink(sidebarGroupApps, sidebarLinkOidc)
	frame.AssertPagePath(tools.Paths.FrontendOidcClients)

	frame.ClickSidebarLink(sidebarGroupAccess, sidebarLinkUsers)
	frame.AssertPagePath(tools.Paths.FrontendUsers)

	frame.ClickSidebarLink(sidebarGroupSystem, sidebarLinkBackups)
	frame.AssertPagePath(tools.Paths.FrontendBackedUpApps)

	frame.ClickSidebarLink(sidebarGroupSystem, sidebarLinkSettings)
	frame.AssertPagePath(tools.Paths.FrontendSettings)

	frame.ClickSidebarLink(sidebarGroupAccess, sidebarLinkGroups)
	frame.AssertPagePath(tools.Paths.FrontendGroups)

	frame.ClickSidebarLink(sidebarGroupSystem, sidebarLinkTerminal)
	frame.AssertPagePath(tools.Paths.FrontendTerminalApps)

	frame.ClickSidebarLink(sidebarGroupSystem, sidebarLinkEmail)
	frame.AssertPagePath(tools.Paths.FrontendEmail)

	frame.ClickSidebarUserLink()
	frame.AssertPagePath(tools.Paths.FrontendAccount)

	frame.ClickSidebarLink(sidebarGroupApps, sidebarLinkInstalledApps)
	frame.AssertPagePath(tools.Paths.FrontendInstalledApps)
}

func TestAccountTabVisibilityForAnonymousAndAuthenticated(t *testing.T) {
	frame := pages.Setup(t)
	defer frame.Client.Test.ResetTestState()

	frame.ClickSidebarUserLink()
	frame.AssertPagePath(tools.Paths.FrontendAccount)

	frame.LogoutViaClient()
	frame.GoToInstalledAppsPage().Frame.AssertPagePath(tools.Paths.FrontendInstalledApps)
	frame.AssertElementNotPresentByID(sidebarUserLink)
	frame.AssertElementNotPresentByID(sidebarLinkFeedback)
}
