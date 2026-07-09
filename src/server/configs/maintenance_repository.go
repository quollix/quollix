package configs

import (
	"fmt"
	"server/tools"
	"strconv"
	"time"

	u "github.com/quollix/common/utils"
)

type MaintenanceConfig struct {
	IanaTimezone               string
	MaintenanceWindowStartHour int
	NextMaintenanceAt          time.Time
}

type MaintenanceRepository interface {
	GetMaintenanceConfig() (*MaintenanceConfig, error)
	IsMaintenanceConfigSet() (bool, error)
	SetMaintenanceConfig(*MaintenanceConfig) error
}

type MaintenanceRepositoryImpl struct {
	DatabaseConnector tools.DatabaseConnector
}

func (r *MaintenanceRepositoryImpl) GetMaintenanceConfig() (*MaintenanceConfig, error) {
	var ianaTimeZone string
	var maintenanceWindowStartHourString string
	var nextMaintenanceAtUtcString string

	err := r.DatabaseConnector.GetDB().QueryRow(`
SELECT
	(SELECT value FROM configs WHERE key = $1),
	(SELECT value FROM configs WHERE key = $2),
	(SELECT value FROM configs WHERE key = $3);
`,
		ConfigKeys.MaintenanceIanaTimeZone,
		ConfigKeys.MaintenanceWindowStartHour,
		ConfigKeys.MaintenanceNextAtUtc,
	).Scan(
		&ianaTimeZone,
		&maintenanceWindowStartHourString,
		&nextMaintenanceAtUtcString,
	)
	if err != nil {
		return nil, u.Logger.NewError(err.Error())
	}

	maintenanceWindowStartHour, err := strconv.Atoi(maintenanceWindowStartHourString)
	if err != nil {
		return nil, u.Logger.NewError(err.Error())
	}

	nextMaintenanceAtUtc, err := time.Parse(time.RFC3339, nextMaintenanceAtUtcString)
	if err != nil {
		return nil, u.Logger.NewError(err.Error())
	}

	return &MaintenanceConfig{
		MaintenanceWindowStartHour: maintenanceWindowStartHour,
		NextMaintenanceAt:          nextMaintenanceAtUtc.UTC(),
		IanaTimezone:               ianaTimeZone,
	}, nil
}

func (r *MaintenanceRepositoryImpl) IsMaintenanceConfigSet() (bool, error) {
	var count int
	err := r.DatabaseConnector.GetDB().QueryRow(`
SELECT COUNT(*) FROM configs
WHERE key IN ($1, $2, $3);
`,
		ConfigKeys.MaintenanceIanaTimeZone,
		ConfigKeys.MaintenanceWindowStartHour,
		ConfigKeys.MaintenanceNextAtUtc,
	).Scan(&count)
	if err != nil {
		return false, u.Logger.NewError(err.Error())
	}
	return count == 3, nil
}

func (r *MaintenanceRepositoryImpl) SetMaintenanceConfig(config *MaintenanceConfig) error {
	_, err := r.DatabaseConnector.GetDB().Exec(`
INSERT INTO configs (key, value)
VALUES
	($1, $2),
	($3, $4),
	($5, $6)
ON CONFLICT (key) DO UPDATE
SET value = EXCLUDED.value;
`,
		ConfigKeys.MaintenanceIanaTimeZone, config.IanaTimezone,
		ConfigKeys.MaintenanceWindowStartHour, fmt.Sprintf("%d", config.MaintenanceWindowStartHour),
		ConfigKeys.MaintenanceNextAtUtc, config.NextMaintenanceAt.UTC().Format(time.RFC3339),
	)
	if err != nil {
		return u.Logger.NewError(err.Error())
	}
	return nil
}
