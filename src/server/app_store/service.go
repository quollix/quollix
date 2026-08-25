package app_store

import (
	"fmt"
	"server/apps_basic"

	"github.com/quollix/common/quollix/api"
	"github.com/quollix/common/store"
	u "github.com/quollix/common/utils"
	"github.com/quollix/common/validation"
)

type AppStoreService interface {
	DownloadVersionByID(versionId int) (*apps_basic.RepoApp, error)
	InstallDownloadedVersion(app *apps_basic.RepoApp) error
	GetVersions(userName, appName string) ([]store.LeanVersionDto, error)
	SearchForApps(sr *store.SearchRequest) ([]store.AppWithLatestVersion, error)
	GetStoreVersionDownloadForBrowser(versionId int) (*api.BinaryFile, error)
}

const (
	AppAlreadyInstalledError                     = "an app with that name is already installed"
	AppFromAnotherMaintainerExistsAlreadyError   = "this app already exists for another maintainer, install is therefore not possible"
	CanNotInstallOlderAppVersionOverNewerOrEqual = "cannot install an older or same app version over an existing newer or same version"
)

type AppStoreServiceImpl struct {
	AppStoreClientLean         AppStoreClientLean
	AppRepo                    apps_basic.AppRepository
	ClientCredentialsGenerator apps_basic.ClientCredentialsGenerator
	AuthHelper                 u.AuthHelper
	VersionFileNameEncoder     apps_basic.VersionFileNameEncoder
	AppServiceHelper           apps_basic.AppServiceHelper
	AppService                 apps_basic.AppService
	VersionValidator           validation.VersionValidator
	VersionVerifier            VersionVerifier
}

func (a *AppStoreServiceImpl) DownloadVersionByID(versionId int) (*apps_basic.RepoApp, error) {
	fullTagInfo, err := a.AppStoreClientLean.DownloadVersionByID(versionId)
	if err != nil {
		return nil, err
	}
	if err := a.VersionValidator.Validate(fullTagInfo.Content, fullTagInfo.Maintainer, fullTagInfo.AppName); err != nil {
		return nil, fmt.Errorf("version validation failed: %w", err)
	}
	if err := a.VersionVerifier.Verify(fullTagInfo); err != nil {
		return nil, err
	}
	clientId, clientSecret, err := a.ClientCredentialsGenerator.Generate()
	if err != nil {
		return nil, err
	}
	appSecret, err := a.AuthHelper.GenerateSecret()
	if err != nil {
		return nil, err
	}

	port, err := a.AppServiceHelper.GetPortFromComposeYaml(fullTagInfo.Content, fullTagInfo.AppName)
	if err != nil {
		return nil, err
	}

	repoApp := apps_basic.NewRepoApp(
		fullTagInfo.Maintainer,
		fullTagInfo.AppName,
		fullTagInfo.VersionName,
		api.Policies.AdminOnlyAccessPolicy,
		port,
		clientId,
		clientSecret,
		appSecret,
		fullTagInfo.VersionCreationTimestamp,
		fullTagInfo.Content,
		false,
		true,
		true,
	)
	return repoApp, nil
}

func (a *AppStoreServiceImpl) InstallDownloadedVersion(app *apps_basic.RepoApp) error {
	doesAppExist, err := a.AppRepo.DoesAppExist(app.AppName)
	if err != nil {
		return err
	}
	if doesAppExist {
		return u.Logger.NewError(AppAlreadyInstalledError)
	}

	app.AccessPolicy = api.Policies.AdminOnlyAccessPolicy
	err = a.AppService.UpsertAppInDatabase(app)
	if err != nil {
		return u.Logger.NewError(err.Error())
	}
	installedApp, err := a.AppRepo.GetAppByName(app.AppName)
	if err != nil {
		return err
	}
	return a.AppService.StartApp(installedApp.AppId)
}

func (a *AppStoreServiceImpl) GetVersions(userName, appName string) ([]store.LeanVersionDto, error) {
	return a.AppStoreClientLean.ListVersions(userName, appName)
}

func (a *AppStoreServiceImpl) SearchForApps(sr *store.SearchRequest) ([]store.AppWithLatestVersion, error) {
	apps, err := a.AppStoreClientLean.SearchForApps(sr.MaintainerSearchTerm, sr.AppSearchTerm, sr.ShowUnofficialApps)
	if err != nil {
		return nil, err
	}
	if len(apps) == 0 {
		return nil, u.Logger.NewError(NoAppsFoundError)
	}
	return apps, nil
}

func (a *AppStoreServiceImpl) GetStoreVersionDownloadForBrowser(versionId int) (*api.BinaryFile, error) {
	repoApp, err := a.DownloadVersionByID(versionId)
	if err != nil {
		return nil, err
	}

	composeArchiveName := &apps_basic.ComposeArchiveName{
		Maintainer:               repoApp.Maintainer,
		AppName:                  repoApp.AppName,
		Version:                  repoApp.VersionName,
		VersionCreationTimestamp: repoApp.VersionCreationTimestamp,
	}
	fileName, err := a.VersionFileNameEncoder.EncodeComposeArchiveName(composeArchiveName)
	if err != nil {
		return nil, err
	}
	versionDownload := api.BinaryFile{
		FileName: fileName,
		Content:  repoApp.VersionContent,
	}
	return &versionDownload, nil
}
