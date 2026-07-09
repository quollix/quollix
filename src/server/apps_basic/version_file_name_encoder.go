package apps_basic

import (
	"path/filepath"
	"strings"
	"time"

	u "github.com/quollix/common/utils"
)

type VersionFileNameEncoder interface {
	EncodeComposeArchiveName(dto *ComposeArchiveName) (string, error)
	DecodeComposeArchiveName(fileName string) (*ComposeArchiveName, error)
}

type VersionFileNameEncoderImpl struct{}

const VersionFileUploadTimestampLayout = "2006-01-02-15-04-05"

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
	baseName := filepath.Base(fileName)
	if !strings.HasSuffix(baseName, ".yml") {
		return nil, u.Logger.NewError("file must end with .yml")
	}

	stem := strings.TrimSuffix(baseName, ".yml")
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
