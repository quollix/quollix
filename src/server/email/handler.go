package email

import (
	"net/http"
	"server/configs"
	"server/users"
	"strconv"

	api "github.com/quollix/common/quollix/api"
	u "github.com/quollix/common/utils"
	"github.com/quollix/common/validation"
)

var expectedUserInvitationEmailErrors = u.MapOf(users.UserAlreadyExistsError, users.EmailAlreadyExistsError, users.ReservedEmailMustMatchUserError, u.EmailServiceNotEnabledErrorMessage)
var expectedPasswordResetEmailErrors = u.MapOf(users.AdminCanNotResetOwnPasswordError, u.EmailServiceNotEnabledErrorMessage)

type EmailHandler struct {
	EmailService     EmailService
	EmailRepository  configs.EmailRepository
	ConfigsRepo      configs.ConfigsRepository
	OidcEmailService configs.OidcEmailExposureService
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

func (e *EmailHandler) SendTestEmail(w http.ResponseWriter, r *http.Request) {
	request, ok := validation.ReadBody[api.TestEmailRequest](w, r)
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

func (e *EmailHandler) ReadOidcEmailExposureConfig(w http.ResponseWriter, r *http.Request) {
	exposeRealEmail, err := e.OidcEmailService.ReadExposeRealEmailInOidcToken()
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}
	u.SendJsonResponse(w, api.SingleBool{Value: exposeRealEmail})
}

func (e *EmailHandler) SaveOidcEmailExposureConfig(w http.ResponseWriter, r *http.Request) {
	request, ok := validation.ReadBody[api.SingleBool](w, r)
	if !ok {
		return
	}
	if err := e.OidcEmailService.SaveExposeRealEmailInOidcToken(request.Value); err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}
}

func (e *EmailHandler) ReadInvitationTemplate(w http.ResponseWriter, r *http.Request) {
	template, err := e.ConfigsRepo.GetConfig(configs.ConfigKeys.InvitationEmailTemplate)
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}
	u.SendJsonResponse(w, api.InvitationEmailTemplateRequest{Template: template})
}

func (e *EmailHandler) SaveInvitationTemplate(w http.ResponseWriter, r *http.Request) {
	request, ok := validation.ReadBody[api.InvitationEmailTemplateRequest](w, r)
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
	userInvitationRequest, ok := validation.ReadBody[api.InviteUserRequest](w, r)
	if !ok {
		return
	}
	if err := e.UserEmailService.InviteUserViaEmail(userInvitationRequest.Username, userInvitationRequest.Email); err != nil {
		u.WriteResponseError(w, expectedUserInvitationEmailErrors, err)
		return
	}
}

func (e *EmailHandler) SendPasswordResetEmailHandler(w http.ResponseWriter, r *http.Request) {
	userIdString, ok := validation.ReadBody[api.NumberString](w, r)
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
