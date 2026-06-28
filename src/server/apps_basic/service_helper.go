package apps_basic

import (
	"fmt"
	"server/tools"
	"strconv"
	"time"

	u "github.com/quollix/common/utils"
	"gopkg.in/yaml.v3"
)

type AppServiceHelper interface {
	ConvertToAppDtos(apps []RepoApp) []AppDto
	IsAppVisibleToUser(userId int, role tools.UserAccessLevel, app RepoApp) bool
	GetPortFromComposeYaml(composeContent []byte, appName string) (string, error)
}

type AppServiceHelperImpl struct {
	Authorizer  Authorizer
	AppDetector AppDetector
}

type ComposeFile struct {
	Services map[string]struct {
		Labels map[string]any `yaml:"labels"`
	} `yaml:"services"`
}

func (a *AppServiceHelperImpl) IsAppVisibleToUser(userId int, role tools.UserAccessLevel, app RepoApp) bool {
	if role == tools.AdminLevel {
		return true
	} else if !app.ShouldBeRunning {
		return false
	} else if a.AppDetector.IsSystemApp(app.AppName) {
		return false
	} else {
		err := a.Authorizer.Authorize(app.AccessPolicy, role, userId, app.AppName)
		return err == nil
	}
}

func (a *AppServiceHelperImpl) ConvertToAppDtos(apps []RepoApp) []AppDto {
	var appDtos []AppDto
	for _, app := range apps {
		appDto := AppDto{
			AppId:                    strconv.Itoa(app.AppId),
			Maintainer:               app.Maintainer,
			AppName:                  app.AppName,
			VersionName:              app.VersionName,
			AccessPolicy:             app.AccessPolicy,
			Port:                     app.Port,
			IsRunning:                app.ShouldBeRunning,
			IsOfficialDatabaseApp:    a.AppDetector.IsOfficialDatabaseApp(app.AppName),
			VersionCreationTimestamp: app.VersionCreationTimestamp,
			VersionContent:           app.VersionContent,
			ClientId:                 app.ClientId,
			ClientSecret:             app.ClientSecret,
			AutomaticBackupsEnabled:  app.AutomaticBackupsEnabled,
			AutomaticUpdatesEnabled:  app.AutomaticUpdatesEnabled,
			IsOfficial:               a.AppDetector.IsOfficialApp(app.Maintainer),
		}
		if a.AppDetector.IsOfficialApp(app.Maintainer) {
			appDto.DocsUrl = tools.InstalledAppDocsUrl(app.AppName)
		} else {
			appDto.DocsUrl = ""
		}
		appDtos = append(appDtos, appDto)
	}
	return appDtos
}

func GetSampleApp() *RepoApp {
	app := NewRepoApp(
		"quollix",
		tools.SampleApp,
		"v1.0.0",
		tools.Policies.PublicAccessPolicy,
		"80",
		"abcdef1234567890",
		"abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
		time.Date(2020, time.January, 1, 10, 0, 0, 0, time.UTC),
		[]byte("initial-content"),
		true,
		true,
		true,
	)
	return app
}

func (a *AppServiceHelperImpl) GetPortFromComposeYaml(composeContent []byte, appName string) (string, error) {
	var composeFile ComposeFile
	if err := yaml.Unmarshal(composeContent, &composeFile); err != nil {
		return "", err
	}
	service, ok := composeFile.Services[appName]
	if !ok {
		return "", u.Logger.NewError("service not found in docker-compose.yml")
	}
	portValue, ok := service.Labels["quollix.port"]
	if !ok {
		return "", u.Logger.NewError("could not find quollix.port label in docker-compose.yml")
	}
	port := fmt.Sprintf("%v", portValue)
	return port, nil
}
