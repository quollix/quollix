package tools

import (
	"path/filepath"

	u "github.com/quollix/common/utils"
)

type DirectoryProvider interface {
	InitializeDirectories() error
	GetCacheDir() string
	GetDockerDir() string
	GetMigrationsDir() string
	GetSampleAppDir() string
	GetResticImageDockerfilePath() string
	GetOfficialDatabaseDockerDir() string
	GetTemplatesPath() string
	GetPublishedAppsDir() string
}

type DirectoryProviderImpl struct {
	projectDir string
	Config     *GlobalConfig
}

func NewDirectoryProvider(config *GlobalConfig) DirectoryProvider {
	return &DirectoryProviderImpl{
		Config: config,
	}
}

func (d *DirectoryProviderImpl) GetCacheDir() string {
	return filepath.Join(d.projectDir, "cache")
}

func (d *DirectoryProviderImpl) InitializeDirectories() error {
	projectDir, err := u.FindDir("server")
	if err != nil {
		return err
	}
	d.projectDir = projectDir
	return nil
}

func (d *DirectoryProviderImpl) getAssetsDir() string {
	return filepath.Join(d.projectDir, "assets")
}

func (d *DirectoryProviderImpl) GetDockerDir() string {
	return d.getAssetsDir() + "/docker"
}

func (d *DirectoryProviderImpl) GetMigrationsDir() string {
	return d.getAssetsDir() + "/migrations"
}

func (d *DirectoryProviderImpl) GetSampleAppDir() string {
	return d.GetDockerDir() + "/sampleapp"
}

func (d *DirectoryProviderImpl) GetResticImageDockerfilePath() string {
	return d.GetDockerDir() + "/Dockerfile.restic"
}

func (d *DirectoryProviderImpl) GetOfficialDatabaseDockerDir() string {
	databaseDir := d.GetDockerDir() + "/postgres"
	if d.Config.DeployOfficialDatabaseWithPortExposed {
		return databaseDir + "/test"
	} else {
		return databaseDir + "/prod"
	}
}

func (d *DirectoryProviderImpl) GetTemplatesPath() string {
	return d.getAssetsDir() + "/templates"
}

func (d *DirectoryProviderImpl) GetPublishedAppsDir() string {
	return d.GetDockerDir() + "/published-apps"
}
