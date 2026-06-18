package setup

import (
	"server/app_store"
	"server/apps_advanced"
	"server/apps_basic"
	"server/backups"
	"server/certificates"
	"server/configs"
	"server/email"
	"server/frontend"
	"server/groups"
	"server/ingress"
	"server/maintenance"
	"server/oidc_provider"
	"server/terminal"

	"mime"
	"net/http"
	"server/backup_server"
	"server/tools"
	"server/users"

	"github.com/go-chi/chi/v5"
)

type HandlerRegisterer struct {
	SshHandler                *backup_server.SshHandler
	AppsHandler               *apps_basic.AppsHandler
	BackupsHandler            *backups.BackupHandler
	SettingsHandler           *configs.SettingsHandler
	AppManager                apps_basic.AppService
	MaintenanceHandler        maintenance.MaintenanceAgent
	SetupHandler              *ingress.SetupHandler
	AppStoreHandler           *app_store.AppStoreHandler
	UserHandler               *users.UserHandler
	RouteRegisterer           *users.RouteRegisterer
	CertificateHandler        *certificates.CertificateHandler
	Router                    chi.Router
	Config                    *tools.GlobalConfig
	TestStateReset            *TestStateResetHandler
	TemplateHandler           *frontend.TemplateHandlerImpl
	BackedUpAppsLoaderHandler *frontend.BackedUpAppsLoaderHandler
	BackupsPageLoaderHandler  *frontend.BackupsPageLoaderHandler
	OidcHandler               *oidc_provider.OidcHandler
	MaintenanceConfigsHandler *maintenance.MaintenanceConfigsHandler
	AppsAdvancedHandler       *apps_advanced.AppsAdvancedHandler
	GroupHandler              *groups.GroupHandler
	EmailHandler              *email.EmailHandler
	TerminalHandler           *terminal.TerminalHandlerImpl
}

func (h *HandlerRegisterer) RegisterApplicationHandlers() error {
	err := h.registerMimeTypes()
	if err != nil {
		return err
	}
	routes := h.aggregateRoutes()
	h.RouteRegisterer.RegisterRoutes(routes)

	return nil
}

func (h *HandlerRegisterer) aggregateRoutes() []users.Route {
	routes := make([]users.Route, 0)
	routes = append(routes, h.anonymousRoutes()...)
	routes = append(routes, h.authenticatedRoutes()...)
	routes = append(routes, h.adminRoutes()...)
	if h.Config.UseLocalAppStoreClient {
		routes = append(routes, h.adminDevelopmentRoutes()...)
	}
	routes = append(routes, h.adminEmailRoutes()...)
	routes = append(routes, h.adminGroupRoutes()...)
	routes = append(routes, h.adminTerminalRoutes()...)
	routes = append(routes, h.frontendRoutes()...)
	routes = append(routes, h.frontendEmailRoutes()...)
	routes = append(routes, h.frontendGroupRoutes()...)
	routes = append(routes, h.frontendTerminalRoutes()...)
	routes = append(routes, h.fallbackRoutes()...)

	if h.Config.OpenTestStateResetEndpoint {
		routes = append(routes,
			// When running the backend with the development profile, it is helpful to run a reset function after each test that simply deletes all artifacts, such as database entries or Docker containers.
			users.Route{Path: tools.Paths.BackendResetTestState, HandlerFunc: h.TestStateReset.ResetTestStateHandler, AccessLevel: tools.AnonymousLevel},
		)
	}
	return routes
}

