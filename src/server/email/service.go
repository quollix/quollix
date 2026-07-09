package email

import (
	"reflect"
	"server/configs"
	"server/tools"

	u "github.com/quollix/common/utils"
)

const (
	SampleTestRecipientEmail = tools.SampleTestRecipientEmail
	SampleTestEmailSubject   = tools.SampleTestEmailSubject
	SampleTestEmailBody      = tools.SampleTestEmailBody
)

type EmailService interface {
	SaveEmailConfig(*u.EmailConfig) error
	SendEmail(to, subject, body string) error
}

type EmailServiceImpl struct {
	EmailClient u.EmailClient
	EmailRepo   configs.EmailRepository
}

func (e *EmailServiceImpl) SendEmail(to, subject, body string) error {
	emailConfig, err := e.EmailRepo.ReadEmailConfig()
	if err != nil {
		return err
	}
	if !emailConfig.IsEnabled {
		return u.Logger.NewError(u.EmailServiceNotEnabledErrorMessage)
	}
	return e.EmailClient.SendEmail(emailConfig, to, subject, body)
}

func (e *EmailServiceImpl) SaveEmailConfig(config *u.EmailConfig) error {
	if config.IsEnabled {
		if err := e.EmailClient.CheckEmailServerConnection(config); err != nil {
			return err
		}
	}
	return e.EmailRepo.SaveEmailConfig(config)
}

func IsSampleEmailConfig(actual *u.EmailConfig) bool {
	return reflect.DeepEqual(actual, configs.GetSampleEmailConfig())
}
