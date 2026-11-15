package apps_basic

import (
	"fmt"
	"maps"

	"server/tools"

	"github.com/quollix/common/quollix/api"
	u "github.com/quollix/common/utils"
	"github.com/quollix/common/validation"
)

func CompleteAppComposeYaml(app *RepoApp, baseDomain, ianaTimeZone string, settings api.ObservabilitySettings, emailSettings u.EmailConfig) ([]byte, error) {
	if validation.IsWgEasyApp(app.Maintainer, app.AppName) {
		return app.VersionContent, nil
	}

	completedEnvVars := map[string]string{
		tools.ComposeEnvVars.BaseDomain:   baseDomain,
		tools.ComposeEnvVars.IanaTimeZone: ianaTimeZone,
		tools.ComposeEnvVars.ClientId:     app.ClientId,
		tools.ComposeEnvVars.ClientSecret: app.ClientSecret,
	}
	maps.Copy(completedEnvVars, app.Secrets)
	if IsOfficialGrafanaApp(app) {
		completedEnvVars[tools.ComposeEnvVars.LogsRetentionDays] = settings.LogsRetentionDays
		completedEnvVars[tools.ComposeEnvVars.LogsMaxDiskSpaceGB] = settings.LogsMaxDiskSpaceGB
		completedEnvVars[tools.ComposeEnvVars.MetricsRetentionDays] = settings.MetricsRetentionDays
		completedEnvVars[tools.ComposeEnvVars.HostMetricsIntervalSeconds] = settings.HostMetricsIntervalSeconds
		if emailSettings.IsEnabled {
			completedEnvVars[tools.ComposeEnvVars.SmtpEnabled] = "true"
			completedEnvVars[tools.ComposeEnvVars.SmtpHost] = fmt.Sprintf("%s:%s", emailSettings.SMTPHost, emailSettings.SMTPPort)
			completedEnvVars[tools.ComposeEnvVars.SmtpUser] = emailSettings.EmailAccountUsername
			completedEnvVars[tools.ComposeEnvVars.SmtpPassword] = emailSettings.EmailAccountPassword
			completedEnvVars[tools.ComposeEnvVars.SmtpFromAddress] = emailSettings.FromEmailAddress
		}
	}

	return validation.CompleteDockerComposeYaml(app.Maintainer, app.AppName, app.VersionContent, completedEnvVars)
}
