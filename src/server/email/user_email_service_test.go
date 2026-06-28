package email

import (
	"server/configs"
	"server/users"
	"testing"

	"github.com/quollix/common/assert"
	u "github.com/quollix/common/utils"
)

func TestValidateInvitationEmailTemplate_HappyPath(t *testing.T) {
	err := ValidateInvitationEmailTemplate(users.DefaultInvitationEmailTemplate)
	assert.Nil(t, err)
}

func TestValidateInvitationEmailTemplate_MissingPlaceholderReturnsError(t *testing.T) {
	err := ValidateInvitationEmailTemplate("Hello {{.Username}} {{.InvitationLink}} {{.ExpirationDate}}")
	assert.Equal(t, "invitation email template is missing required placeholder", u.ExtractError(err))
}

func TestValidateInvitationEmailTemplate_UnsupportedPlaceholderReturnsError(t *testing.T) {
	err := ValidateInvitationEmailTemplate("Hello {{.Username}} {{.InvitationLink}} {{.ExpirationDate}} {{.ServerUrl}} {{.Host}}")
	assert.Equal(t, "invitation email template contains unsupported placeholder", u.ExtractError(err))
}

func TestValidateInvitationEmailTemplate_UnsupportedSyntaxReturnsError(t *testing.T) {
	err := ValidateInvitationEmailTemplate("{{if .Username}}x{{end}} {{.Username}} {{.InvitationLink}} {{.ExpirationDate}} {{.ServerUrl}}")
	assert.Equal(t, "invitation email template contains unsupported template syntax", u.ExtractError(err))
}

func TestSaveInvitationEmailTemplate_ValidatesBeforePersisting(t *testing.T) {
	configRepo := configs.NewConfigsRepositoryMock(t)
	service := &UserEmailServiceImpl{
		ConfigRepo: configRepo,
	}

	err := service.SaveInvitationEmailTemplate("Hello {{.Username}}")
	assert.Equal(t, "invitation email template is missing required placeholder", u.ExtractError(err))
}

func TestRenderInvitationEmailTemplate_RendersAllFields(t *testing.T) {
	configRepo := configs.NewConfigsRepositoryMock(t)
	configRepo.EXPECT().GetConfig(configs.ConfigKeys.InvitationEmailTemplate).Return("User={{.Username}} Link={{.InvitationLink}} ValidUntil={{.ExpirationDate}} Server={{.ServerUrl}}", nil)

	service := &UserEmailServiceImpl{
		ConfigRepo: configRepo,
	}

	rendered, err := service.RenderInvitationEmailTemplate(InvitationTemplateRenderData{
		Username:       "alice",
		InvitationLink: "https://quollix.example.com/set-password?token=abc",
		ExpirationDate: "2026-01-01 10:00:00",
		ServerUrl:      "https://quollix.example.com",
	})
	assert.Nil(t, err)
	assert.Equal(t, "User=alice Link=https://quollix.example.com/set-password?token=abc ValidUntil=2026-01-01 10:00:00 Server=https://quollix.example.com", rendered)
}
