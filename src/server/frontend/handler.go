package frontend

import (
	"bytes"
	"mime"
	"net/http"
	"path"
	"server/apps_basic"
	"server/frontend/assets"
	frontendpages "server/frontend/pages"
	"server/frontend/renderer"
	"server/tools"
	"server/users"
	"strconv"
	"time"

	u "github.com/quollix/common/utils"
	"github.com/quollix/common/validation"
)

type TemplateHandlerImpl struct {
	TemplateService           renderer.TemplateService
	PageRenderer              frontendpages.PageRenderer
	PageDataBuilder           FrontendPageDataBuilder
	AppsHandler               *apps_basic.AppsHandler
	AssetStore                assets.AssetStore
	BackedUpAppsLoaderHandler *BackedUpAppsLoaderHandler
}

func (t *TemplateHandlerImpl) LoginHandler(w http.ResponseWriter, r *http.Request) {
	t.PageRenderer.RenderPage(w, r, "login", "", nil)
}

func (t *TemplateHandlerImpl) SettingsHandler(w http.ResponseWriter, r *http.Request) {
	content, err := t.PageDataBuilder.BuildSettingsPage()
	if err != nil {
		t.pageCreationFailed(w, err)
		return
	}
	t.PageRenderer.RenderPage(w, r, "settings", tools.Links.UsageDocs.Settings, content)
}

func (t *TemplateHandlerImpl) InstalledAppsHandler(w http.ResponseWriter, r *http.Request) {
	usersId, role, err := t.AppsHandler.UserService.GetUserIdAndRoleFromQuollixRequest(r)
	if err != nil {
		t.pageCreationFailed(w, err)
		return
	}

	content, err := t.PageDataBuilder.BuildInstalledAppsPage(usersId, role)
	if err != nil {
		t.pageCreationFailed(w, err)
		return
	}

	t.PageRenderer.RenderPage(w, r, "installed-apps", tools.Links.UsageDocs.InstalledApps, content)
}

func (t *TemplateHandlerImpl) UsersHandler(w http.ResponseWriter, r *http.Request) {
	content, err := t.PageDataBuilder.BuildUsersPage()
	if err != nil {
		t.pageCreationFailed(w, err)
		return
	}
	t.PageRenderer.RenderPage(w, r, "users", tools.Links.UsageDocs.Users, content)
}

func (t *TemplateHandlerImpl) StoreHandler(w http.ResponseWriter, r *http.Request) {
	maintainerName := r.URL.Query().Get("maintainer_name")
	appName := r.URL.Query().Get("app_name")
	isSearch := r.URL.Query().Get("is_search") == "true"

	if err := validation.Validate("maintainer_name", validation.FieldSearchTerm, maintainerName); err != nil {
		t.pageCreationFailed(w, err)
		return
	}
	if err := validation.Validate("app_name", validation.FieldSearchTerm, appName); err != nil {
		t.pageCreationFailed(w, err)
		return
	}

	showUnofficialString := r.URL.Query().Get("show_unofficial")
	showUnofficial := showUnofficialString == "true"

	content, err := t.PageDataBuilder.BuildStorePage(maintainerName, appName, showUnofficial, isSearch)
	if err != nil {
		t.pageCreationFailed(w, err)
		return
	}

	t.PageRenderer.RenderPage(w, r, "app-store", tools.Links.UsageDocs.AppStore, content)
}

func (t *TemplateHandlerImpl) VersionsHandler(w http.ResponseWriter, r *http.Request) {
	maintainer := r.URL.Query().Get("maintainer")
	app := r.URL.Query().Get("app")

	if err := validation.Validate("maintainer", validation.FieldDefault, maintainer); err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}
	if err := validation.Validate("app", validation.FieldDefault, app); err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}

	content, err := t.PageDataBuilder.BuildVersionsPage(maintainer, app)
	if err != nil {
		t.pageCreationFailed(w, err)
		return
	}

	t.PageRenderer.RenderPage(w, r, "versions", tools.Links.UsageDocs.StoreVersions, content)
}

func (t *TemplateHandlerImpl) BackedUpAppsHandler(w http.ResponseWriter, r *http.Request) {
	// Backup pages intentionally render an empty SSR shell and load data with page-local JavaScript.
	//
	// Facts:
	// - Quollix runs inside Docker and normally keeps the browser request open until SSR has built the full page.
	// - Reading backup repository data starts helper containers that connect to the backup server.
	//
	// Issue:
	// - If SSR waits for the backup lookup, the browser connection stays open while Docker starts or stops helper containers.
	// - Those container lifecycle changes can refresh Docker networking and interrupt the open browser connection, which shows up as a network changed error.
	//
	// Solution:
	// - First return a stable page shell without backup data.
	// - Then start the backup lookup asynchronously and let the frontend poll a JSON endpoint until the data is ready.
	content, err := t.PageDataBuilder.BuildBackedUpAppsPage()
	if err != nil {
		t.pageCreationFailed(w, err)
		return
	}
	if content.IsBackupEnabled {
		t.BackedUpAppsLoaderHandler.StartLoading()
	}
	t.PageRenderer.RenderPage(w, r, "backed-up-apps", tools.Links.UsageDocs.Backups, content)
}

