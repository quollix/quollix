package apps_basic

import (
	"strings"
	"time"

	u "github.com/quollix/common/utils"
	"gopkg.in/yaml.v3"
)

type VersionFileNameEncoder interface {
	EncodeComposeArchiveName(dto *ComposeArchiveName) (string, error)
	DecodeComposeArchiveName(fileName string) (*ComposeArchiveName, error)
	DecodeTestAppDefinitionName(fileName string, composeContent []byte, createdAt time.Time) (*ComposeArchiveName, error)
}

type VersionFileNameEncoderImpl struct{}

const VersionFileUploadTimestampLayout = "2006-01-02-15-04-05"
const TestAppDefinitionVersion = "1.0"

type testAppDefinitionComposeFile struct {
	Services map[string]struct {
		ContainerName string `yaml:"container_name"`
	} `yaml:"services"`
}

type ComposeArchiveName struct {
	Maintainer               string
	AppName                  string
	Version                  string
	VersionCreationTimestamp time.Time
}

func (v *VersionFileNameEncoderImpl) EncodeComposeArchiveName(dto *ComposeArchiveName) (string, error) {
	if dto.Maintainer == "" || dto.AppName == "" || dto.Version == "" {
		return "", u.Logger.NewError("maintainer, appName, version must be non-empty")
	}
	if dto.VersionCreationTimestamp.IsZero() {
		return "", u.Logger.NewError("createdAt must be set")
	}

	timestampPart := dto.VersionCreationTimestamp.UTC().Format(VersionFileUploadTimestampLayout)
	return dto.Maintainer + "_" + dto.AppName + "_" + dto.Version + "_" + timestampPart + ".yml", nil
}

func (v *VersionFileNameEncoderImpl) DecodeComposeArchiveName(fileName string) (*ComposeArchiveName, error) {
	if !strings.HasSuffix(fileName, ".yml") {
		return nil, u.Logger.NewError("file must end with .yml")
	}

	stem := strings.TrimSuffix(fileName, ".yml")
	parts := strings.Split(stem, "_")
	if len(parts) != 4 {
		return nil, u.Logger.NewError("expected 4 underscore-separated parts: maintainer_app_version_YYYY-MM-DD-HH-MM-SS.yml")
	}

	createdAt, err := time.ParseInLocation(VersionFileUploadTimestampLayout, parts[3], time.UTC)
	if err != nil {
		return nil, u.Logger.NewError("invalid timestamp, expected YYYY-MM-DD-HH-MM-SS")
	}

	if parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return nil, u.Logger.NewError("maintainer, appName, version must be non-empty")
	}

	return &ComposeArchiveName{
		Maintainer:               parts[0],
		AppName:                  parts[1],
		Version:                  parts[2],
		VersionCreationTimestamp: createdAt.UTC(),
	}, nil
}

func (v *VersionFileNameEncoderImpl) DecodeTestAppDefinitionName(fileName string, composeContent []byte, createdAt time.Time) (*ComposeArchiveName, error) {
	if !strings.HasSuffix(fileName, ".yml") {
		return nil, u.Logger.NewError("file must end with .yml")
	}

	appName := strings.TrimSuffix(fileName, ".yml")
	maintainer, err := inferMaintainerFromMainServiceContainerName(appName, composeContent)
	if err != nil {
		return nil, err
	}

	return &ComposeArchiveName{
		Maintainer:               maintainer,
		AppName:                  appName,
		Version:                  TestAppDefinitionVersion,
		VersionCreationTimestamp: createdAt.UTC(),
	}, nil
}

func inferMaintainerFromMainServiceContainerName(appName string, composeContent []byte) (string, error) {
	var composeFile testAppDefinitionComposeFile
	if err := yaml.Unmarshal(composeContent, &composeFile); err != nil {
		return "", err
	}

	service, ok := composeFile.Services[appName]
	if !ok {
		return "", u.Logger.NewError("main service must be defined")
	}

	containerNameParts := strings.Split(service.ContainerName, "_")
	if len(containerNameParts) != 3 || containerNameParts[1] != appName || containerNameParts[2] != appName {
		return "", u.Logger.NewError("main service has invalid container_name")
	}

	return containerNameParts[0], nil
}
