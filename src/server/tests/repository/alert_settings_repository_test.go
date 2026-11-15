//go:build integration

package repository

import (
	"testing"

	"server/alerts"

	"github.com/quollix/common/assert"
)

func TestAlertSettingsRepository_ReadWrite(t *testing.T) {
	InitDeps()
	defer ConfigRepo.Wipe()

	expected := alerts.DefaultAlertSettings()
	assert.Nil(t, AlertSettingsRepo.Write(expected))

	actual, err := AlertSettingsRepo.Read()
	assert.Nil(t, err)
	assert.Equal(t, expected, actual)

	expected.CPUUsageEnabled = false
	expected.MemoryUsageEnabled = false
	expected.DiskUsageEnabled = false
	expected.CertificateExpiryEnabled = false
	expected.QuollixErrorsEnabled = false
	expected.HostUsageTriggerPercent = 95
	expected.HostUsageRecoveryPercent = 75
	expected.HostUsagePendingMinutes = 0
	assert.Nil(t, AlertSettingsRepo.Write(expected))
	actual, err = AlertSettingsRepo.Read()
	assert.Nil(t, err)
	assert.Equal(t, expected, actual)
}
