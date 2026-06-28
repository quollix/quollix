//go:build acceptance

package acceptance

import (
	"testing"

	"server/tests/frontend_pages"
	"server/tools"
)

const (
	sidebarGroupApps         = "sidebar-group-apps"
	sidebarGroupIdentity     = "sidebar-group-identity"
	sidebarGroupFederation   = "sidebar-group-federation"
	sidebarGroupSystem       = "sidebar-group-system"
	sidebarLinkInstalledApps = "sidebar-link-installed-apps"
	sidebarLinkStore         = "sidebar-link-store"
	sidebarLinkMaintenance   = "sidebar-link-maintenance"
	sidebarLinkAppSso        = "sidebar-link-app-sso"
	sidebarLinkProviders     = "sidebar-link-providers"
	sidebarLinkClients       = "sidebar-link-clients"
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
	frame := frontend_pages.Setup(t)
	defer frame.Client.Test.ResetTestState()

	frame.Pages.OpenInstalledAppsPage()
	frame.Assert.PagePath(tools.Paths.FrontendInstalledApps)

	frame.Browser.ClickSidebarLink(sidebarGroupApps, sidebarLinkStore)
	frame.Assert.PagePath(tools.Paths.FrontendStore)

	frame.Browser.ClickSidebarLink(sidebarGroupApps, sidebarLinkMaintenance)
	frame.Assert.PagePath(tools.Paths.FrontendMaintenance)

	frame.Browser.ClickSidebarLink(sidebarGroupApps, sidebarLinkAppSso)
	frame.Assert.PagePath(tools.Paths.FrontendAppSso)

	frame.Browser.ClickSidebarLink(sidebarGroupIdentity, sidebarLinkUsers)
	frame.Assert.PagePath(tools.Paths.FrontendUsers)

	frame.Browser.ClickSidebarLink(sidebarGroupFederation, sidebarLinkProviders)
	frame.Assert.PagePath(tools.Paths.FrontendProviders)

	frame.Browser.ClickSidebarLink(sidebarGroupFederation, sidebarLinkClients)
	frame.Assert.PagePath(tools.Paths.FrontendClients)

	frame.Browser.ClickSidebarLink(sidebarGroupSystem, sidebarLinkBackups)
	frame.Assert.PagePath(tools.Paths.FrontendBackedUpApps)

	frame.Browser.ClickSidebarLink(sidebarGroupSystem, sidebarLinkSettings)
	frame.Assert.PagePath(tools.Paths.FrontendSettings)

	frame.Browser.ClickSidebarLink(sidebarGroupIdentity, sidebarLinkGroups)
	frame.Assert.PagePath(tools.Paths.FrontendGroups)

	frame.Browser.ClickSidebarLink(sidebarGroupSystem, sidebarLinkTerminal)
	frame.Assert.PagePath(tools.Paths.FrontendTerminalApps)

	frame.Browser.ClickSidebarLink(sidebarGroupSystem, sidebarLinkEmail)
	frame.Assert.PagePath(tools.Paths.FrontendEmail)

	frame.Browser.ClickSidebarUserLink()
	frame.Assert.PagePath(tools.Paths.FrontendAccount)

	frame.Browser.ClickSidebarLink(sidebarGroupApps, sidebarLinkInstalledApps)
	frame.Assert.PagePath(tools.Paths.FrontendInstalledApps)
}

func TestAccountTabVisibilityForAnonymousAndAuthenticated(t *testing.T) {
	frame := frontend_pages.Setup(t)
	defer frame.Client.Test.ResetTestState()

	frame.Browser.ClickSidebarUserLink()
	frame.Assert.PagePath(tools.Paths.FrontendAccount)

	frame.Session.SignOutViaClient()
	frame.Pages.GoToInstalledAppsPage().Frame.Assert.PagePath(tools.Paths.FrontendInstalledApps)
	frame.Assert.ElementNotPresent("#" + sidebarUserLink)
	frame.Assert.ElementNotPresent("#" + sidebarLinkFeedback)
}
