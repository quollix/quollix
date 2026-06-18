//go:build acceptance

package pages

import (
	"fmt"
	"server/tools"
	"server/users"

	"github.com/quollix/common/assert"
	u "github.com/quollix/common/utils"
)

type EmailPage struct {
	Frame *FrameType
}

func (e *EmailPage) FillConfig(cfg *u.EmailConfig) *EmailPage {
	SetInputValue(e.Frame.TestingT(), e.Frame.Page(), "#email-smtp-host-input", cfg.SMTPHost)
	SetInputValue(e.Frame.TestingT(), e.Frame.Page(), "#email-smtp-port-input", cfg.SMTPPort)
	SetInputValue(e.Frame.TestingT(), e.Frame.Page(), "#email-from-address-input", cfg.FromEmailAddress)
	SetInputValue(e.Frame.TestingT(), e.Frame.Page(), "#email-account-username-input", cfg.EmailAccountUsername)
	SetInputValue(e.Frame.TestingT(), e.Frame.Page(), "#email-account-password-input", cfg.EmailAccountPassword)
	SetCheckboxValue(e.Frame.TestingT(), e.Frame.Page(), "#email-enabled-checkbox", cfg.IsEnabled)
	return e
}

func (e *EmailPage) EnterPassword(password string) *EmailPage {
	SetInputValue(e.Frame.TestingT(), e.Frame.Page(), "#email-account-password-input", password)
	return e
}

func (e *EmailPage) SetTestRecipient(recipient string) *EmailPage {
	SetInputValue(e.Frame.TestingT(), e.Frame.Page(), "#email-test-recipient-input", recipient)
	return e
}

func (e *EmailPage) EnableEmail() *EmailPage {
	SetCheckboxValue(e.Frame.TestingT(), e.Frame.Page(), "#email-enabled-checkbox", true)
	return e
}

func (e *EmailPage) Save() *EmailPage {
	saveButton, err := e.Frame.Page().Element("#email-save-config-button")
	assert.Nil(e.Frame.TestingT(), err)
	saveButton.MustClick()
	e.Frame.AssertSnackbarVisibleWithTextEventually("Email settings saved.")
	e.Frame.AssertPagePath(tools.Paths.FrontendEmail)
	return e
}

func (e *EmailPage) TestConnection() *EmailPage {
	button, err := e.Frame.Page().Element("#email-test-connection-button")
	assert.Nil(e.Frame.TestingT(), err)
	button.MustClick()
	e.Frame.AssertSnackbarVisibleWithTextEventually("Email connection test successful.")
	return e
}

func (e *EmailPage) SendTestEmail() *EmailPage {
	button, err := e.Frame.Page().Element("#email-send-test-button")
	assert.Nil(e.Frame.TestingT(), err)
	button.MustClick()
	e.Frame.AssertSnackbarVisibleWithTextEventually("Test email sent.")
	return e
}

func (e *EmailPage) Reset() *EmailPage {
	button, err := e.Frame.Page().Element("#email-reset-config-button")
	assert.Nil(e.Frame.TestingT(), err)
	button.MustClick()
	e.Frame.ConfirmDialog()
	e.Frame.AssertSnackbarVisibleWithTextEventually("Email settings have been reset.")
	e.Frame.AssertPagePath(tools.Paths.FrontendEmail)
	return e
}

func (e *EmailPage) SetInvitationTemplate(template string) *EmailPage {
	SetInputValue(e.Frame.TestingT(), e.Frame.Page(), "#invitation-email-template-input", template)
	return e
}

func (e *EmailPage) SaveInvitationTemplate() *EmailPage {
	button, err := e.Frame.Page().Element("#email-save-template-button")
	assert.Nil(e.Frame.TestingT(), err)
	button.MustClick()
	e.Frame.AssertSnackbarVisibleWithTextEventually("Invitation email template saved.")
	return e
}

