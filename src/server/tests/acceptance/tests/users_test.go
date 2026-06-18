//go:build acceptance

package acceptance

import (
	"server/configs"
	"server/tests/acceptance/pages"
	"server/tests/component"
	"server/tools"
	"testing"

	"github.com/quollix/common/assert"
)

func TestUsersPageInvitationStateChangesAfterSetPasswordViaClient(t *testing.T) {
	frame := pages.Setup(t)
	defer frame.Client.Test.ResetTestState()

	page := frame.GoToUsersPage()
	users := page.ListUsers()
	assert.Equal(t, 1, len(users))

	admin := page.GetRequiredUser(tools.DefaultAdminName)
	assert.Equal(t, tools.DefaultAdminName, admin.Name)
	assert.Equal(t, tools.DefaultAdminEmail, admin.Email)
	assert.Equal(t, "Admin", admin.Role)
	assertFrontendTimestampSet(t, admin.Created)
	assert.Equal(t, "—", admin.InvitationExpiration)
	assert.False(t, admin.PasswordLinkPresent)
	assert.Equal(t, "—", admin.PasswordLinkCellText)
	assert.True(t, admin.EditButtonPresent)
	assert.False(t, admin.ResetButtonPresent)
	assert.False(t, admin.DeleteButtonPresent)

	page.CreateUser(username, sampleUserEmail).AssertUserInList(username, sampleUserEmail)
	users = page.ListUsers()
	assert.Equal(t, 2, len(users))

	user := page.GetRequiredUser(username)
	assert.Equal(t, username, user.Name)
	assert.Equal(t, sampleUserEmail, user.Email)
	assert.Equal(t, "User", user.Role)
	assertFrontendTimestampSet(t, user.Created)
	assertFrontendTimestampSet(t, user.InvitationExpiration)
	assert.True(t, user.PasswordLinkPresent)
	assert.True(t, user.EditButtonPresent)
	assert.True(t, user.ResetButtonPresent)
	assert.True(t, user.DeleteButtonPresent)

	userFromBackend := frame.Client.Users.GetByUsername(username)
	anonymousClient := component.GetQuollixClient(t)
	err := anonymousClient.Users.SetPasswordViaToken(component.SampleUserPassword, userFromBackend.SetPasswordToken)
	assert.Nil(t, err)

	frame.ReloadPage()
	user = page.GetRequiredUser(username)
	assert.Equal(t, "—", user.InvitationExpiration)
	assert.False(t, user.PasswordLinkPresent)
	assert.Equal(t, "—", user.PasswordLinkCellText)
	assert.True(t, user.EditButtonPresent)
	assert.True(t, user.ResetButtonPresent)
	assert.True(t, user.DeleteButtonPresent)
}

func TestUserEditPageUpdatesAdminUsernameAndEmail(t *testing.T) {
	frame := pages.Setup(t)
	defer frame.Client.Test.ResetTestState()

	usersPage := frame.GoToUsersPage()
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
	frame := pages.Setup(t)
	defer frame.Client.Test.ResetTestState()

	usersPage := frame.GoToUsersPage()
	usersPage.CreateUser(username, sampleUserEmail).AssertUserInList(username, sampleUserEmail)

	user := usersPage.GetRequiredUser(username)
	assert.False(t, user.SendPasswordResetEmailButtonPresent)

	assert.Nil(t, frame.Client.Email.SaveConfig(configs.GetSampleEmailConfig()))

	frame.GoToUsersPage()
	user = frame.UsersPage.GetRequiredUser(username)
	assert.True(t, user.SendPasswordResetEmailButtonPresent)

	frame.UsersPage.SendPasswordResetEmail(username)
}
