package apps_basic

import (
	"maps"
	"server/tools"

	api "github.com/quollix/common/quollix/api"
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
	ListAppsForRole(userId int, role tools.UserAccessLevel) ([]api.AdminAppDto, error)
	ListAppsForNonAdmin(userId int, role tools.UserAccessLevel) ([]api.NonAdminAppDto, error)
	UpdateAppAutoMaintenanceSettings(appId int, autoUpdateEnabled, autoBackupEnabled bool) error
	RegenerateOidcClientCredentials(appId int) error
	RegenerateAppSecret(appId int, secretName string) error
}

type AppServiceImpl struct {
	AppRepo                    AppRepository
	DockerService              tools.DockerService
	AppServiceHelper           AppServiceHelper
	AppDetector                AppDetector
	ComposeExtractor           ComposeExtractorImpl
	ComposeSecretExtractor     ComposeSecretExtractor
	ClientCredentialsGenerator ClientCredentialsGenerator
	DatabaseIndependentRuntime DatabaseIndependentRuntime
	VersionFileNameEncoder     VersionFileNameEncoder
	AuthHelper                 u.AuthHelper
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

func (a *AppServiceImpl) RegenerateAppSecret(appId int, secretName string) error {
	app, err := a.AppRepo.GetAppById(appId)
	if err != nil {
		return err
	}
	if _, exists := app.Secrets[secretName]; !exists {
		return u.Logger.NewError("app secret not found", tools.AppIdField, appId, "secret_name", secretName)
	}

	app.Secrets[secretName], err = a.AuthHelper.GenerateSecret()
	if err != nil {
		return err
	}
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
		if err = a.ensureAppSecrets(app, repoApp); err != nil {
			return err
		}
		err = a.AppRepo.UpdateApp(app)
		if err != nil {
			return err
		}
	} else {
		if err = a.ensureAppSecrets(app, nil); err != nil {
			return err
		}
		_, err = a.AppRepo.CreateApp(app)
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *AppServiceImpl) ensureAppSecrets(app *RepoApp, existingApp *RepoApp) error {
	var existingSecrets map[string]string
	if existingApp != nil {
		existingSecrets = existingApp.Secrets
	}

	secrets := copyAppSecrets(existingSecrets)
	maps.Copy(secrets, app.Secrets)

	requiredSecrets, err := a.ComposeSecretExtractor.Extract(app.VersionContent)
	if err != nil {
		return err
	}
	for _, secretName := range requiredSecrets {
		if _, exists := secrets[secretName]; exists {
			continue
		}
		if legacyValue, exists := legacySecretValue(existingApp, secretName); exists {
			secrets[secretName] = legacyValue
			continue
		}

		secrets[secretName], err = a.AuthHelper.GenerateSecret()
		if err != nil {
			return err
		}
	}

	app.Secrets = secrets
	return nil
}

func copyAppSecrets(secrets map[string]string) map[string]string {
	copiedSecrets := map[string]string{}
	maps.Copy(copiedSecrets, secrets)
	return copiedSecrets
}

func (a *AppServiceImpl) ListAppsForRole(userId int, role tools.UserAccessLevel) ([]api.AdminAppDto, error) {
	filteredApps, err := a.listVisibleRepoApps(userId, role)
	if err != nil {
		return nil, err
	}
	appDtos := a.AppServiceHelper.ConvertToAdminAppDtos(filteredApps)
	if role != tools.AdminLevel {
		clearSensitiveAppDtoFields(appDtos)
	}
	return appDtos, nil
}

func (a *AppServiceImpl) ListAppsForNonAdmin(userId int, role tools.UserAccessLevel) ([]api.NonAdminAppDto, error) {
	filteredApps, err := a.listVisibleRepoApps(userId, role)
	if err != nil {
		return nil, err
	}
	return a.AppServiceHelper.ConvertToNonAdminAppDtos(filteredApps), nil
}

func (a *AppServiceImpl) listVisibleRepoApps(userId int, role tools.UserAccessLevel) ([]RepoApp, error) {
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
	return filteredApps, nil
}

func clearSensitiveAppDtoFields(apps []api.AdminAppDto) {
	for i := range apps {
		apps[i].AppId = ""
		apps[i].Port = ""
		apps[i].ClientId = ""
		apps[i].ClientSecret = ""
		apps[i].AppSecret = ""
		apps[i].VersionContent = nil
		apps[i].Secrets = nil
		apps[i].AutomaticBackupsEnabled = false
		apps[i].AutomaticUpdatesEnabled = false
	}
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
