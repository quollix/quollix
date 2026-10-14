package apps_basic

import (
	"maps"

	"server/tools"

	"github.com/quollix/common/quollix/api"
	"github.com/quollix/common/validation"
)

func CompleteAppComposeYaml(app *RepoApp, baseDomain, ianaTimeZone string, settings *api.ObservabilitySettings) ([]byte, error) {
	if validation.IsWgEasyApp(app.Maintainer, app.AppName) {
		return app.VersionContent, nil
	}

	completedEnvVars := map[string]string{
		tools.ComposeEnvVars.BaseDomain:                 baseDomain,
		tools.ComposeEnvVars.IanaTimeZone:               ianaTimeZone,
		tools.ComposeEnvVars.LogsRetentionDays:          settings.LogsRetentionDays,
		tools.ComposeEnvVars.LogsMaxDiskSpaceGB:         settings.LogsMaxDiskSpaceGB,
		tools.ComposeEnvVars.MetricsRetentionDays:       settings.MetricsRetentionDays,
		tools.ComposeEnvVars.HostMetricsIntervalSeconds: settings.HostMetricsIntervalSeconds,
		tools.ComposeEnvVars.ClientId:                   app.ClientId,
		tools.ComposeEnvVars.ClientSecret:               app.ClientSecret,
	}
	maps.Copy(completedEnvVars, app.Secrets)

	return validation.CompleteDockerComposeYaml(app.Maintainer, app.AppName, app.VersionContent, completedEnvVars)
}
