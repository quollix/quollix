package email

import (
	"strings"

	"server/configs"
	"server/users"

	u "github.com/quollix/common/utils"
)

const (
	PartialEmailDeliveryError                = "At least one user did not receive this email. See error logs or Grafana for details."
	BroadcastEmailSubjectCannotBeEmptyError  = "broadcast email subject cannot be empty"
	BroadcastEmailBodyCannotBeEmptyError     = "broadcast email body cannot be empty"
	BroadcastEmailTemplateCannotBeEmptyError = "broadcast email template cannot be empty"
)

type MultiUserEmailSender interface {
	SendEmailToUsersIfEnabled(subject, body string, onlyAdmins bool) error
	SendEmailToUsersWithConfig(emailConfig *u.EmailConfig, subject, body string, onlyAdmins bool) error
}

type MultiUserEmailSenderImpl struct {
	UserRepo     users.UserRepository
	EmailRepo    configs.EmailRepository
	EmailService EmailService
}

func (s *MultiUserEmailSenderImpl) SendEmailToUsersIfEnabled(subject, body string, onlyAdmins bool) error {
	emailConfig, err := s.EmailRepo.ReadEmailConfig()
	if err != nil {
		return err
	}
	if !emailConfig.IsEnabled {
		u.Logger.Info("skipping multi-user email because SMTP is not enabled", "subject", subject)
		return nil
	}
	return s.SendEmailToUsersWithConfig(emailConfig, subject, body, onlyAdmins)
}

func (s *MultiUserEmailSenderImpl) SendEmailToUsersWithConfig(emailConfig *u.EmailConfig, subject, body string, onlyAdmins bool) error {
	allUsers, err := s.UserRepo.ListUsers()
	if err != nil {
		return err
	}

	failedDelivery := false
	for _, user := range allUsers {
		address := strings.TrimSpace(user.Email)
		if !user.IsEnabled || address == "" || strings.HasSuffix(strings.ToLower(address), "@"+users.ReservedEmailDomain) || onlyAdmins && !user.IsAdmin {
			continue
		}
		if err := s.EmailService.SendEmailWithConfig(emailConfig, address, subject, body); err != nil {
			u.Logger.Error(err, "recipient", address, "subject", subject)
			failedDelivery = true
			continue
		}
	}
	if failedDelivery {
		return u.Logger.NewError(PartialEmailDeliveryError)
	}
	return nil
}