func (h *HandlerRegisterer) anonymousRoutes() []users.Route {
	return []users.Route{
		{Path: tools.Paths.BackendSettingsHostRead, HandlerFunc: h.SettingsHandler.ReadHostHandler, AccessLevel: tools.AnonymousLevel},
		{Path: tools.Paths.BackendAppsList, HandlerFunc: h.AppsHandler.AppListHandler, AccessLevel: tools.AnonymousLevel},
		{Path: tools.Paths.BackendLogin, HandlerFunc: h.UserHandler.LoginHandler, AccessLevel: tools.AnonymousLevel},
		{Path: tools.Paths.BackendCheckAuth, HandlerFunc: h.UserHandler.CheckAuthHandler, AccessLevel: tools.AnonymousLevel},
		{Path: tools.Paths.BackendUsersSetPassword, HandlerFunc: h.UserHandler.AcceptNewPasswordViaTokenHandler, AccessLevel: tools.AnonymousLevel},
		{Path: tools.Paths.BackendOidcWellKnown, HandlerFunc: h.OidcHandler.HandleDiscovery, AccessLevel: tools.AnonymousLevel},
		{Path: tools.Paths.BackendOidcJwks, HandlerFunc: h.OidcHandler.HandleJWKS, AccessLevel: tools.AnonymousLevel},
		{Path: tools.Paths.BackendOidcAuthorize, HandlerFunc: h.OidcHandler.HandleAuthorize, AccessLevel: tools.AnonymousLevel},
		{Path: tools.Paths.BackendOidcToken, HandlerFunc: h.OidcHandler.HandleToken, AccessLevel: tools.AnonymousLevel},
		{Path: tools.Paths.BackendOidcUserinfo, HandlerFunc: h.OidcHandler.HandleUserinfo, AccessLevel: tools.AnonymousLevel},
		{Path: tools.Paths.BackendApi + "/*", HandlerFunc: EndpointDoesNotExistHandler, AccessLevel: tools.AnonymousLevel},
	}
}

func (h *HandlerRegisterer) authenticatedRoutes() []users.Route {
	return []users.Route{
		{Path: tools.Paths.BackendUsersLogout, HandlerFunc: h.UserHandler.LogoutHandler, AccessLevel: tools.UserLevel},
		{Path: tools.Paths.BackendUsersChangeOwnPassword, HandlerFunc: h.UserHandler.UserChangesOwnPasswordHandler, AccessLevel: tools.UserLevel},
		{Path: tools.Paths.BackendSecret, HandlerFunc: h.UserHandler.SecretHandler, AccessLevel: tools.UserLevel},
	}
}

