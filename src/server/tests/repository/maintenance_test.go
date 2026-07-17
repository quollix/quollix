//go:build integration

package repository

import (
	"testing"
	"time"

	"github.com/quollix/common/assert"
	"github.com/quollix/common/quollix/api"
)

func TestMaintenanceRepository(t *testing.T) {
	InitDeps()
	defer ConfigRepo.Wipe()

	isSet, err := MaintenanceRepo.IsMaintenanceConfigSet()
	assert.Nil(t, err)
	assert.False(t, isSet)
	_, err = MaintenanceRepo.GetMaintenanceConfig()
	assert.NotNil(t, err)

	maintenanceConfig := &api.MaintenanceConfig{
		MaintenanceWindowStartHour: 6,
		NextMaintenanceAt:          time.Date(2024, time.June, 30, 6, 0, 0, 0, time.UTC),
		IanaTimezone:               "Europe/Berlin",
	}
	err = MaintenanceRepo.SetMaintenanceConfig(maintenanceConfig)
	assert.Nil(t, err)

	isSet, err = MaintenanceRepo.IsMaintenanceConfigSet()
	assert.Nil(t, err)
	assert.True(t, isSet)

	actualConfig, err := MaintenanceRepo.GetMaintenanceConfig()
	assert.Nil(t, err)
	assert.Equal(t, maintenanceConfig, actualConfig)
}
