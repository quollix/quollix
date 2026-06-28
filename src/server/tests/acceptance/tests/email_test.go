//go:build acceptance

package acceptance

import (
	"server/configs"
	"server/email"
	"server/tests/frontend_pages"
	"testing"

	"github.com/quollix/common/assert"
)

func TestEmailPageSaveAndResetConfig(t *testing.T) {
	frame := frontend_pages.Setup(t)
	defer frame.Client.Test.ResetTestState()

	fillConfig := configs.GetSampleEmailConfig()
	fillConfig.IsEnabled = false

	emailPage := frame.Pages.GoToEmailPage()

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

	frame.Browser.ReloadPage()
	frame.Pages.EmailPage.
		AssertFormValues(fillConfig).
		AssertTestRecipientValue("")

	frame.Pages.EmailPage.Reset().
		AssertFormValues(configs.GetEmptyEmailConfig()).
		AssertTestRecipientValue("")

	resetConfig, err := frame.Client.Email.ReadConfig()
	assert.Nil(t, err)
	assert.Equal(t, *configs.GetEmptyEmailConfig(), *resetConfig)
}

func TestEmailPageSaveAndResetInvitationTemplate(t *testing.T) {
	frame := frontend_pages.Setup(t)
	defer frame.Client.Test.ResetTestState()

	customTemplate := "Username={{.Username}}\nLink={{.InvitationLink}}\nValidUntil={{.ExpirationDate}}\nServer={{.ServerUrl}}"

	emailPage := frame.Pages.GoToEmailPage()
	emailPage.AssertDefaultInvitationTemplate()

	emailPage.
		SetInvitationTemplate(customTemplate).
		SaveInvitationTemplate().
		AssertInvitationTemplateValue(customTemplate)

	frame.Browser.ReloadPage()
	frame.Pages.EmailPage.AssertInvitationTemplateValue(customTemplate)

	frame.Pages.EmailPage.
		ResetInvitationTemplate().
		AssertDefaultInvitationTemplate()
}

func TestEmailPageOidcEmailExposureToggle(t *testing.T) {
	frame := frontend_pages.Setup(t)
	defer frame.Client.Test.ResetTestState()

	frame.Pages.GoToEmailPage().
		AssertExposeRealEmailInOidcToken(false).
		SetExposeRealEmailInOidcToken(true)

	frame.Browser.ReloadPage()
	frame.Pages.EmailPage.AssertExposeRealEmailInOidcToken(true)

	frame.Pages.EmailPage.SetExposeRealEmailInOidcToken(false)

	frame.Browser.ReloadPage()
	frame.Pages.EmailPage.AssertExposeRealEmailInOidcToken(false)
}

func TestEmailPasswordVisibilityToggle(t *testing.T) {
	frame := frontend_pages.Setup(t)
	defer frame.Client.Test.ResetTestState()

	frame.Pages.GoToEmailPage().
		EnterPassword("visible-secret").
		AssertPasswordVisibility(false).
		AssertPasswordValue("visible-secret").
		TogglePasswordVisibility().
		AssertPasswordVisibility(true).
		AssertPasswordValue("visible-secret")
}
