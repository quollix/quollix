package apps_basic

import (
	"os"
	"path/filepath"
	"server/configs"
	"server/tools"

	u "github.com/quollix/common/utils"
	"github.com/quollix/common/validation"
)

type AppRuntimeSpec struct {
	App *RepoApp

	ServerHost   string
	IanaTimeZone string

	ClientId     string
	ClientSecret string

	EmailSettings *u.EmailConfig
}

// Why does this exist? The problem is, when doing app operations on postgres app like backup or restore, we need to stop the postgres container temporarily. During this time, we can not read any data from the database, so we must assert to have all required runtime information before conducting the operation.
type DatabaseIndependentRuntime interface {
	StartApp(spec *AppRuntimeSpec) error
	StopApp(app *RepoApp) error
	CollectAppSpec(appId int) (*AppRuntimeSpec, error)
}

type DatabaseIndependentRuntimeImpl struct {
	DockerService   tools.DockerService
	AppDetector     AppDetector
	AppRepo         AppRepository
	ConfigsRepo     configs.ConfigsRepository
	MaintenanceRepo configs.MaintenanceRepository
	EmailRepository configs.EmailRepository
}

func (r *DatabaseIndependentRuntimeImpl) StartApp(spec *AppRuntimeSpec) error {
	r.DockerService.CreateDockerNetwork(spec.App.Maintainer, spec.App.AppName)
	r.DockerService.AttachBrandAppToNetwork(spec.App.Maintainer, spec.App.AppName, spec.ServerHost)

	composeDir, err := os.MkdirTemp("", "app-compose-")
	if err != nil {
		return u.Logger.NewError(err.Error())
	}
	defer u.RemoveDir(composeDir)

	envVars := map[string]string{
		tools.ComposeEnvVars.ServerHost:   spec.ServerHost,
		tools.ComposeEnvVars.ClientId:     spec.App.ClientId,
		tools.ComposeEnvVars.ClientSecret: spec.App.ClientSecret,
		tools.ComposeEnvVars.IanaTimeZone: spec.IanaTimeZone,
	}

	composeYamlPath := filepath.Join(composeDir, "docker-compose.yml")
	completedComposeContent, err := validation.CompleteDockerComposeYaml(spec.App.Maintainer, spec.App.AppName, spec.App.VersionContent, envVars)
	if err != nil {
		return err
	}
	err = os.WriteFile(composeYamlPath, completedComposeContent, 0600)
	if err != nil {
		return u.Logger.NewError(err.Error())
	}

	return r.DockerService.StartAppContainer(spec.App.Maintainer, spec.App.AppName, composeYamlPath, envVars)
}

func (r *DatabaseIndependentRuntimeImpl) StopApp(app *RepoApp) error {
	composeDir, err := os.MkdirTemp("", "app-compose-")
	if err != nil {
		return u.Logger.NewError(err.Error())
	}
	defer u.RemoveDir(composeDir)

	composeYamlPath := filepath.Join(composeDir, "docker-compose.yml")

	err = os.WriteFile(composeYamlPath, app.VersionContent, 0600)
	if err != nil {
		return u.Logger.NewError(err.Error())
	}

	r.DockerService.StopAppContainer(app.Maintainer, app.AppName, composeYamlPath)
	r.DockerService.DetachBrandAppFromNetwork(app.Maintainer, app.AppName)
	r.DockerService.RemoveNetwork(app.Maintainer, app.AppName)
	return nil
}

func (a *DatabaseIndependentRuntimeImpl) CollectAppSpec(appId int) (*AppRuntimeSpec, error) {
	spec := &AppRuntimeSpec{}

	var err error
	spec.App, err = a.AppRepo.GetAppById(appId)
	if err != nil {
		return nil, err
	}

	spec.ServerHost, err = a.ConfigsRepo.GetConfig(configs.ConfigKeys.ServerHost)
	if err != nil {
		return nil, err
	}

	maintenanceConfig, err := a.MaintenanceRepo.GetMaintenanceConfig()
	if err != nil {
		return nil, err
	}
	spec.IanaTimeZone = maintenanceConfig.IanaTimezone

	spec.EmailSettings, err = a.EmailRepository.ReadEmailConfig()
	if err != nil {
		return nil, err
	}
	return spec, nil
}
