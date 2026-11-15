//go:build integration

package repository

import (
	"testing"
	"time"

	"server/alerts"

	"github.com/quollix/common/assert"
)

func TestAlertRuntimeStateRepository_ReadWrite(t *testing.T) {
	InitDeps()
	defer ConfigRepo.Wipe()

	expected := alerts.DefaultRuntimeState()
	expected.CPUUsage = alerts.AlertRuntimeState{
		Status:       alerts.AlertStatusPending,
		PendingSince: time.Date(2026, time.November, 7, 12, 30, 0, 0, time.UTC),
	}
	expected.MemoryUsage.Status = alerts.AlertStatusFiring
	expected.CertificateExpiry.Status = alerts.AlertStatusFiring
	assert.Nil(t, AlertStateRepo.Save(expected))

	actual, err := AlertStateRepo.Read()
	assert.Nil(t, err)
	assert.Equal(t, expected, actual)
}
