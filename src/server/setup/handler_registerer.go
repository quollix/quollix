package setup

import (
	"mime"
	"net/http"

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
	"server/oidc_client"
	"server/oidc_provider"
	"server/terminal"

	"server/backup_server"
	"server/tools"
	"server/users"

	"github.com/go-chi/chi/v5"
	"github.com/quollix/common/quollix/api"
	u "github.com/quollix/common/utils"
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
	OidcRelyingPartyHandler   *oidc_provider.OidcRelyingPartyHandler
	OidcClientHandler         *oidc_client.OidcClientHandler
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
	if h.Config.ExposeDevelopmentRoutes {
		// Development-only endpoints must not be exposed in production, even to admins.
		routes = append(routes, h.developmentRoutes()...)
	}
	routes = append(routes, h.adminEmailRoutes()...)
	routes = append(routes, h.adminGroupRoutes()...)
	routes = append(routes, h.adminTerminalRoutes()...)
	routes = append(routes, h.frontendRoutes()...)
	routes = append(routes, h.frontendEmailRoutes()...)
	routes = append(routes, h.frontendGroupRoutes()...)
	routes = append(routes, h.frontendTerminalRoutes()...)
	routes = append(routes, h.fallbackRoutes()...)

	return routes
}

func (h *HandlerRegisterer) anonymousRoutes() []users.Route {
	return []users.Route{
		{Path: api.Paths.BackendSettingsBaseDomainRead, HandlerFunc: h.SettingsHandler.ReadBaseDomainHandler, AccessLevel: tools.AnonymousLevel},
		{Path: api.Paths.BackendAppsList, HandlerFunc: h.AppsHandler.AppListHandler, AccessLevel: tools.AnonymousLevel},
		{Path: api.Paths.BackendSignIn, HandlerFunc: h.UserHandler.SignInHandler, AccessLevel: tools.AnonymousLevel},
		{Path: api.Paths.BackendCheckAuth, HandlerFunc: h.UserHandler.CheckAuthHandler, AccessLevel: tools.AnonymousLevel},
		{Path: api.Paths.BackendHealth, HandlerFunc: HealthHandler, AccessLevel: tools.AnonymousLevel},
		{Path: api.Paths.BackendUsersSetPassword, HandlerFunc: h.UserHandler.AcceptNewPasswordViaTokenHandler, AccessLevel: tools.AnonymousLevel},
		{Path: api.Paths.BackendOidcWellKnown, HandlerFunc: h.OidcHandler.HandleDiscovery, AccessLevel: tools.AnonymousLevel},
		{Path: api.Paths.BackendOidcJwks, HandlerFunc: h.OidcHandler.HandleJWKS, AccessLevel: tools.AnonymousLevel},
		{Path: api.Paths.BackendOidcAuthorize, HandlerFunc: h.OidcHandler.HandleAuthorize, AccessLevel: tools.AnonymousLevel},
		{Path: api.Paths.BackendOidcToken, HandlerFunc: h.OidcHandler.HandleToken, AccessLevel: tools.AnonymousLevel},
		{Path: api.Paths.BackendOidcUserinfo, HandlerFunc: h.OidcHandler.HandleUserinfo, AccessLevel: tools.AnonymousLevel},
		{Path: api.Paths.BackendOidcSignIn, HandlerFunc: h.OidcClientHandler.StartLogin, AccessLevel: tools.AnonymousLevel},
		{Path: api.Paths.BackendOidcCallback, HandlerFunc: h.OidcClientHandler.Callback, AccessLevel: tools.AnonymousLevel},
		{Path: api.Paths.BackendApi + "/*", HandlerFunc: EndpointDoesNotExistHandler, AccessLevel: tools.AnonymousLevel},
	}
}

