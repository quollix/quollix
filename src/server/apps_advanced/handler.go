package apps_advanced

import (
	"fmt"
	"net/http"
	"server/apps_basic"
	"server/tools"
	"strconv"

	u "github.com/quollix/common/utils"
	"github.com/quollix/common/validation"
)

var (
	ExpectedAppUploadErrors = u.MapOf(AppFromAnotherMaintainerExistsAlreadyError, CanNotUploadOlderAppVersionOverNewer)
	CantUpdateAppError      = "can't update app because the latest version is already installed"
	ExpectedUpdateErrors    = u.MapOf(CantUpdateAppError, apps_basic.OperationNotAllowedOnSystemAppError)
)

type AppsAdvancedHandler struct {
	VersionFileNameEncoder apps_basic.VersionFileNameEncoder
	VersionValidator       validation.VersionValidator
	AppServiceAdvanced     AppsServiceAdvanced
	AppRepo                apps_basic.AppRepository
	OperationRegistry      apps_basic.OperationRegistry
}

func (a *AppsAdvancedHandler) UploadVersionFileToApplicationHandler(w http.ResponseWriter, r *http.Request) {
	versionFile, ok := validation.ReadBody[tools.BinaryFile](w, r)
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
		u.WriteResponseError(w, ExpectedAppUploadErrors, err)
		return
	}
}

func (a *AppsAdvancedHandler) DownloadVersionFileFromApplicationHandler(w http.ResponseWriter, r *http.Request) {
	appIdString, ok := validation.ReadBody[tools.NumberString](w, r)
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
	appIdString, ok := validation.ReadBody[tools.NumberString](w, r)
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
		u.WriteResponseError(w, ExpectedUpdateErrors, err)
		return
	}
}
