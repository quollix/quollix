package app_store

import (
	"fmt"
	"server/apps_basic"
	"server/tools"

	"github.com/quollix/common/store"
	u "github.com/quollix/common/utils"
	"github.com/quollix/common/validation"
)

type AppStoreService interface {
	DownloadVersion(maintainerName, appName, versionName string) (*apps_basic.RepoApp, error)
	DownloadAndInstallVersion(versionTree *store.VersionTree) error
	GetVersions(userName, appName string) ([]store.LeanVersionDto, error)
	SearchForApps(sr *store.SearchRequest) ([]store.AppWithLatestVersion, error)
	GetVersionDownload(versionTree *store.VersionTree) (*tools.BinaryFile, error)
}

var AppAlreadyInstalledError = "an app with that name is already installed"

type AppStoreServiceImpl struct {
	AppStoreClientLean         AppStoreClientLean
	AppRepo                    apps_basic.AppRepository
	ClientCredentialsGenerator apps_basic.ClientCredentialsGenerator
	VersionFileNameEncoder     apps_basic.VersionFileNameEncoder
	AppServiceHelper           apps_basic.AppServiceHelper
	VersionValidator           validation.VersionValidator
	VersionVerifier            VersionVerifier
}

func (a *AppStoreServiceImpl) DownloadVersion(maintainerName, appName, versionName string) (*apps_basic.RepoApp, error) {
	fullTagInfo, err := a.AppStoreClientLean.DownloadVersion(maintainerName, appName, versionName)
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

	port, err := a.AppServiceHelper.GetPortFromComposeYaml(fullTagInfo.Content, appName)
	if err != nil {
		return nil, err
	}

	repoApp := apps_basic.NewRepoApp(
		fullTagInfo.Maintainer,
		fullTagInfo.AppName,
		fullTagInfo.VersionName,
		tools.Policies.AdminOnlyAccessPolicy,
		port,
		clientId,
		clientSecret,
		fullTagInfo.VersionCreationTimestamp,
		fullTagInfo.Content,
		false,
		true,
		true,
	)
	return repoApp, nil
}

func (a *AppStoreServiceImpl) DownloadAndInstallVersion(versionTree *store.VersionTree) error {
	doesAppExist, err := a.AppRepo.DoesAppExist(versionTree.AppName)
	if err != nil {
		return err
	}
	if doesAppExist {
		return u.Logger.NewError(AppAlreadyInstalledError)
	}

	app, err := a.DownloadVersion(versionTree.Maintainer, versionTree.AppName, versionTree.VersionName)
	if err != nil {
		return err
	}

	app.AccessPolicy = tools.Policies.AdminOnlyAccessPolicy
	_, err = a.AppRepo.CreateApp(app)
	if err != nil {
		return u.Logger.NewError(err.Error())
	}

	return nil
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

func (a *AppStoreServiceImpl) GetVersionDownload(versionTree *store.VersionTree) (*tools.BinaryFile, error) {
	repoApp, err := a.DownloadVersion(versionTree.Maintainer, versionTree.AppName, versionTree.VersionName)
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
	versionDownload := tools.BinaryFile{
		FileName: fileName,
		Content:  repoApp.VersionContent,
	}
	return &versionDownload, nil
}
