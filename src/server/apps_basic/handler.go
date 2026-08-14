package apps_basic

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"server/tools"
	"server/users"

	api "github.com/quollix/common/quollix/api"
	u "github.com/quollix/common/utils"
	"github.com/quollix/common/validation"
)

var (
	InvalidAccessPolicyError      = "invalid access policy"
	ExpectedAppStartErrors        = u.MapOf(tools.DockerHubRateLimitReachedErrorMessage, tools.DockerImageUnsupportedPlatformErrorMessage)
	AppSecretEditExpectedErrors   = u.MapOf(AppSecretNotFoundError)
	AppSecretDeleteExpectedErrors = u.MapOf(AppSecretNotFoundError, AppSecretInUseError)
)

type AppsHandler struct {
	OperationRegistry      OperationRegistry
	AppService             AppService
	AppRepo                AppRepository
	UserRepo               users.UserRepository
	SecretStorage          users.SecretAndCookieStorage
	Authorizer             Authorizer
	AuthHelper             u.AuthHelper
	AppDetector            AppDetector
	DatabaseConnector      tools.DatabaseConnector
	UserService            users.UserService
	VersionFileNameEncoder VersionFileNameEncoder
	VersionValidator       validation.VersionValidator
}

func (a *AppsHandler) AppListForAdminHandler(w http.ResponseWriter, r *http.Request) {
	appDtos, err := a.AppService.ListAppsForAdmin()
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}

	u.SendJsonResponse(w, appDtos)
}

func (a *AppsHandler) AppListForNonAdminHandler(w http.ResponseWriter, r *http.Request) {
	userId, role, err := a.UserService.GetUserIdAndRoleFromQuollixRequest(r)
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}

	appDtos, err := a.AppService.ListAppsForNonAdmin(userId, role)
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}

	u.SendJsonResponse(w, appDtos)
}

func (a *AppsHandler) SecretHandler(w http.ResponseWriter, r *http.Request) {
	u.Logger.Debug("SecretHandler called")
	request, ok := validation.ReadBody[api.AppAccessSecretRequest](w, r)
	if !ok {
		return
	}

	cookie, err := r.Cookie(api.BrandAppAuthCookieName)
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}

	userId, role, err := a.UserService.GetUserIdAndRoleFromQuollixRequest(r)
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}

	app, err := a.AppRepo.GetAppRequestData(request.AppName)
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}

	err = a.Authorizer.Authorize(app.AccessPolicy, role, userId, app.AppName)
	if err != nil {
		u.WriteResponseError(w, u.MapOf(AccessDeniedError), err)
		return
	}

	secret, err := a.SecretStorage.GenerateSecretForCookie(cookie.Value, app.AppName)
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}
	u.SendJsonResponse(w, secret)
}

func (a *AppsHandler) AppStartHandler(w http.ResponseWriter, r *http.Request) {
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
	appIdString, ok := validation.ReadBody[api.NumberString](w, r)
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

var allowedAccessPolicies = map[string]any{
	api.Policies.PublicAccessPolicy:          nil,
	api.Policies.AuthenticatedAccessPolicy:   nil,
	api.Policies.GroupRestrictedAccessPolicy: nil,
	api.Policies.AdminOnlyAccessPolicy:       nil,
}

func (a *AppsHandler) ChangeAccessPolicyHandler(w http.ResponseWriter, r *http.Request) {
	var req api.ChangeAccessPolicyRequest
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
	appIdString, ok := validation.ReadBody[api.NumberString](w, r)
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

func (a *AppsHandler) AppOperationInfoHandler(w http.ResponseWriter, r *http.Request) {
	operations := a.OperationRegistry.ListOperations()
	response := struct {
		Operations []string `json:"operations"`
		IsOngoing  bool     `json:"is_ongoing"`
	}{
		Operations: operations,
		IsOngoing:  len(operations) > 0,
	}
	u.SendJsonResponse(w, response)
}

func WriteConcurrentOperationError(w http.ResponseWriter, attemptedOperation string, err error) {
	u.Logger.Info(concurrentOperationErrorMessage, tools.AttemptedOperationField, attemptedOperation, "error", err)
	http.Error(w, concurrentOperationErrorMessage, http.StatusBadRequest)
}

func (a *AppsHandler) UpdateAutomaticMaintenanceSettingsHandler(w http.ResponseWriter, r *http.Request) {
	autoMaintenanceSettings, ok := validation.ReadBody[api.AutoMaintenanceSettingsResponse](w, r)
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
	appIdString, ok := validation.ReadBody[api.NumberString](w, r)
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

func (a *AppsHandler) RegenerateAppSecretHandler(w http.ResponseWriter, r *http.Request) {
	request, ok := validation.ReadBody[api.AppSecretRequest](w, r)
	if !ok {
		return
	}

	appId, err := strconv.Atoi(request.AppId)
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}

	err = a.AppService.RegenerateAppSecret(appId, request.Name)
	if err != nil {
		u.WriteResponseError(w, AppSecretEditExpectedErrors, err)
		return
	}
}

func (a *AppsHandler) UpdateAppSecretHandler(w http.ResponseWriter, r *http.Request) {
	request, ok := validation.ReadBody[api.AppSecretUpdateRequest](w, r)
	if !ok {
		return
	}

	appId, err := strconv.Atoi(request.AppId)
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}

	err = a.AppService.UpdateAppSecret(appId, request.Name, request.Value)
	if err != nil {
		u.WriteResponseError(w, AppSecretEditExpectedErrors, err)
		return
	}
}

func (a *AppsHandler) DeleteAppSecretHandler(w http.ResponseWriter, r *http.Request) {
	request, ok := validation.ReadBody[api.AppSecretRequest](w, r)
	if !ok {
		return
	}

	appId, err := strconv.Atoi(request.AppId)
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}

	err = a.AppService.DeleteUnusedAppSecret(appId, request.Name)
	if err != nil {
		u.WriteResponseError(w, AppSecretDeleteExpectedErrors, err)
		return
	}
}

func (a *AppsHandler) IsDatabaseAvailableHandler(w http.ResponseWriter, r *http.Request) {
	if err := a.DatabaseConnector.GetDB().Ping(); err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}
}
