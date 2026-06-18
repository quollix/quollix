package apps_basic

import (
	u "github.com/quollix/common/utils"
	"gopkg.in/yaml.v3"
)

const (
	InvalidComposeServiceConfigError = "service config is not a map"
	MissingContainerNameError        = "missing container_name"
)

type ComposeExtractor interface {
	Extract(data []byte) ([]string, []string, error)
}

type ComposeExtractorImpl struct{}

type DockerComposeYaml struct {
	Services map[string]any `yaml:"services"`
	Volumes  map[string]any `yaml:"volumes"`
}

func (e *ComposeExtractorImpl) Extract(data []byte) ([]string, []string, error) {
	var config DockerComposeYaml
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, nil, err
	}

	volumes := []string{}
	for volumeName := range config.Volumes {
		volumes = append(volumes, volumeName)
	}

	containerNames := []string{}
	for serviceName, serviceConfigAny := range config.Services {
		serviceConfigMap, isMap := serviceConfigAny.(map[string]any)
		if !isMap {
			return nil, nil, u.Logger.NewError(InvalidComposeServiceConfigError, "service_name", serviceName)
		}

		containerName, isString := serviceConfigMap["container_name"].(string)
		if !isString || containerName == "" {
			return nil, nil, u.Logger.NewError(MissingContainerNameError, "service_name", serviceName)
		}

		containerNames = append(containerNames, containerName)
	}

	return volumes, containerNames, nil
}
