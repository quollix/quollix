//go:build component

package component

import (
	"fmt"
	"server/tools"
	"server/users"
	"testing"
	"time"

	"github.com/quollix/common/assert"
	u "github.com/quollix/common/utils"
)

func TestUsers(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	userList := client.Users.List()
	assert.Equal(t, 1, len(userList))
	adminUser := userList[0]
	assert.Equal(t, "admin", adminUser.Username)
	assert.True(t, adminUser.IsAdmin)
	assert.Equal(t, tools.DefaultAdminEmail, adminUser.Email)
	assert.Equal(t, tools.DefaultTime, adminUser.SetPasswordTokenExpirationDate)
	assert.True(t, adminUser.CreationDate.After(time.Now().UTC().Add(-12*time.Hour)))
	assert.True(t, adminUser.CreationDate.Before(time.Now().UTC().Add(5*time.Minute)))

	InviteUserAndSetPassword(client, SampleUsername, "userpassword", SampleUserEmail)
	userList = client.Users.List()
	assert.Equal(t, 2, len(userList))
	normalUser, err := getNormalUser(userList)
	assert.Nil(t, err)
	assert.Equal(t, SampleUsername, normalUser.Username)
	assert.False(t, normalUser.IsAdmin)

	assert.Nil(t, client.Users.Delete(normalUser.Id))
	userList = client.Users.List()
	assert.Equal(t, 1, len(userList))
	adminUser = userList[0]
	assert.Equal(t, "admin", adminUser.Username)
	assert.True(t, adminUser.IsAdmin)
}

func getNormalUser(userList []tools.User) (*tools.User, error) {
	for _, user := range userList {
		if !user.IsAdmin {
			userCopy := user
			return &userCopy, nil
		}
	}
	return nil, fmt.Errorf("normal user not found")
}

func TestLogout(t *testing.T) {
	client := GetClientAndLogin(t)
	assert.Nil(t, client.Auth.Logout())
	err := client.Auth.Logout()
	assert.NotNil(t, err)
	u.AssertDeepStackErrorFromRequest(t, err, "cookie not found")
}

func TestGetCurrentUser_ReturnsLoggedInUser(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()
	user, err := client.Auth.GetCurrentUser()
	assert.Nil(t, err)
	assert.Equal(t, "admin", user.Username)
	assert.True(t, user.IsAdmin)

	InviteUserAndSetPassword(client, SampleUsername, "userpassword", SampleUserEmail)
	assert.Nil(t, client.Auth.Login(SampleUsername, "userpassword"))
	user, err = client.Auth.GetCurrentUser()
	assert.Nil(t, err)
	assert.Equal(t, SampleUsername, user.Username)
	assert.False(t, user.IsAdmin)
}

func TestAdminCantDeleteHisOwnAccount(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()
	userList := client.Users.List()
	assert.Equal(t, 1, len(userList))
	adminUserId := userList[0].Id
	err := client.Users.Delete(adminUserId)
	assert.NotNil(t, err)
	u.AssertDeepStackErrorFromRequest(t, err, users.AdminCanNotDeleteOwnAccountError)
}

func TestAdminCantResetHisOwnPassword(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()
	userList := client.Users.List()
	assert.Equal(t, 1, len(userList))
	adminUserId := userList[0].Id
	err := client.Users.ResetPassword(adminUserId)
	assert.NotNil(t, err)
	u.AssertDeepStackErrorFromRequest(t, err, users.AdminCanNotResetOwnPasswordError)
}

func TestChangePassword(t *testing.T) {
	adminClient := GetClientAndLogin(t)
	defer adminClient.Test.ResetTestState()

	InviteUserAndSetPassword(adminClient, SampleUsername, "userpassword", SampleUserEmail)
	userClient := GetQuollixClient(t)
	assert.Nil(t, userClient.Auth.Login(SampleUsername, "userpassword"))

	oldPassword := "userpassword"
	newPassword := oldPassword + "2"

	err := userClient.Auth.ChangePassword("wrongpassword", newPassword)
	u.AssertDeepStackErrorFromRequest(t, err, users.IncorrectCurrentPasswordError)

	assert.Nil(t, userClient.Auth.ChangePassword(oldPassword, newPassword))
	assert.NotNil(t, userClient.Auth.Login(SampleUsername, oldPassword))
	assert.Nil(t, userClient.Auth.Login(SampleUsername, newPassword))
}

