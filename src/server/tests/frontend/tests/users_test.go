//go:build frontend

package frontend

import (
	"fmt"
	"server/configs"
	"server/tests/component"
	"server/tests/frontend_pages"
	"server/tools"
	"testing"

	"github.com/quollix/common/quollix/api_client"

	"github.com/quollix/common/assert"
	u "github.com/quollix/common/utils"
)

func TestUsersPageInvitationStateChangesAfterSetPasswordViaClient(t *testing.T) {
	frame := frontend_pages.Setup(t)
	defer frame.Client.Test.ResetTestState()

	page := frame.Pages.GoToUsersPage()
	users := page.ListUsers()
	assert.Equal(t, 1, len(users))

	admin := page.GetRequiredUser(tools.DefaultAdminName)
	assert.Equal(t, tools.DefaultAdminName, admin.Name)
	assert.Equal(t, tools.DefaultAdminEmail, admin.Email)
	assertFrontendTimestampSet(t, admin.Created)
	assert.Equal(t, "—", admin.InvitationExpiration)
	assert.False(t, admin.PasswordLinkPresent)
	assert.Equal(t, "—", admin.PasswordLinkCellText)
	assert.True(t, admin.EditButtonPresent)
	assert.False(t, admin.ResetButtonPresent)

	page.CreateUser(username, sampleUserEmail).AssertUserInList(username, sampleUserEmail)
	users = page.ListUsers()
	assert.Equal(t, 2, len(users))

	user := page.GetRequiredUser(username)
	assert.Equal(t, username, user.Name)
	assert.Equal(t, sampleUserEmail, user.Email)
	assertFrontendTimestampSet(t, user.Created)
	assertFrontendTimestampSet(t, user.InvitationExpiration)
	assert.True(t, user.PasswordLinkPresent)
	assert.True(t, user.EditButtonPresent)
	assert.True(t, user.ResetButtonPresent)

	userFromBackend := component.GetRequiredUserByUsername(t, frame.Client, username)
	anonymousClient := api_client.NewQuollixClient()
	err := anonymousClient.Users.SetPasswordViaToken(component.SampleUserPassword, userFromBackend.SetPasswordToken)
	assert.Nil(t, err)

	frame.Browser.ReloadPage()
	user = page.GetRequiredUser(username)
	assert.Equal(t, "—", user.InvitationExpiration)
	assert.False(t, user.PasswordLinkPresent)
	assert.Equal(t, "—", user.PasswordLinkCellText)
	assert.True(t, user.EditButtonPresent)
	assert.True(t, user.ResetButtonPresent)
}

func TestUsersPageRestrictsAdminActionsOnOwnAccount(t *testing.T) {
	frame := frontend_pages.Setup(t)
	defer frame.Client.Test.ResetTestState()

	page := frame.Pages.GoToUsersPage()
	admin := page.GetRequiredUser(tools.DefaultAdminName)
	assert.Equal(t, "admin", admin.Role)
	assert.True(t, admin.IsEnabled)
	assert.False(t, admin.CanChangeRole)
	assert.False(t, admin.CanEnableOrDisableUser)
	assert.False(t, admin.CanDeleteUser)
	assert.False(t, admin.ResetButtonPresent)
}

func TestUsersPageCanPromoteAndDemoteOtherUsersBeforeDeactivation(t *testing.T) {
	frame := frontend_pages.Setup(t)
	defer frame.Client.Test.ResetTestState()

	page := frame.Pages.GoToUsersPage()

	page.CreateUser(username, sampleUserEmail)
	user := page.GetRequiredUser(username)
	assert.Equal(t, "user", user.Role)
	assert.True(t, user.IsEnabled)
	assert.True(t, user.CanChangeRole)
	assert.True(t, user.CanEnableOrDisableUser)
	assert.True(t, user.CanDeleteUser)

	page.SetUserRole(username, "Admin")
	user = page.GetRequiredUser(username)
	assert.Equal(t, "admin", user.Role)
	assert.False(t, user.CanEnableOrDisableUser)
	assert.True(t, user.CanDeleteUser)
	assert.True(t, user.ResetButtonPresent)

	page.SetUserRole(username, "User")
	user = page.GetRequiredUser(username)
	assert.Equal(t, "user", user.Role)
	assert.True(t, user.CanEnableOrDisableUser)
	page.SetUserEnabled(username, false)
	user = page.GetRequiredUser(username)
	assert.False(t, user.IsEnabled)
	assert.False(t, user.CanPromoteToAdmin)
}

