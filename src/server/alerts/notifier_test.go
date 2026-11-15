package alerts

import (
	"errors"
	"testing"

	"server/configs"
	"server/email"

	"github.com/quollix/common/assert"
)

type notifierTestObjects struct {
	Notifier    *NotifierImpl
	ConfigsRepo *configs.ConfigsRepositoryMock
	EmailSender *email.MultiUserEmailSenderMock
}

func newNotifierTestObjects(t *testing.T) notifierTestObjects {
	configsRepo := configs.NewConfigsRepositoryMock(t)
	emailSender := email.NewMultiUserEmailSenderMock(t)
	notifier := &NotifierImpl{
		ConfigsRepo:          configsRepo,
		MultiUserEmailSender: emailSender,
	}

	return notifierTestObjects{
		Notifier:    notifier,
		ConfigsRepo: configsRepo,
		EmailSender: emailSender,
	}
}

func TestNotifier_HandlesPartialEmailDeliveryError(t *testing.T) {
	testObjects := newNotifierTestObjects(t)
	notification := Notification{AlertName: "test alert", Firing: true, Details: "details"}
	subject, body := formatNotificationEmail(notification, "example.com")

	testObjects.ConfigsRepo.EXPECT().GetConfig(configs.ConfigKeys.BaseDomain).Return("example.com", nil)
	testObjects.EmailSender.EXPECT().SendEmailToUsersIfEnabled(subject, body, true).Return(errors.New(email.PartialEmailDeliveryError))

	err := testObjects.Notifier.Send(notification)
	assert.Nil(t, err)
}

func TestNotifier_SendsRecoveryNotification(t *testing.T) {
	testObjects := newNotifierTestObjects(t)
	notification := Notification{AlertName: "disk usage", Firing: false, Details: "Disk usage returned to normal."}

	subject, body := formatNotificationEmail(notification, "example.com")
	testObjects.ConfigsRepo.EXPECT().GetConfig(configs.ConfigKeys.BaseDomain).Return("example.com", nil)
	testObjects.EmailSender.EXPECT().SendEmailToUsersIfEnabled(subject, body, true).Return(nil)

	err := testObjects.Notifier.Send(notification)
	assert.Nil(t, err)
}
