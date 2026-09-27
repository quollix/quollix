package apps_basic

import (
	"testing"

	"github.com/quollix/common/assert"
	u "github.com/quollix/common/utils"
	"github.com/quollix/common/validation"
)

func TestCompleteAppComposeYaml_WgEasyReturnsDefinitionUnchanged(t *testing.T) {
	content := []byte("wg-easy compose content")
	app := &RepoApp{Maintainer: u.OfficialMaintainer, AppName: "wgeasy", VersionContent: content}

	completedContent, envVars, err := CompleteAppComposeYaml(app, "example.com", "Etc/UTC")

	assert.Nil(t, err)
	assert.Equal(t, content, completedContent)
	assert.Equal(t, map[string]string{}, envVars)
	assert.True(t, validation.IsWgEasyApp(app.Maintainer, app.AppName))
}