func (h *HandlerRegisterer) adminRoutes() []users.Route {
	return []users.Route{
		{Path: tools.Paths.BackendAppsRegenerateOidcCredentials, HandlerFunc: h.AppsHandler.RegenerateOidcClientCredentials, AccessLevel: tools.AdminLevel},
		{Path: tools.Paths.BackendSettingsHostSave, HandlerFunc: h.SettingsHandler.SaveHostHandler, AccessLevel: tools.AdminLevel},
		{Path: tools.Paths.BackendMaintenanceConfigsRead, HandlerFunc: h.MaintenanceConfigsHandler.ReadMaintenanceConfigs, AccessLevel: tools.AdminLevel},
		{Path: tools.Paths.BackendMaintenanceConfigsSave, HandlerFunc: h.MaintenanceConfigsHandler.SaveMaintenanceConfigs, AccessLevel: tools.AdminLevel},
		{Path: tools.Paths.BackendMaintenanceRetentionPolicyRead, HandlerFunc: h.MaintenanceConfigsHandler.ReadRetentionPolicyHandler, AccessLevel: tools.AdminLevel},
		{Path: tools.Paths.BackendMaintenanceRetentionPolicySave, HandlerFunc: h.MaintenanceConfigsHandler.SaveRetentionPolicyHandler, AccessLevel: tools.AdminLevel},
		{Path: tools.Paths.BackendMaintenanceTriggerMaintenanceJob, HandlerFunc: h.MaintenanceConfigsHandler.RunMaintenanceJobHandler, AccessLevel: tools.AdminLevel},
		{Path: tools.Paths.BackendSettingsStartDns01CertificateChallenge, HandlerFunc: h.CertificateHandler.WildcardCertificateGenerationHandler, AccessLevel: tools.AdminLevel},
		{Path: tools.Paths.BackendSettingsCertificateUpload, HandlerFunc: h.CertificateHandler.CertificateUploadHandler, AccessLevel: tools.AdminLevel},
		{Path: tools.Paths.BackendSettingsCertificateDownload, HandlerFunc: h.CertificateHandler.CertificateDownloadHandler, AccessLevel: tools.AdminLevel},
		{Path: tools.Paths.BackendSettingsCertificateReset, HandlerFunc: h.CertificateHandler.ResetCertificateHandler, AccessLevel: tools.AdminLevel},
		{Path: tools.Paths.BackendSettingsCertificateOperationStatus, HandlerFunc: h.CertificateHandler.GetOperationMonitorStatus, AccessLevel: tools.AdminLevel},
		{Path: tools.Paths.BackendSettingsSshSave, HandlerFunc: h.SshHandler.SaveSshSettingsHandler, AccessLevel: tools.AdminLevel},
		{Path: tools.Paths.BackendSettingsSshRead, HandlerFunc: h.SshHandler.ReadSshSettingsHandler, AccessLevel: tools.AdminLevel},
		{Path: tools.Paths.BackendSettingsSshConfigsReset, HandlerFunc: h.SshHandler.ResetSshSettingsHandler, AccessLevel: tools.AdminLevel},
		{Path: tools.Paths.BackendSettingsSshTestAccess, HandlerFunc: h.SshHandler.TestSshAccessHandler, AccessLevel: tools.AdminLevel},
		{Path: tools.Paths.BackendSettingsGetSshKnownHosts, HandlerFunc: h.SshHandler.GetKnownHostsHandler, AccessLevel: tools.AdminLevel},
		{Path: tools.Paths.BackendBackupsPurgeBackupServer, HandlerFunc: h.SshHandler.PurgeBackupServerHandler, AccessLevel: tools.AdminLevel},
		{Path: tools.Paths.BackendAppsStart, HandlerFunc: h.AppsHandler.AppStartHandler, AccessLevel: tools.AdminLevel},
		{Path: tools.Paths.BackendAppsStop, HandlerFunc: h.AppsHandler.AppStopHandler, AccessLevel: tools.AdminLevel},
		{Path: tools.Paths.BackendAppsChangeAccessPolicy, HandlerFunc: h.AppsHandler.ChangeAccessPolicyHandler, AccessLevel: tools.AdminLevel},
		{Path: tools.Paths.BackendAppsDelete, HandlerFunc: h.AppsHandler.AppPruneHandler, AccessLevel: tools.AdminLevel},
		{Path: tools.Paths.BackendAppOperationInfo, HandlerFunc: h.AppsHandler.AppOperationInfoHandler, AccessLevel: tools.AdminLevel},
		{Path: tools.Paths.BackendIsDatabaseAvailable, HandlerFunc: h.AppsHandler.IsDatabaseAvailableHandler, AccessLevel: tools.AdminLevel},
		{Path: tools.Paths.BackendAppUploadToApplication, HandlerFunc: h.AppsAdvancedHandler.UploadVersionFileToApplicationHandler, AccessLevel: tools.AdminLevel},
		{Path: tools.Paths.BackendAppDownloadFromApplication, HandlerFunc: h.AppsAdvancedHandler.DownloadVersionFileFromApplicationHandler, AccessLevel: tools.AdminLevel},
		{Path: tools.Paths.BackendAppsUpdate, HandlerFunc: h.AppsAdvancedHandler.VersionUpdateHandler, AccessLevel: tools.AdminLevel},
		{Path: tools.Paths.BackendAppAutomaticMaintenanceSettings, HandlerFunc: h.AppsHandler.UpdateAutomaticMaintenanceSettingsHandler, AccessLevel: tools.AdminLevel},
		{Path: tools.Paths.BackendStoreSearch, HandlerFunc: h.AppStoreHandler.SearchAppsHandler, AccessLevel: tools.AdminLevel},
		{Path: tools.Paths.BackendStoreVersionsInstall, HandlerFunc: h.AppStoreHandler.VersionInstallationHandler, AccessLevel: tools.AdminLevel},
		{Path: tools.Paths.BackendStoreVersionsDownload, HandlerFunc: h.AppStoreHandler.VersionDownloadHandler, AccessLevel: tools.AdminLevel},
		{Path: tools.Paths.BackendStoreVersionsList, HandlerFunc: h.AppStoreHandler.GetVersionsHandler, AccessLevel: tools.AdminLevel},
		{Path: tools.Paths.BackendBackedUpAppsPage, HandlerFunc: h.BackedUpAppsLoaderHandler.Read, AccessLevel: tools.AdminLevel},
		{Path: tools.Paths.BackendBackupsPage, HandlerFunc: h.BackupsPageLoaderHandler.Read, AccessLevel: tools.AdminLevel},
		{Path: tools.Paths.BackendBackupsCreate, HandlerFunc: h.BackupsHandler.CreateBackupHandler, AccessLevel: tools.AdminLevel},
		{Path: tools.Paths.BackendBackupsList, HandlerFunc: h.BackupsHandler.ListBackupsHandler, AccessLevel: tools.AdminLevel},
		{Path: tools.Paths.BackendBackupsRestore, HandlerFunc: h.BackupsHandler.RestoreBackupHandler, AccessLevel: tools.AdminLevel},
		{Path: tools.Paths.BackendBackupsDelete, HandlerFunc: h.BackupsHandler.DeleteBackupHandler, AccessLevel: tools.AdminLevel},
		{Path: tools.Paths.BackendBackupsListApps, HandlerFunc: h.BackupsHandler.ListAppsOfBackupRepository, AccessLevel: tools.AdminLevel},
		{Path: tools.Paths.BackendUsersList, HandlerFunc: h.UserHandler.ListUsersHandler, AccessLevel: tools.AdminLevel},
		{Path: tools.Paths.BackendUsersDelete, HandlerFunc: h.UserHandler.DeleteUserHandler, AccessLevel: tools.AdminLevel},
		{Path: tools.Paths.BackendUsersInviteUser, HandlerFunc: h.UserHandler.InviteUserHandler, AccessLevel: tools.AdminLevel},
		{Path: tools.Paths.BackendUsersResetPassword, HandlerFunc: h.UserHandler.ResetPasswordAndCreateTokenHandler, AccessLevel: tools.AdminLevel},
		{Path: tools.Paths.BackendUsersChangeUsername, HandlerFunc: h.UserHandler.ChangeUsernameHandler, AccessLevel: tools.AdminLevel},
		{Path: tools.Paths.BackendUsersChangeEmail, HandlerFunc: h.UserHandler.ChangeEmailHandler, AccessLevel: tools.AdminLevel},
	}
}