func (h *HandlerRegisterer) authenticatedRoutes() []users.Route {
	return []users.Route{
		{Path: api.Paths.BackendUsersSignOut, HandlerFunc: h.UserHandler.SignOutHandler, AccessLevel: tools.UserLevel},
		{Path: api.Paths.BackendUsersSetOwnPassword, HandlerFunc: h.UserHandler.UserSetsOwnPasswordHandler, AccessLevel: tools.UserLevel},
		{Path: api.Paths.BackendUsersChangeOwnPassword, HandlerFunc: h.UserHandler.UserChangesOwnPasswordHandler, AccessLevel: tools.UserLevel},
		{Path: api.Paths.BackendSecret, HandlerFunc: h.UserHandler.SecretHandler, AccessLevel: tools.UserLevel},
	}
}

func (h *HandlerRegisterer) adminRoutes() []users.Route {
	routes := make([]users.Route, 0)
	routes = append(routes, h.adminAppRoutes()...)
	routes = append(routes, h.adminAdvancedAppRoutes()...)
	routes = append(routes, h.adminStoreRoutes()...)
	routes = append(routes, h.adminBackupRoutes()...)
	routes = append(routes, h.adminSettingsRoutes()...)
	routes = append(routes, h.adminMaintenanceRoutes()...)
	routes = append(routes, h.adminUserRoutes()...)
	routes = append(routes, h.adminOidcProviderRoutes()...)
	routes = append(routes, h.adminOidcClientRoutes()...)
	return routes
}

func (h *HandlerRegisterer) adminAppRoutes() []users.Route {
	return []users.Route{
		{Path: api.Paths.BackendAppsStart, HandlerFunc: h.AppsHandler.AppStartHandler, AccessLevel: tools.AdminLevel},
		{Path: api.Paths.BackendAppsStop, HandlerFunc: h.AppsHandler.AppStopHandler, AccessLevel: tools.AdminLevel},
		{Path: api.Paths.BackendAppsChangeAccessPolicy, HandlerFunc: h.AppsHandler.ChangeAccessPolicyHandler, AccessLevel: tools.AdminLevel},
		{Path: api.Paths.BackendAppsDelete, HandlerFunc: h.AppsHandler.AppPruneHandler, AccessLevel: tools.AdminLevel},
		{Path: api.Paths.BackendAppOperationInfo, HandlerFunc: h.AppsHandler.AppOperationInfoHandler, AccessLevel: tools.AdminLevel},
		{Path: api.Paths.BackendIsDatabaseAvailable, HandlerFunc: h.AppsHandler.IsDatabaseAvailableHandler, AccessLevel: tools.AdminLevel},
		{Path: api.Paths.BackendAppAutomaticMaintenanceSettings, HandlerFunc: h.AppsHandler.UpdateAutomaticMaintenanceSettingsHandler, AccessLevel: tools.AdminLevel},
		{Path: api.Paths.BackendAppsRegenerateOidcCredentials, HandlerFunc: h.AppsHandler.RegenerateOidcClientCredentials, AccessLevel: tools.AdminLevel},
	}
}

func (h *HandlerRegisterer) adminAdvancedAppRoutes() []users.Route {
	return []users.Route{
		{Path: api.Paths.BackendAppUploadToApplication, HandlerFunc: h.AppsAdvancedHandler.UploadVersionFileToApplicationHandler, AccessLevel: tools.AdminLevel},
		{Path: api.Paths.BackendAppDownloadFromApplication, HandlerFunc: h.AppsAdvancedHandler.DownloadVersionFileFromApplicationHandler, AccessLevel: tools.AdminLevel},
		{Path: api.Paths.BackendAppsUpdate, HandlerFunc: h.AppsAdvancedHandler.VersionUpdateHandler, AccessLevel: tools.AdminLevel},
	}
}

