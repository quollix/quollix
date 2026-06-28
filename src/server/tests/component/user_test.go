//go:build component

package component

import (
	"fmt"
	"server/apps_basic"
	"server/tests/api_client"
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

	userList := ListUsers(t, client)
	assert.Equal(t, 1, len(userList))
	adminUser := userList[0]
	assert.Equal(t, "admin", adminUser.Username)
	assert.True(t, adminUser.IsAdmin)
	assert.True(t, adminUser.IsEnabled)
	assert.Equal(t, tools.DefaultAdminEmail, adminUser.Email)
	assert.Equal(t, tools.DefaultTime, adminUser.SetPasswordTokenExpirationDate)
	assert.True(t, adminUser.CreationDate.After(time.Now().UTC().Add(-12*time.Hour)))
	assert.True(t, adminUser.CreationDate.Before(time.Now().UTC().Add(5*time.Minute)))

	InviteUserAndSetPassword(t, client, SampleUsername, "userpassword", SampleUserEmail)
	userList = ListUsers(t, client)
	assert.Equal(t, 2, len(userList))
	normalUser, err := getNormalUser(userList)
	assert.Nil(t, err)
	assert.Equal(t, SampleUsername, normalUser.Username)
	assert.False(t, normalUser.IsAdmin)
	assert.True(t, normalUser.IsEnabled)

	assert.Nil(t, client.Users.Delete(normalUser.Id))
	userList = ListUsers(t, client)
	assert.Equal(t, 1, len(userList))
	adminUser = userList[0]
	assert.Equal(t, "admin", adminUser.Username)
	assert.True(t, adminUser.IsAdmin)
}

func TestDisablingUserBlocksLoginAndInvalidatesSessions(t *testing.T) {
	adminClient := GetClientAndLogin(t)
	defer adminClient.Test.ResetTestState()
	prepareAuthenticatedSampleApp(t, adminClient)

	InviteUserAndSetPassword(t, adminClient, SampleUsername, SampleUserPassword, SampleUserEmail)
	user := GetRequiredUserByUsername(t, adminClient, SampleUsername)
	userClient := api_client.NewQuollixClient()
	assert.Nil(t, userClient.Auth.SignIn(SampleUsername, SampleUserPassword))
	appClient := GetAppClient(t, userClient)
	appCookie := *appClient.Parent.Cookie

	assert.Nil(t, adminClient.Users.SetEnabled(user.Id, false))
	disabledUser := GetRequiredUserByUsername(t, adminClient, SampleUsername)
	assert.False(t, disabledUser.IsEnabled)
	u.AssertDeepStackErrorFromRequest(t, userClient.Auth.SignIn(SampleUsername, SampleUserPassword), users.UserDisabledError)
	_, err := userClient.Auth.GetCurrentUser()
	// Disabling a user deletes existing sessions, so active sessions fail as missing cookies rather than disabled users.
	u.AssertDeepStackErrorFromRequest(t, err, users.CookieNotFoundError)
	err = AssertSampleAppContentWithCookie(&appCookie)
	assert.NotNil(t, err)
	u.AssertDeepStackErrorFromRequest(t, err, apps_basic.AccessDeniedError)

	assert.Nil(t, adminClient.Users.SetEnabled(user.Id, true))
	enabledUser := GetRequiredUserByUsername(t, adminClient, SampleUsername)
	assert.True(t, enabledUser.IsEnabled)
	assert.Nil(t, userClient.Auth.SignIn(SampleUsername, SampleUserPassword))
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
	assert.Nil(t, client.Auth.SignOut())
	err := client.Auth.SignOut()
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

	InviteUserAndSetPassword(t, client, SampleUsername, "userpassword", SampleUserEmail)
	assert.Nil(t, client.Auth.SignIn(SampleUsername, "userpassword"))
	user, err = client.Auth.GetCurrentUser()
	assert.Nil(t, err)
	assert.Equal(t, SampleUsername, user.Username)
	assert.False(t, user.IsAdmin)
}