func (e *EmailPage) ResetInvitationTemplate() *EmailPage {
	button, err := e.Frame.Page().Element("#email-reset-template-button")
	assert.Nil(e.Frame.TestingT(), err)
	button.MustClick()
	e.Frame.ConfirmDialog()
	e.Frame.AssertSnackbarVisibleWithTextEventually("Invitation email template has been reset.")
	e.Frame.AssertPagePath(tools.Paths.FrontendEmail)
	return e
}

func (e *EmailPage) TogglePasswordVisibility() *EmailPage {
	button, err := e.Frame.Page().Element("#email-account-password-toggle")
	assert.Nil(e.Frame.TestingT(), err)
	button.MustClick()
	return e
}

func (e *EmailPage) AssertFormValues(cfg *u.EmailConfig) *EmailPage {
	err := tools.Eventually(func() error {
		if actual := GetInputValue(e.Frame.TestingT(), e.Frame.Page(), "#email-smtp-host-input"); actual != cfg.SMTPHost {
			return fmt.Errorf("unexpected smtp host: %q", actual)
		}
		if actual := GetInputValue(e.Frame.TestingT(), e.Frame.Page(), "#email-smtp-port-input"); actual != cfg.SMTPPort {
			return fmt.Errorf("unexpected smtp port: %q", actual)
		}
		if actual := GetInputValue(e.Frame.TestingT(), e.Frame.Page(), "#email-from-address-input"); actual != cfg.FromEmailAddress {
			return fmt.Errorf("unexpected from address: %q", actual)
		}
		if actual := GetInputValue(e.Frame.TestingT(), e.Frame.Page(), "#email-account-username-input"); actual != cfg.EmailAccountUsername {
			return fmt.Errorf("unexpected account username: %q", actual)
		}
		if actual := GetInputValue(e.Frame.TestingT(), e.Frame.Page(), "#email-account-password-input"); actual != cfg.EmailAccountPassword {
			return fmt.Errorf("unexpected account password: %q", actual)
		}
		if actual := GetCheckboxValue(e.Frame.TestingT(), e.Frame.Page(), "#email-enabled-checkbox"); actual != cfg.IsEnabled {
			return fmt.Errorf("unexpected enabled checkbox state: %t", actual)
		}
		return nil
	})
	assert.Nil(e.Frame.TestingT(), err)
	return e
}

func (e *EmailPage) AssertTestRecipientValue(expected string) *EmailPage {
	err := tools.Eventually(func() error {
		if actual := GetInputValue(e.Frame.TestingT(), e.Frame.Page(), "#email-test-recipient-input"); actual != expected {
			return fmt.Errorf("unexpected test recipient: %q", actual)
		}
		return nil
	})
	assert.Nil(e.Frame.TestingT(), err)
	return e
}

func (e *EmailPage) AssertPasswordVisibility(visible bool) *EmailPage {
	expectedType := "password"
	if visible {
		expectedType = "text"
	}

	err := tools.Eventually(func() error {
		actual := GetInputType(e.Frame.TestingT(), e.Frame.Page(), "#email-account-password-input")
		if actual != expectedType {
			return fmt.Errorf("unexpected password input type: %q", actual)
		}
		return nil
	})
	assert.Nil(e.Frame.TestingT(), err)
	return e
}

func (e *EmailPage) AssertPasswordValue(expected string) *EmailPage {
	err := tools.Eventually(func() error {
		if actual := GetInputValue(e.Frame.TestingT(), e.Frame.Page(), "#email-account-password-input"); actual != expected {
			return fmt.Errorf("unexpected password value: %q", actual)
		}
		return nil
	})
	assert.Nil(e.Frame.TestingT(), err)
	return e
}

func (e *EmailPage) AssertInvitationTemplateValue(expected string) *EmailPage {
	err := tools.Eventually(func() error {
		if actual := GetInputValue(e.Frame.TestingT(), e.Frame.Page(), "#invitation-email-template-input"); actual != expected {
			return fmt.Errorf("unexpected invitation email template: %q", actual)
		}
		return nil
	})
	assert.Nil(e.Frame.TestingT(), err)
	return e
}

func (e *EmailPage) AssertDefaultInvitationTemplate() *EmailPage {
	return e.AssertInvitationTemplateValue(users.DefaultInvitationEmailTemplate)
}