func TestUserInvitationFlow(t *testing.T) {
	adminClient := GetClientAndLogin(t)
	defer adminClient.Test.ResetTestState()

	assert.Nil(t, adminClient.Users.Invite(SampleUsername, SampleUserEmail))
	user := adminClient.Users.GetByUsername(SampleUsername)
	assert.Equal(t, 64, len(user.SetPasswordToken))

	anonymousClient := GetQuollixClient(t)
	assert.NotNil(t, anonymousClient.Auth.Login(SampleUsername, SampleUserPassword))
	assert.Nil(t, anonymousClient.Users.SetPasswordViaToken(SampleUserPassword, user.SetPasswordToken))
	err := anonymousClient.Users.SetPasswordViaToken("another-password", user.SetPasswordToken)
	u.AssertDeepStackErrorFromRequest(t, err, users.UserNotFoundError)
	assert.Nil(t, anonymousClient.Auth.Login(SampleUsername, SampleUserPassword))

	assert.Nil(t, adminClient.Users.ResetPassword(user.Id))
	assert.NotNil(t, anonymousClient.Auth.Login(SampleUsername, SampleUserPassword))
	user = adminClient.Users.GetByUsername(SampleUsername)
	assert.Nil(t, anonymousClient.Users.SetPasswordViaToken("new-password", user.SetPasswordToken))
	err = anonymousClient.Users.SetPasswordViaToken("another-new-password", user.SetPasswordToken)
	u.AssertDeepStackErrorFromRequest(t, err, users.UserNotFoundError)
	assert.Nil(t, anonymousClient.Auth.Login(SampleUsername, "new-password"))
}

func TestInvitingExistingUsernameFails(t *testing.T) {
	adminClient := GetClientAndLogin(t)
	defer adminClient.Test.ResetTestState()

	assert.Nil(t, adminClient.Users.Invite(SampleUsername, SampleUserEmail))
	err := adminClient.Users.Invite(SampleUsername, SampleUserEmail)
	u.AssertDeepStackErrorFromRequest(t, err, users.UserAlreadyExistsError)
}

func TestInvitingExistingUserEmailFails(t *testing.T) {
	adminClient := GetClientAndLogin(t)
	defer adminClient.Test.ResetTestState()

	assert.Nil(t, adminClient.Users.Invite(SampleUsername, SampleUserEmail))
	err := adminClient.Users.Invite("user2", SampleUserEmail)
	u.AssertDeepStackErrorFromRequest(t, err, users.EmailAlreadyExistsError)
}

func TestChangeUsername(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	userList := client.Users.List()
	assert.Equal(t, 1, len(userList))
	adminUser := userList[0]

	assert.Nil(t, client.Users.ChangeUsername(adminUser.Id, "newadmin"))
	userList = client.Users.List()
	assert.Equal(t, 1, len(userList))
	assert.Equal(t, "newadmin", userList[0].Username)

	assert.Nil(t, client.Users.Invite(SampleUsername, SampleUserEmail))
	err := client.Users.ChangeUsername(adminUser.Id, SampleUsername)
	u.AssertDeepStackErrorFromRequest(t, err, users.UserAlreadyExistsError)
}

func TestChangeEmail(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	userList := client.Users.List()
	assert.Equal(t, 1, len(userList))
	adminUser := userList[0]

	assert.Nil(t, client.Users.ChangeEmail(adminUser.Id, "newadmin@example.invalid"))
	userList = client.Users.List()
	assert.Equal(t, 1, len(userList))
	assert.Equal(t, "newadmin@example.invalid", userList[0].Email)

	assert.Nil(t, client.Users.Invite(SampleUsername, SampleUserEmail))
	err := client.Users.ChangeEmail(adminUser.Id, SampleUserEmail)
	u.AssertDeepStackErrorFromRequest(t, err, users.EmailAlreadyExistsError)
}
