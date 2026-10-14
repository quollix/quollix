package apps_basic

import (
	"os"
	"path/filepath"

	"server/configs"
	"server/tools"

	"github.com/quollix/common/quollix/api"
	u "github.com/quollix/common/utils"
)

type AppRuntimeSpec struct {
	App *RepoApp

	BaseDomain   string
	IanaTimeZone string

	ClientId     string
	ClientSecret string

	EmailSettings         *u.EmailConfig
	ObservabilitySettings *api.ObservabilitySettings
}

// Why does this exist? The problem is, when doing app operations on postgres app like backup or restore, we need to stop the postgres container temporarily. During this time, we can not read any data from the database, so we must assert to have all required runtime information before conducting the operation.
type DatabaseIndependentRuntime interface {
	StartApp(spec *AppRuntimeSpec) error
	StopApp(spec *AppRuntimeSpec) error
	CollectAppSpec(appId int) (*AppRuntimeSpec, error)
}

type ContainerLogMonitor interface {
	StartMonitoring()
	StopMonitoring()
}

type ContainerMetricsMonitor interface {
	RunCycle() error
	StartMonitoring()
	StopMonitoring()
}

type VolumeMetricsMonitor interface {
	RunCycle() error
	StartMonitoring()
	StopMonitoring()
}

type DatabaseIndependentRuntimeImpl struct {
	ContainerMetricsMonitor ContainerMetricsMonitor
	VolumeMetricsMonitor    VolumeMetricsMonitor
	ContainerLogMonitor     ContainerLogMonitor
	DockerService           tools.DockerService
	AppDetector             AppDetector
	AppRepo                 AppRepository
	ConfigsService          configs.ConfigsService
	MaintenanceRepo         configs.MaintenanceRepository
	EmailRepository         configs.EmailRepository
	ObservabilityRepository configs.ObservabilityRepository
}

func (r *DatabaseIndependentRuntimeImpl) StartApp(spec *AppRuntimeSpec) error {
	r.DockerService.CreateDockerNetwork(spec.App.Maintainer, spec.App.AppName)
	r.DockerService.AttachBrandAppToNetwork(spec.App.Maintainer, spec.App.AppName, spec.BaseDomain)

	composeDir, err := os.MkdirTemp("", "app-compose-")
	if err != nil {
		return u.Logger.NewError(err.Error())
	}
	defer u.RemoveDir(composeDir)

	composeYamlPath := filepath.Join(composeDir, "docker-compose.yml")
	completedComposeContent, err := CompleteAppComposeYaml(spec.App, spec.BaseDomain, spec.IanaTimeZone, spec.ObservabilitySettings)
	if err != nil {
		return err
	}
	err = os.WriteFile(composeYamlPath, completedComposeContent, 0o600)
	if err != nil {
		return u.Logger.NewError(err.Error())
	}

	if err := r.DockerService.StartAppContainer(spec.App.Maintainer, spec.App.AppName, composeYamlPath); err != nil {
		return err
	}
	if isOfficialGrafanaApp(spec.App) {
		r.ContainerLogMonitor.StartMonitoring()
		r.ContainerMetricsMonitor.StartMonitoring()
		r.VolumeMetricsMonitor.StartMonitoring()
	}
	return nil
}

func (r *DatabaseIndependentRuntimeImpl) StopApp(spec *AppRuntimeSpec) error {
	app := spec.App
	composeDir, err := os.MkdirTemp("", "app-compose-")
	if err != nil {
		return u.Logger.NewError(err.Error())
	}
	defer u.RemoveDir(composeDir)

	composeYamlPath := filepath.Join(composeDir, "docker-compose.yml")

	completedContent, err := CompleteAppComposeYaml(app, spec.BaseDomain, spec.IanaTimeZone, spec.ObservabilitySettings)
	if err != nil {
		return err
	}
	err = os.WriteFile(composeYamlPath, completedContent, 0o600)
	if err != nil {
		return u.Logger.NewError(err.Error())
	}

	if isOfficialGrafanaApp(app) {
		r.ContainerMetricsMonitor.StopMonitoring()
		r.VolumeMetricsMonitor.StopMonitoring()
		r.ContainerLogMonitor.StopMonitoring()
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

	spec.BaseDomain, err = a.ConfigsService.GetBaseDomain()
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
	spec.ObservabilitySettings, err = a.ObservabilityRepository.Read()
	if err != nil {
		return nil, err
	}
	return spec, nil
}

func isOfficialGrafanaApp(app *RepoApp) bool {
	return app.Maintainer == u.OfficialMaintainer && app.AppName == tools.GrafanaAppName
}