func (t *TemplateHandlerImpl) EmailPageHandler(w http.ResponseWriter, r *http.Request) {
	content, err := t.PageDataBuilder.BuildEmailPage()
	if err != nil {
		t.pageCreationFailed(w, err)
		return
	}

	t.PageRenderer.RenderPage(w, r, "email", tools.Links.UsageDocs.Email, content)
}

func (t *TemplateHandlerImpl) TerminalAppsPageHandler(w http.ResponseWriter, r *http.Request) {
	content, err := t.PageDataBuilder.BuildTerminalAppsPage()
	if err != nil {
		t.pageCreationFailed(w, err)
		return
	}

	t.PageRenderer.RenderPage(w, r, "terminal-apps", tools.Links.UsageDocs.Terminal, content)
}

func (t *TemplateHandlerImpl) TerminalServicesPageHandler(w http.ResponseWriter, r *http.Request) {
	selectedMaintainer := r.URL.Query().Get("maintainer")
	selectedAppName := r.URL.Query().Get("appName")

	if err := validation.Validate("selected_maintainer", validation.FieldDefaultOrEmpty, selectedMaintainer); err != nil {
		t.pageCreationFailed(w, err)
		return
	}
	if err := validation.Validate("selected_app_name", validation.FieldDefaultOrEmpty, selectedAppName); err != nil {
		t.pageCreationFailed(w, err)
		return
	}

	content, err := t.PageDataBuilder.BuildTerminalServicesPage(selectedMaintainer, selectedAppName)
	if err != nil {
		t.pageCreationFailed(w, err)
		return
	}

	t.PageRenderer.RenderPage(w, r, "terminal-services", tools.Links.UsageDocs.Terminal, content)
}

func (t *TemplateHandlerImpl) TerminalViewPageHandler(w http.ResponseWriter, r *http.Request) {
	selectedMaintainer := r.URL.Query().Get("maintainer")
	selectedAppName := r.URL.Query().Get("appName")
	selectedServiceName := r.URL.Query().Get("serviceName")

	if err := validation.Validate("selected_maintainer", validation.FieldDefaultOrEmpty, selectedMaintainer); err != nil {
		t.pageCreationFailed(w, err)
		return
	}
	if err := validation.Validate("selected_app_name", validation.FieldDefaultOrEmpty, selectedAppName); err != nil {
		t.pageCreationFailed(w, err)
		return
	}
	if err := validation.Validate("selected_service_name", validation.FieldDefaultOrEmpty, selectedServiceName); err != nil {
		t.pageCreationFailed(w, err)
		return
	}

	content, err := t.PageDataBuilder.BuildTerminalViewPage(selectedMaintainer, selectedAppName, selectedServiceName)
	if err != nil {
		t.pageCreationFailed(w, err)
		return
	}

	t.PageRenderer.RenderPage(w, r, "terminal-view", tools.Links.UsageDocs.Terminal, content)
}

func (t *TemplateHandlerImpl) GroupsPageHandler(w http.ResponseWriter, r *http.Request) {
	content, err := t.PageDataBuilder.BuildGroupsPage()
	if err != nil {
		t.pageCreationFailed(w, err)
		return
	}

	t.PageRenderer.RenderPage(w, r, "groups", tools.Links.UsageDocs.Groups, content)
}

func (t *TemplateHandlerImpl) GroupMembersPageHandler(w http.ResponseWriter, r *http.Request) {
	groupId, err := strconv.Atoi(r.URL.Query().Get("group-id"))
	if err != nil {
		t.pageCreationFailed(w, err)
		return
	}

	content, err := t.PageDataBuilder.BuildGroupMembersPage(groupId)
	if err != nil {
		t.pageCreationFailed(w, err)
		return
	}

	t.PageRenderer.RenderPage(w, r, "group-members", "", content)
}

