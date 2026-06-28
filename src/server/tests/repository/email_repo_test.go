//go:build integration

package repository

import (
	"server/configs"
	"testing"

	"github.com/quollix/common/assert"
	u "github.com/quollix/common/utils"
)

func TestEmailRepository_SaveReadAndOverwrite(t *testing.T) {
	InitDeps()
	defer ConfigRepo.Wipe()

	isPresent, err := EmailRepo.IsEmailConfigPresent()
	assert.Nil(t, err)
	assert.False(t, isPresent)

	firstConfig := &u.EmailConfig{
		SMTPHost:             "smtp-1.example.com",
		SMTPPort:             "587",
		FromEmailAddress:     "noreply-1@example.com",
		EmailAccountUsername: "user-1",
		EmailAccountPassword: "password-1",
		IsEnabled:            true,
	}
	assert.Nil(t, EmailRepo.SaveEmailConfig(firstConfig))

	isPresent, err = EmailRepo.IsEmailConfigPresent()
	assert.Nil(t, err)
	assert.True(t, isPresent)

	actualConfig, err := EmailRepo.ReadEmailConfig()
	assert.Nil(t, err)
	assert.Equal(t, firstConfig, actualConfig)

	emptyConfig := configs.GetEmptyEmailConfig()
	assert.Nil(t, EmailRepo.SaveEmailConfig(emptyConfig))

	actualConfig, err = EmailRepo.ReadEmailConfig()
	assert.Nil(t, err)
	assert.Equal(t, emptyConfig, actualConfig)

	isPresent, err = EmailRepo.IsEmailConfigPresent()
	assert.Nil(t, err)
	assert.True(t, isPresent)
}