func (h *HandlerRegisterer) adminStoreRoutes() []users.Route {
	return []users.Route{
		{Path: api.Paths.BackendStoreSearch, HandlerFunc: h.AppStoreHandler.SearchAppsHandler, AccessLevel: tools.AdminLevel},
		{Path: api.Paths.BackendStoreVersionsInstall, HandlerFunc: h.AppStoreHandler.VersionInstallationHandler, AccessLevel: tools.AdminLevel},
		{Path: api.Paths.BackendStoreVersionsDownload, HandlerFunc: h.AppStoreHandler.VersionDownloadHandler, AccessLevel: tools.AdminLevel},
		{Path: api.Paths.BackendStoreVersionsList, HandlerFunc: h.AppStoreHandler.GetVersionsHandler, AccessLevel: tools.AdminLevel},
	}
}

func (h *HandlerRegisterer) adminBackupRoutes() []users.Route {
	return []users.Route{
		{Path: api.Paths.BackendBackedUpAppsPage, HandlerFunc: h.BackedUpAppsLoaderHandler.Read, AccessLevel: tools.AdminLevel},
		{Path: api.Paths.BackendBackupsPage, HandlerFunc: h.BackupsPageLoaderHandler.Read, AccessLevel: tools.AdminLevel},
		{Path: api.Paths.BackendBackupsCreate, HandlerFunc: h.BackupsHandler.CreateBackupHandler, AccessLevel: tools.AdminLevel},
		{Path: api.Paths.BackendBackupsList, HandlerFunc: h.BackupsHandler.ListBackupsHandler, AccessLevel: tools.AdminLevel},
		{Path: api.Paths.BackendBackupsRestore, HandlerFunc: h.BackupsHandler.RestoreBackupHandler, AccessLevel: tools.AdminLevel},
		{Path: api.Paths.BackendBackupsDelete, HandlerFunc: h.BackupsHandler.DeleteBackupHandler, AccessLevel: tools.AdminLevel},
		{Path: api.Paths.BackendBackupsListApps, HandlerFunc: h.BackupsHandler.ListAppsOfBackupRepository, AccessLevel: tools.AdminLevel},
		{Path: api.Paths.BackendBackupsPurgeBackupServer, HandlerFunc: h.SshHandler.PurgeBackupServerHandler, AccessLevel: tools.AdminLevel},
	}
}

func (h *HandlerRegisterer) adminSettingsRoutes() []users.Route {
	return []users.Route{
		{Path: api.Paths.BackendSettingsBaseDomainSave, HandlerFunc: h.SettingsHandler.SaveBaseDomainHandler, AccessLevel: tools.AdminLevel},
		{Path: api.Paths.BackendSettingsStartDns01CertificateChallenge, HandlerFunc: h.CertificateHandler.WildcardCertificateGenerationHandler, AccessLevel: tools.AdminLevel},
		{Path: api.Paths.BackendSettingsCertificateUpload, HandlerFunc: h.CertificateHandler.CertificateUploadHandler, AccessLevel: tools.AdminLevel},
		{Path: api.Paths.BackendSettingsCertificateDownload, HandlerFunc: h.CertificateHandler.CertificateDownloadHandler, AccessLevel: tools.AdminLevel},
		{Path: api.Paths.BackendSettingsCertificateReset, HandlerFunc: h.CertificateHandler.ResetCertificateHandler, AccessLevel: tools.AdminLevel},
		{Path: api.Paths.BackendSettingsCertificateOperationStatus, HandlerFunc: h.CertificateHandler.GetOperationMonitorStatus, AccessLevel: tools.AdminLevel},
		{Path: api.Paths.BackendSettingsAcmeAccountPrivateKeyUpload, HandlerFunc: h.CertificateHandler.AcmeAccountPrivateKeyUploadHandler, AccessLevel: tools.AdminLevel},
		{Path: api.Paths.BackendSettingsAcmeAccountPrivateKeyDownload, HandlerFunc: h.CertificateHandler.AcmeAccountPrivateKeyDownloadHandler, AccessLevel: tools.AdminLevel},
		{Path: api.Paths.BackendSettingsSshSave, HandlerFunc: h.SshHandler.SaveSshSettingsHandler, AccessLevel: tools.AdminLevel},
		{Path: api.Paths.BackendSettingsSshRead, HandlerFunc: h.SshHandler.ReadSshSettingsHandler, AccessLevel: tools.AdminLevel},
		{Path: api.Paths.BackendSettingsSshConfigsReset, HandlerFunc: h.SshHandler.ResetSshSettingsHandler, AccessLevel: tools.AdminLevel},
		{Path: api.Paths.BackendSettingsSshTestAccess, HandlerFunc: h.SshHandler.TestSshAccessHandler, AccessLevel: tools.AdminLevel},
		{Path: api.Paths.BackendSettingsGetSshKnownHosts, HandlerFunc: h.SshHandler.GetKnownHostsHandler, AccessLevel: tools.AdminLevel},
	}
}

