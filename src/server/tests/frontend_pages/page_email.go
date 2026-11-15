package frontend_pages

import (
	"fmt"
	"server/users"
	"time"

	"github.com/quollix/common/assert"
	"github.com/quollix/common/quollix/api"
	u "github.com/quollix/common/utils"
)

type EmailPage struct {
	Frame *FrameType
}

func (e *EmailPage) FillConfig(cfg *u.EmailConfig) *EmailPage {
	e.Frame.Controls.SetInputValue("#email-smtp-host-input", cfg.SMTPHost)
	e.Frame.Controls.SetInputValue("#email-smtp-port-input", cfg.SMTPPort)
	e.Frame.Controls.SetInputValue("#email-from-address-input", cfg.FromEmailAddress)
	e.Frame.Controls.SetInputValue("#email-account-username-input", cfg.EmailAccountUsername)
	e.Frame.Controls.SetInputValue("#email-account-password-input", cfg.EmailAccountPassword)
	e.Frame.Controls.SetCheckboxValue("#email-enabled-checkbox", cfg.IsEnabled)
	return e
}

func (e *EmailPage) EnterPassword(password string) *EmailPage {
	e.Frame.Controls.SetInputValue("#email-account-password-input", password)
	return e
}

func (e *EmailPage) SetTestRecipient(recipient string) *EmailPage {
	e.Frame.Controls.SetInputValue("#email-test-recipient-input", recipient)
	return e
}

func (e *EmailPage) EnableEmail() *EmailPage {
	e.Frame.Controls.SetCheckboxValue("#email-enabled-checkbox", true)
	return e
}

func (e *EmailPage) Save() *EmailPage {
	saveButton, err := e.Frame.Page.Element("#email-save-config-button")
	assert.Nil(e.Frame.T, err)
	e.Frame.Browser.DoAndWaitDOMContentLoaded(func() {
		saveButton.MustClick()
	})
	e.Frame.Assert.SnackbarVisibleWithTextEventually("Email settings saved.")
	e.Frame.Assert.PagePath(api.Paths.FrontendEmail)
	return e
}

func (e *EmailPage) AssertBroadcastAndInvitationSectionsVisible(visible bool) *EmailPage {
	if visible {
		e.Frame.Assert.ElementPresent("#broadcast-email-form")
		e.Frame.Assert.ElementPresent("#email-template-form")
		return e
	}
	err := u.EventuallyWithTimeout(browserTimeout, 50*time.Millisecond, func() error {
		for _, selector := range []string{"#broadcast-email-form", "#email-template-form"} {
			exists, _, err := e.Frame.Page.Has(selector)
			if err != nil {
				return err
			}
			if exists {
				return fmt.Errorf("element %q is still present", selector)
			}
		}
		return nil
	})
	assert.Nil(e.Frame.T, err)
	return e
}

func (e *EmailPage) TestConnection() *EmailPage {
	button, err := e.Frame.Page.Element("#email-test-connection-button")
	assert.Nil(e.Frame.T, err)
	button.MustClick()
	e.Frame.Assert.SnackbarVisibleWithTextEventually("Email connection test successful.")
	return e
}

func (e *EmailPage) SendTestEmail() *EmailPage {
	button, err := e.Frame.Page.Element("#email-send-test-button")
	assert.Nil(e.Frame.T, err)
	button.MustClick()
	e.Frame.Assert.SnackbarVisibleWithTextEventually("Test email sent.")
	return e
}

func (e *EmailPage) Reset() *EmailPage {
	button, err := e.Frame.Page.Element("#email-reset-config-button")
	assert.Nil(e.Frame.T, err)
	button.MustClick()
	e.Frame.Browser.ConfirmDialog()
	e.Frame.Assert.SnackbarVisibleWithTextEventually("Email settings have been reset.")
	e.Frame.Assert.PagePath(api.Paths.FrontendEmail)
	return e
}

func (e *EmailPage) SetInvitationTemplate(template string) *EmailPage {
	e.Frame.Controls.SetInputValue("#invitation-email-template-input", template)
	return e
}

func (e *EmailPage) SaveInvitationTemplate() *EmailPage {
	button, err := e.Frame.Page.Element("#email-save-template-button")
	assert.Nil(e.Frame.T, err)
	button.MustClick()
	e.Frame.Assert.SnackbarVisibleWithTextEventually("Invitation email template saved.")
	return e
}

func (e *EmailPage) ResetInvitationTemplate() *EmailPage {
	button, err := e.Frame.Page.Element("#email-reset-template-button")
	assert.Nil(e.Frame.T, err)
	button.MustClick()
	e.Frame.Browser.ConfirmDialog()
	e.Frame.Assert.SnackbarVisibleWithTextEventually("Invitation email template has been reset.")
	e.Frame.Assert.PagePath(api.Paths.FrontendEmail)
	return e
}

func (e *EmailPage) SetExposeRealEmailInOidcToken(exposeRealEmail bool) *EmailPage {
	e.Frame.Controls.SetCheckboxValue("#expose-real-email-in-oidc-token-checkbox", exposeRealEmail)
	if exposeRealEmail {
		e.Frame.Browser.ConfirmDialog()
	}
	e.Frame.Assert.SnackbarVisibleWithTextEventually("OIDC email exposure setting saved.")
	return e
}

func (e *EmailPage) TogglePasswordVisibility() *EmailPage {
	button, err := e.Frame.Page.Element("#email-account-password-toggle")
	assert.Nil(e.Frame.T, err)
	button.MustClick()
	return e
}

