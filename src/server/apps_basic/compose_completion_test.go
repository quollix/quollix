package apps_basic

import (
	"testing"

	"server/configs"

	"github.com/quollix/common/assert"
	u "github.com/quollix/common/utils"
	"github.com/quollix/common/validation"
	"gopkg.in/yaml.v3"
)

func TestCompleteAppComposeYaml_WgEasyReturnsDefinitionUnchanged(t *testing.T) {
	content := []byte("wg-easy compose content")
	app := &RepoApp{Maintainer: u.OfficialMaintainer, AppName: "wgeasy", VersionContent: content}

	completedContent, err := CompleteAppComposeYaml(app, "example.com", "Etc/UTC", nil)

	assert.Nil(t, err)
	assert.Equal(t, content, completedContent)
	assert.True(t, validation.IsWgEasyApp(app.Maintainer, app.AppName))
}

func TestCompleteAppComposeYaml_InjectsObservabilitySettings(t *testing.T) {
	app := &RepoApp{Maintainer: "maintainer", AppName: "app", VersionContent: []byte(`
services:
  app:
    command:
      - -logsRetention=${LOGS_RETENTION_DAYS}d
      - -logsStorage=${LOGS_MAX_DISK_SPACE_GB}GB
      - -metricsRetention=${METRICS_RETENTION_DAYS}d
      - -hostMetricsInterval=${HOST_METRICS_INTERVAL_SECONDS}s
`)}
	settings := configs.NewObservabilitySettings(14, 10, 60, 420, 660, 5400)
	completed, err := CompleteAppComposeYaml(app, "example.com", "Etc/UTC", settings)
	assert.Nil(t, err)
	var definition struct {
		Services map[string]struct {
			Command []string `yaml:"command"`
		} `yaml:"services"`
	}
	assert.Nil(t, yaml.Unmarshal(completed, &definition))
	assert.Equal(t, []string{"-logsRetention=14d", "-logsStorage=10GB", "-metricsRetention=60d", "-hostMetricsInterval=420s"}, definition.Services["app"].Command)
}
