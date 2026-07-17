package frontend_pages

import (
	"server/tools"
	"strings"
	"time"

	"github.com/quollix/common/assert"
	"github.com/quollix/common/quollix/api"
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
		if navigationErr := p.Frame.Page.Navigate(url); navigationErr != nil {
			if isNetworkChangedNavigationError(navigationErr) {
				return navigationErr
			}
			fatalNavigationErr = navigationErr
			return nil
		}
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
	p.Visit(api.Paths.FrontendAccount)
	return p.AccountPage
}

func (p *FramePages) OpenInstalledAppsPage() *InstalledAppsPage {
	p.Visit(api.Paths.FrontendInstalledApps)
	return p.InstalledAppsPage
}

func (p *FramePages) OpenUsersPage() *UsersPage {
	p.Visit(api.Paths.FrontendUsers)
	return p.UsersPage
}

func (p *FramePages) VisitSignInPage() *SignInPage {
	p.Visit(api.Paths.FrontendSignIn)
	return p.SignInPage
}

func (p *FramePages) GoToInstalledAppsPage() *InstalledAppsPage {
	p.Visit(api.Paths.FrontendInstalledApps)
	return p.InstalledAppsPage
}

func (p *FramePages) GoToStorePage() *StorePage {
	p.Visit(api.Paths.FrontendStore)
	return p.StorePage
}

func (p *FramePages) OpenMaintenancePage() *MaintenancePage {
	p.Visit(api.Paths.FrontendMaintenance)
	return p.MaintenancePage
}

func (p *FramePages) GoToMaintenancePage() *FrameType {
	p.Visit(api.Paths.FrontendMaintenance)
	return p.Frame
}

func (p *FramePages) OpenOidcClientsPage() *OidcClientsPage {
	p.Visit(api.Paths.FrontendAppSso)
	return p.OidcClientsPage
}

func (p *FramePages) OpenProvidersPage() *ProvidersPage {
	p.Visit(api.Paths.FrontendProviders)
	return p.ProvidersPage
}

func (p *FramePages) OpenClientsPage() *ClientsPage {
	p.Visit(api.Paths.FrontendClients)
	return p.ClientsPage
}

func (p *FramePages) GoToOidcPage() *FrameType {
	p.Visit(api.Paths.FrontendAppSso)
	return p.Frame
}

func (p *FramePages) GoToUsersPage() *UsersPage {
	p.Visit(api.Paths.FrontendUsers)
	return p.UsersPage
}

func (p *FramePages) GoToBackupsPage() *FrameType {
	p.Visit(api.Paths.FrontendBackedUpApps)
	return p.Frame
}

func (p *FramePages) OpenBackedUpAppsPage() *BackedUpAppsPage {
	p.Visit(api.Paths.FrontendBackedUpApps)
	return p.BackedUpAppsPage
}

func (p *FramePages) GoToSettingsPage() *SettingsPage {
	p.Visit(api.Paths.FrontendSettings)
	return p.SettingsPage
}

func (p *FramePages) GoToEmailPage() *EmailPage {
	p.Visit(api.Paths.FrontendEmail)
	p.Frame.Assert.PagePath(api.Paths.FrontendEmail)
	return p.EmailPage
}

func (p *FramePages) GoToTerminalPage() *TerminalAppsPage {
	p.Visit(api.Paths.FrontendTerminalApps)
	p.Frame.Assert.PagePath(api.Paths.FrontendTerminalApps)
	return p.TerminalAppsPage
}

func (p *FramePages) OpenGroupsPage() *GroupsPage {
	p.Visit(api.Paths.FrontendGroups)
	p.Frame.Assert.PagePath(api.Paths.FrontendGroups)
	return p.GroupsPage
}