func (h *HandlerRegisterer) adminDevelopmentRoutes() []users.Route {
	return []users.Route{
		{Path: tools.Paths.BackendStoreReloadPublishedApps, HandlerFunc: h.AppStoreHandler.ReloadLocalAppsHandler, AccessLevel: tools.AdminLevel},
	}
}

func (h *HandlerRegisterer) adminEmailRoutes() []users.Route {
	return []users.Route{
		{Path: tools.Paths.BackendEmailSaveConfig, HandlerFunc: h.EmailHandler.SaveEmailConfig, AccessLevel: tools.AdminLevel},
		{Path: tools.Paths.BackendEmailReadConfig, HandlerFunc: h.EmailHandler.ReadEmailConfig, AccessLevel: tools.AdminLevel},
		{Path: tools.Paths.BackendEmailTestConnection, HandlerFunc: h.EmailHandler.TestEmailServerConnection, AccessLevel: tools.AdminLevel},
		{Path: tools.Paths.BackendEmailSendTestEmail, HandlerFunc: h.EmailHandler.SendTestEmail, AccessLevel: tools.AdminLevel},
		{Path: tools.Paths.BackendEmailResetConfig, HandlerFunc: h.EmailHandler.ResetEmailConfig, AccessLevel: tools.AdminLevel},
		{Path: tools.Paths.BackendEmailReadInvitationTemplate, HandlerFunc: h.EmailHandler.ReadInvitationTemplate, AccessLevel: tools.AdminLevel},
		{Path: tools.Paths.BackendEmailSaveInvitationTemplate, HandlerFunc: h.EmailHandler.SaveInvitationTemplate, AccessLevel: tools.AdminLevel},
		{Path: tools.Paths.BackendEmailResetInvitationTemplate, HandlerFunc: h.EmailHandler.ResetInvitationTemplate, AccessLevel: tools.AdminLevel},
		{Path: tools.Paths.BackendUsersInviteUserViaEmail, HandlerFunc: h.EmailHandler.InviteUserViaEmailHandler, AccessLevel: tools.AdminLevel},
		{Path: tools.Paths.BackendUsersResetPasswordViaEmail, HandlerFunc: h.EmailHandler.SendPasswordResetEmailHandler, AccessLevel: tools.AdminLevel},
	}
}

