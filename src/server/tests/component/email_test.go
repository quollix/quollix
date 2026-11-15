//go:build component

package component

import (
	"server/configs"
	"server/email"
	"server/tools"
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

func TestBroadcastEmailTemplateAndSendToUsers(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	readTemplate, err := client.Email.ReadBroadcastTemplate()
	assert.Nil(t, err)
	assert.Equal(t, users.DefaultBroadcastEmailTemplate, readTemplate)
	assert.Nil(t, client.Email.SaveBroadcastTemplate(email.TestBroadcastEmailBody))
	readTemplate, err = client.Email.ReadBroadcastTemplate()
	assert.Nil(t, err)
	assert.Equal(t, email.TestBroadcastEmailBody, readTemplate)

	assert.Nil(t, client.Email.SaveConfig(configs.GetSampleEmailConfig()))
	InviteUserAndSetPassword(t, client, SampleUsername, SampleUserPassword, tools.SampleRealUserEmail)
	assert.Nil(t, client.Email.SendBroadcast(email.TestBroadcastEmailSubject, email.TestBroadcastEmailBody))

	assert.Nil(t, client.Email.ResetBroadcastTemplate())
	readTemplate, err = client.Email.ReadBroadcastTemplate()
	assert.Nil(t, err)
	assert.Equal(t, users.DefaultBroadcastEmailTemplate, readTemplate)
}

func TestBroadcastEmailReturnsExpectedErrors(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	err := client.Email.SaveBroadcastTemplate("")
	assert.NotNil(t, err)
	u.AssertDeepStackErrorFromRequest(t, err, email.BroadcastEmailTemplateCannotBeEmptyError)

	assert.Nil(t, client.Email.SaveBroadcastTemplate("Plain text without placeholders."))

	assert.Nil(t, client.Email.SaveConfig(configs.GetSampleEmailConfig()))
	err = client.Email.SendBroadcast(" ", email.TestBroadcastEmailBody)
	assert.NotNil(t, err)
	u.AssertDeepStackErrorFromRequest(t, err, email.BroadcastEmailSubjectCannotBeEmptyError)

	err = client.Email.SendBroadcast(email.TestBroadcastEmailSubject, "\t")
	assert.NotNil(t, err)
	u.AssertDeepStackErrorFromRequest(t, err, email.BroadcastEmailBodyCannotBeEmptyError)

	assert.Nil(t, client.Email.ResetConfig())
	err = client.Email.SendBroadcast(email.TestBroadcastEmailSubject, email.TestBroadcastEmailBody)
	assert.NotNil(t, err)
	u.AssertDeepStackErrorFromRequest(t, err, u.EmailServiceNotEnabledErrorMessage)
}

func TestBroadcastEmailReturnsPartialDeliveryError(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	assert.Nil(t, client.Email.SaveConfig(configs.GetSampleEmailConfig()))
	assert.Nil(t, client.Email.SaveBroadcastTemplate(email.TestBroadcastEmailBody))
	InviteUserAndSetPassword(t, client, "broadcast-success", SampleUserPassword, tools.SampleRealUserEmail)
	InviteUserAndSetPassword(t, client, "broadcast-failure", SampleUserPassword, "broadcast-failure@example.com")

	err := client.Email.SendBroadcast(email.TestBroadcastEmailSubject, email.TestBroadcastEmailBody)
	assert.NotNil(t, err)
	u.AssertDeepStackErrorFromRequest(t, err, email.PartialEmailDeliveryError)
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

func TestInviteViaEmailRejectsEmptyEmail(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	assert.Nil(t, client.Email.SaveConfig(configs.GetSampleEmailConfig()))
	err := client.Email.InviteViaEmail(SampleUsername, "")
	assert.NotNil(t, err)
	u.AssertDeepStackErrorFromRequest(t, err, "Invalid input. The content of the field Email must be a valid email address.")
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
	InviteUserAndSetPassword(t, client, SampleUsername, SampleUserPassword, SampleUserEmail)

	err := client.Email.InviteViaEmail(SampleUsername, "another@example.com")
	assert.NotNil(t, err)
	u.AssertDeepStackErrorFromRequest(t, err, users.UserAlreadyExistsError)
}

func TestInviteViaEmailReturnsEmailAlreadyExistsError(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	assert.Nil(t, client.Email.SaveConfig(configs.GetSampleEmailConfig()))
	InviteUserAndSetPassword(t, client, SampleUsername, SampleUserPassword, tools.SampleRealUserEmail)

	err := client.Email.InviteViaEmail("another-username", tools.SampleRealUserEmail)
	assert.NotNil(t, err)
	u.AssertDeepStackErrorFromRequest(t, err, users.EmailAlreadyExistsError)
}

func TestSendPasswordResetEmail(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	InviteUserAndSetPassword(t, client, SampleUsername, SampleUserPassword, SampleUserEmail)
	user := GetRequiredUserByUsername(t, client, SampleUsername)
	firstToken := user.SetPasswordToken

	err := client.Email.ResetPasswordViaEmail(user.Id)
	assert.NotNil(t, err)
	u.AssertDeepStackErrorFromRequest(t, err, u.EmailServiceNotEnabledErrorMessage)

	assert.Nil(t, client.Email.SaveConfig(configs.GetSampleEmailConfig()))

	err = client.Email.ResetPasswordViaEmail(user.Id)
	assert.Nil(t, err)

	user = GetRequiredUserByUsername(t, client, SampleUsername)
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
