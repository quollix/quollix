package observability

import (
	"strconv"

	"server/apps_basic"
	"server/configs"
	"server/tools"

	"github.com/quollix/common/quollix/api"
	u "github.com/quollix/common/utils"
)

const InvalidSettingsError = "Retention must be between 1 and 36500 days, log storage must be between 1 and 1000000 GB, and collection intervals must be between 1 and 604800 seconds"

type SettingsService interface {
	Save(*api.ObservabilitySettings) error
}

type SettingsServiceImpl struct {
	Repository    configs.ObservabilityRepository
	AppRepository apps_basic.AppRepository
	AppService    apps_basic.AppService
}

func (s *SettingsServiceImpl) Save(settings *api.ObservabilitySettings) error {
	for _, field := range []struct {
		value   string
		maximum uint64
	}{
		{settings.LogsRetentionDays, 36500},
		{settings.MetricsRetentionDays, 36500},
		{settings.LogsMaxDiskSpaceGB, 1000000},
		{settings.HostMetricsIntervalSeconds, 604800},
		{settings.ContainerMetricsIntervalSeconds, 604800},
		{settings.VolumeMetricsIntervalSeconds, 604800},
	} {
		value, err := strconv.ParseUint(field.value, 10, 64)
		if err != nil || value < 1 || value > field.maximum {
			return u.Logger.NewError(InvalidSettingsError)
		}
	}
	if err := s.Repository.Save(settings); err != nil {
		return err
	}
	apps, err := s.AppRepository.ListApps()
	if err != nil {
		return err
	}
	for _, app := range apps {
		if app.Maintainer == u.OfficialMaintainer && app.AppName == tools.GrafanaAppName && app.ShouldBeRunning {
			return s.AppService.StartApp(app.AppId)
		}
	}
	return nil
}