func TestAdminCantDeleteHisOwnAccount(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()
	userList := ListUsers(t, client)
	assert.Equal(t, 1, len(userList))
	adminUserId := userList[0].Id
	err := client.Users.Delete(adminUserId)
	assert.NotNil(t, err)
	u.AssertDeepStackErrorFromRequest(t, err, users.AdminCanNotDeleteOwnAccountError)
}

func TestAdminCantResetHisOwnPassword(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()
	userList := ListUsers(t, client)
	assert.Equal(t, 1, len(userList))
	adminUserId := userList[0].Id
	err := client.Users.ResetPassword(adminUserId)
	assert.NotNil(t, err)
	u.AssertDeepStackErrorFromRequest(t, err, users.AdminCanNotResetOwnPasswordError)
}

func TestAdminCantDisableHisOwnAccount(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()
	userList := ListUsers(t, client)
	assert.Equal(t, 1, len(userList))
	adminUserId := userList[0].Id
	err := client.Users.SetEnabled(adminUserId, false)
	assert.NotNil(t, err)
	u.AssertDeepStackErrorFromRequest(t, err, users.AdminCanNotDisableOwnAccountError)
}

func TestChangePassword(t *testing.T) {
	adminClient := GetClientAndLogin(t)
	defer adminClient.Test.ResetTestState()

	InviteUserAndSetPassword(t, adminClient, SampleUsername, "userpassword", SampleUserEmail)
	userClient := api_client.NewQuollixClient()
	assert.Nil(t, userClient.Auth.SignIn(SampleUsername, "userpassword"))

	oldPassword := "userpassword"
	newPassword := oldPassword + "2"

	err := userClient.Auth.ChangePassword("wrongpassword", newPassword)
	u.AssertDeepStackErrorFromRequest(t, err, users.IncorrectCurrentPasswordError)

	assert.Nil(t, userClient.Auth.ChangePassword(oldPassword, newPassword))
	assert.NotNil(t, userClient.Auth.SignIn(SampleUsername, oldPassword))
	assert.Nil(t, userClient.Auth.SignIn(SampleUsername, newPassword))
}

func TestUserInvitationFlow(t *testing.T) {
	adminClient := GetClientAndLogin(t)
	defer adminClient.Test.ResetTestState()

	assert.Nil(t, adminClient.Users.Invite(SampleUsername, SampleUserEmail))
	user := GetRequiredUserByUsername(t, adminClient, SampleUsername)
	assert.Equal(t, 64, len(user.SetPasswordToken))

	anonymousClient := api_client.NewQuollixClient()
	assert.NotNil(t, anonymousClient.Auth.SignIn(SampleUsername, SampleUserPassword))
	assert.Nil(t, anonymousClient.Users.SetPasswordViaToken(SampleUserPassword, user.SetPasswordToken))
	err := anonymousClient.Users.SetPasswordViaToken("another-password", user.SetPasswordToken)
	u.AssertDeepStackErrorFromRequest(t, err, users.UserNotFoundError)
	assert.Nil(t, anonymousClient.Auth.SignIn(SampleUsername, SampleUserPassword))

	assert.Nil(t, adminClient.Users.ResetPassword(user.Id))
	assert.NotNil(t, anonymousClient.Auth.SignIn(SampleUsername, SampleUserPassword))
	user = GetRequiredUserByUsername(t, adminClient, SampleUsername)
	assert.Nil(t, anonymousClient.Users.SetPasswordViaToken("new-password", user.SetPasswordToken))
	err = anonymousClient.Users.SetPasswordViaToken("another-new-password", user.SetPasswordToken)
	u.AssertDeepStackErrorFromRequest(t, err, users.UserNotFoundError)
	assert.Nil(t, anonymousClient.Auth.SignIn(SampleUsername, "new-password"))
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

	assert.Nil(t, adminClient.Users.Invite(SampleUsername, SampleRealUserEmail))
	err := adminClient.Users.Invite("user2", SampleRealUserEmail)
	u.AssertDeepStackErrorFromRequest(t, err, users.EmailAlreadyExistsError)
}

