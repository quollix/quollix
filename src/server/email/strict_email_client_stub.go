package email

import (
	"regexp"
	"server/configs"
	"server/tools"
	"strings"

	u "github.com/quollix/common/utils"
)

const (
	invitationEmailSubject    = "User invitation"
	passwordResetEmailSubject = "Password reset"
)

type StrictEmailClientStub struct{}

func (e *StrictEmailClientStub) SendEmail(emailConfig *u.EmailConfig, to, subject, body string) error {
	if err := assertExpectedStrictEmailConfig(emailConfig); err != nil {
		return err
	}
	if subject == tools.SampleTestEmailSubject {
		return e.assertTestEmail(to, subject, body)
	}
	if subject == invitationEmailSubject {
		return e.assertInvitationEmail(to, body)
	}
	if subject == passwordResetEmailSubject {
		return e.assertPasswordResetEmail(to, body)
	}
	return u.Logger.NewError("unexpected email subject", "actual", subject)
}

func (e *StrictEmailClientStub) assertTestEmail(to, subject, body string) error {
	if to != tools.SampleTestRecipientEmail {
		return u.Logger.NewError("unexpected test email recipient", "actual", to, "expected", tools.SampleTestRecipientEmail)
	}
	if subject != tools.SampleTestEmailSubject {
		return u.Logger.NewError("unexpected test email subject", "actual", subject, "expected", tools.SampleTestEmailSubject)
	}
	if body != tools.SampleTestEmailBody {
		return u.Logger.NewError("unexpected test email body", "actual", body, "expected", tools.SampleTestEmailBody)
	}
	return nil
}

func (e *StrictEmailClientStub) assertInvitationEmail(to, body string) error {
	if ok, err := regexp.MatchString(`^[^@\s]+@[^@\s]+\.[^@\s]+$`, to); err != nil {
		return u.Logger.NewError(err.Error())
	} else if !ok {
		return u.Logger.NewError("unexpected invitation email recipient", "actual", to)
	}
	if !strings.Contains(body, "https://quollix.") {
		return u.Logger.NewError("unexpected invitation email body", "actual", body)
	}
	if !strings.Contains(body, "set-password?token=") {
		return u.Logger.NewError("unexpected invitation email body", "actual", body)
	}
	if strings.Contains(body, "{{.") {
		return u.Logger.NewError("unexpected invitation email body", "actual", body)
	}
	return nil
}

func (e *StrictEmailClientStub) assertPasswordResetEmail(to, body string) error {
	if ok, err := regexp.MatchString(`^[^@\s]+@[^@\s]+\.[^@\s]+$`, to); err != nil {
		return u.Logger.NewError(err.Error())
	} else if !ok {
		return u.Logger.NewError("unexpected password reset email recipient", "actual", to)
	}
	if !strings.Contains(body, "https://quollix.") {
		return u.Logger.NewError("unexpected password reset email body", "actual", body)
	}
	if !strings.Contains(body, "set-password?token=") {
		return u.Logger.NewError("unexpected password reset email body", "actual", body)
	}
	if !strings.Contains(body, "A password reset was requested for your account") {
		return u.Logger.NewError("unexpected password reset email body", "actual", body)
	}
	return nil
}

func (e *StrictEmailClientStub) CheckEmailServerConnection(emailConfig *u.EmailConfig) error {
	return assertExpectedStrictEmailConfig(emailConfig)
}

func assertExpectedStrictEmailConfig(actual *u.EmailConfig) error {
	expected := configs.GetSampleEmailConfig()
	if *actual == *expected {
		return nil
	}
	return u.Logger.NewError("unexpected email config", "actual", actual, "expected", expected)
}