func (h *HandlerRegisterer) adminMaintenanceRoutes() []users.Route {
	return []users.Route{
		{Path: api.Paths.BackendMaintenanceConfigsRead, HandlerFunc: h.MaintenanceConfigsHandler.ReadMaintenanceConfigs, AccessLevel: tools.AdminLevel},
		{Path: api.Paths.BackendMaintenanceConfigsSave, HandlerFunc: h.MaintenanceConfigsHandler.SaveMaintenanceConfigs, AccessLevel: tools.AdminLevel},
		{Path: api.Paths.BackendMaintenanceRetentionPolicyRead, HandlerFunc: h.MaintenanceConfigsHandler.ReadRetentionPolicyHandler, AccessLevel: tools.AdminLevel},
		{Path: api.Paths.BackendMaintenanceRetentionPolicySave, HandlerFunc: h.MaintenanceConfigsHandler.SaveRetentionPolicyHandler, AccessLevel: tools.AdminLevel},
		{Path: api.Paths.BackendMaintenanceTriggerMaintenanceJob, HandlerFunc: h.MaintenanceConfigsHandler.RunMaintenanceJobHandler, AccessLevel: tools.AdminLevel},
	}
}

func (h *HandlerRegisterer) adminUserRoutes() []users.Route {
	return []users.Route{
		{Path: api.Paths.BackendUsersList, HandlerFunc: h.UserHandler.ListUsersHandler, AccessLevel: tools.AdminLevel},
		{Path: api.Paths.BackendUsersDelete, HandlerFunc: h.UserHandler.DeleteUserHandler, AccessLevel: tools.AdminLevel},
		{Path: api.Paths.BackendUsersInviteUser, HandlerFunc: h.UserHandler.InviteUserHandler, AccessLevel: tools.AdminLevel},
		{Path: api.Paths.BackendUsersResetPassword, HandlerFunc: h.UserHandler.ResetPasswordAndCreateTokenHandler, AccessLevel: tools.AdminLevel},
		{Path: api.Paths.BackendUsersChangeUsername, HandlerFunc: h.UserHandler.ChangeUsernameHandler, AccessLevel: tools.AdminLevel},
		{Path: api.Paths.BackendUsersChangeEmail, HandlerFunc: h.UserHandler.ChangeEmailHandler, AccessLevel: tools.AdminLevel},
		{Path: api.Paths.BackendUsersSetEnabled, HandlerFunc: h.UserHandler.SetUserEnabledHandler, AccessLevel: tools.AdminLevel},
	}
}

