package api_client

import (
	"encoding/json"
	emailpkg "server/email"
	"server/tools"
	"strconv"

	u "github.com/quollix/common/utils"
)

type EmailClient struct {
	quollix *QuollixClient
}

type invitationEmailTemplateRequest struct {
	Template string `json:"template"`
}

func (c *EmailClient) InviteViaEmail(username, email string) error {
	request := emailpkg.InviteUserRequest{Username: username, Email: email}
	_, err := c.quollix.Parent.DoRequest(tools.Paths.BackendUsersInviteUserViaEmail, request)
	return err
}

func (c *EmailClient) ResetPasswordViaEmail(userId int) error {
	_, err := c.quollix.Parent.DoRequest(tools.Paths.BackendUsersResetPasswordViaEmail, tools.NumberString{Value: strconv.Itoa(userId)})
	return err
}

func (c *EmailClient) SaveConfig(cfg *u.EmailConfig) error {
	_, err := c.quollix.Parent.DoRequest(tools.Paths.BackendEmailSaveConfig, cfg)
	return err
}

func (c *EmailClient) ReadConfig() (*u.EmailConfig, error) {
	body, err := c.quollix.Parent.DoRequest(tools.Paths.BackendEmailReadConfig, nil)
	if err != nil {
		return nil, err
	}
	var emailConfig u.EmailConfig
	err = json.Unmarshal(body, &emailConfig)
	if err != nil {
		return nil, err
	}
	return &emailConfig, nil
}

func (c *EmailClient) TestConnection(cfg *u.EmailConfig) error {
	_, err := c.quollix.Parent.DoRequest(tools.Paths.BackendEmailTestConnection, cfg)
	return err
}

func (c *EmailClient) SendTestEmail(cfg *u.EmailConfig, toEmail string) error {
	request := emailpkg.TestEmailRequest{
		EmailConfig: *cfg,
		ToEmail:     toEmail,
	}
	_, err := c.quollix.Parent.DoRequest(tools.Paths.BackendEmailSendTestEmail, request)
	return err
}

func (c *EmailClient) ResetConfig() error {
	_, err := c.quollix.Parent.DoRequest(tools.Paths.BackendEmailResetConfig, nil)
	return err
}

func (c *EmailClient) SaveExposeRealEmailInOidcToken(exposeRealEmail bool) error {
	_, err := c.quollix.Parent.DoRequest(tools.Paths.BackendEmailSaveOidcEmailExposure, tools.SingleBool{Value: exposeRealEmail})
	return err
}

func (c *EmailClient) ReadExposeRealEmailInOidcToken() (bool, error) {
	body, err := c.quollix.Parent.DoRequest(tools.Paths.BackendEmailReadOidcEmailExposure, nil)
	if err != nil {
		return false, err
	}

	var response tools.SingleBool
	err = json.Unmarshal(body, &response)
	if err != nil {
		return false, err
	}
	return response.Value, nil
}

func (c *EmailClient) SaveInvitationTemplate(template string) error {
	_, err := c.quollix.Parent.DoRequest(tools.Paths.BackendEmailSaveInvitationTemplate, invitationEmailTemplateRequest{
		Template: template,
	})
	return err
}

func (c *EmailClient) ReadInvitationTemplate() (string, error) {
	body, err := c.quollix.Parent.DoRequest(tools.Paths.BackendEmailReadInvitationTemplate, nil)
	if err != nil {
		return "", err
	}
	var response invitationEmailTemplateRequest
	err = json.Unmarshal(body, &response)
	if err != nil {
		return "", err
	}
	return response.Template, nil
}

func (c *EmailClient) ResetInvitationTemplate() error {
	_, err := c.quollix.Parent.DoRequest(tools.Paths.BackendEmailResetInvitationTemplate, nil)
	return err
}
