package email

import (
	"net/http"
	"server/configs"
	"server/tools"
	"server/users"
	"strconv"

	u "github.com/quollix/common/utils"
	"github.com/quollix/common/validation"
)

var expectedUserInvitationEmailErrors = u.MapOf(users.UserAlreadyExistsError, users.EmailAlreadyExistsError, u.EmailServiceNotEnabledErrorMessage)
var expectedPasswordResetEmailErrors = u.MapOf(users.AdminCanNotResetOwnPasswordError, u.EmailServiceNotEnabledErrorMessage)

type EmailHandler struct {
	EmailService     EmailService
	EmailRepository  configs.EmailRepository
	ConfigsRepo      configs.ConfigsRepository
	EmailClient      u.EmailClient
	UserEmailService UserEmailService
}

func (e *EmailHandler) SaveEmailConfig(w http.ResponseWriter, r *http.Request) {
	settings, ok := validation.ReadBody[u.EmailConfig](w, r)
	if !ok {
		return
	}
	if err := e.EmailService.SaveEmailConfig(settings); err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}
}

func (e *EmailHandler) ReadEmailConfig(w http.ResponseWriter, r *http.Request) {
	settings, err := e.EmailRepository.ReadEmailConfig()
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}
	u.SendJsonResponse(w, settings)
}

func (e *EmailHandler) TestEmailServerConnection(w http.ResponseWriter, r *http.Request) {
	config, ok := validation.ReadBody[u.EmailConfig](w, r)
	if !ok {
		return
	}
	config.IsEnabled = true
	if err := e.EmailClient.CheckEmailServerConnection(config); err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}
}

type TestEmailRequest struct {
	EmailConfig u.EmailConfig `json:"email_config"`
	ToEmail     string        `json:"to_email" validate:"email"`
}

func (e *EmailHandler) SendTestEmail(w http.ResponseWriter, r *http.Request) {
	request, ok := validation.ReadBody[TestEmailRequest](w, r)
	if !ok {
		return
	}
	if err := e.EmailClient.SendEmail(&request.EmailConfig, request.ToEmail, SampleTestEmailSubject, SampleTestEmailBody); err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}
}

func (e *EmailHandler) ResetEmailConfig(w http.ResponseWriter, r *http.Request) {
	if err := e.EmailService.SaveEmailConfig(configs.GetEmptyEmailConfig()); err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}
}

type InvitationEmailTemplateRequest struct {
	Template string `json:"template" validate:"ignore"`
}

type InviteUserRequest struct {
	Username string `json:"username" validate:"username"`
	Email    string `json:"email" validate:"email"`
}

func (e *EmailHandler) ReadInvitationTemplate(w http.ResponseWriter, r *http.Request) {
	template, err := e.ConfigsRepo.GetConfig(configs.ConfigKeys.InvitationEmailTemplate)
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}
	u.SendJsonResponse(w, InvitationEmailTemplateRequest{Template: template})
}

func (e *EmailHandler) SaveInvitationTemplate(w http.ResponseWriter, r *http.Request) {
	request, ok := validation.ReadBody[InvitationEmailTemplateRequest](w, r)
	if !ok {
		return
	}
	if err := e.UserEmailService.SaveInvitationEmailTemplate(request.Template); err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}
}

func (e *EmailHandler) ResetInvitationTemplate(w http.ResponseWriter, r *http.Request) {
	if err := e.UserEmailService.ResetInvitationEmailTemplate(); err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}
}

func (e *EmailHandler) InviteUserViaEmailHandler(w http.ResponseWriter, r *http.Request) {
	userInvitationRequest, ok := validation.ReadBody[InviteUserRequest](w, r)
	if !ok {
		return
	}
	if err := e.UserEmailService.InviteUserViaEmail(userInvitationRequest.Username, userInvitationRequest.Email); err != nil {
		u.WriteResponseError(w, expectedUserInvitationEmailErrors, err)
		return
	}
}

func (e *EmailHandler) SendPasswordResetEmailHandler(w http.ResponseWriter, r *http.Request) {
	userIdString, ok := validation.ReadBody[tools.NumberString](w, r)
	if !ok {
		return
	}
	userIdToResetPassword, err := strconv.Atoi(userIdString.Value)
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}
	currentlyLoggedInUser, err := users.GetAuthFromContext(r)
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}

	err = e.UserEmailService.SendPasswordResetEmail(userIdToResetPassword, currentlyLoggedInUser.Id)
	if err != nil {
		u.WriteResponseError(w, expectedPasswordResetEmailErrors, err)
		return
	}
}
