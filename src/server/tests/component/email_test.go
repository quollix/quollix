//go:build component

package component

import (
	"server/configs"
	"server/email"
	"server/users"
	"testing"

	"github.com/quollix/common/assert"
	u "github.com/quollix/common/utils"
)

func TestInvitationEmailTemplateAndInviteViaEmail(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	readTemplate, err := client.Email.ReadInvitationTemplate()
	assert.Nil(t, err)
	assert.Equal(t, users.DefaultInvitationEmailTemplate, readTemplate)

	template := "Username={{.Username}}\nLink={{.InvitationLink}}\nValidUntil={{.ExpirationDate}}\nServer={{.ServerUrl}}"
	assert.Nil(t, client.Email.SaveInvitationTemplate(template))

	readTemplate, err = client.Email.ReadInvitationTemplate()
	assert.Nil(t, err)
	assert.Equal(t, template, readTemplate)

	assert.Nil(t, client.Email.SaveConfig(configs.GetSampleEmailConfig()))
	assert.Nil(t, client.Email.InviteViaEmail(SampleUsername, SampleUserEmail))

	assert.Nil(t, client.Email.ResetInvitationTemplate())
	resetTemplate, err := client.Email.ReadInvitationTemplate()
	assert.Nil(t, err)
	assert.Equal(t, users.DefaultInvitationEmailTemplate, resetTemplate)
}

func TestEmailConfigWorkflow(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	readEmailConfig, err := client.Email.ReadConfig()
	assert.Nil(t, err)
	assert.Equal(t, *configs.GetEmptyEmailConfig(), *readEmailConfig)

	assert.Nil(t, client.Email.SaveConfig(configs.GetSampleEmailConfig()))

	readEmailConfig, err = client.Email.ReadConfig()
	assert.Nil(t, err)
	assert.Equal(t, *configs.GetSampleEmailConfig(), *readEmailConfig)

	assert.Nil(t, client.Email.ResetConfig())

	readEmailConfig, err = client.Email.ReadConfig()
	assert.Nil(t, err)
	assert.Equal(t, *configs.GetEmptyEmailConfig(), *readEmailConfig)
}

func TestInviteViaEmailFailsWhenEmailIsDisabled(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	err := client.Email.InviteViaEmail(SampleUsername, SampleUserEmail)
	assert.NotNil(t, err)
	u.AssertDeepStackErrorFromRequest(t, err, u.EmailServiceNotEnabledErrorMessage)
}

func TestSendTestEmail(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	assert.Nil(t, client.Email.SaveConfig(configs.GetSampleEmailConfig()))
	assert.Nil(t, client.Email.SendTestEmail(configs.GetSampleEmailConfig(), email.SampleTestRecipientEmail))
}

func TestInviteViaEmailReturnsUserAlreadyExistsError(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	assert.Nil(t, client.Email.SaveConfig(configs.GetSampleEmailConfig()))
	InviteUserAndSetPassword(client, SampleUsername, SampleUserPassword, SampleUserEmail)

	err := client.Email.InviteViaEmail(SampleUsername, "another@example.com")
	assert.NotNil(t, err)
	u.AssertDeepStackErrorFromRequest(t, err, users.UserAlreadyExistsError)
}

func TestInviteViaEmailReturnsEmailAlreadyExistsError(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	assert.Nil(t, client.Email.SaveConfig(configs.GetSampleEmailConfig()))
	InviteUserAndSetPassword(client, SampleUsername, SampleUserPassword, SampleUserEmail)

	err := client.Email.InviteViaEmail("another-username", SampleUserEmail)
	assert.NotNil(t, err)
	u.AssertDeepStackErrorFromRequest(t, err, users.EmailAlreadyExistsError)
}

func TestSendPasswordResetEmail(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	InviteUserAndSetPassword(client, SampleUsername, SampleUserPassword, SampleUserEmail)
	user := client.Users.GetByUsername(SampleUsername)
	firstToken := user.SetPasswordToken

	err := client.Email.ResetPasswordViaEmail(user.Id)
	assert.NotNil(t, err)
	u.AssertDeepStackErrorFromRequest(t, err, u.EmailServiceNotEnabledErrorMessage)

	assert.Nil(t, client.Email.SaveConfig(configs.GetSampleEmailConfig()))

	err = client.Email.ResetPasswordViaEmail(user.Id)
	assert.Nil(t, err)

	user = client.Users.GetByUsername(SampleUsername)
	assert.Equal(t, 64, len(user.SetPasswordToken))
	assert.True(t, user.SetPasswordToken != firstToken)
}

func TestSendPasswordResetEmailReturnsOwnPasswordError(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	assert.Nil(t, client.Email.SaveConfig(configs.GetSampleEmailConfig()))

	adminUser, err := client.Auth.GetCurrentUser()
	assert.Nil(t, err)
	err = client.Email.ResetPasswordViaEmail(adminUser.Id)
	assert.NotNil(t, err)
	u.AssertDeepStackErrorFromRequest(t, err, users.AdminCanNotResetOwnPasswordError)
}
