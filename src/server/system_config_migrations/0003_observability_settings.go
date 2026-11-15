package system_config_migrations

import (
	"strconv"

	"server/configs"
	"server/tools"

	u "github.com/quollix/common/utils"
)

func (s *SystemConfigMigrationsProviderImpl) createObservabilitySettings() error {
	u.Logger.Info(systemConfigMigrationLogMessage, "step", 3, "action", "create observability settings")
	for key, value := range map[string]string{
		configs.ConfigKeys.LogsRetentionDays:               strconv.Itoa(tools.DefaultLogsRetentionDays),
		configs.ConfigKeys.LogsMaxDiskSpaceGB:              strconv.Itoa(tools.DefaultLogsMaxDiskSpaceGB),
		configs.ConfigKeys.MetricsRetentionDays:            strconv.Itoa(tools.DefaultMetricsRetentionDays),
		configs.ConfigKeys.HostMetricsIntervalSeconds:      strconv.Itoa(tools.DefaultHostMetricsIntervalSeconds),
		configs.ConfigKeys.ContainerMetricsIntervalSeconds: strconv.Itoa(tools.DefaultContainerMetricsIntervalSeconds),
		configs.ConfigKeys.VolumeMetricsIntervalSeconds:    strconv.Itoa(tools.DefaultVolumeMetricsIntervalSeconds),
	} {
		exists, err := s.ConfigsRepo.IsConfigSet(key)
		if err != nil {
			return err
		}
		if exists {
			continue
		}
		if err := s.ConfigsRepo.SetConfig(key, value); err != nil {
			return err
		}
	}
	return nil
}