func (h *HandlerRegisterer) adminOidcProviderRoutes() []users.Route {
	return []users.Route{
		{Path: api.Paths.BackendOidcAuthProvidersCreate, HandlerFunc: h.OidcClientHandler.CreateAuthProvider, AccessLevel: tools.AdminLevel},
		{Path: api.Paths.BackendOidcAuthProvidersUpdate, HandlerFunc: h.OidcClientHandler.UpdateAuthProvider, AccessLevel: tools.AdminLevel},
		{Path: api.Paths.BackendOidcAuthProvidersList, HandlerFunc: h.OidcClientHandler.ListAuthProviders, AccessLevel: tools.AdminLevel},
		{Path: api.Paths.BackendOidcAuthProvidersDelete, HandlerFunc: h.OidcClientHandler.DeleteAuthProvider, AccessLevel: tools.AdminLevel},
		{Path: api.Paths.BackendOidcAuthProvidersTestDiscovery, HandlerFunc: h.OidcClientHandler.TestAuthProviderDiscovery, AccessLevel: tools.AdminLevel},
	}
}

func (h *HandlerRegisterer) adminOidcClientRoutes() []users.Route {
	return []users.Route{
		{Path: api.Paths.BackendOidcRelyingPartiesCreate, HandlerFunc: h.OidcRelyingPartyHandler.CreateClient, AccessLevel: tools.AdminLevel},
		{Path: api.Paths.BackendOidcRelyingPartiesUpdate, HandlerFunc: h.OidcRelyingPartyHandler.UpdateClient, AccessLevel: tools.AdminLevel},
		{Path: api.Paths.BackendOidcRelyingPartiesList, HandlerFunc: h.OidcRelyingPartyHandler.ListClients, AccessLevel: tools.AdminLevel},
		{Path: api.Paths.BackendOidcRelyingPartiesDelete, HandlerFunc: h.OidcRelyingPartyHandler.DeleteClient, AccessLevel: tools.AdminLevel},
		{Path: api.Paths.BackendOidcRelyingPartiesRegenerate, HandlerFunc: h.OidcRelyingPartyHandler.RegenerateClientCredentials, AccessLevel: tools.AdminLevel},
	}
}

func (h *HandlerRegisterer) developmentRoutes() []users.Route {
	return []users.Route{
		{Path: api.Paths.BackendReloadFrontendTemplatesFromFileSystem, HandlerFunc: h.TemplateHandler.ReloadFrontendTemplatesFromFileSystemHandler, AccessLevel: tools.AnonymousLevel},
		{Path: api.Paths.BackendResetTestState, HandlerFunc: h.TestStateReset.ResetTestStateHandler, AccessLevel: tools.AnonymousLevel},
		{Path: api.Paths.BackendStoreReloadPublishedApps, HandlerFunc: h.AppStoreHandler.ReloadLocalAppsHandler, AccessLevel: tools.AdminLevel},
	}
}

func (h *HandlerRegisterer) adminEmailRoutes() []users.Route {
	return []users.Route{
		{Path: api.Paths.BackendEmailSaveConfig, HandlerFunc: h.EmailHandler.SaveEmailConfig, AccessLevel: tools.AdminLevel},
		{Path: api.Paths.BackendEmailReadConfig, HandlerFunc: h.EmailHandler.ReadEmailConfig, AccessLevel: tools.AdminLevel},
		{Path: api.Paths.BackendEmailTestConnection, HandlerFunc: h.EmailHandler.TestEmailServerConnection, AccessLevel: tools.AdminLevel},
		{Path: api.Paths.BackendEmailSendTestEmail, HandlerFunc: h.EmailHandler.SendTestEmail, AccessLevel: tools.AdminLevel},
		{Path: api.Paths.BackendEmailResetConfig, HandlerFunc: h.EmailHandler.ResetEmailConfig, AccessLevel: tools.AdminLevel},
		{Path: api.Paths.BackendEmailReadOidcEmailExposure, HandlerFunc: h.EmailHandler.ReadOidcEmailExposureConfig, AccessLevel: tools.AdminLevel},
		{Path: api.Paths.BackendEmailSaveOidcEmailExposure, HandlerFunc: h.EmailHandler.SaveOidcEmailExposureConfig, AccessLevel: tools.AdminLevel},
		{Path: api.Paths.BackendEmailReadInvitationTemplate, HandlerFunc: h.EmailHandler.ReadInvitationTemplate, AccessLevel: tools.AdminLevel},
		{Path: api.Paths.BackendEmailSaveInvitationTemplate, HandlerFunc: h.EmailHandler.SaveInvitationTemplate, AccessLevel: tools.AdminLevel},
		{Path: api.Paths.BackendEmailResetInvitationTemplate, HandlerFunc: h.EmailHandler.ResetInvitationTemplate, AccessLevel: tools.AdminLevel},
		{Path: api.Paths.BackendUsersInviteUserViaEmail, HandlerFunc: h.EmailHandler.InviteUserViaEmailHandler, AccessLevel: tools.AdminLevel},
		{Path: api.Paths.BackendUsersResetPasswordViaEmail, HandlerFunc: h.EmailHandler.SendPasswordResetEmailHandler, AccessLevel: tools.AdminLevel},
	}
}

