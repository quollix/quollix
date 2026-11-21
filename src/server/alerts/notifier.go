package alerts

import (
	"fmt"

	"server/configs"
	"server/email"

	u "github.com/quollix/common/utils"
)

type Notification struct {
	AlertName string
	Firing    bool
	Details   string
}

const (
	alertEmailSubjectTemplate = "Quollix alert %s: %s"
	alertEmailBodyTemplate    = `%s

Quollix instance: https://quollix.%s
`
)

type Notifier interface {
	Send(Notification) error
}

type NotifierImpl struct {
	ConfigsRepo          configs.ConfigsRepository
	MultiUserEmailSender email.MultiUserEmailSender
}

func (n *NotifierImpl) Send(notification Notification) error {
	baseDomain, err := n.ConfigsRepo.GetConfig(configs.ConfigKeys.BaseDomain)
	if err != nil {
		return err
	}
	subject, body := formatNotificationEmail(notification, baseDomain)

	err = n.MultiUserEmailSender.SendEmailToUsersIfEnabled(subject, body, true)
	if u.ExtractError(err) == email.PartialEmailDeliveryError {
		// Treat partial delivery as handled so the alert agent advances state instead of resending to every admin next check.
		return nil
	}
	return err
}

func formatNotificationEmail(notification Notification, baseDomain string) (string, string) {
	status := "recovered"
	if notification.Firing {
		status = "firing"
	}
	subject := fmt.Sprintf(alertEmailSubjectTemplate, status, notification.AlertName)
	body := fmt.Sprintf(alertEmailBodyTemplate, notification.Details, baseDomain)
	return subject, body
}
