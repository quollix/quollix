package email

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	texttemplate "text/template"

	"server/configs"
	"server/tools"
	"server/users"

	u "github.com/quollix/common/utils"
)

var invitationEmailPlaceholderPattern = regexp.MustCompile(`{{\s*\.([A-Za-z0-9_]+)\s*}}`)

var invitationEmailPlaceholders = map[string]string{
	"Username":       "non-empty username",
	"InvitationLink": "invitation link",
	"ExpirationDate": "expiration date",
	"ServerUrl":      "server URL",
}

type InvitationTemplateRenderData struct {
	Username       string
	InvitationLink string
	ExpirationDate string
	ServerUrl      string
}

type UserEmailService interface {
	InviteUserViaEmail(username, targetEmail string) error
	SendPasswordResetEmail(userIdToResetPassword, currentlyLoggedInUserId int) error
	ResetInvitationEmailTemplate() error
	SaveInvitationEmailTemplate(template string) error
	RenderInvitationEmailTemplate(data InvitationTemplateRenderData) (string, error)
}

type UserEmailServiceImpl struct {
	ConfigRepo     configs.ConfigsRepository
	ConfigsService configs.ConfigsService
	EmailRepo      configs.EmailRepository
	UserService    users.UserService
	UserRepo       users.UserRepository
	EmailService   EmailService
}

func (e *UserEmailServiceImpl) verifyEmailConfigEnabled() error {
	emailConfig, err := e.EmailRepo.ReadEmailConfig()
	if err != nil {
		return err
	}
	if !emailConfig.IsEnabled {
		return u.Logger.NewError(u.EmailServiceNotEnabledErrorMessage)
	}
	return nil
}

func (e *UserEmailServiceImpl) buildServerURL() (string, error) {
	baseDomain, err := e.ConfigsService.GetBaseDomain()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("https://quollix.%s", baseDomain), nil
}

func (e *UserEmailServiceImpl) buildSetPasswordLink(token string) (string, error) {
	serverURL, err := e.buildServerURL()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s%s?token=%s", serverURL, tools.Paths.FrontendSetPassword, token), nil
}

func (e *UserEmailServiceImpl) InviteUserViaEmail(username, targetEmail string) error {
	if err := e.verifyEmailConfigEnabled(); err != nil {
		return err
	}

	invitationDetails, err := e.UserService.InviteUser(username, targetEmail)
	if err != nil {
		return err
	}
	serverURL, err := e.buildServerURL()
	if err != nil {
		return err
	}
	invitationLink, err := e.buildSetPasswordLink(invitationDetails.Token)
	if err != nil {
		return err
	}
	body, err := e.RenderInvitationEmailTemplate(InvitationTemplateRenderData{
		Username:       username,
		InvitationLink: invitationLink,
		ExpirationDate: invitationDetails.ExpirationDate.Format(tools.PrettyFrontendTimeLayout),
		ServerUrl:      serverURL,
	})
	if err != nil {
		return err
	}
	return e.EmailService.SendEmail(targetEmail, users.InvitationEmailSubject, body)
}

func (e *UserEmailServiceImpl) SendPasswordResetEmail(userIdToResetPassword, currentlyLoggedInUserId int) error {
	if err := e.verifyEmailConfigEnabled(); err != nil {
		return err
	}

	token, err := e.UserService.ResetPasswordOfUser(userIdToResetPassword, currentlyLoggedInUserId)
	if err != nil {
		return err
	}

	user, err := e.UserRepo.GetUserById(userIdToResetPassword)
	if err != nil {
		return err
	}
	serverURL, err := e.buildServerURL()
	if err != nil {
		return err
	}
	resetLink, err := e.buildSetPasswordLink(token)
	if err != nil {
		return err
	}
	body := fmt.Sprintf(
		users.DefaultPasswordResetEmailBody,
		serverURL,
		user.Username,
		resetLink,
		user.SetPasswordTokenExpirationDate.Format(tools.PrettyFrontendTimeLayout),
	)
	return e.EmailService.SendEmail(user.Email, users.PasswordResetEmailSubject, body)
}

func (e *UserEmailServiceImpl) ResetInvitationEmailTemplate() error {
	return e.ConfigRepo.SetConfig(configs.ConfigKeys.InvitationEmailTemplate, users.DefaultInvitationEmailTemplate)
}

func (e *UserEmailServiceImpl) SaveInvitationEmailTemplate(template string) error {
	if err := ValidateInvitationEmailTemplate(template); err != nil {
		return err
	}
	return e.ConfigRepo.SetConfig(configs.ConfigKeys.InvitationEmailTemplate, template)
}

func (e *UserEmailServiceImpl) RenderInvitationEmailTemplate(data InvitationTemplateRenderData) (string, error) {
	templateText, err := e.ConfigRepo.GetConfig(configs.ConfigKeys.InvitationEmailTemplate)
	if err != nil {
		return "", err
	}
	tmpl, err := texttemplate.New("invitation_email").Option("missingkey=error").Parse(templateText)
	if err != nil {
		return "", u.Logger.NewError(err.Error())
	}
	var rendered bytes.Buffer
	if err := tmpl.Execute(&rendered, data); err != nil {
		return "", u.Logger.NewError(err.Error())
	}
	return rendered.String(), nil
}

func ValidateInvitationEmailTemplate(template string) error {
	placeholderNames := invitationEmailPlaceholderPattern.FindAllStringSubmatch(template, -1)
	found := map[string]bool{}
	for _, match := range placeholderNames {
		found[match[1]] = true
	}

	for placeholder := range invitationEmailPlaceholders {
		if !found[placeholder] {
			return u.Logger.NewError("invitation email template is missing required placeholder", "placeholder", placeholder)
		}
	}

	for placeholder := range found {
		if _, ok := invitationEmailPlaceholders[placeholder]; !ok {
			return u.Logger.NewError("invitation email template contains unsupported placeholder", "placeholder", placeholder)
		}
	}

	cleaned := invitationEmailPlaceholderPattern.ReplaceAllString(template, "")
	if strings.Contains(cleaned, "{{") || strings.Contains(cleaned, "}}") {
		return u.Logger.NewError("invitation email template contains unsupported template syntax")
	}
	return nil
}
