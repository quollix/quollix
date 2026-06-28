package apps_basic

import (
	"fmt"
	"net/http"
	"server/tools"

	u "github.com/quollix/common/utils"
)

var (
	OperationNotAllowedOnOfficialDatabaseAppError = fmt.Sprintf("operation not allowed on %s app", u.OfficialDatabaseAppName)
	OfficialDatabaseAppErrorMap                   = u.MapOf(OperationNotAllowedOnOfficialDatabaseAppError)
	SetAccessPolicyExpectedErrors                 = u.MapOf(OperationNotAllowedOnOfficialDatabaseAppError)

	OperationNotAllowedOnSystemAppError = "operation not allowed on system app"
)

type AppDetector interface {
	IsOfficialDatabaseApp(appName string) bool
	IsSystemApp(appName string) bool
	IsOfficialApp(maintainerName string) bool
	WriteErrorIfOfficialDatabaseAppIsAddressed(w http.ResponseWriter, appId int) bool
}

type AppDetectorImpl struct {
	AppRepo AppRepository
}

func (a *AppDetectorImpl) IsOfficialApp(maintainerName string) bool {
	return maintainerName == u.OfficialMaintainer
}

func (a *AppDetectorImpl) IsOfficialDatabaseApp(appName string) bool {
	return appName == u.OfficialDatabaseAppName
}

func (a *AppDetectorImpl) IsSystemApp(appName string) bool {
	return appName == u.OfficialDatabaseAppName || appName == u.OfficialBrandAppName || appName == tools.QuollogAppName
}

func (a *AppDetectorImpl) WriteErrorIfOfficialDatabaseAppIsAddressed(w http.ResponseWriter, appId int) bool {
	app, err := a.AppRepo.GetAppById(appId)
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return true
	}
	if a.IsOfficialDatabaseApp(app.AppName) {
		u.WriteResponseError(w, OfficialDatabaseAppErrorMap, u.Logger.NewError(OperationNotAllowedOnOfficialDatabaseAppError))
		return true
	}
	return false
}
