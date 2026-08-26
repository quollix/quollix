package apps_advanced

import (
	"fmt"
	"net/http"
	"server/app_migrations"
	"server/app_store"
	"server/apps_basic"
	"strconv"

	"github.com/quollix/common/quollix/api"
	"github.com/quollix/common/store"
	u "github.com/quollix/common/utils"
	"github.com/quollix/common/validation"
)

var (
	CantUpdateAppError = "can't update app because the latest version is already installed"

	ExpectedAppMutationErrors = map[string]any{
		AppFromAnotherMaintainerExistsAlreadyError:             nil,
		CanNotUploadOlderAppVersionOverNewer:                   nil,
		CantUpdateAppError:                                     nil,
		app_store.AppAlreadyInstalledError:                     nil,
		app_store.AppFromAnotherMaintainerExistsAlreadyError:   nil,
		app_store.CanNotInstallOlderAppVersionOverNewerOrEqual: nil,
		app_store.InvalidPackageSigningError:                   nil,
		apps_basic.OperationNotAllowedOnSystemAppError:         nil,
		app_migrations.InvalidPostgresEnvironmentError:         nil,
		app_migrations.MissingPostgresDataVolumeError:          nil,
		app_migrations.MultiplePostgresServicesError:           nil,
		app_migrations.MultipleRabbitMQServicesError:           nil,
		app_migrations.PostgresDataVolumeChangedError:          nil,
		app_migrations.PostgresUserChangedError:                nil,
	}
)

type AppsAdvancedHandler struct {
	VersionFileNameEncoder apps_basic.VersionFileNameEncoder
	VersionValidator       validation.VersionValidator
	AppServiceAdvanced     AppsServiceAdvanced
	AppStoreService        app_store.AppStoreService
	AppRepo                apps_basic.AppRepository
	OperationRegistry      apps_basic.OperationRegistry
}

func (a *AppsAdvancedHandler) UploadVersionFileToApplicationHandler(w http.ResponseWriter, r *http.Request) {
	versionFile, ok := validation.ReadBody[api.BinaryFile](w, r)
	if !ok {
		return
	}

	composeArchive, err := a.VersionFileNameEncoder.DecodeComposeArchiveName(versionFile.FileName)
	if err != nil {
		u.WriteResponseErrorAlways(w, err)
		return
	}

	if err := a.VersionValidator.Validate(versionFile.Content, composeArchive.Maintainer, composeArchive.AppName); err != nil {
		u.WriteResponseErrorAlways(w, err)
		return
	}

	err = a.AppServiceAdvanced.UploadAppToApplication(versionFile, composeArchive)
	if err != nil {
		u.WriteResponseError(w, ExpectedAppMutationErrors, err)
		return
	}
}

func (a *AppsAdvancedHandler) DownloadVersionFileFromApplicationHandler(w http.ResponseWriter, r *http.Request) {
	appIdString, ok := validation.ReadBody[api.NumberString](w, r)
	if !ok {
		return
	}

	appId, err := strconv.Atoi(appIdString.Value)
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}

	versionFile, err := a.AppServiceAdvanced.DownloadAppFromApplication(appId)
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}
	u.SendJsonResponse(w, versionFile)
}

func (a *AppsAdvancedHandler) VersionUpdateHandler(w http.ResponseWriter, r *http.Request) {
	appIdString, ok := validation.ReadBody[api.NumberString](w, r)
	if !ok {
		return
	}

	appId, err := strconv.Atoi(appIdString.Value)
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}

	app, err := a.AppRepo.GetAppById(appId)
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}

	operation := fmt.Sprintf("updating '%s'", app.AppName)
	handle, err := a.OperationRegistry.TryBlockAppOperation(app.AppName, operation)
	if err != nil {
		apps_basic.WriteConcurrentOperationError(w, operation, err)
		return
	}
	defer handle.Done()

	err = a.AppServiceAdvanced.UpdateAppViaAppStore(appId)
	if err != nil {
		u.WriteResponseError(w, ExpectedAppMutationErrors, err)
		return
	}
}

func (a *AppsAdvancedHandler) InstallOrUpdateAppFromStoreVersionHandler(w http.ResponseWriter, r *http.Request) {
	versionID, ok := validation.ReadBody[store.VersionID](w, r)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	downloadedRepoApp, err := a.AppStoreService.DownloadVersionByID(versionID.VersionId)
	if err != nil {
		u.WriteResponseError(w, ExpectedAppMutationErrors, err)
		return
	}

	doesAppExist, err := a.AppRepo.DoesAppExist(downloadedRepoApp.AppName)
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}
	if !doesAppExist {
		operation := fmt.Sprintf("installing '%s'", downloadedRepoApp.AppName)
		handle, err := a.OperationRegistry.TryBlockAppOperation(downloadedRepoApp.AppName, operation)
		if err != nil {
			apps_basic.WriteConcurrentOperationError(w, operation, err)
			return
		}
		defer handle.Done()

		err = a.AppStoreService.InstallDownloadedVersion(downloadedRepoApp)
		if err != nil {
			u.WriteResponseError(w, ExpectedAppMutationErrors, err)
			return
		}
		return
	}

	app, err := a.AppRepo.GetAppByName(downloadedRepoApp.AppName)
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}
	if app.Maintainer != downloadedRepoApp.Maintainer {
		u.WriteResponseError(w, ExpectedAppMutationErrors, u.Logger.NewError(app_store.AppFromAnotherMaintainerExistsAlreadyError))
		return
	}

	operation := fmt.Sprintf("updating '%s'", app.AppName)
	handle, err := a.OperationRegistry.TryBlockAppOperation(app.AppName, operation)
	if err != nil {
		apps_basic.WriteConcurrentOperationError(w, operation, err)
		return
	}
	defer handle.Done()

	err = a.AppServiceAdvanced.UpdateAppToStoreVersion(app.AppId, downloadedRepoApp)
	if err != nil {
		u.WriteResponseError(w, ExpectedAppMutationErrors, err)
		return
	}
}
