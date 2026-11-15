package email

import (
	"errors"
	"testing"

	"server/configs"
	"server/users"

	"github.com/quollix/common/assert"
	api "github.com/quollix/common/quollix/api"
	u "github.com/quollix/common/utils"
)

type multiUserEmailSenderTestObjects struct {
	Sender       MultiUserEmailSender
	EmailService *EmailServiceMock
	EmailRepo    *configs.EmailRepositoryMock
	UserRepo     *users.UserRepositoryMock
}

func newMultiUserEmailSenderTestObjects(t *testing.T) multiUserEmailSenderTestObjects {
	emailService := NewEmailServiceMock(t)
	emailRepo := configs.NewEmailRepositoryMock(t)
	userRepo := users.NewUserRepositoryMock(t)
	sender := &MultiUserEmailSenderImpl{
		UserRepo:     userRepo,
		EmailRepo:    emailRepo,
		EmailService: emailService,
	}
	return multiUserEmailSenderTestObjects{
		Sender:       sender,
		EmailService: emailService,
		EmailRepo:    emailRepo,
		UserRepo:     userRepo,
	}
}

func assertMultiUserEmailSenderExpectations(t *testing.T, testObjects multiUserEmailSenderTestObjects) {
	testObjects.EmailService.AssertExpectations(t)
	testObjects.EmailRepo.AssertExpectations(t)
	testObjects.UserRepo.AssertExpectations(t)
}

func TestMultiUserEmailSender_SendEmailToUsers_ContinuesAfterDeliveryFailure(t *testing.T) {
	testObjects := newMultiUserEmailSenderTestObjects(t)
	defer assertMultiUserEmailSenderExpectations(t, testObjects)

	testObjects.UserRepo.EXPECT().ListUsers().Return([]api.User{
		{Email: "first@example.com", IsEnabled: true},
		{Email: "second@example.com", IsEnabled: true},
	}, nil)
	emailConfig := configs.GetSampleEmailConfig()
	testObjects.EmailRepo.EXPECT().ReadEmailConfig().Return(emailConfig, nil)
	testObjects.EmailService.EXPECT().SendEmailWithConfig(emailConfig, "first@example.com", "Subject", "Body").Return(errors.New("smtp unavailable"))
	testObjects.EmailService.EXPECT().SendEmailWithConfig(emailConfig, "second@example.com", "Subject", "Body").Return(nil)

	err := testObjects.Sender.SendEmailToUsersIfEnabled("Subject", "Body", false)
	assert.Equal(t, PartialEmailDeliveryError, u.ExtractError(err))
}

func TestMultiUserEmailSender_SendEmailToUsers_OnlySendsToEligibleAdmins(t *testing.T) {
	testObjects := newMultiUserEmailSenderTestObjects(t)
	defer assertMultiUserEmailSenderExpectations(t, testObjects)

	testObjects.UserRepo.EXPECT().ListUsers().Return([]api.User{
		{Email: "admin@example.com", IsAdmin: true, IsEnabled: true},
		{Email: "user@example.com", IsEnabled: true},
		{Email: "disabled@example.com", IsAdmin: true},
		{Email: "reserved@example.invalid", IsAdmin: true, IsEnabled: true},
	}, nil)
	emailConfig := configs.GetSampleEmailConfig()
	testObjects.EmailRepo.EXPECT().ReadEmailConfig().Return(emailConfig, nil)
	testObjects.EmailService.EXPECT().SendEmailWithConfig(emailConfig, "admin@example.com", "Subject", "Body").Return(nil)

	err := testObjects.Sender.SendEmailToUsersIfEnabled("Subject", "Body", true)
	assert.Nil(t, err)
}

func TestMultiUserEmailSender_SendEmailToUsers_WhenSmtpIsDisabled_SkipsSending(t *testing.T) {
	testObjects := newMultiUserEmailSenderTestObjects(t)
	defer assertMultiUserEmailSenderExpectations(t, testObjects)

	testObjects.EmailRepo.EXPECT().ReadEmailConfig().Return(configs.GetEmptyEmailConfig(), nil)

	err := testObjects.Sender.SendEmailToUsersIfEnabled("Subject", "Body", true)
	assert.Nil(t, err)
}
