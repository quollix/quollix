package configs

import (
	"strconv"

	"server/tools"

	"github.com/quollix/common/quollix/api"
	u "github.com/quollix/common/utils"
)

type ObservabilityRepository interface {
	Read() (*api.ObservabilitySettings, error)
	Save(*api.ObservabilitySettings) error
}

type ObservabilityRepositoryImpl struct {
	DatabaseConnector tools.DatabaseConnector
}

func (r *ObservabilityRepositoryImpl) Read() (*api.ObservabilitySettings, error) {
	settings := &api.ObservabilitySettings{}
	err := r.DatabaseConnector.GetDB().QueryRow(`
SELECT
 (SELECT value FROM configs WHERE key = $1),
 (SELECT value FROM configs WHERE key = $2),
 (SELECT value FROM configs WHERE key = $3),
 (SELECT value FROM configs WHERE key = $4),
 (SELECT value FROM configs WHERE key = $5),
 (SELECT value FROM configs WHERE key = $6);
`, ConfigKeys.LogsRetentionDays, ConfigKeys.LogsMaxDiskSpaceGB, ConfigKeys.MetricsRetentionDays, ConfigKeys.HostMetricsIntervalSeconds, ConfigKeys.ContainerMetricsIntervalSeconds, ConfigKeys.VolumeMetricsIntervalSeconds).
		Scan(&settings.LogsRetentionDays, &settings.LogsMaxDiskSpaceGB, &settings.MetricsRetentionDays, &settings.HostMetricsIntervalSeconds, &settings.ContainerMetricsIntervalSeconds, &settings.VolumeMetricsIntervalSeconds)
	if err != nil {
		return nil, u.Logger.NewError(err.Error())
	}
	return settings, nil
}

func (r *ObservabilityRepositoryImpl) Save(settings *api.ObservabilitySettings) error {
	_, err := r.DatabaseConnector.GetDB().Exec(`
INSERT INTO configs (key, value) VALUES ($1, $2), ($3, $4), ($5, $6), ($7, $8), ($9, $10), ($11, $12)
ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value;
`, ConfigKeys.LogsRetentionDays, settings.LogsRetentionDays,
		ConfigKeys.LogsMaxDiskSpaceGB, settings.LogsMaxDiskSpaceGB,
		ConfigKeys.MetricsRetentionDays, settings.MetricsRetentionDays,
		ConfigKeys.HostMetricsIntervalSeconds, settings.HostMetricsIntervalSeconds,
		ConfigKeys.ContainerMetricsIntervalSeconds, settings.ContainerMetricsIntervalSeconds,
		ConfigKeys.VolumeMetricsIntervalSeconds, settings.VolumeMetricsIntervalSeconds)
	if err != nil {
		return u.Logger.NewError(err.Error())
	}
	return nil
}

func NewObservabilitySettings(logsRetentionDays, logsMaxDiskSpaceGB, metricsRetentionDays, hostMetricsIntervalSeconds, containerMetricsIntervalSeconds, volumeMetricsIntervalSeconds int) *api.ObservabilitySettings {
	return &api.ObservabilitySettings{
		LogsRetentionDays:               strconv.Itoa(logsRetentionDays),
		LogsMaxDiskSpaceGB:              strconv.Itoa(logsMaxDiskSpaceGB),
		MetricsRetentionDays:            strconv.Itoa(metricsRetentionDays),
		HostMetricsIntervalSeconds:      strconv.Itoa(hostMetricsIntervalSeconds),
		ContainerMetricsIntervalSeconds: strconv.Itoa(containerMetricsIntervalSeconds),
		VolumeMetricsIntervalSeconds:    strconv.Itoa(volumeMetricsIntervalSeconds),
	}
}
