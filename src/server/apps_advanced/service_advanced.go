package apps_advanced

import (
	"maps"

	"server/app_migrations"
	"server/app_store"
	"server/apps_basic"
	"server/backup_server"
	"server/backups"
	"server/tools"

	"github.com/quollix/common/quollix/api"
	"github.com/quollix/common/store"
	u "github.com/quollix/common/utils"
)

const (
	AppFromAnotherMaintainerExistsAlreadyError = "this app already exists for another maintainer, upload is therefore not possible"
	CanNotUploadOlderAppVersionOverNewer       = "cannot upload an older app version over an existing newer version"
)

type AppsServiceAdvanced interface {
	UploadAppToApplication(versionFile *api.BinaryFile, composeArchive *apps_basic.ComposeArchiveName) error
	DownloadAppFromApplication(appId int) (*api.BinaryFile, error)
	UpdateAppViaAppStore(appId int) error
}

type AppsServiceAdvancedImpl struct {
	AppServiceHelper           apps_basic.AppServiceHelper
	AppRepo                    apps_basic.AppRepository
	ClientCredentialsGenerator apps_basic.ClientCredentialsGenerator
	AuthHelper                 u.AuthHelper
	BackupsService             backups.BackupService
	SshRepo                    backup_server.SshRepository
	VersionFileNameEncoder     apps_basic.VersionFileNameEncoder
	AppService                 apps_basic.AppService
	AppStoreService            app_store.AppStoreService
	SshRepositoryService       backup_server.SshRepositoryService
	AppDetector                apps_basic.AppDetector
	AppDatabaseMigrator        app_migrations.AppDatabaseMigrator
}

func (a *AppsServiceAdvancedImpl) UploadAppToApplication(versionFile *api.BinaryFile, composeArchive *apps_basic.ComposeArchiveName) error {
	port, err := a.AppServiceHelper.GetPortFromComposeYaml(versionFile.Content, composeArchive.AppName)
	if err != nil {
		return err
	}

	doesAppWithMaintainerExist, err := a.AppRepo.DoesAppWithMaintainerExist(composeArchive.Maintainer, composeArchive.AppName)
	if err != nil {
		return err
	}

	if doesAppWithMaintainerExist {
		return a.updateAppFromUploadedVersion(versionFile, composeArchive, port)
	}

	doesAppExist, err := a.AppRepo.DoesAppExist(composeArchive.AppName)
	if err != nil {
		return err
	}
	if doesAppExist {
		return u.Logger.NewError(AppFromAnotherMaintainerExistsAlreadyError)
	}

	clientId, clientSecret, err := a.ClientCredentialsGenerator.Generate()
	if err != nil {
		return err
	}
	appSecret, err := a.AuthHelper.GenerateSecret()
	if err != nil {
		return err
	}

	app := apps_basic.NewRepoApp(
		composeArchive.Maintainer,
		composeArchive.AppName,
		composeArchive.Version,
		api.Policies.AdminOnlyAccessPolicy,
		port,
		clientId,
		clientSecret,
		appSecret,
		composeArchive.VersionCreationTimestamp,
		versionFile.Content,
		false,
		false,
		true,
	)

	return a.AppService.UpsertAppInDatabase(app)
}

func (a *AppsServiceAdvancedImpl) updateAppFromUploadedVersion(versionFile *api.BinaryFile, composeArchive *apps_basic.ComposeArchiveName, port string) error {
	appFromDatabase, err := a.AppRepo.GetAppByName(composeArchive.AppName)
	if err != nil {
		return err
	}

	if composeArchive.VersionCreationTimestamp.Before(appFromDatabase.VersionCreationTimestamp) {
		return u.Logger.NewError(CanNotUploadOlderAppVersionOverNewer)
	}

	uploadedRepoApp := *appFromDatabase
	uploadedRepoApp.VersionName = composeArchive.Version
	uploadedRepoApp.VersionCreationTimestamp = composeArchive.VersionCreationTimestamp
	uploadedRepoApp.VersionContent = versionFile.Content
	uploadedRepoApp.Port = port
	return a.replaceInstalledAppVersion(appFromDatabase, &uploadedRepoApp)
}

