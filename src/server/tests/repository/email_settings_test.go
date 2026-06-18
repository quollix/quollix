//go:build integration

package repository

import (
	"server/configs"
	"testing"

	"github.com/quollix/common/assert"
)

func TestEmailSettingsOperations(t *testing.T) {
	InitDeps()
	ConfigRepo.Wipe()

	_, err := EmailRepo.ReadEmailConfig()
	assert.NotNil(t, err)

	err = EmailRepo.SaveEmailConfig(configs.GetSampleEmailConfig())
	assert.Nil(t, err)

	settings, err := EmailRepo.ReadEmailConfig()
	assert.Nil(t, err)
	assert.Equal(t, *configs.GetSampleEmailConfig(), *settings)

	updated := configs.GetSampleEmailConfig()
	updated.SMTPHost = "smtp.updated.example.com"
	updated.IsEnabled = false

	err = EmailRepo.SaveEmailConfig(updated)
	assert.Nil(t, err)

	settings, err = EmailRepo.ReadEmailConfig()
	assert.Nil(t, err)
	assert.Equal(t, *updated, *settings)
}