func TestChangeUsername(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	userList := ListUsers(t, client)
	assert.Equal(t, 1, len(userList))
	adminUser := userList[0]

	assert.Nil(t, client.Users.ChangeUsername(adminUser.Id, "newadmin"))
	userList = ListUsers(t, client)
	assert.Equal(t, 1, len(userList))
	assert.Equal(t, "newadmin", userList[0].Username)

	assert.Nil(t, client.Users.Invite(SampleUsername, SampleUserEmail))
	err := client.Users.ChangeUsername(adminUser.Id, SampleUsername)
	u.AssertDeepStackErrorFromRequest(t, err, users.UserAlreadyExistsError)
}

func TestChangeEmail(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	userList := ListUsers(t, client)
	assert.Equal(t, 1, len(userList))
	adminUser := userList[0]

	assert.Nil(t, client.Users.ChangeEmail(adminUser.Id, "newadmin@example.com"))
	userList = ListUsers(t, client)
	assert.Equal(t, 1, len(userList))
	assert.Equal(t, "newadmin@example.com", userList[0].Email)

	assert.Nil(t, client.Users.Invite(SampleUsername, SampleRealUserEmail))
	err := client.Users.ChangeEmail(adminUser.Id, SampleRealUserEmail)
	u.AssertDeepStackErrorFromRequest(t, err, users.EmailAlreadyExistsError)
}

func TestInviteUserRejectsMismatchedReservedEmail(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	err := client.Users.Invite("tom", "alice@example.invalid")
	u.AssertDeepStackErrorFromRequest(t, err, users.ReservedEmailMustMatchUserError)
}

func TestInviteUserAcceptsMatchingReservedEmail(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	assert.Nil(t, client.Users.Invite("tom", "tom@example.invalid"))
	user := GetRequiredUserByUsername(t, client, "tom")
	assert.Equal(t, "tom@example.invalid", user.Email)
}

func TestChangeEmailRejectsMismatchedReservedEmail(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	userList := ListUsers(t, client)
	assert.Equal(t, 1, len(userList))
	adminUser := userList[0]
	err := client.Users.ChangeEmail(adminUser.Id, "tom@example.invalid")
	u.AssertDeepStackErrorFromRequest(t, err, users.ReservedEmailMustMatchUserError)
}

func TestChangeEmailAcceptsMatchingReservedEmail(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	assert.Nil(t, client.Users.Invite("tom", "tom@example.com"))
	user := GetRequiredUserByUsername(t, client, "tom")
	assert.Nil(t, client.Users.ChangeEmail(user.Id, "tom@example.invalid"))

	user = GetRequiredUserByUsername(t, client, "tom")
	assert.Equal(t, "tom@example.invalid", user.Email)
}

func TestChangeUsernameUpdatesReservedEmail(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	assert.Nil(t, client.Users.Invite("tom", "tom@example.invalid"))
	user := GetRequiredUserByUsername(t, client, "tom")
	assert.Nil(t, client.Users.ChangeUsername(user.Id, "thomas"))

	user = GetRequiredUserByUsername(t, client, "thomas")
	assert.Equal(t, "thomas@example.invalid", user.Email)
}

func TestChangeUsernameKeepsRealEmail(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	assert.Nil(t, client.Users.Invite("tom", "tom@example.com"))
	user := GetRequiredUserByUsername(t, client, "tom")
	assert.Nil(t, client.Users.ChangeUsername(user.Id, "thomas"))

	user = GetRequiredUserByUsername(t, client, "thomas")
	assert.Equal(t, "tom@example.com", user.Email)
}
