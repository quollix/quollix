package apps_basic

import (
	"maps"

	"server/tools"

	"github.com/quollix/common/validation"
)

func CompleteAppComposeYaml(app *RepoApp, baseDomain, ianaTimeZone string) ([]byte, map[string]string, error) {
	completedEnvVars := map[string]string{
		tools.ComposeEnvVars.BaseDomain:       baseDomain,
		tools.ComposeEnvVars.LegacyServerHost: baseDomain,
		tools.ComposeEnvVars.IanaTimeZone:     ianaTimeZone,
	}
	completedEnvVars[tools.ComposeEnvVars.ClientId] = app.ClientId
	completedEnvVars[tools.ComposeEnvVars.ClientSecret] = app.ClientSecret
	completedEnvVars[tools.ComposeEnvVars.AppSecret] = app.AppSecret
	maps.Copy(completedEnvVars, app.Secrets)

	completedComposeContent, err := validation.CompleteDockerComposeYaml(app.Maintainer, app.AppName, app.VersionContent, completedEnvVars)
	return completedComposeContent, completedEnvVars, err
}