func (h *HandlerRegisterer) adminGroupRoutes() []users.Route {
	return []users.Route{
		{Path: tools.Paths.BackendListAllGroups, HandlerFunc: h.GroupHandler.ListAllGroupsHandler, AccessLevel: tools.AdminLevel},
		{Path: tools.Paths.BackendGroupsCreate, HandlerFunc: h.GroupHandler.CreateGroupHandler, AccessLevel: tools.AdminLevel},
		{Path: tools.Paths.BackendGroupsDelete, HandlerFunc: h.GroupHandler.DeleteGroupHandler, AccessLevel: tools.AdminLevel},
		{Path: tools.Paths.BackendGroupsAddUsers, HandlerFunc: h.GroupHandler.AddUserToGroupHandler, AccessLevel: tools.AdminLevel},
		{Path: tools.Paths.BackendGroupsRemoveUsers, HandlerFunc: h.GroupHandler.RemoveUserFromGroupHandler, AccessLevel: tools.AdminLevel},
		{Path: tools.Paths.BackendGroupsGrantGroupAccessToApps, HandlerFunc: h.GroupHandler.GrantAppAccessHandler, AccessLevel: tools.AdminLevel},
		{Path: tools.Paths.BackendGroupsRevokeGroupAccessToApps, HandlerFunc: h.GroupHandler.RevokeAppAccessHandler, AccessLevel: tools.AdminLevel},
		{Path: tools.Paths.BackendGroupsListUsersByMembership, HandlerFunc: h.GroupHandler.ListUsersByGroupMembership, AccessLevel: tools.AdminLevel},
		{Path: tools.Paths.BackendGroupsListAppsAccessByGroup, HandlerFunc: h.GroupHandler.ListAppsAccessByGroup, AccessLevel: tools.AdminLevel},
	}
}

func (h *HandlerRegisterer) adminTerminalRoutes() []users.Route {
	return []users.Route{
		{Path: tools.Paths.BackendTerminal, HandlerFunc: h.TerminalHandler.ServeTerminalWebsocket, AccessLevel: tools.AdminLevel},
	}
}