func (e *EmailPage) AssertExposeRealEmailInOidcToken(expected bool) *EmailPage {
	err := u.Eventually(func() error {
		actual := e.Frame.Controls.GetCheckboxValue("#expose-real-email-in-oidc-token-checkbox")
		if actual != expected {
			return fmt.Errorf("unexpected OIDC email exposure checkbox state: %t", actual)
		}
		return nil
	})
	assert.Nil(e.Frame.T, err)
	return e
}

func (e *EmailPage) AssertFormValues(cfg *u.EmailConfig) *EmailPage {
	err := u.Eventually(func() error {
		if actual := e.Frame.Controls.GetInputValue("#email-smtp-host-input"); actual != cfg.SMTPHost {
			return fmt.Errorf("unexpected smtp host: %q", actual)
		}
		if actual := e.Frame.Controls.GetInputValue("#email-smtp-port-input"); actual != cfg.SMTPPort {
			return fmt.Errorf("unexpected smtp port: %q", actual)
		}
		if actual := e.Frame.Controls.GetInputValue("#email-from-address-input"); actual != cfg.FromEmailAddress {
			return fmt.Errorf("unexpected from address: %q", actual)
		}
		if actual := e.Frame.Controls.GetInputValue("#email-account-username-input"); actual != cfg.EmailAccountUsername {
			return fmt.Errorf("unexpected account username: %q", actual)
		}
		if actual := e.Frame.Controls.GetInputValue("#email-account-password-input"); actual != cfg.EmailAccountPassword {
			return fmt.Errorf("unexpected account password: %q", actual)
		}
		if actual := e.Frame.Controls.GetCheckboxValue("#email-enabled-checkbox"); actual != cfg.IsEnabled {
			return fmt.Errorf("unexpected enabled checkbox state: %t", actual)
		}
		return nil
	})
	assert.Nil(e.Frame.T, err)
	return e
}

func (e *EmailPage) AssertTestRecipientValue(expected string) *EmailPage {
	err := u.Eventually(func() error {
		if actual := e.Frame.Controls.GetInputValue("#email-test-recipient-input"); actual != expected {
			return fmt.Errorf("unexpected test recipient: %q", actual)
		}
		return nil
	})
	assert.Nil(e.Frame.T, err)
	return e
}

func (e *EmailPage) AssertPasswordVisibility(visible bool) *EmailPage {
	expectedType := "password"
	if visible {
		expectedType = "text"
	}

	err := u.Eventually(func() error {
		actual := e.Frame.Controls.GetInputType("#email-account-password-input")
		if actual != expectedType {
			return fmt.Errorf("unexpected password input type: %q", actual)
		}
		return nil
	})
	assert.Nil(e.Frame.T, err)
	return e
}

func (e *EmailPage) AssertPasswordValue(expected string) *EmailPage {
	err := u.Eventually(func() error {
		if actual := e.Frame.Controls.GetInputValue("#email-account-password-input"); actual != expected {
			return fmt.Errorf("unexpected password value: %q", actual)
		}
		return nil
	})
	assert.Nil(e.Frame.T, err)
	return e
}

func (e *EmailPage) AssertInvitationTemplateValue(expected string) *EmailPage {
	err := u.Eventually(func() error {
		if actual := e.Frame.Controls.GetInputValue("#invitation-email-template-input"); actual != expected {
			return fmt.Errorf("unexpected invitation email template: %q", actual)
		}
		return nil
	})
	assert.Nil(e.Frame.T, err)
	return e
}

func (e *EmailPage) AssertDefaultInvitationTemplate() *EmailPage {
	return e.AssertInvitationTemplateValue(users.DefaultInvitationEmailTemplate)
}

func (e *EmailPage) SetBroadcastBody(body string) *EmailPage {
	e.Frame.Controls.SetInputValue("#broadcast-email-body-input", body)
	return e
}

func (e *EmailPage) SaveBroadcastTemplate() *EmailPage {
	button, err := e.Frame.Page.Element("#email-save-broadcast-template-button")
	assert.Nil(e.Frame.T, err)
	button.MustClick()
	e.Frame.Assert.SnackbarVisibleWithTextEventually("Broadcast email template saved.")
	return e
}

func (e *EmailPage) ResetBroadcastTemplate() *EmailPage {
	button, err := e.Frame.Page.Element("#email-reset-broadcast-template-button")
	assert.Nil(e.Frame.T, err)
	button.MustClick()
	e.Frame.Browser.ConfirmDialog()
	e.Frame.Assert.SnackbarVisibleWithTextEventually("Broadcast email template has been reset.")
	e.Frame.Assert.PagePath(api.Paths.FrontendEmail)
	return e
}

func (e *EmailPage) AssertBroadcastTemplateValue(expected string) *EmailPage {
	err := u.Eventually(func() error {
		if actual := e.Frame.Controls.GetInputValue("#broadcast-email-body-input"); actual != expected {
			return fmt.Errorf("unexpected broadcast email body: %q", actual)
		}
		return nil
	})
	assert.Nil(e.Frame.T, err)
	return e
}

func (e *EmailPage) SetBroadcastSubject(subject string) *EmailPage {
	e.Frame.Controls.SetInputValue("#broadcast-email-subject-input", subject)
	return e
}

func (e *EmailPage) SendBroadcastEmail() *EmailPage {
	button, err := e.Frame.Page.Element("#send-broadcast-email-button")
	assert.Nil(e.Frame.T, err)
	e.Frame.Browser.DoAndWaitDOMContentLoaded(func() {
		button.MustClick()
		e.Frame.Browser.ConfirmDialog()
	})
	e.Frame.Assert.SnackbarVisibleWithTextEventually("Email sent to all eligible users.")
	return e
}