func (h *HandlerRegisterer) adminGroupRoutes() []users.Route {
	return []users.Route{
		{Path: api.Paths.BackendListAllGroups, HandlerFunc: h.GroupHandler.ListAllGroupsHandler, AccessLevel: tools.AdminLevel},
		{Path: api.Paths.BackendGroupsCreate, HandlerFunc: h.GroupHandler.CreateGroupHandler, AccessLevel: tools.AdminLevel},
		{Path: api.Paths.BackendGroupsDelete, HandlerFunc: h.GroupHandler.DeleteGroupHandler, AccessLevel: tools.AdminLevel},
		{Path: api.Paths.BackendGroupsAddUsers, HandlerFunc: h.GroupHandler.AddUserToGroupHandler, AccessLevel: tools.AdminLevel},
		{Path: api.Paths.BackendGroupsRemoveUsers, HandlerFunc: h.GroupHandler.RemoveUserFromGroupHandler, AccessLevel: tools.AdminLevel},
		{Path: api.Paths.BackendGroupsGrantGroupAccessToApps, HandlerFunc: h.GroupHandler.GrantAppAccessHandler, AccessLevel: tools.AdminLevel},
		{Path: api.Paths.BackendGroupsRevokeGroupAccessToApps, HandlerFunc: h.GroupHandler.RevokeAppAccessHandler, AccessLevel: tools.AdminLevel},
		{Path: api.Paths.BackendGroupsListUsersByMembership, HandlerFunc: h.GroupHandler.ListUsersByGroupMembership, AccessLevel: tools.AdminLevel},
		{Path: api.Paths.BackendGroupsListAppsAccessByGroup, HandlerFunc: h.GroupHandler.ListAppsAccessByGroup, AccessLevel: tools.AdminLevel},
	}
}

func (h *HandlerRegisterer) adminTerminalRoutes() []users.Route {
	return []users.Route{
		{Path: api.Paths.BackendTerminal, HandlerFunc: h.TerminalHandler.ServeTerminalWebsocket, AccessLevel: tools.AdminLevel},
	}
}

