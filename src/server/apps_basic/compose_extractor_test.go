package apps_basic

import (
	"testing"

	"github.com/quollix/common/assert"
	u "github.com/quollix/common/utils"
)

func TestComposeExtractorImpl_Extract_UsesContainerNames(t *testing.T) {
	composeYaml := `
services:
  sampleapp:
    image: nginx:latest
    container_name: quollix_sampleapp_sampleapp
    volumes:
      - quollix_sampleapp_data:/data
volumes:
  quollix_sampleapp_data:
`

	extractor := &ComposeExtractorImpl{}
	volumes, containerNames, err := extractor.Extract([]byte(composeYaml))
	assert.Nil(t, err)

	assert.Equal(t, []string{"quollix_sampleapp_data"}, volumes)
	assert.Equal(t, []string{"quollix_sampleapp_sampleapp"}, containerNames)
}

func TestComposeExtractorImpl_Extract_NoVolumesReturnsEmptyVolumes(t *testing.T) {
	composeYaml := `
services:
  sampleapp:
    image: nginx:latest
    container_name: quollix_sampleapp_sampleapp
`

	extractor := &ComposeExtractorImpl{}
	volumes, containerNames, err := extractor.Extract([]byte(composeYaml))
	assert.Nil(t, err)

	assert.Equal(t, []string{}, volumes)
	assert.Equal(t, []string{"quollix_sampleapp_sampleapp"}, containerNames)
}

func TestComposeExtractorImpl_Extract_InvalidServiceConfigReturnsError(t *testing.T) {
	composeYaml := `
services:
  invalid-service-config: service-string
`

	extractor := &ComposeExtractorImpl{}
	volumes, containerNames, err := extractor.Extract([]byte(composeYaml))

	assert.Equal(t, InvalidComposeServiceConfigError, u.ExtractError(err))
	assert.Nil(t, volumes)
	assert.Nil(t, containerNames)
}

func TestComposeExtractorImpl_Extract_MissingContainerNameReturnsError(t *testing.T) {
	composeYaml := `
services:
  sampleapp:
    image: nginx:latest
`

	extractor := &ComposeExtractorImpl{}
	volumes, containerNames, err := extractor.Extract([]byte(composeYaml))

	assert.Equal(t, MissingContainerNameError, u.ExtractError(err))
	assert.Nil(t, volumes)
	assert.Nil(t, containerNames)
}
