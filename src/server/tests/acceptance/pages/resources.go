//go:build acceptance

package pages

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"server/tests/component"
	"server/tools"
	"strings"
	"testing"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/quollix/common/assert"
)

var (
	browser                          *rod.Browser
	wasFrontendReloadedDuringThisRun = false
)

const (
	defaultTimeout         = 3 * time.Second
	backupOperationTimeout = 1 * time.Minute
)

func Setup(t *testing.T) *FrameType {
	if browser == nil {
		headless := os.Getenv("HEADFUL") != "true"
		// GitHub-hosted Linux runners can block Chromium's sandbox (userns/AppArmor), so acceptance tests disable browser sandboxing in CI.
		noSandbox := os.Getenv("CI") == "true"
		u := launcher.New().Headless(headless).NoSandbox(noSandbox).MustLaunch()
		// Initialize the shared root Rod client once. It owns the underlying Chrome process connection.
		browser = rod.New().ControlURL(u).MustConnect().MustIgnoreCertErrors(true)
	}

	// Create a per-test incognito browser-context client.
	// This is a lightweight Rod handle for isolation; it does not start a new Chrome process.
	testBrowser, err := browser.Incognito()
	assert.Nil(t, err)

	frame := NewFrameType(t)
	frame.browser = testBrowser
	frame.page = frame.browser.MustPage()
	t.Cleanup(func() {
		if frame.page != nil {
			assert.Nil(t, frame.page.Close())
			frame.page = nil
		}
		if frame.browser != nil {
			// Closing the per-test incognito context guarantees isolated cookies and storage.
			assert.Nil(t, frame.browser.Close())
			frame.browser = nil
		}
	})

	frame.Client = component.GetQuollixClient(t)
	frame.LoginAsAdminViaClient()

	if !wasFrontendReloadedDuringThisRun {
		// this means, we can make changes to frontend and simply re-run the acceptance tests with latest changes, without having to re-redeploy the quollix container
		frame.Client.Frontend.Reload()
		wasFrontendReloadedDuringThisRun = true
	}
	return frame
}

func CloseBrowser() {
	if browser == nil {
		return
	}
	// This closes the shared root Rod client and therefore terminates the underlying Chrome process.
	if err := browser.Close(); err != nil {
		panic(err.Error())
	}
	browser = nil
}

type FrameType struct {
	t                    *testing.T
	browser              *rod.Browser
	InstalledAppsPage    *InstalledAppsPage
	LoginPage            *LoginPage
	OidcClientsPage      *OidcClientsPage
	SettingsPage         *SettingsPage
	EmailPage            *EmailPage
	page                 *rod.Page
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
	Client               *component.QuollixClient
}

func (f *FrameType) OpenInstalledAppsPage() *InstalledAppsPage {
	f.visit(tools.Paths.FrontendInstalledApps)
	return f.InstalledAppsPage
}

func (f *FrameType) OpenUsersPage() *UsersPage {
	f.visit(tools.Paths.FrontendUsers)
	return f.UsersPage
}

func (f *FrameType) VisitLoginPage() *LoginPage {
	f.visit(tools.Paths.FrontendLogin)
	return f.LoginPage
}

func (f *FrameType) visit(path string) {
	f.DoAndWaitDOMContentLoaded(func() {
		f.page.MustNavigate("https://quollix.localhost" + path)
	})
}

func (f *FrameType) DoAndWaitDOMContentLoaded(action func()) {
	waitForNavigation := f.page.WaitNavigation(proto.PageLifecycleEventNameDOMContentLoaded)
	action()
	waitForNavigation()
}

func (f *FrameType) GoToInstalledAppsPage() *InstalledAppsPage {
	f.visit(tools.Paths.FrontendInstalledApps)
	return f.InstalledAppsPage
}

func (f *FrameType) GoToStorePage() *StorePage {
	f.visit(tools.Paths.FrontendStore)
	return f.StorePage
}

func (f *FrameType) OpenMaintenancePage() *MaintenancePage {
	f.visit(tools.Paths.FrontendMaintenance)
	return f.MaintenancePage
}

func (f *FrameType) GoToMaintenancePage() *FrameType {
	f.visit(tools.Paths.FrontendMaintenance)
	return f
}

func (f *FrameType) OpenOidcClientsPage() *OidcClientsPage {
	f.visit(tools.Paths.FrontendOidcClients)
	return f.OidcClientsPage
}

func (f *FrameType) GoToOidcPage() *FrameType {
	f.visit(tools.Paths.FrontendOidcClients)
	return f
}

func (f *FrameType) GoToUsersPage() *UsersPage {
	f.visit(tools.Paths.FrontendUsers)
	return f.UsersPage
}