func (t *TemplateHandlerImpl) GroupAppsPageHandler(w http.ResponseWriter, r *http.Request) {
	groupId, err := strconv.Atoi(r.URL.Query().Get("group-id"))
	if err != nil {
		t.pageCreationFailed(w, err)
		return
	}

	content, err := t.PageDataBuilder.BuildGroupAppsPage(groupId)
	if err != nil {
		t.pageCreationFailed(w, err)
		return
	}

	t.PageRenderer.RenderPage(w, r, "group-apps", "", content)
}

func (t *TemplateHandlerImpl) pageCreationFailed(w http.ResponseWriter, err error) {
	t.PageRenderer.PageCreationFailed(w, err)
}

func (t *TemplateHandlerImpl) ListBackupsHandler(w http.ResponseWriter, r *http.Request) {
	request := tools.MaintainerAndApp{
		Maintainer: r.URL.Query().Get("maintainer"),
		AppName:    r.URL.Query().Get("app"),
	}
	if err := validation.Validate("Maintainer", validation.FieldDefault, request.Maintainer); err != nil {
		t.pageCreationFailed(w, err)
		return
	}
	if err := validation.Validate("AppName", validation.FieldDefault, request.AppName); err != nil {
		t.pageCreationFailed(w, err)
		return
	}

	content, err := t.PageDataBuilder.BuildBackupsPage(request)
	if err != nil {
		t.pageCreationFailed(w, err)
		return
	}

	t.PageRenderer.RenderPage(w, r, "backups", "", content)
}

func (t *TemplateHandlerImpl) SetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")

	if err := validation.Validate("token", validation.FieldSecret, token); err != nil {
		t.pageCreationFailed(w, err)
		return
	}

	content, err := t.PageDataBuilder.BuildSetPasswordPage(token)
	if err != nil {
		t.pageCreationFailed(w, err)
		return
	}

	t.PageRenderer.RenderPage(w, r, "set-password", "", content)
}

func (t *TemplateHandlerImpl) OidcClientsHandler(w http.ResponseWriter, r *http.Request) {
	content, err := t.PageDataBuilder.BuildOidcClientsPage()
	if err != nil {
		t.pageCreationFailed(w, err)
		return
	}
	t.PageRenderer.RenderPage(w, r, "oidc-clients", tools.Links.UsageDocs.OidcClients, content)
}

func (t *TemplateHandlerImpl) ReloadFrontendTemplatesFromFileSystemHandler(w http.ResponseWriter, r *http.Request) {
	err := t.TemplateService.ReloadTemplateFromFileSystem()
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}
}

func (t *TemplateHandlerImpl) MaintenancePageHandler(w http.ResponseWriter, r *http.Request) {
	content, err := t.PageDataBuilder.BuildMaintenancePage()
	if err != nil {
		t.pageCreationFailed(w, err)
		return
	}

	t.PageRenderer.RenderPage(w, r, "maintenance", tools.Links.UsageDocs.Maintenance, content)
}

func (t *TemplateHandlerImpl) PageNotFoundHandler(w http.ResponseWriter, r *http.Request) {
	t.pageCreationFailed(w, u.Logger.NewError("page creation failed", "path", r.URL.Path))
}

func (t *TemplateHandlerImpl) WebResourcesProviderHandler(w http.ResponseWriter, r *http.Request) {
	fileServer := http.StripPrefix(tools.FrontendResourcesPathWithSlash, http.FileServer(http.FS(tools.FrontendResourceFilesystem)))

	if contentType := mime.TypeByExtension(path.Ext(r.URL.Path)); contentType != "" {
		w.Header().Set("Content-Type", contentType)
	}

	if contentBytes, ok := t.AssetStore.Get(r.URL.Path); ok {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		http.ServeContent(w, r, r.URL.Path, time.Time{}, bytes.NewReader(contentBytes))
		return
	}

	w.Header().Set("Cache-Control", "no-cache")
	fileServer.ServeHTTP(w, r)
}

func (t *TemplateHandlerImpl) AccountPageHandler(w http.ResponseWriter, r *http.Request) {
	user, err := users.GetAuthFromContext(r)
	if err != nil {
		t.pageCreationFailed(w, err)
		return
	}
	accountPageData := t.PageDataBuilder.BuildAccountPageData(user)
	t.PageRenderer.RenderPage(w, r, "account", "", accountPageData)
}

func (t *TemplateHandlerImpl) UserEditPageHandler(w http.ResponseWriter, r *http.Request) {
	userId := r.URL.Query().Get("user-id")
	if _, err := strconv.Atoi(userId); err != nil {
		t.pageCreationFailed(w, err)
		return
	}

	page, err := t.PageDataBuilder.BuildUserEditPageData(userId)
	if err != nil {
		t.pageCreationFailed(w, err)
		return
	}
	t.PageRenderer.RenderPage(w, r, "edit-user", "", page)
}