func (h *HandlerRegisterer) frontendRoutes() []users.Route {
	return []users.Route{
		{Path: tools.FrontendResourcesPathWithSlash + "*", HandlerFunc: h.TemplateHandler.WebResourcesProviderHandler, AccessLevel: tools.AnonymousLevel},
		{Path: tools.Paths.FrontendInstalledApps, HandlerFunc: h.TemplateHandler.InstalledAppsHandler, AccessLevel: tools.AnonymousLevel},
		{Path: tools.Paths.FrontendIndex, HandlerFunc: h.TemplateHandler.InstalledAppsHandler, AccessLevel: tools.AnonymousLevel},
		{Path: tools.Paths.FrontendLogin, HandlerFunc: h.TemplateHandler.LoginHandler, AccessLevel: tools.AnonymousLevel},
		{Path: tools.Paths.FrontendAccount, HandlerFunc: h.TemplateHandler.AccountPageHandler, AccessLevel: tools.UserLevel},
		{Path: tools.Paths.FrontendSettings, HandlerFunc: h.TemplateHandler.SettingsHandler, AccessLevel: tools.AdminLevel},
		{Path: tools.Paths.FrontendUsers, HandlerFunc: h.TemplateHandler.UsersHandler, AccessLevel: tools.AdminLevel},
		{Path: tools.Paths.FrontendStore, HandlerFunc: h.TemplateHandler.StoreHandler, AccessLevel: tools.AdminLevel},
		{Path: tools.Paths.FrontendVersions, HandlerFunc: h.TemplateHandler.VersionsHandler, AccessLevel: tools.AdminLevel},
		{Path: tools.Paths.FrontendBackedUpApps, HandlerFunc: h.TemplateHandler.BackedUpAppsHandler, AccessLevel: tools.AdminLevel},
		{Path: tools.Paths.FrontendListBackups, HandlerFunc: h.TemplateHandler.ListBackupsHandler, AccessLevel: tools.AdminLevel},
		{Path: tools.Paths.FrontendSetPassword, HandlerFunc: h.TemplateHandler.SetPasswordHandler, AccessLevel: tools.AnonymousLevel},
		{Path: tools.Paths.FrontendOidcClients, HandlerFunc: h.TemplateHandler.OidcClientsHandler, AccessLevel: tools.AdminLevel},
		{Path: tools.Paths.FrontendMaintenance, HandlerFunc: h.TemplateHandler.MaintenancePageHandler, AccessLevel: tools.AdminLevel},
		{Path: tools.Paths.FrontendUserEdit, HandlerFunc: h.TemplateHandler.UserEditPageHandler, AccessLevel: tools.AdminLevel},
		{Path: tools.Paths.BackendReloadFrontendTemplatesFromFileSystem, HandlerFunc: h.TemplateHandler.ReloadFrontendTemplatesFromFileSystemHandler, AccessLevel: tools.AdminLevel},
	}
}

func (h *HandlerRegisterer) frontendEmailRoutes() []users.Route {
	return []users.Route{
		{Path: tools.Paths.FrontendEmail, HandlerFunc: h.TemplateHandler.EmailPageHandler, AccessLevel: tools.AdminLevel},
	}
}

func (h *HandlerRegisterer) frontendGroupRoutes() []users.Route {
	return []users.Route{
		{Path: tools.Paths.FrontendGroups, HandlerFunc: h.TemplateHandler.GroupsPageHandler, AccessLevel: tools.AdminLevel},
		{Path: tools.Paths.FrontendGroupMembers, HandlerFunc: h.TemplateHandler.GroupMembersPageHandler, AccessLevel: tools.AdminLevel},
		{Path: tools.Paths.FrontendGroupApps, HandlerFunc: h.TemplateHandler.GroupAppsPageHandler, AccessLevel: tools.AdminLevel},
	}
}

func (h *HandlerRegisterer) frontendTerminalRoutes() []users.Route {
	return []users.Route{
		{Path: tools.Paths.FrontendTerminalApps, HandlerFunc: h.TemplateHandler.TerminalAppsPageHandler, AccessLevel: tools.AdminLevel},
		{Path: tools.Paths.FrontendTerminalServices, HandlerFunc: h.TemplateHandler.TerminalServicesPageHandler, AccessLevel: tools.AdminLevel},
		{Path: tools.Paths.FrontendTerminalView, HandlerFunc: h.TemplateHandler.TerminalViewPageHandler, AccessLevel: tools.AdminLevel},
	}
}

func (h *HandlerRegisterer) fallbackRoutes() []users.Route {
	return []users.Route{
		{Path: "/*", HandlerFunc: h.TemplateHandler.PageNotFoundHandler, AccessLevel: tools.AnonymousLevel},
	}
}

func (h *HandlerRegisterer) registerMimeTypes() error {
	// Ensure consistent Content-Type for static assets across OS/env. Go's mime.TypeByExtension may rely on system mappings (e.g., Windows registry), which can be missing or incorrect (commonly .js -> text/plain), breaking JS/fonts in browsers.
	if err := mime.AddExtensionType(".js", "text/javascript"); err != nil {
		return err
	}
	if err := mime.AddExtensionType(".woff", "font/woff"); err != nil {
		return err
	}
	if err := mime.AddExtensionType(".woff2", "font/woff2"); err != nil {
		return err
	}
	return nil
}

func EndpointDoesNotExistHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "endpoint does not exist, please inform the developers about this problem", http.StatusNotFound)
}