func (a *AppsServiceAdvancedImpl) DownloadAppFromApplication(appId int) (*api.BinaryFile, error) {
	app, err := a.AppRepo.GetAppById(appId)
	if err != nil {
		return nil, err
	}
	composeArchiveName := &apps_basic.ComposeArchiveName{
		Maintainer:               app.Maintainer,
		AppName:                  app.AppName,
		Version:                  app.VersionName,
		VersionCreationTimestamp: app.VersionCreationTimestamp,
	}

	fileName, err := a.VersionFileNameEncoder.EncodeComposeArchiveName(composeArchiveName)
	if err != nil {
		return nil, err
	}

	versionFile := &api.BinaryFile{
		FileName: fileName,
		Content:  app.VersionContent,
	}
	return versionFile, nil
}

func (b *AppsServiceAdvancedImpl) UpdateAppViaAppStore(appId int) error {
	app, err := b.AppRepo.GetAppById(appId)
	if err != nil {
		return err
	}

	if b.AppDetector.IsSystemApp(app.AppName) {
		return u.Logger.NewError(apps_basic.OperationNotAllowedOnSystemAppError)
	}

	versions, err := b.AppStoreService.GetVersions(app.Maintainer, app.AppName)
	if err != nil {
		return err
	}
	if len(versions) == 0 {
		return u.Logger.NewError("no versions found for app")
	}
	latestVersionInAppStore := getLatestVersion(versions)
	if latestVersionInAppStore.CreationTimestamp.After(app.VersionCreationTimestamp) {
		err = b.updateAppFromStoreVersion(app, latestVersionInAppStore.Name)
		if err != nil {
			return err
		}
	} else {
		return u.Logger.NewError(CantUpdateAppError)
	}
	return nil
}

func (b *AppsServiceAdvancedImpl) updateAppFromStoreVersion(app *apps_basic.RepoApp, versionName string) error {
	downloadedRepoApp, err := b.AppStoreService.DownloadVersion(app.Maintainer, app.AppName, versionName)
	if err != nil {
		return err
	}
	return b.replaceInstalledAppVersion(app, downloadedRepoApp)
}

func (b *AppsServiceAdvancedImpl) replaceInstalledAppVersion(app *apps_basic.RepoApp, newApp *apps_basic.RepoApp) error {
	u.Logger.Info("updating app", tools.AppField, app.AppName)
	shouldBeRunning := app.ShouldBeRunning

	newApp.ShouldBeRunning = shouldBeRunning
	newApp.ClientId = app.ClientId
	newApp.ClientSecret = app.ClientSecret
	newApp.AppSecret = app.AppSecret
	newApp.Secrets = maps.Clone(app.Secrets)

	oldComposeContent, _, err := apps_basic.CompleteAppComposeYaml(app, "", "")
	if err != nil {
		return err
	}
	newComposeContent, _, err := apps_basic.CompleteAppComposeYaml(newApp, "", "")
	if err != nil {
		return err
	}
	if err = b.AppDatabaseMigrator.ValidateForAppUpdate(oldComposeContent, newComposeContent); err != nil {
		return err
	}

	isBackupEnabled, err := b.SshRepositoryService.IsBackupEnabled()
	if err != nil {
		return err
	}
	if isBackupEnabled {
		err = b.BackupsService.CreateBackup(app.AppId, tools.PreUpdateBackupDescription)
		if err != nil {
			return err
		}
	}

	err = b.stopAppAndApplyInstalledAppVersionReplacement(app, newApp, oldComposeContent, newComposeContent, shouldBeRunning)
	if err != nil {
		return err
	}
	return nil
}

func (b *AppsServiceAdvancedImpl) stopAppAndApplyInstalledAppVersionReplacement(app *apps_basic.RepoApp, newApp *apps_basic.RepoApp, oldComposeContent, newComposeContent []byte, shouldBeRunning bool) error {
	err := b.AppService.StopApp(app.AppId)
	if err != nil {
		return err
	}
	if err = b.AppDatabaseMigrator.MigrateForAppUpdate(app.Maintainer, app.AppName, oldComposeContent, newComposeContent); err != nil {
		return err
	}

	err = b.AppService.UpsertAppInDatabase(newApp)
	if err != nil {
		return err
	}
	if shouldBeRunning {
		return b.AppService.StartApp(app.AppId)
	}
	return nil
}

func getLatestVersion(versions []store.LeanVersionDto) store.LeanVersionDto {
	latest := versions[0]
	for _, version := range versions {
		if version.CreationTimestamp.After(latest.CreationTimestamp) {
			latest = version
		}
	}
	return latest
}