func (h *HandlerRegisterer) frontendRoutes() []users.Route {
	return []users.Route{
		{Path: tools.FrontendResourcesPathWithSlash + "*", HandlerFunc: h.TemplateHandler.WebResourcesProviderHandler, AccessLevel: tools.AnonymousLevel},
		{Path: api.Paths.FrontendInstalledApps, HandlerFunc: h.TemplateHandler.InstalledAppsHandler, AccessLevel: tools.AnonymousLevel},
		{Path: api.Paths.FrontendIndex, HandlerFunc: h.TemplateHandler.InstalledAppsHandler, AccessLevel: tools.AnonymousLevel},
		{Path: api.Paths.FrontendSignIn, HandlerFunc: h.TemplateHandler.SignInHandler, AccessLevel: tools.AnonymousLevel},
		{Path: api.Paths.FrontendAppOpen, HandlerFunc: h.TemplateHandler.OpenInstalledAppHandler, AccessLevel: tools.UserLevel},
		{Path: api.Paths.FrontendAccount, HandlerFunc: h.TemplateHandler.AccountPageHandler, AccessLevel: tools.UserLevel},
		{Path: api.Paths.FrontendSettings, HandlerFunc: h.TemplateHandler.SettingsHandler, AccessLevel: tools.AdminLevel},
		{Path: api.Paths.FrontendUsers, HandlerFunc: h.TemplateHandler.UsersHandler, AccessLevel: tools.AdminLevel},
		{Path: api.Paths.FrontendStore, HandlerFunc: h.TemplateHandler.StoreHandler, AccessLevel: tools.AdminLevel},
		{Path: api.Paths.FrontendVersions, HandlerFunc: h.TemplateHandler.VersionsHandler, AccessLevel: tools.AdminLevel},
		{Path: api.Paths.FrontendBackedUpApps, HandlerFunc: h.TemplateHandler.BackedUpAppsHandler, AccessLevel: tools.AdminLevel},
		{Path: api.Paths.FrontendListBackups, HandlerFunc: h.TemplateHandler.ListBackupsHandler, AccessLevel: tools.AdminLevel},
		{Path: api.Paths.FrontendSetPassword, HandlerFunc: h.TemplateHandler.SetPasswordHandler, AccessLevel: tools.AnonymousLevel},
		{Path: api.Paths.FrontendAppSso, HandlerFunc: h.TemplateHandler.AppSsoHandler, AccessLevel: tools.AdminLevel},
		{Path: api.Paths.FrontendProviders, HandlerFunc: h.TemplateHandler.ProvidersHandler, AccessLevel: tools.AdminLevel},
		{Path: api.Paths.FrontendClients, HandlerFunc: h.TemplateHandler.OidcClientsHandler, AccessLevel: tools.AdminLevel},
		{Path: api.Paths.FrontendMaintenance, HandlerFunc: h.TemplateHandler.MaintenancePageHandler, AccessLevel: tools.AdminLevel},
		{Path: api.Paths.FrontendUserEdit, HandlerFunc: h.TemplateHandler.UserEditPageHandler, AccessLevel: tools.AdminLevel},
	}
}

func (h *HandlerRegisterer) frontendEmailRoutes() []users.Route {
	return []users.Route{
		{Path: api.Paths.FrontendEmail, HandlerFunc: h.TemplateHandler.EmailPageHandler, AccessLevel: tools.AdminLevel},
	}
}

func (h *HandlerRegisterer) frontendGroupRoutes() []users.Route {
	return []users.Route{
		{Path: api.Paths.FrontendGroups, HandlerFunc: h.TemplateHandler.GroupsPageHandler, AccessLevel: tools.AdminLevel},
		{Path: api.Paths.FrontendGroupMembers, HandlerFunc: h.TemplateHandler.GroupMembersPageHandler, AccessLevel: tools.AdminLevel},
		{Path: api.Paths.FrontendGroupApps, HandlerFunc: h.TemplateHandler.GroupAppsPageHandler, AccessLevel: tools.AdminLevel},
	}
}

func (h *HandlerRegisterer) frontendTerminalRoutes() []users.Route {
	return []users.Route{
		{Path: api.Paths.FrontendTerminalApps, HandlerFunc: h.TemplateHandler.TerminalAppsPageHandler, AccessLevel: tools.AdminLevel},
		{Path: api.Paths.FrontendTerminalServices, HandlerFunc: h.TemplateHandler.TerminalServicesPageHandler, AccessLevel: tools.AdminLevel},
		{Path: api.Paths.FrontendTerminalView, HandlerFunc: h.TemplateHandler.TerminalViewPageHandler, AccessLevel: tools.AdminLevel},
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

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	u.SendJsonResponse(w, map[string]string{"status": "ok"})
}
