package apps_basic

import (
	"server/tools"

	u "github.com/quollix/common/utils"
)

const (
	RunningAppState    = "Running"
	NotRunningAppState = "Not running"
)

type AppService interface {
	StartApp(appId int) error
	StopApp(appId int) error
	StartAppsThatShouldBeRunning()
	DeleteAppAndArtifacts(appId int) error

	SetAppShouldBeRunning(appId int, shouldBeRunning bool) error
	SetAccessPolicy(appId int, policy string) error
	UpsertAppInDatabase(app *RepoApp) error
	ListAppsForRole(userId int, role tools.UserAccessLevel) ([]AppDto, error)
	ListAppsForAdmin() ([]AppDto, error)
	UpdateAppAutoMaintenanceSettings(appId int, autoUpdateEnabled, autoBackupEnabled bool) error
	RegenerateOidcClientCredentials(appId int) error
}

type AppServiceImpl struct {
	AppRepo                    AppRepository
	DockerService              tools.DockerService
	AppServiceHelper           AppServiceHelper
	AppDetector                AppDetector
	ComposeExtractor           ComposeExtractorImpl
	ClientCredentialsGenerator ClientCredentialsGenerator
	DatabaseIndependentRuntime DatabaseIndependentRuntime
	VersionFileNameEncoder     VersionFileNameEncoder
}

func (a *AppServiceImpl) RegenerateOidcClientCredentials(appId int) error {
	app, err := a.AppRepo.GetAppById(appId)
	if err != nil {
		return err
	}

	newClientId, newClientSecret, err := a.ClientCredentialsGenerator.Generate()
	if err != nil {
		return err
	}

	app.ClientId = newClientId
	app.ClientSecret = newClientSecret

	return a.AppRepo.UpdateApp(app)
}

func (a *AppServiceImpl) SetAccessPolicy(appId int, policy string) error {
	app, err := a.AppRepo.GetAppById(appId)
	if err != nil {
		return err
	}
	app.AccessPolicy = policy
	return a.AppRepo.UpdateApp(app)
}

func (a *AppServiceImpl) StartApp(appId int) error {
	spec, err := a.DatabaseIndependentRuntime.CollectAppSpec(appId)
	if err != nil {
		return err
	}

	if err := a.DatabaseIndependentRuntime.StartApp(spec); err != nil {
		return err
	}

	return a.SetAppShouldBeRunning(appId, true)
}

func (a *AppServiceImpl) StopApp(appId int) error {
	err := a.SetAppShouldBeRunning(appId, false)
	if err != nil {
		return err
	}
	app, err := a.AppRepo.GetAppById(appId)
	if err != nil {
		return err
	}

	return a.DatabaseIndependentRuntime.StopApp(app)
}

func (a *AppServiceImpl) SetAppShouldBeRunning(appId int, shouldBeRunning bool) error {
	app, err := a.AppRepo.GetAppById(appId)
	if err != nil {
		return err
	}
	app.ShouldBeRunning = shouldBeRunning
	err = a.AppRepo.UpdateApp(app)
	if err != nil {
		return err
	}
	return nil
}

func (a *AppServiceImpl) StartAppsThatShouldBeRunning() {
	apps, err := a.AppRepo.ListApps()
	if err != nil {
		u.Logger.Error(err)
	}

	var idsOfRunningApps []int
	for _, app := range apps {
		// database is started by a separate function
		if a.AppDetector.IsOfficialDatabaseApp(app.AppName) {
			continue
		}
		if app.ShouldBeRunning {
			idsOfRunningApps = append(idsOfRunningApps, app.AppId)
		}
	}

	for _, appId := range idsOfRunningApps {
		err = a.StartApp(appId)
		if err != nil {
			u.Logger.Error(err, tools.AppIdField, appId)
		}
	}
}

func (a *AppServiceImpl) DeleteAppAndArtifacts(appId int) error {
	app, err := a.AppRepo.GetAppById(appId)
	if err != nil {
		return err
	}

	if a.AppDetector.IsOfficialDatabaseApp(app.AppName) {
		return u.Logger.NewError(OperationNotAllowedOnOfficialDatabaseAppError)
	}

	err = a.StopApp(appId)
	if err != nil {
		return err
	}

	volumes, _, err := a.ComposeExtractor.Extract(app.VersionContent)
	if err != nil {
		return err
	}
	a.DockerService.RemoveVolumes(volumes)

	err = a.AppRepo.DeleteApp(appId)
	if err != nil {
		return err
	}

	return nil
}

func (a *AppServiceImpl) UpsertAppInDatabase(app *RepoApp) error {
	doesAppExist, err := a.AppRepo.DoesAppExist(app.AppName)
	if err != nil {
		return err
	}

	if doesAppExist {
		repoApp, err := a.AppRepo.GetAppByName(app.AppName)
		if err != nil {
			return err
		}
		app.AppId = repoApp.AppId
		err = a.AppRepo.UpdateApp(app)
		if err != nil {
			return err
		}
	} else {
		_, err = a.AppRepo.CreateApp(app)
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *AppServiceImpl) ListAppsForRole(userId int, role tools.UserAccessLevel) ([]AppDto, error) {
	repoApps, err := a.AppRepo.ListApps()
	if err != nil {
		return nil, err
	}
	var filteredApps []RepoApp
	for _, app := range repoApps {
		if isVisible := a.AppServiceHelper.IsAppVisibleToUser(userId, role, app); isVisible {
			filteredApps = append(filteredApps, app)
		}
	}
	appDtos := a.AppServiceHelper.ConvertToAppDtos(filteredApps)
	return appDtos, nil
}

func (a *AppServiceImpl) ListAppsForAdmin() ([]AppDto, error) {
	repoApps, err := a.AppRepo.ListApps()
	if err != nil {
		return nil, err
	}
	appDtos := a.AppServiceHelper.ConvertToAppDtos(repoApps)
	return appDtos, nil
}

func (a *AppServiceImpl) UpdateAppAutoMaintenanceSettings(appId int, autoUpdateEnabled, autoBackupEnabled bool) error {
	app, err := a.AppRepo.GetAppById(appId)
	if err != nil {
		return err
	}

	if a.AppDetector.IsOfficialDatabaseApp(app.AppName) && autoUpdateEnabled {
		return u.Logger.NewError(OperationNotAllowedOnOfficialDatabaseAppError)
	}

	app.AutomaticUpdatesEnabled = autoUpdateEnabled
	app.AutomaticBackupsEnabled = autoBackupEnabled

	return a.AppRepo.UpdateApp(app)
}
