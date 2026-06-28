package frontend_pages

import (
	"server/tools"
	"strings"
	"time"

	"github.com/go-rod/rod/lib/proto"
	"github.com/quollix/common/assert"
)

type FramePages struct {
	Frame                *FrameType
	AccountPage          *AccountPage
	InstalledAppsPage    *InstalledAppsPage
	SignInPage           *SignInPage
	OidcClientsPage      *OidcClientsPage
	ProvidersPage        *ProvidersPage
	ClientsPage          *ClientsPage
	SettingsPage         *SettingsPage
	EmailPage            *EmailPage
	StorePage            *StorePage
	VersionsPage         *VersionsPage
	MaintenancePage      *MaintenancePage
	UsersPage            *UsersPage
	UserEditPage         *UserEditPage
	TerminalAppsPage     *TerminalAppsPage
	TerminalServicesPage *TerminalServicesPage
	TerminalViewPage     *TerminalViewPage
	GroupsPage           *GroupsPage
	GroupMembersPage     *GroupMembersPage
	GroupAppsPage        *GroupAppsPage
	BackedUpAppsPage     *BackedUpAppsPage
	BackupsPage          *BackupsPage
}

func newFramePages(frame *FrameType) *FramePages {
	return &FramePages{
		Frame:                frame,
		AccountPage:          &AccountPage{Frame: frame},
		InstalledAppsPage:    &InstalledAppsPage{Frame: frame},
		SignInPage:           &SignInPage{Frame: frame},
		OidcClientsPage:      &OidcClientsPage{Frame: frame},
		ProvidersPage:        &ProvidersPage{Frame: frame},
		ClientsPage:          &ClientsPage{Frame: frame},
		SettingsPage:         &SettingsPage{Frame: frame},
		EmailPage:            &EmailPage{Frame: frame},
		StorePage:            &StorePage{Frame: frame},
		VersionsPage:         &VersionsPage{Frame: frame},
		MaintenancePage:      &MaintenancePage{Frame: frame},
		UsersPage:            &UsersPage{Frame: frame},
		UserEditPage:         &UserEditPage{Frame: frame},
		TerminalAppsPage:     &TerminalAppsPage{Frame: frame},
		TerminalServicesPage: &TerminalServicesPage{Frame: frame},
		TerminalViewPage:     &TerminalViewPage{Frame: frame},
		GroupsPage:           &GroupsPage{Frame: frame},
		GroupMembersPage:     &GroupMembersPage{Frame: frame},
		GroupAppsPage:        &GroupAppsPage{Frame: frame},
		BackedUpAppsPage:     &BackedUpAppsPage{Frame: frame},
		BackupsPage:          &BackupsPage{Frame: frame},
	}
}

func (p *FramePages) Visit(path string) *FrameType {
	url := p.Frame.BaseUrl + path
	var fatalNavigationErr error
	err := tools.EventuallyWithTimeout(browserTimeout, 100*time.Millisecond, func() error {
		waitForNavigation := p.Frame.Page.WaitNavigation(proto.PageLifecycleEventNameDOMContentLoaded)
		if navigationErr := p.Frame.Page.Navigate(url); navigationErr != nil {
			if isNetworkChangedNavigationError(navigationErr) {
				return navigationErr
			}
			fatalNavigationErr = navigationErr
			return nil
		}
		waitForNavigation()
		return nil
	})
	assert.Nil(p.Frame.T, fatalNavigationErr)
	assert.Nil(p.Frame.T, err)
	return p.Frame
}

func isNetworkChangedNavigationError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "NETWORK_CHANGED")
}

func (p *FramePages) GoToAccountPage() *AccountPage {
	p.Visit(tools.Paths.FrontendAccount)
	return p.AccountPage
}

func (p *FramePages) OpenInstalledAppsPage() *InstalledAppsPage {
	p.Visit(tools.Paths.FrontendInstalledApps)
	return p.InstalledAppsPage
}

func (p *FramePages) OpenUsersPage() *UsersPage {
	p.Visit(tools.Paths.FrontendUsers)
	return p.UsersPage
}

func (p *FramePages) VisitSignInPage() *SignInPage {
	p.Visit(tools.Paths.FrontendSignIn)
	return p.SignInPage
}

func (p *FramePages) GoToInstalledAppsPage() *InstalledAppsPage {
	p.Visit(tools.Paths.FrontendInstalledApps)
	return p.InstalledAppsPage
}

func (p *FramePages) GoToStorePage() *StorePage {
	p.Visit(tools.Paths.FrontendStore)
	return p.StorePage
}

func (p *FramePages) OpenMaintenancePage() *MaintenancePage {
	p.Visit(tools.Paths.FrontendMaintenance)
	return p.MaintenancePage
}

func (p *FramePages) GoToMaintenancePage() *FrameType {
	p.Visit(tools.Paths.FrontendMaintenance)
	return p.Frame
}

func (p *FramePages) OpenOidcClientsPage() *OidcClientsPage {
	p.Visit(tools.Paths.FrontendAppSso)
	return p.OidcClientsPage
}

func (p *FramePages) OpenProvidersPage() *ProvidersPage {
	p.Visit(tools.Paths.FrontendProviders)
	return p.ProvidersPage
}

func (p *FramePages) OpenClientsPage() *ClientsPage {
	p.Visit(tools.Paths.FrontendClients)
	return p.ClientsPage
}

func (p *FramePages) GoToOidcPage() *FrameType {
	p.Visit(tools.Paths.FrontendAppSso)
	return p.Frame
}

func (p *FramePages) GoToUsersPage() *UsersPage {
	p.Visit(tools.Paths.FrontendUsers)
	return p.UsersPage
}

func (p *FramePages) GoToBackupsPage() *FrameType {
	p.Visit(tools.Paths.FrontendBackedUpApps)
	return p.Frame
}

func (p *FramePages) OpenBackedUpAppsPage() *BackedUpAppsPage {
	p.Visit(tools.Paths.FrontendBackedUpApps)
	return p.BackedUpAppsPage
}

func (p *FramePages) GoToSettingsPage() *SettingsPage {
	p.Visit(tools.Paths.FrontendSettings)
	return p.SettingsPage
}

func (p *FramePages) GoToEmailPage() *EmailPage {
	p.Visit(tools.Paths.FrontendEmail)
	p.Frame.Assert.PagePath(tools.Paths.FrontendEmail)
	return p.EmailPage
}

func (p *FramePages) GoToTerminalPage() *TerminalAppsPage {
	p.Visit(tools.Paths.FrontendTerminalApps)
	p.Frame.Assert.PagePath(tools.Paths.FrontendTerminalApps)
	return p.TerminalAppsPage
}

func (p *FramePages) OpenGroupsPage() *GroupsPage {
	p.Visit(tools.Paths.FrontendGroups)
	p.Frame.Assert.PagePath(tools.Paths.FrontendGroups)
	return p.GroupsPage
}
