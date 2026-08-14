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

	AppSecretNotFoundError = "app secret not found" // #nosec G101 (CWE-798): Potential hardcoded credentials
	AppSecretInUseError    = "app secret is in use"
)

type AppService interface {
	StartApp(appId int) error
	StopApp(appId int) error
	StartAppsThatShouldBeRunning()
	DeleteAppAndArtifacts(appId int) error

	SetAppShouldBeRunning(appId int, shouldBeRunning bool) error
	SetAccessPolicy(appId int, policy string) error
	UpsertAppInDatabase(app *RepoApp) error
	ListAppsForAdmin() ([]api.AdminAppDto, error)
	ListAppsForNonAdmin(userId int, role tools.UserAccessLevel) ([]api.NonAdminAppDto, error)
	UpdateAppAutoMaintenanceSettings(appId int, autoUpdateEnabled, autoBackupEnabled bool) error
	RegenerateOidcClientCredentials(appId int) error
	UpdateAppSecret(appId int, secretName, value string) error
	RegenerateAppSecret(appId int, secretName string) error
	DeleteUnusedAppSecret(appId int, secretName string) error
}

type AppServiceImpl struct {
	AppRepo                    AppRepository
	DockerService              tools.DockerService
	AppServiceHelper           AppServiceHelper
	AppDetector                AppDetector
	ComposeExtractor           ComposeExtractor
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
		return u.Logger.NewError(AppSecretNotFoundError, tools.AppIdField, appId, "secret_name", secretName)
	}

	app.Secrets[secretName], err = a.AuthHelper.GenerateSecret()
	if err != nil {
		return err
	}
	return a.AppRepo.UpdateApp(app)
}

func (a *AppServiceImpl) UpdateAppSecret(appId int, secretName, value string) error {
	app, err := a.AppRepo.GetAppById(appId)
	if err != nil {
		return err
	}
	if _, exists := app.Secrets[secretName]; !exists {
		return u.Logger.NewError(AppSecretNotFoundError, tools.AppIdField, appId, "secret_name", secretName)
	}

	app.Secrets[secretName] = value
	return a.AppRepo.UpdateApp(app)
}

func (a *AppServiceImpl) DeleteUnusedAppSecret(appId int, secretName string) error {
	app, err := a.AppRepo.GetAppById(appId)
	if err != nil {
		return err
	}
	if _, exists := app.Secrets[secretName]; !exists {
		return u.Logger.NewError(AppSecretNotFoundError, tools.AppIdField, appId, "secret_name", secretName)
	}

	requiredSecrets, err := a.ComposeSecretExtractor.ExtractSecretSet(app.VersionContent)
	if err != nil {
		return err
	}
	if requiredSecrets[secretName] {
		return u.Logger.NewError(AppSecretInUseError, tools.AppIdField, appId, "secret_name", secretName)
	}

	delete(app.Secrets, secretName)
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
	secrets := maps.Clone(app.Secrets)
	if existingApp != nil {
		secrets = maps.Clone(existingApp.Secrets)
		maps.Copy(secrets, app.Secrets)
	}
	if secrets == nil {
		u.Logger.Warn("this path should never be triggered, secrets should never be nil but rather an empty map")
		secrets = map[string]string{}
	}

	requiredSecrets, err := a.ComposeSecretExtractor.ExtractSecrets(app.VersionContent)
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

func (a *AppServiceImpl) ListAppsForAdmin() ([]api.AdminAppDto, error) {
	repoApps, err := a.AppRepo.ListApps()
	if err != nil {
		return nil, err
	}
	return a.AppServiceHelper.ConvertToAdminAppDtos(repoApps), nil
}

func (a *AppServiceImpl) ListAppsForNonAdmin(userId int, role tools.UserAccessLevel) ([]api.NonAdminAppDto, error) {
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
	return a.AppServiceHelper.ConvertToNonAdminAppDtos(filteredApps), nil
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
