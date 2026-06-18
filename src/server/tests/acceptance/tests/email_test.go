//go:build acceptance

package acceptance

import (
	"server/configs"
	"server/email"
	"server/tests/acceptance/pages"
	"testing"

	"github.com/quollix/common/assert"
)

func TestEmailPageSaveAndResetConfig(t *testing.T) {
	frame := pages.Setup(t)
	defer frame.Client.Test.ResetTestState()

	fillConfig := configs.GetSampleEmailConfig()
	fillConfig.IsEnabled = false

	emailPage := frame.GoToEmailPage()

	emailPage.
		AssertFormValues(configs.GetEmptyEmailConfig()).
		AssertTestRecipientValue("")

	emailPage.FillConfig(fillConfig).
		TestConnection().
		SetTestRecipient(email.SampleTestRecipientEmail).
		SendTestEmail().
		EnableEmail().
		Save()
	fillConfig.IsEnabled = true

	savedConfig, err := frame.Client.Email.ReadConfig()
	assert.Nil(t, err)
	assert.Equal(t, *fillConfig, *savedConfig)

	frame.ReloadPage()
	frame.EmailPage.
		AssertFormValues(fillConfig).
		AssertTestRecipientValue("")

	frame.EmailPage.Reset().
		AssertFormValues(configs.GetEmptyEmailConfig()).
		AssertTestRecipientValue("")

	resetConfig, err := frame.Client.Email.ReadConfig()
	assert.Nil(t, err)
	assert.Equal(t, *configs.GetEmptyEmailConfig(), *resetConfig)
}

func TestEmailPageSaveAndResetInvitationTemplate(t *testing.T) {
	frame := pages.Setup(t)
	defer frame.Client.Test.ResetTestState()

	customTemplate := "Username={{.Username}}\nLink={{.InvitationLink}}\nValidUntil={{.ExpirationDate}}\nServer={{.ServerUrl}}"

	emailPage := frame.GoToEmailPage()
	emailPage.AssertDefaultInvitationTemplate()

	emailPage.
		SetInvitationTemplate(customTemplate).
		SaveInvitationTemplate().
		AssertInvitationTemplateValue(customTemplate)

	frame.ReloadPage()
	frame.EmailPage.AssertInvitationTemplateValue(customTemplate)

	frame.EmailPage.
		ResetInvitationTemplate().
		AssertDefaultInvitationTemplate()
}

func TestEmailPasswordVisibilityToggle(t *testing.T) {
	frame := pages.Setup(t)
	defer frame.Client.Test.ResetTestState()

	frame.GoToEmailPage().
		EnterPassword("visible-secret").
		AssertPasswordVisibility(false).
		AssertPasswordValue("visible-secret").
		TogglePasswordVisibility().
		AssertPasswordVisibility(true).
		AssertPasswordValue("visible-secret")
}
