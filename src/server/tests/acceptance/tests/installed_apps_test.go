//go:build acceptance

package acceptance

import (
	"server/tests/acceptance/pages"
	"server/tests/component"
	"server/tools"
	"strings"
	"testing"

	"github.com/quollix/common/assert"
)

func TestInstalledAppPage(t *testing.T) {
	frame := pages.Setup(t)
	defer frame.Client.Test.ResetTestState()

	frame.Client.Apps.InstallSample("2.0")

	installedAppsPage := frame.OpenInstalledAppsPage().AssertHeaderColumnCount(8)
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
	frame := pages.Setup(t)
	defer frame.Client.Test.ResetTestState()

	_, err := frame.Client.Apps.InstallSample("1.0")
	assert.Nil(t, err)

	sampleApp := frame.Client.Apps.GetInstalledSample()
	assert.Equal(t, "1.0", sampleApp.VersionName)

	frame.OpenInstalledAppsPage().UpdateAppViaOperations("sampleapp")
	updatedSampleApp := frame.Client.Apps.GetInstalledSample()
	assert.Equal(t, "2.0", updatedSampleApp.VersionName)
}

func TestSampleAppDeleteViaGui(t *testing.T) {
	frame := pages.Setup(t)
	defer frame.Client.Test.ResetTestState()

	_, err := frame.Client.Apps.InstallSample("2.0")
	assert.Nil(t, err)
	assert.Equal(t, 2, len(frame.Client.Apps.ListInstalled()))

	frame.OpenInstalledAppsPage().DeleteAppViaOperations("sampleapp")
	installedApps := frame.Client.Apps.ListInstalled()
	assert.Equal(t, 1, len(installedApps))
}

func TestInstalledAppsTableColumnsByRole(t *testing.T) {
	frame := pages.Setup(t)
	defer frame.Client.Test.ResetTestState()

	_, err := frame.Client.Apps.InstallSample("2.0")
	assert.Nil(t, err)

	installedAppsPage := frame.OpenInstalledAppsPage().AssertHeaderColumnCount(8)
	installedAppsPage.
		StartAppViaOperations("sampleapp").
		AssertAppStatusAndOpenButtonEventually("sampleapp", true, true).
		SetAccessPolicyPublic("sampleapp")

	users := frame.Client.Users.List()
	for _, user := range users {
		if user.Username == username {
			err = frame.Client.Users.Delete(user.Id)
			assert.Nil(t, err)
			break
		}
	}
	component.InviteUserAndSetPassword(frame.Client, username, component.SampleUserPassword, sampleUserEmail)
	frame.LoginViaClient(username, component.SampleUserPassword)
	frame.OpenInstalledAppsPage().AssertHeaderColumnCount(2)

	frame.LogoutViaClient()
	frame.GoToInstalledAppsPage().AssertHeaderColumnCount(2)
}

func TestOpenRunningAppViaOpenButtonAsAuthenticatedAndAnonymous(t *testing.T) {
	frame := pages.Setup(t)
	defer frame.Client.Test.ResetTestState()

	_, err := frame.Client.Apps.InstallSample("2.0")
	assert.Nil(t, err)

	frame.OpenInstalledAppsPage().
		StartAppViaOperations("sampleapp").
		AssertAppStatusAndOpenButtonEventually("sampleapp", true, true).
		OpenSampleAppInNewTabAndAssertContent().
		SetAccessPolicyPublic("sampleapp")
	frame.Logout().GoToInstalledAppsPage().OpenSampleAppInNewTabAndAssertContent()
}

func TestSampleAppAccessPolicySelection(t *testing.T) {
	frame := pages.Setup(t)
	defer frame.Client.Test.ResetTestState()

	_, err := frame.Client.Apps.InstallSample("2.0")
	assert.Nil(t, err)

	frame.OpenInstalledAppsPage().
		AssertNoAccessPolicySelector("postgres").
		AssertAccessPolicyOptionPresent("sampleapp", "Group restricted").
		AssertSelectedAccessPolicyEventually("sampleapp", "Admin only").
		SetAccessPolicyGroupRestricted("sampleapp").AssertSelectedAccessPolicyEventually("sampleapp", "Group restricted").
		SetAccessPolicyAuthenticated("sampleapp").AssertSelectedAccessPolicyEventually("sampleapp", "Authenticated").
		SetAccessPolicyPublic("sampleapp").AssertSelectedAccessPolicyEventually("sampleapp", "Public").
		SetAccessPolicyAdminOnly("sampleapp").AssertSelectedAccessPolicyEventually("sampleapp", "Admin only")
}