func (f *FrameType) GoToBackupsPage() *FrameType {
	f.visit(tools.Paths.FrontendBackedUpApps)
	return f
}

func (f *FrameType) OpenBackedUpAppsPage() *BackedUpAppsPage {
	f.visit(tools.Paths.FrontendBackedUpApps)
	return f.BackedUpAppsPage
}

func (f *FrameType) GoToSettingsPage() *SettingsPage {
	f.visit(tools.Paths.FrontendSettings)
	return f.SettingsPage
}

func (f *FrameType) GoToEmailPage() *EmailPage {
	f.visit(tools.Paths.FrontendEmail)
	f.AssertPagePath(tools.Paths.FrontendEmail)
	return f.EmailPage
}

func (f *FrameType) GoToTerminalPage() *TerminalAppsPage {
	f.visit(tools.Paths.FrontendTerminalApps)
	f.AssertPagePath(tools.Paths.FrontendTerminalApps)
	return f.TerminalAppsPage
}

func (f *FrameType) OpenGroupsPage() *GroupsPage {
	f.visit(tools.Paths.FrontendGroups)
	f.AssertPagePath(tools.Paths.FrontendGroups)
	return f.GroupsPage
}

func (f *FrameType) GoToAccountPage() *FrameType {
	f.visit(tools.Paths.FrontendAccount)
	return f
}

func (f *FrameType) ReloadPage() *FrameType {
	f.DoAndWaitDOMContentLoaded(func() {
		f.page.MustReload()
	})
	return f
}

func (f *FrameType) Logout() *FrameType {
	f.DoAndWaitDOMContentLoaded(func() {
		f.page.MustElement("#logout-tab a").MustClick()
	})
	f.Client.Parent.Cookie = nil
	f.clearBrowserCookies()
	return f
}

func (f *FrameType) LogoutViaClient() *FrameType {
	err := f.Client.Auth.Logout()
	assert.Nil(f.t, err)
	f.Client.Parent.Cookie = nil
	f.clearBrowserCookies()
	return f
}

func (f *FrameType) ClickSidebarLink(groupId, itemId string) {
	element, err := f.page.Element("#" + groupId)
	assert.Nil(f.t, err)
	element.MustClick()

	element, err = f.page.Element("#" + itemId)
	assert.Nil(f.t, err)
	f.DoAndWaitDOMContentLoaded(func() {
		element.MustClick()
	})
}

func (f *FrameType) ClickSidebarUserLink() {
	f.DoAndWaitDOMContentLoaded(func() {
		f.page.MustElement("#sidebar-user-link").MustClick()
	})
}

func (f *FrameType) ClickSidebarTopLevelLink(itemID string) {
	f.DoAndWaitDOMContentLoaded(func() {
		f.page.MustElement("#" + itemID).MustClick()
	})
}

func (f *FrameType) AssertElementNotPresentByID(elementID string) *FrameType {
	exists, _, err := f.page.Has("#" + elementID)
	assert.Nil(f.t, err)
	assert.False(f.t, exists)
	return f
}

func (f *FrameType) ConfirmDialog() *FrameType {
	err := tools.Eventually(func() error {
		confirmButton, findErr := f.page.Element("#confirm-button")
		if findErr != nil {
			return findErr
		}
		return confirmButton.Click(proto.InputMouseButtonLeft, 1)
	})
	assert.Nil(f.t, err)
	return f
}

func (f *FrameType) AssertSnackbarVisibleWithTextEventually(expectedText string) *FrameType {
	return f.AssertSnackbarVisibleWithTextEventuallyWithin(expectedText, defaultTimeout)
}

func (f *FrameType) AssertSnackbarVisibleWithTextEventuallyWithin(expectedText string, timeout time.Duration) *FrameType {
	err := tools.EventuallyWithTimeout(timeout, 50*time.Millisecond, func() error {
		snackbar, findErr := f.page.Element("#snackbar")
		if findErr != nil {
			return findErr
		}

		visibleAttr, attrErr := snackbar.Attribute("data-visible")
		if attrErr != nil {
			return attrErr
		}
		if visibleAttr == nil || *visibleAttr != "true" {
			return fmt.Errorf("snackbar is not visible yet")
		}

		text, textErr := snackbar.Text()
		if textErr != nil {
			return textErr
		}
		if strings.TrimSpace(text) != expectedText {
			return fmt.Errorf("unexpected snackbar text: %q", strings.TrimSpace(text))
		}
		return nil
	})
	assert.Nil(f.t, err)
	return f
}

