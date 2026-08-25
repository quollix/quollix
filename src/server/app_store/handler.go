package app_store

import (
	"net/http"

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

var storeVersionBrowserDownloadExpectedErrors = u.MapOf(InvalidPackageSigningError)

func (a *AppStoreHandler) DownloadStoreVersionForBrowserHandler(w http.ResponseWriter, r *http.Request) {
	versionID, ok := validation.ReadBody[store.VersionID](w, r)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	versionDownload, err := a.AppStoreService.GetStoreVersionDownloadForBrowser(versionID.VersionId)
	if err != nil {
		u.WriteResponseError(w, storeVersionBrowserDownloadExpectedErrors, err)
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
