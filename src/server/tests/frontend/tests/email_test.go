//go:build frontend

package frontend

import (
	"server/configs"
	"server/email"
	"server/tests/frontend_pages"
	"server/tools"
	"server/users"
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
		AssertTestRecipientValue("").
		AssertBroadcastAndInvitationSectionsVisible(false)

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

	frame.Pages.EmailPage.
		AssertFormValues(fillConfig).
		AssertTestRecipientValue("").
		AssertBroadcastAndInvitationSectionsVisible(true)

	fillConfig.IsEnabled = false
	frame.Pages.EmailPage.FillConfig(fillConfig).
		Save().
		AssertFormValues(fillConfig).
		AssertTestRecipientValue("").
		AssertBroadcastAndInvitationSectionsVisible(false)

	disabledConfig, err := frame.Client.Email.ReadConfig()
	assert.Nil(t, err)
	assert.Equal(t, *fillConfig, *disabledConfig)

	frame.Pages.EmailPage.Reset().
		AssertFormValues(configs.GetEmptyEmailConfig()).
		AssertTestRecipientValue("").
		AssertBroadcastAndInvitationSectionsVisible(false)

	resetConfig, err := frame.Client.Email.ReadConfig()
	assert.Nil(t, err)
	assert.Equal(t, *configs.GetEmptyEmailConfig(), *resetConfig)
}

func TestEmailPageSaveAndResetInvitationTemplate(t *testing.T) {
	frame := frontend_pages.Setup(t)
	defer frame.Client.Test.ResetTestState()
	assert.Nil(t, frame.Client.Email.SaveConfig(configs.GetSampleEmailConfig()))

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

func TestEmailPageSaveAndResetBroadcastTemplate(t *testing.T) {
	frame := frontend_pages.Setup(t)
	defer frame.Client.Test.ResetTestState()
	assert.Nil(t, frame.Client.Email.SaveConfig(configs.GetSampleEmailConfig()))

	customTemplate := "Hello,\n\nAn important update for all Quollix users.\n\nBest regards,\nThe administrators"
	emailPage := frame.Pages.GoToEmailPage()
	emailPage.AssertBroadcastTemplateValue(users.DefaultBroadcastEmailTemplate)

	emailPage.
		SetBroadcastBody(customTemplate).
		SaveBroadcastTemplate().
		AssertBroadcastTemplateValue(customTemplate)

	frame.Browser.ReloadPage()
	frame.Pages.EmailPage.AssertBroadcastTemplateValue(customTemplate)
	frame.Pages.EmailPage.
		ResetBroadcastTemplate().
		AssertBroadcastTemplateValue(users.DefaultBroadcastEmailTemplate)
}

func TestEmailPageSendBroadcastEmailShowsSnackbarAndRestoresTemplate(t *testing.T) {
	frame := frontend_pages.Setup(t)
	defer frame.Client.Test.ResetTestState()

	const testUsername = "broadcast-test-user"

	assert.Nil(t, frame.Client.Email.SaveConfig(configs.GetSampleEmailConfig()))
	savedTemplate := "Hello,\n\nThis is the saved broadcast template.\n\nBest regards,\nThe administrators"
	assert.Nil(t, frame.Client.Email.SaveBroadcastTemplate(savedTemplate))
	assert.Nil(t, frame.Client.Users.Invite(testUsername, tools.SampleRealUserEmail))

	frame.Pages.GoToEmailPage().
		SetBroadcastBody(email.TestBroadcastEmailBody).
		SetBroadcastSubject(email.TestBroadcastEmailSubject).
		SendBroadcastEmail().
		AssertBroadcastTemplateValue(savedTemplate)
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