func (f *FrameType) AssertPagePath(expectedPath string) {
	lastPath := ""
	err := tools.Eventually(func() error {
		info, infoErr := f.page.Info()
		if infoErr != nil {
			return infoErr
		}
		currentURL, parseError := url.Parse(info.URL)
		if parseError != nil {
			return parseError
		}
		lastPath = currentURL.Path
		if lastPath != expectedPath {
			return fmt.Errorf("expected path %s but got %s", expectedPath, lastPath)
		}
		return nil
	})
	assert.Nil(f.t, err)
}

func (f *FrameType) AssertAppOperationStarted() *FrameType {
	err := tools.EventuallyWithTimeout(backupOperationTimeout, 50*time.Millisecond, func() error {
		_, isOngoing := f.Client.Apps.GetCurrentOperations()
		if !isOngoing {
			return fmt.Errorf("app operation did not start yet")
		}
		return nil
	})
	assert.Nil(f.t, err)
	return f
}

func (f *FrameType) AssertAppOperationFinished() *FrameType {
	err := tools.EventuallyWithTimeout(backupOperationTimeout, 50*time.Millisecond, func() error {
		_, isOngoing := f.Client.Apps.GetCurrentOperations()
		if isOngoing {
			return fmt.Errorf("app operation is still ongoing")
		}
		return nil
	})
	assert.Nil(f.t, err)
	return f
}

// Waiting for the operation to start and then finish usually prevents transient browser navigation failures such as net::ERR_NETWORK_CHANGED in subsequent acceptance steps.
func (f *FrameType) AssertAppOperationStartedAndFinished() *FrameType {
	return f.AssertAppOperationStarted().AssertAppOperationFinished()
}

func (f *FrameType) Page() *rod.Page {
	return f.page
}

func (f *FrameType) TestingT() *testing.T {
	return f.t
}

func (f *FrameType) LoginAsAdminViaClient() *FrameType {
	return f.LoginViaClient(tools.DefaultAdminName, tools.DefaultAdminPassword)
}

func (f *FrameType) LoginViaClient(username, password string) *FrameType {
	err := f.Client.Auth.Login(username, password)
	assert.Nil(f.t, err)
	f.syncBrowserCookieFromClient()
	f.visit(tools.Paths.FrontendInstalledApps)
	return f
}

func (f *FrameType) GetAuthCookie() *http.Cookie {
	cookies := f.page.MustCookies("https://quollix.localhost")
	for _, cookie := range cookies {
		if cookie.Name == tools.BrandAppAuthCookieName {
			return &http.Cookie{ // #nosec G124: acceptance tests reconstruct the browser cookie only for local test replay
				Name:     cookie.Name,
				Value:    cookie.Value,
				Path:     cookie.Path,
				Secure:   cookie.Secure,
				HttpOnly: cookie.HTTPOnly,
			}
		}
	}
	return nil
}

func (f *FrameType) syncClientCookieFromBrowser() {
	f.Client.Parent.Cookie = f.GetAuthCookie()
}

func (f *FrameType) syncBrowserCookieFromClient() {
	f.clearBrowserCookies()
	cookie := f.Client.Parent.Cookie
	if cookie == nil {
		return
	}
	f.page.MustSetCookies(&proto.NetworkCookieParam{
		Name:     cookie.Name,
		Value:    cookie.Value,
		URL:      "https://quollix.localhost",
		Path:     "/",
		Secure:   cookie.Secure,
		HTTPOnly: cookie.HttpOnly,
	})
}

func (f *FrameType) clearBrowserCookies() {
	f.page.MustSetCookies()
}

func NewFrameType(t *testing.T) *FrameType {
	frame := &FrameType{t: t}
	frame.InstalledAppsPage = &InstalledAppsPage{frame}
	frame.LoginPage = &LoginPage{frame}
	frame.OidcClientsPage = &OidcClientsPage{frame}
	frame.SettingsPage = &SettingsPage{frame}
	frame.EmailPage = &EmailPage{Frame: frame}
	frame.StorePage = &StorePage{frame}
	frame.VersionsPage = &VersionsPage{frame}
	frame.MaintenancePage = &MaintenancePage{frame}
	frame.UsersPage = &UsersPage{frame}
	frame.UserEditPage = &UserEditPage{frame}
	frame.TerminalAppsPage = &TerminalAppsPage{Frame: frame}
	frame.TerminalServicesPage = &TerminalServicesPage{Frame: frame}
	frame.TerminalViewPage = &TerminalViewPage{Frame: frame}
	frame.GroupsPage = &GroupsPage{Frame: frame}
	frame.GroupMembersPage = &GroupMembersPage{Frame: frame}
	frame.GroupAppsPage = &GroupAppsPage{Frame: frame}
	frame.BackedUpAppsPage = &BackedUpAppsPage{frame}
	frame.BackupsPage = &BackupsPage{frame}
	return frame
}
