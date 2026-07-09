package apps_advanced

import (
	"server/app_store"
	"server/apps_basic"
	"server/backup_server"
	"server/backups"
	"server/tools"

	"github.com/quollix/common/store"
	u "github.com/quollix/common/utils"
)

const (
	AppFromAnotherMaintainerExistsAlreadyError = "this app already exists for another maintainer, upload is therefore not possible"
	CanNotUploadOlderAppVersionOverNewer       = "cannot upload an older app version over an existing newer version"
)

type AppsServiceAdvanced interface {
	UploadAppToApplication(versionFile *tools.BinaryFile, composeArchive *apps_basic.ComposeArchiveName) error
	DownloadAppFromApplication(appId int) (*tools.BinaryFile, error)
	UpdateAppViaAppStore(appId int) error
}

type AppsServiceAdvancedImpl struct {
	AppServiceHelper           apps_basic.AppServiceHelper
	AppRepo                    apps_basic.AppRepository
	ClientCredentialsGenerator apps_basic.ClientCredentialsGenerator
	BackupsService             backups.BackupService
	SshRepo                    backup_server.SshRepository
	VersionFileNameEncoder     apps_basic.VersionFileNameEncoder
	AppService                 apps_basic.AppService
	AppStoreService            app_store.AppStoreService
	SshRepositoryService       backup_server.SshRepositoryService
	AppDetector                apps_basic.AppDetector
}

func (a *AppsServiceAdvancedImpl) UploadAppToApplication(versionFile *tools.BinaryFile, composeArchive *apps_basic.ComposeArchiveName) error {
	port, err := a.AppServiceHelper.GetPortFromComposeYaml(versionFile.Content, composeArchive.AppName)
	if err != nil {
		return err
	}

	doesAppWithMaintainerExist, err := a.AppRepo.DoesAppWithMaintainerExist(composeArchive.Maintainer, composeArchive.AppName)
	if err != nil {
		return err
	}

	if doesAppWithMaintainerExist {
		return a.conductAppUpdate(versionFile, composeArchive, port)
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

	app := apps_basic.NewRepoApp(
		composeArchive.Maintainer,
		composeArchive.AppName,
		composeArchive.Version,
		tools.Policies.AdminOnlyAccessPolicy,
		port,
		clientId, clientSecret,
		composeArchive.VersionCreationTimestamp,
		versionFile.Content,
		false,
		false,
		true,
	)

	_, err = a.AppRepo.CreateApp(app)
	return err
}

func (a *AppsServiceAdvancedImpl) conductAppUpdate(versionFile *tools.BinaryFile, composeArchive *apps_basic.ComposeArchiveName, port string) error {
	appFromDatabase, err := a.AppRepo.GetAppByName(composeArchive.AppName)
	if err != nil {
		return err
	}

	if composeArchive.VersionCreationTimestamp.Before(appFromDatabase.VersionCreationTimestamp) {
		return u.Logger.NewError(CanNotUploadOlderAppVersionOverNewer)
	}

	isEnabled, err := a.SshRepo.IsRemoteBackupEnabled()
	if err != nil {
		return err
	}
	if isEnabled {
		err := a.BackupsService.CreateBackup(appFromDatabase.AppId, tools.PreUpdateBackupDescription)
		if err != nil {
			return err
		}
	}

	appFromDatabase.VersionName = composeArchive.Version
	appFromDatabase.VersionCreationTimestamp = composeArchive.VersionCreationTimestamp
	appFromDatabase.VersionContent = versionFile.Content
	appFromDatabase.Port = port
	return a.AppRepo.UpdateApp(appFromDatabase)
}

func (a *AppsServiceAdvancedImpl) DownloadAppFromApplication(appId int) (*tools.BinaryFile, error) {
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

	versionFile := &tools.BinaryFile{
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
		app.VersionName = latestVersionInAppStore.Name
		err = b.installNewVersion(app)
		if err != nil {
			return err
		}
	} else {
		return u.Logger.NewError(CantUpdateAppError)
	}
	return nil
}

func (b *AppsServiceAdvancedImpl) installNewVersion(app *apps_basic.RepoApp) error {
	u.Logger.Info("updating app", tools.AppField, app.AppName)
	err := b.AppService.StopApp(app.AppId)
	if err != nil {
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

	downloadedRepoApp, err := b.AppStoreService.DownloadVersion(app.Maintainer, app.AppName, app.VersionName)
	if err != nil {
		return err
	}
	downloadedRepoApp.ShouldBeRunning = app.ShouldBeRunning
	downloadedRepoApp.ClientId = app.ClientId
	downloadedRepoApp.ClientSecret = app.ClientSecret
	err = b.AppService.UpsertAppInDatabase(downloadedRepoApp)
	if err != nil {
		return err
	}
	if app.ShouldBeRunning {
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
