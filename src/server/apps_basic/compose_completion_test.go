package apps_basic

import (
	"testing"

	"server/configs"
	"server/tools"

	"github.com/quollix/common/assert"
	"github.com/quollix/common/quollix/api"
	u "github.com/quollix/common/utils"
	"github.com/quollix/common/validation"
	"gopkg.in/yaml.v3"
)

func TestCompleteAppComposeYaml_WgEasyReturnsDefinitionUnchanged(t *testing.T) {
	content := []byte("wg-easy compose content")
	app := &RepoApp{Maintainer: u.OfficialMaintainer, AppName: "wgeasy", VersionContent: content}

	completedContent, err := CompleteAppComposeYaml(app, "example.com", "Etc/UTC", api.ObservabilitySettings{}, u.EmailConfig{})

	assert.Nil(t, err)
	assert.Equal(t, content, completedContent)
	assert.True(t, validation.IsWgEasyApp(app.Maintainer, app.AppName))
}

func TestCompleteAppComposeYaml_InjectsSettingsForOfficialGrafana(t *testing.T) {
	app := &RepoApp{Maintainer: u.OfficialMaintainer, AppName: tools.GrafanaAppName, VersionContent: []byte(`
services:
  grafana:
    command:
      - -logsRetention=${LOGS_RETENTION_DAYS}d
      - -logsStorage=${LOGS_MAX_DISK_SPACE_GB}GB
      - -metricsRetention=${METRICS_RETENTION_DAYS}d
      - -hostMetricsInterval=${HOST_METRICS_INTERVAL_SECONDS}s
    environment:
      GF_SMTP_ENABLED: "${SMTP_ENABLED:-false}"
      GF_SMTP_HOST: "${SMTP_HOST:-localhost:25}"
      GF_SMTP_USER: "${SMTP_USER:-}"
      GF_SMTP_PASSWORD: "${SMTP_PASSWORD:-}"
      GF_SMTP_FROM_ADDRESS: "${SMTP_FROM_ADDRESS:-admin@grafana.localhost}"
`)}
	emailSettings := configs.GetSampleEmailConfig()
	observabilitySettings := configs.NewObservabilitySettings(14, 10, 60, 420, 660, 5400)

	completed, err := CompleteAppComposeYaml(app, "example.com", "Etc/UTC", *observabilitySettings, *emailSettings)
	assert.Nil(t, err)
	var definition struct {
		Services map[string]struct {
			Command     []string          `yaml:"command"`
			Environment map[string]string `yaml:"environment"`
		} `yaml:"services"`
	}
	assert.Nil(t, yaml.Unmarshal(completed, &definition))
	assert.Equal(t, []string{"-logsRetention=14d", "-logsStorage=10GB", "-metricsRetention=60d", "-hostMetricsInterval=420s"}, definition.Services["grafana"].Command)
	assert.Equal(t, map[string]string{
		"GF_SMTP_ENABLED":      "true",
		"GF_SMTP_HOST":         emailSettings.SMTPHost + ":" + emailSettings.SMTPPort,
		"GF_SMTP_USER":         emailSettings.EmailAccountUsername,
		"GF_SMTP_PASSWORD":     emailSettings.EmailAccountPassword,
		"GF_SMTP_FROM_ADDRESS": emailSettings.FromEmailAddress,
	}, definition.Services["grafana"].Environment)

	completed, err = CompleteAppComposeYaml(app, "example.com", "Etc/UTC", *observabilitySettings, u.EmailConfig{})
	assert.Nil(t, err)
	assert.Nil(t, yaml.Unmarshal(completed, &definition))
	assert.Equal(t, map[string]string{
		"GF_SMTP_ENABLED":      "${SMTP_ENABLED:-false}",
		"GF_SMTP_HOST":         "${SMTP_HOST:-localhost:25}",
		"GF_SMTP_USER":         "${SMTP_USER:-}",
		"GF_SMTP_PASSWORD":     "${SMTP_PASSWORD:-}",
		"GF_SMTP_FROM_ADDRESS": "${SMTP_FROM_ADDRESS:-admin@grafana.localhost}",
	}, definition.Services["grafana"].Environment)
}
