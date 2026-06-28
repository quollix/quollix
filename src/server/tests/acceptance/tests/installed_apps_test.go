//go:build acceptance

package acceptance

import (
	"server/tests/component"
	"server/tests/frontend_pages"
	"server/tools"
	"strings"
	"testing"

	"github.com/quollix/common/assert"
)

func TestInstalledAppPage(t *testing.T) {
	frame := frontend_pages.Setup(t)
	defer frame.Client.Test.ResetTestState()

	_, err := component.InstallSample(t, frame.Client, "2.0")
	assert.Nil(t, err)

	installedAppsPage := frame.Pages.OpenInstalledAppsPage().AssertHeaderColumnCount(8)
	postgres := installedAppsPage.GetApp("postgres")
	assert.Equal(t, "quollix", postgres.Maintainer)
	assert.Equal(t, "postgres", postgres.AppName)
	assert.True(t, strings.HasPrefix(postgres.Version, "17.5"))
	assert.Equal(t, "—", postgres.VersionCreated)
	assert.False(t, postgres.OpenButtonPresent)
	installedAppsPage.AssertDocsLinkHref("postgres", tools.InstalledAppDocsUrl("postgres"))

	sampleApp := installedAppsPage.GetApp("sampleapp")
	assert.Equal(t, "samplemaintainer", sampleApp.Maintainer)
	assert.Equal(t, "sampleapp", sampleApp.AppName)
	assert.True(t, strings.HasPrefix(sampleApp.Version, "2.0"))
	assertFrontendRelativeTimeSet(t, sampleApp.VersionCreated)
	installedAppsPage.AssertVersionCreatedTooltip("sampleapp", tools.SampleAppVersion2CreationTimestamp.Format(tools.PrettyFrontendTimeLayout))
	assert.True(t, sampleApp.OpenButtonPresent)
	assert.False(t, sampleApp.OpenButtonEnabled)
	assert.False(t, sampleApp.IsRunning)
	assert.Equal(t, "Not running", sampleApp.Status)

	installedAppsPage.StartAppViaOperations("sampleapp").
		AssertAppStatusAndOpenButtonEventually("sampleapp", true, true).
		StopAppViaOperations("sampleapp").
		AssertAppStatusAndOpenButtonEventually("sampleapp", false, false)
}

func TestSampleAppUpdateViaGui(t *testing.T) {
	frame := frontend_pages.Setup(t)
	defer frame.Client.Test.ResetTestState()

	_, err := component.InstallSample(t, frame.Client, "1.0")
	assert.Nil(t, err)

	sampleApp := component.GetInstalledSample(t, frame.Client)
	assert.Equal(t, "1.0", sampleApp.VersionName)

	frame.Pages.OpenInstalledAppsPage().UpdateAppViaOperations("sampleapp")
	updatedSampleApp := component.GetInstalledSample(t, frame.Client)
	assert.Equal(t, "2.0", updatedSampleApp.VersionName)
}

func TestSampleAppDeleteViaGui(t *testing.T) {
	frame := frontend_pages.Setup(t)
	defer frame.Client.Test.ResetTestState()

	_, err := component.InstallSample(t, frame.Client, "2.0")
	assert.Nil(t, err)
	assert.Equal(t, 2, len(component.ListInstalledApps(t, frame.Client)))

	frame.Pages.OpenInstalledAppsPage().DeleteAppViaOperations("sampleapp")
	installedApps := component.ListInstalledApps(t, frame.Client)
	assert.Equal(t, 1, len(installedApps))
}

func TestInstalledAppsTableColumnsByRole(t *testing.T) {
	frame := frontend_pages.Setup(t)
	defer frame.Client.Test.ResetTestState()

	_, err := component.InstallSample(t, frame.Client, "2.0")
	assert.Nil(t, err)

	installedAppsPage := frame.Pages.OpenInstalledAppsPage().AssertHeaderColumnCount(8)
	installedAppsPage.
		StartAppViaOperations("sampleapp").
		AssertAppStatusAndOpenButtonEventually("sampleapp", true, true).
		SetAccessPolicyPublic("sampleapp")

	users, err := frame.Client.Users.List()
	assert.Nil(t, err)
	for _, user := range users {
		if user.Username == username {
			err = frame.Client.Users.Delete(user.Id)
			assert.Nil(t, err)
			break
		}
	}
	component.InviteUserAndSetPassword(t, frame.Client, username, component.SampleUserPassword, sampleUserEmail)
	frame.Session.SignInViaClient(username, component.SampleUserPassword)
	frame.Pages.OpenInstalledAppsPage().AssertHeaderColumnCount(2)

	frame.Session.SignOutViaClient()
	frame.Pages.GoToInstalledAppsPage().AssertHeaderColumnCount(2)
}

func TestOpenRunningAppViaOpenButtonAsAuthenticatedAndAnonymous(t *testing.T) {
	frame := frontend_pages.Setup(t)
	defer frame.Client.Test.ResetTestState()

	_, err := component.InstallSample(t, frame.Client, "2.0")
	assert.Nil(t, err)

	frame.Pages.OpenInstalledAppsPage().
		StartAppViaOperations("sampleapp").
		AssertAppStatusAndOpenButtonEventually("sampleapp", true, true).
		OpenSampleAppInNewTabAndAssertSampleAppContent().
		SetAccessPolicyPublic("sampleapp")
	frame.Session.SignOut().Pages.GoToInstalledAppsPage().OpenSampleAppInNewTabAndAssertSampleAppContent()
}

func TestSampleAppAccessPolicySelection(t *testing.T) {
	frame := frontend_pages.Setup(t)
	defer frame.Client.Test.ResetTestState()

	_, err := component.InstallSample(t, frame.Client, "2.0")
	assert.Nil(t, err)

	frame.Pages.OpenInstalledAppsPage().
		AssertNoAccessPolicySelector("postgres").
		AssertAccessPolicyOptionPresent("sampleapp", "Group restricted").
		AssertSelectedAccessPolicyEventually("sampleapp", "Admin only").
		SetAccessPolicyGroupRestricted("sampleapp").AssertSelectedAccessPolicyEventually("sampleapp", "Group restricted").
		SetAccessPolicyAuthenticated("sampleapp").AssertSelectedAccessPolicyEventually("sampleapp", "Authenticated").
		SetAccessPolicyPublic("sampleapp").AssertSelectedAccessPolicyEventually("sampleapp", "Public").
		SetAccessPolicyAdminOnly("sampleapp").AssertSelectedAccessPolicyEventually("sampleapp", "Admin only")
}
