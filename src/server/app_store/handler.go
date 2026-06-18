package app_store

import (
	"net/http"
	"server/apps_basic"

	"github.com/quollix/common/store"
	u "github.com/quollix/common/utils"
	"github.com/quollix/common/validation"
)

var NoAppsFoundError = "no apps_basic found"
var PublishedAppsDirectoryDoesNotExistError = "published apps directory does not exist"
var PublishedAppsDirectoryIsEmptyError = "published apps directory is empty"
var expectedNoAppsFoundErrors = u.MapOf(NoAppsFoundError)
var expectedPublishedAppsReloadErrors = u.MapOf(
	PublishedAppsDirectoryDoesNotExistError,
	PublishedAppsDirectoryIsEmptyError,
)

type AppStoreHandler struct {
	AppStoreService AppStoreService
	AppService      apps_basic.AppService
	AppRepo         apps_basic.AppRepository
	AppStoreClient  AppStoreClientLean
}

func (a *AppStoreHandler) GetVersionsHandler(w http.ResponseWriter, r *http.Request) {
	appTree, ok := validation.ReadBody[store.AppTree](w, r)
	if !ok {
		return
	}

	versions, err := a.AppStoreService.GetVersions(appTree.Maintainer, appTree.AppName)
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}
	u.SendJsonResponse(w, versions)
}

func (a *AppStoreHandler) SearchAppsHandler(w http.ResponseWriter, r *http.Request) {
	searchRequest, ok := validation.ReadBody[store.SearchRequest](w, r)
	if !ok {
		return
	}

	apps, err := a.AppStoreService.SearchForApps(searchRequest)
	if err != nil {
		u.WriteResponseError(w, expectedNoAppsFoundErrors, err)
		return
	}
	u.SendJsonResponse(w, apps)
}

var versionInstallationExpectedErrors = u.MapOf(AppAlreadyInstalledError, InvalidPackageSigningError)
var versionDownloadExpectedErrors = u.MapOf(InvalidPackageSigningError)

func (a *AppStoreHandler) VersionInstallationHandler(w http.ResponseWriter, r *http.Request) {
	versionTree, ok := validation.ReadBody[store.VersionTree](w, r)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err := a.AppStoreService.DownloadAndInstallVersion(versionTree)
	if err != nil {
		u.WriteResponseError(w, versionInstallationExpectedErrors, err)
		return
	}
}

func (a *AppStoreHandler) VersionDownloadHandler(w http.ResponseWriter, r *http.Request) {
	versionTree, ok := validation.ReadBody[store.VersionTree](w, r)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	versionDownload, err := a.AppStoreService.GetVersionDownload(versionTree)
	if err != nil {
		u.WriteResponseError(w, versionDownloadExpectedErrors, err)
		return
	}
	u.SendJsonResponse(w, versionDownload)
}

func (a *AppStoreHandler) ReloadLocalAppsHandler(w http.ResponseWriter, r *http.Request) {
	u.Logger.Info("Reloading local store apps from disk into app store client")
	err := a.AppStoreClient.ReloadLocalApps()
	if err != nil {
		u.WriteResponseError(w, expectedPublishedAppsReloadErrors, err)
		return
	}
}
