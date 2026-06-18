package apps_basic

import (
	"encoding/json"
	"fmt"
	"net/http"
	"server/tools"
	"server/users"
	"strconv"

	u "github.com/quollix/common/utils"
	"github.com/quollix/common/validation"
)

var (
	InvalidAccessPolicyError = "invalid access policy"
	ExpectedAppStartErrors   = u.MapOf(tools.DockerHubRateLimitReachedErrorMessage, tools.DockerImageUnsupportedPlatformErrorMessage)
)

type AppsHandler struct {
	OperationRegistry      OperationRegistry
	AppService             AppService
	AppRepo                AppRepository
	UserRepo               users.UserRepository
	AuthHelper             u.AuthHelper
	AppDetector            AppDetector
	DatabaseConnector      tools.DatabaseConnector
	UserService            users.UserService
	VersionFileNameEncoder VersionFileNameEncoder
	VersionValidator       validation.VersionValidator
}

func (a *AppsHandler) AppListHandler(w http.ResponseWriter, r *http.Request) {
	userId, role, err := a.UserService.GetUserIdAndRoleFromQuollixRequest(r)
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}

	appDtos, err := a.AppService.ListAppsForRole(userId, role)
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}

	u.SendJsonResponse(w, appDtos)
}

func (a *AppsHandler) AppStartHandler(w http.ResponseWriter, r *http.Request) {
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

	operation := fmt.Sprintf("starting '%s'", app.AppName)
	handle, err := a.OperationRegistry.TryBlockAppOperation(app.AppName, operation)
	if err != nil {
		WriteConcurrentOperationError(w, operation, err)
		return
	}
	defer handle.Done()

	if a.AppDetector.WriteErrorIfOfficialDatabaseAppIsAddressed(w, appId) {
		return
	}

	err = a.AppService.StartApp(appId)
	if err != nil {
		u.WriteResponseError(w, ExpectedAppStartErrors, err)
		return
	}
}

func (a *AppsHandler) AppStopHandler(w http.ResponseWriter, r *http.Request) {
	appIdString, ok := validation.ReadBody[tools.NumberString](w, r)
	if !ok {
		return
	}

	appId, err := strconv.Atoi(appIdString.Value)
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}

	if a.AppDetector.WriteErrorIfOfficialDatabaseAppIsAddressed(w, appId) {
		return
	}

	app, err := a.AppRepo.GetAppById(appId)
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}

	operation := fmt.Sprintf("stopping '%s'", app.AppName)
	handle, err := a.OperationRegistry.TryBlockAppOperation(app.AppName, operation)
	if err != nil {
		WriteConcurrentOperationError(w, operation, err)
		return
	}
	defer handle.Done()

	err = a.AppService.StopApp(appId)
	if err != nil {
		u.WriteResponseError(w, OfficialDatabaseAppErrorMap, err)
		return
	}
}

type ChangeAccessPolicyRequest struct {
	AppId        string `json:"app_id" validate:"number"`
	AccessPolicy string `json:"access_policy" validate:"access_policy"`
}

var allowedAccessPolicies = map[string]any{
	tools.Policies.PublicAccessPolicy:          nil,
	tools.Policies.AuthenticatedAccessPolicy:   nil,
	tools.Policies.GroupRestrictedAccessPolicy: nil,
	tools.Policies.AdminOnlyAccessPolicy:       nil,
}

func (a *AppsHandler) ChangeAccessPolicyHandler(w http.ResponseWriter, r *http.Request) {
	var req ChangeAccessPolicyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid input", http.StatusBadRequest)
		return
	}

	_, ok := allowedAccessPolicies[req.AccessPolicy]
	if !ok {
		http.Error(w, InvalidAccessPolicyError, http.StatusBadRequest)
		return
	}

	appId, err := strconv.Atoi(req.AppId)
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}

	if a.AppDetector.WriteErrorIfOfficialDatabaseAppIsAddressed(w, appId) {
		return
	}

	err = a.AppService.SetAccessPolicy(appId, req.AccessPolicy)
	if err != nil {
		u.WriteResponseError(w, SetAccessPolicyExpectedErrors, err)
		return
	}
}

func (a *AppsHandler) AppPruneHandler(w http.ResponseWriter, r *http.Request) {
	appIdString, ok := validation.ReadBody[tools.NumberString](w, r)
	if !ok {
		return
	}

	appId, err := strconv.Atoi(appIdString.Value)
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}

	if a.AppDetector.WriteErrorIfOfficialDatabaseAppIsAddressed(w, appId) {
		return
	}

	app, err := a.AppRepo.GetAppById(appId)
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}

	operation := fmt.Sprintf("deleting '%s'", app.AppName)
	handle, err := a.OperationRegistry.TryBlockAppOperation(app.AppName, operation)
	if err != nil {
		WriteConcurrentOperationError(w, operation, err)
		return
	}
	defer handle.Done()

	err = a.AppService.DeleteAppAndArtifacts(appId)
	if err != nil {
		u.WriteResponseError(w, OfficialDatabaseAppErrorMap, err)
		return
	}

}

type AppOperationInfoResponse struct {
	Operations            []string `json:"operations"`
	IsOngoing             bool     `json:"is_ongoing"`
	AppOperationsFinished []string `json:"app_operations_finished"`
}

func (a *AppsHandler) AppOperationInfoHandler(w http.ResponseWriter, r *http.Request) {
	operations := a.OperationRegistry.ListOperations()
	response := AppOperationInfoResponse{
		Operations:            operations,
		IsOngoing:             len(operations) > 0,
		AppOperationsFinished: a.OperationRegistry.ListFinishedAppOperations(),
	}
	u.SendJsonResponse(w, response)
}

func WriteConcurrentOperationError(w http.ResponseWriter, attemptedOperation string, err error) {
	u.Logger.Info(concurrentOperationErrorMessage, tools.AttemptedOperationField, attemptedOperation, "error", err)
	http.Error(w, concurrentOperationErrorMessage, http.StatusBadRequest)
}

type AutoMaintenanceSettingsResponse struct {
	AppId                   string `json:"app_id" validate:"number"`
	AutomaticUpdatesEnabled bool   `json:"automatic_updates_enabled"`
	AutomaticBackupsEnabled bool   `json:"automatic_backups_enabled"`
}

func (a *AppsHandler) UpdateAutomaticMaintenanceSettingsHandler(w http.ResponseWriter, r *http.Request) {
	autoMaintenanceSettings, ok := validation.ReadBody[AutoMaintenanceSettingsResponse](w, r)
	if !ok {
		return
	}

	appId, err := strconv.Atoi(autoMaintenanceSettings.AppId)
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}

	err = a.AppService.UpdateAppAutoMaintenanceSettings(appId, autoMaintenanceSettings.AutomaticUpdatesEnabled, autoMaintenanceSettings.AutomaticBackupsEnabled)
	if err != nil {
		u.WriteResponseError(w, OfficialDatabaseAppErrorMap, err)
		return
	}
}

func (a *AppsHandler) RegenerateOidcClientCredentials(w http.ResponseWriter, r *http.Request) {
	appIdString, ok := validation.ReadBody[tools.NumberString](w, r)
	if !ok {
		return
	}

	appId, err := strconv.Atoi(appIdString.Value)
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}

	err = a.AppService.RegenerateOidcClientCredentials(appId)
	if err != nil {
		u.WriteResponseError(w, OfficialDatabaseAppErrorMap, err)
		return
	}
}

func (a *AppsHandler) IsDatabaseAvailableHandler(w http.ResponseWriter, r *http.Request) {
	if err := a.DatabaseConnector.GetDB().Ping(); err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}
}