func TestUsersPageDirectInviteAllowsEmptyEmail(t *testing.T) {
	frame := frontend_pages.Setup(t)
	defer frame.Client.Test.ResetTestState()

	page := frame.Pages.GoToUsersPage()
	page.
		AssertInviteEmailRequired(false).
		CreateUser(username, "").
		AssertUserInList(username, username+"@example.invalid")
}

func TestUsersPageCanDisableAndEnableUser(t *testing.T) {
	frame := frontend_pages.Setup(t)
	defer frame.Client.Test.ResetTestState()

	page := frame.Pages.GoToUsersPage()
	page.CreateUser(username, sampleUserEmail).AssertUserInList(username, sampleUserEmail)

	page.SetUserEnabled(username, false)
	user := page.GetRequiredUser(username)
	assert.False(t, user.IsEnabled)
	assert.False(t, user.CanPromoteToAdmin)

	waitForBackendUserEnabledState(t, frame.Client, username, false)

	page.SetUserEnabled(username, true)
	user = page.GetRequiredUser(username)
	assert.True(t, user.IsEnabled)
	assert.True(t, user.CanPromoteToAdmin)

	waitForBackendUserEnabledState(t, frame.Client, username, true)
}

func waitForBackendUserEnabledState(t *testing.T, client *api_client.QuollixClient, username string, expectedIsEnabled bool) {
	err := u.Eventually(func() error {
		user, exists, err := client.Users.GetByUsername(username)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("user does not exist: %s", username)
		}
		if user.IsEnabled != expectedIsEnabled {
			return fmt.Errorf("expected user enabled state to be %v, got %v", expectedIsEnabled, user.IsEnabled)
		}
		return nil
	})
	assert.Nil(t, err)
}

func TestUserEditPageUpdatesAdminUsernameAndEmail(t *testing.T) {
	frame := frontend_pages.Setup(t)
	defer frame.Client.Test.ResetTestState()

	usersPage := frame.Pages.GoToUsersPage()
	users := usersPage.ListUsers()
	assert.Equal(t, 1, len(users))

	userEditPage := usersPage.OpenEditPageForUser(tools.DefaultAdminName)
	userEditPage.AssertCurrentValues(tools.DefaultAdminName, tools.DefaultAdminEmail)

	newUsername := "newadmin"
	userEditPage.ChangeUsername(newUsername)
	assert.Equal(t, newUsername, userEditPage.GetUsername())

	newEmail := "newadmin@example.invalid"
	userEditPage.ChangeEmail(newEmail)
	assert.Equal(t, newEmail, userEditPage.GetEmail())
}

func TestUsersPageSendPasswordResetEmail(t *testing.T) {
	frame := frontend_pages.Setup(t)
	defer frame.Client.Test.ResetTestState()

	usersPage := frame.Pages.GoToUsersPage()
	usersPage.CreateUser(username, sampleUserEmail).AssertUserInList(username, sampleUserEmail)

	user := usersPage.GetRequiredUser(username)
	assert.False(t, user.SendPasswordResetEmailButtonPresent)

	assert.Nil(t, frame.Client.Email.SaveConfig(configs.GetSampleEmailConfig()))

	frame.Pages.GoToUsersPage()
	usersPage = frame.Pages.UsersPage
	usersPage.SetUserRole(username, "Admin")
	user = usersPage.GetRequiredUser(username)
	assert.True(t, user.ResetButtonPresent)
	assert.True(t, user.SendPasswordResetEmailButtonPresent)

	usersPage.SendPasswordResetEmail(username)
}

func TestUsersPageInviteViaEmailButtonRequiresEmail(t *testing.T) {
	frame := frontend_pages.Setup(t)
	defer frame.Client.Test.ResetTestState()

	assert.Nil(t, frame.Client.Email.SaveConfig(configs.GetSampleEmailConfig()))

	page := frame.Pages.GoToUsersPage()
	page.AssertInviteViaEmailButtonDisabled(true)
	page.SetInviteEmail(sampleUserEmail).AssertInviteViaEmailButtonDisabled(false)
}
