package tools

type PathsType struct {
	FrontendInstalledApps    string
	FrontendIndex            string
	FrontendSignIn           string
	FrontendSettings         string
	FrontendUsers            string
	FrontendUserEdit         string
	FrontendStore            string
	FrontendVersions         string
	FrontendBackedUpApps     string
	FrontendListBackups      string
	FrontendSetPassword      string
	FrontendGroups           string
	FrontendGroupMembers     string
	FrontendGroupApps        string
	FrontendAccount          string
	FrontendLogs             string
	FrontendMetrics          string
	FrontendReports          string
	FrontendAppSso           string
	FrontendProviders        string
	FrontendClients          string
	FrontendCompliance       string
	FrontendMaintenance      string
	FrontendTerminalApps     string
	FrontendTerminalServices string
	FrontendTerminalView     string
	FrontendEmail            string

	BackendReloadFrontendTemplatesFromFileSystem string

	WebResourcesPath          string
	WebResourcesImages        string
	WebResourcesGroupPagePath string
	WebResourcesVendor        string
	WebResourcesVendorCss     string
	WebResourcesVendorJs      string
	WebResourcesVendorFonts   string

	BackendApi            string
	BackendSecret         string
	BackendSignIn         string
	BackendResetTestState string
	BackendCheckAuth      string

	BackendUsers                      string
	BackendUsersSignOut               string
	BackendUsersList                  string
	BackendUsersDelete                string
	BackendUsersSetOwnPassword        string
	BackendUsersChangeOwnPassword     string
	BackendUsersInviteUser            string
	BackendUsersInviteUserViaEmail    string
	BackendUsersSetPassword           string
	BackendUsersResetPassword         string
	BackendUsersResetPasswordViaEmail string
	BackendUsersChangeUsername        string
	BackendUsersChangeEmail           string
	BackendUsersSetEnabled            string

	BackendApps                            string
	BackendAppsList                        string
	BackendAppsDelete                      string
	BackendAppsUpdate                      string
	BackendAppsStart                       string
	BackendAppsStop                        string
	BackendAppsChangeAccessPolicy          string
	BackendAppOperationInfo                string
	BackendAppAutomaticMaintenanceSettings string
	BackendAppDownloadFromApplication      string
	BackendAppUploadToApplication          string

	BackendStore                    string
	BackendStoreSearch              string
	BackendStoreVersions            string
	BackendStoreVersionsInstall     string
	BackendStoreVersionsDownload    string
	BackendStoreVersionsList        string
	BackendStoreReloadPublishedApps string

	BackendBackups                  string
	BackendBackedUpAppsPage         string
	BackendBackupsPage              string
	BackendBackupsCreate            string
	BackendBackupsList              string
	BackendBackupsRestore           string
	BackendBackupsDelete            string
	BackendBackupsListApps          string
	BackendBackupsPurgeBackupServer string

	BackendSettings                               string
	BackendSettingsBaseDomain                     string
	BackendSettingsBaseDomainSave                 string
	BackendSettingsBaseDomainRead                 string
	BackendSettingsCertificate                    string
	BackendSettingsCertificateUpload              string
	BackendSettingsCertificateDownload            string
	BackendSettingsCertificateReset               string
	BackendSettingsCertificateOperationStatus     string
	BackendSettingsStartDns01CertificateChallenge string
	BackendSettingsSsh                            string
	BackendSettingsSshRead                        string
	BackendSettingsSshSave                        string
	BackendSettingsSshTestAccess                  string
	BackendSettingsGetSshKnownHosts               string
	BackendSettingsSshConfigsReset                string

	BackendOidcWellKnown string
	BackendOidcJwks      string
	BackendOidcAuthorize string
	BackendOidcToken     string
	BackendOidcUserinfo  string
	BackendOidcSignIn    string
	BackendOidcCallback  string

	BackendGroups                        string
	BackendGroupsCreate                  string
	BackendGroupsDelete                  string
	BackendGroupsAddUsers                string
	BackendGroupsRemoveUsers             string
	BackendGroupsGrantGroupAccessToApps  string
	BackendGroupsRevokeGroupAccessToApps string
	BackendGroupsListUsersByMembership   string
	BackendGroupsListAppsAccessByGroup   string
	BackendListAllGroups                 string

	BackendEmail                        string
	BackendEmailSaveConfig              string
	BackendEmailReadConfig              string
	BackendEmailTestConnection          string
	BackendEmailSendTestEmail           string
	BackendEmailResetConfig             string
	BackendEmailReadOidcEmailExposure   string
	BackendEmailSaveOidcEmailExposure   string
	BackendEmailReadInvitationTemplate  string
	BackendEmailSaveInvitationTemplate  string
	BackendEmailResetInvitationTemplate string

	BackendOidcClients                    string
	BackendAppsRegenerateOidcCredentials  string
	BackendOidcAuthProvidersCreate        string
	BackendOidcAuthProvidersUpdate        string
	BackendOidcAuthProvidersList          string
	BackendOidcAuthProvidersDelete        string
	BackendOidcAuthProvidersTestDiscovery string
	BackendOidcRelyingPartiesCreate       string
	BackendOidcRelyingPartiesUpdate       string
	BackendOidcRelyingPartiesList         string
	BackendOidcRelyingPartiesDelete       string
	BackendOidcRelyingPartiesRegenerate   string

	BackendTerminal string

	BackendMaintenance                      string
	BackendMaintenanceConfigsRead           string
	BackendMaintenanceConfigsSave           string
	BackendMaintenanceRetentionPolicyRead   string
	BackendMaintenanceRetentionPolicySave   string
	BackendMaintenanceTriggerMaintenanceJob string

	BackendIsDatabaseAvailable string
}

var Paths = func() PathsType {
	p := PathsType{}

	p.FrontendInstalledApps = "/installed-apps"
	p.FrontendIndex = "/"
	p.FrontendSignIn = "/sign-in"
	p.FrontendSettings = "/settings"
	p.FrontendUsers = "/users"
	p.FrontendUserEdit = "/edit-user"
	p.FrontendStore = "/store"
	p.FrontendVersions = "/versions"
	p.FrontendSetPassword = "/set-password"
	p.FrontendAccount = "/account"
	p.FrontendLogs = "/logs"
	p.FrontendMetrics = "/metrics"
	p.FrontendReports = "/reports"
	p.FrontendAppSso = "/sso"
	p.FrontendProviders = "/providers"
	p.FrontendClients = "/clients"
	p.FrontendCompliance = "/compliance"
	p.FrontendMaintenance = "/maintenance"

	p.FrontendBackedUpApps = "/backed-up-apps"
	p.FrontendListBackups = "/backups"

	p.FrontendGroups = "/groups"
	p.FrontendGroupMembers = "/group-members"
	p.FrontendGroupApps = "/group-apps"

	p.FrontendTerminalApps = "/terminal-apps"
	p.FrontendTerminalServices = "/terminal-services"
	p.FrontendTerminalView = "/terminal-view"
	p.FrontendEmail = "/email"

	p.WebResourcesPath = "/frontend/resources"

	p.WebResourcesGroupPagePath = p.WebResourcesPath + "/pages/groups"
	p.WebResourcesImages = p.WebResourcesPath + "/images"
	p.WebResourcesVendor = p.WebResourcesPath + "/vendor"
	p.WebResourcesVendorCss = p.WebResourcesVendor + "/css"
	p.WebResourcesVendorJs = p.WebResourcesVendor + "/js"
	p.WebResourcesVendorFonts = p.WebResourcesVendor + "/fonts"

	p.BackendApi = "/api"
	p.BackendReloadFrontendTemplatesFromFileSystem = p.BackendApi + "/reload-frontend-templates-from-file-system"
	p.BackendSecret = p.BackendApi + "/secret"
	p.BackendSignIn = p.BackendApi + "/sign-in"
	p.BackendResetTestState = p.BackendApi + "/reset-test-state"
	p.BackendCheckAuth = p.BackendApi + "/check-auth"

	p.BackendUsers = p.BackendApi + "/users"
	p.BackendUsersSignOut = p.BackendUsers + "/sign-out"
	p.BackendUsersList = p.BackendUsers + "/list"
	p.BackendUsersDelete = p.BackendUsers + "/delete"
	p.BackendUsersSetOwnPassword = p.BackendUsers + "/set-own-password"
	p.BackendUsersChangeOwnPassword = p.BackendUsers + "/change-own-password"
	p.BackendUsersInviteUser = p.BackendUsers + "/invite"
	p.BackendUsersInviteUserViaEmail = p.BackendUsers + "/invite-via-email"
	p.BackendUsersSetPassword = p.BackendUsers + "/set-password"
	p.BackendUsersResetPassword = p.BackendUsers + "/reset-password"
	p.BackendUsersResetPasswordViaEmail = p.BackendUsers + "/reset-password-via-email"
	p.BackendUsersChangeUsername = p.BackendUsers + "/change-username"
	p.BackendUsersChangeEmail = p.BackendUsers + "/change-email"
	p.BackendUsersSetEnabled = p.BackendUsers + "/set-enabled"

	p.BackendApps = p.BackendApi + "/apps"
	p.BackendAppsList = p.BackendApps + "/list"
	p.BackendAppsDelete = p.BackendApps + "/delete"
	p.BackendAppsUpdate = p.BackendApps + "/update"
	p.BackendAppsStart = p.BackendApps + "/start"
	p.BackendAppsStop = p.BackendApps + "/stop"
	p.BackendAppsChangeAccessPolicy = p.BackendApps + "/change-access-policy"
	p.BackendAppOperationInfo = p.BackendApps + "/operation-info"
	p.BackendAppAutomaticMaintenanceSettings = p.BackendApps + "/automatic-maintenance-settings"
	p.BackendAppDownloadFromApplication = p.BackendApps + "/download-from-application"
	p.BackendAppUploadToApplication = p.BackendApps + "/upload-to-application"

	p.BackendStore = p.BackendApi + "/store"
	p.BackendStoreSearch = p.BackendStore + "/search"
	p.BackendStoreVersions = p.BackendStore + "/versions"
	p.BackendStoreVersionsInstall = p.BackendStoreVersions + "/install"
	p.BackendStoreVersionsDownload = p.BackendStoreVersions + "/download"
	p.BackendStoreVersionsList = p.BackendStoreVersions + "/list"
	p.BackendStoreReloadPublishedApps = p.BackendStore + "/reload-local-store-apps"

	p.BackendBackups = p.BackendApi + "/backups"
	p.BackendBackedUpAppsPage = p.BackendApi + "/backed-up-apps-page"
	p.BackendBackupsPage = p.BackendApi + "/backups-page"
	p.BackendBackupsCreate = p.BackendBackups + "/create"
	p.BackendBackupsList = p.BackendBackups + "/list"
	p.BackendBackupsRestore = p.BackendBackups + "/restore"
	p.BackendBackupsDelete = p.BackendBackups + "/delete"
	p.BackendBackupsListApps = p.BackendBackups + "/list-apps_basic"
	p.BackendBackupsPurgeBackupServer = p.BackendBackups + "/purge-backup-server"

	p.BackendSettings = p.BackendApi + "/settings"
	p.BackendSettingsBaseDomain = p.BackendSettings + "/base-domain"
	p.BackendSettingsBaseDomainSave = p.BackendSettingsBaseDomain + "/save"
	p.BackendSettingsBaseDomainRead = p.BackendSettingsBaseDomain + "/read"

	p.BackendSettingsCertificate = p.BackendSettings + "/certificate"
	p.BackendSettingsCertificateUpload = p.BackendSettingsCertificate + "/upload"
	p.BackendSettingsCertificateDownload = p.BackendSettingsCertificate + "/download"
	p.BackendSettingsCertificateReset = p.BackendSettingsCertificate + "/reset"
	p.BackendSettingsCertificateOperationStatus = p.BackendSettingsCertificate + "/operation-status"
	p.BackendSettingsStartDns01CertificateChallenge = p.BackendSettingsCertificate + "/generate"

	p.BackendSettingsSsh = p.BackendSettings + "/ssh"
	p.BackendSettingsSshRead = p.BackendSettingsSsh + "/read"
	p.BackendSettingsSshSave = p.BackendSettingsSsh + "/save"
	p.BackendSettingsSshTestAccess = p.BackendSettingsSsh + "/test-access"
	p.BackendSettingsGetSshKnownHosts = p.BackendSettingsSsh + "/known-hosts"
	p.BackendSettingsSshConfigsReset = p.BackendSettings + "/reset-ssh-configs"

	p.BackendOidcWellKnown = "/.well-known/openid-configuration"
	p.BackendOidcJwks = p.BackendApi + "/jwks"
	p.BackendOidcAuthorize = p.BackendApi + "/authorize"
	p.BackendOidcToken = p.BackendApi + "/token"
	p.BackendOidcUserinfo = p.BackendApi + "/userinfo"
	p.BackendOidcSignIn = p.BackendApi + "/auth/oidc/sign-in"
	p.BackendOidcCallback = p.BackendApi + "/auth/oidc/callback"

	p.BackendGroups = p.BackendApi + "/groups"
	p.BackendGroupsCreate = p.BackendGroups + "/create"
	p.BackendGroupsDelete = p.BackendGroups + "/delete"
	p.BackendGroupsAddUsers = p.BackendGroups + "/add-users-to-group"
	p.BackendGroupsRemoveUsers = p.BackendGroups + "/remove-users-from-group"
	p.BackendGroupsGrantGroupAccessToApps = p.BackendGroups + "/grant-group-access-to-app"
	p.BackendGroupsRevokeGroupAccessToApps = p.BackendGroups + "/revoke-group-access-to-app"
	p.BackendGroupsListUsersByMembership = p.BackendGroups + "/list-users-by-membership"
	p.BackendGroupsListAppsAccessByGroup = p.BackendGroups + "/list-apps_basic-access-by-group"
	p.BackendListAllGroups = p.BackendGroups + "/list-all-groups"

	p.BackendEmail = p.BackendApi + "/email"
	p.BackendEmailSaveConfig = p.BackendEmail + "/save-configs"
	p.BackendEmailReadConfig = p.BackendEmail + "/read-configs"
	p.BackendEmailTestConnection = p.BackendEmail + "/test-connection"
	p.BackendEmailSendTestEmail = p.BackendEmail + "/send-test-email"
	p.BackendEmailResetConfig = p.BackendEmail + "/reset-configs"
	p.BackendEmailReadOidcEmailExposure = p.BackendEmail + "/read-oidc-email-exposure"
	p.BackendEmailSaveOidcEmailExposure = p.BackendEmail + "/save-oidc-email-exposure"
	p.BackendEmailReadInvitationTemplate = p.BackendEmail + "/read-invitation-template"
	p.BackendEmailSaveInvitationTemplate = p.BackendEmail + "/save-invitation-template"
	p.BackendEmailResetInvitationTemplate = p.BackendEmail + "/reset-invitation-template"

	p.BackendOidcClients = p.BackendApi + "/oidc-clients"
	p.BackendAppsRegenerateOidcCredentials = p.BackendOidcClients + "/regenerate-oidc-credentials"
	p.BackendOidcAuthProvidersCreate = p.BackendOidcClients + "/auth-providers/create"
	p.BackendOidcAuthProvidersUpdate = p.BackendOidcClients + "/auth-providers/update"
	p.BackendOidcAuthProvidersList = p.BackendOidcClients + "/auth-providers/list"
	p.BackendOidcAuthProvidersDelete = p.BackendOidcClients + "/auth-providers/delete"
	p.BackendOidcAuthProvidersTestDiscovery = p.BackendOidcClients + "/auth-providers/test-discovery"
	p.BackendOidcRelyingPartiesCreate = p.BackendOidcClients + "/relying-parties/create"
	p.BackendOidcRelyingPartiesUpdate = p.BackendOidcClients + "/relying-parties/update"
	p.BackendOidcRelyingPartiesList = p.BackendOidcClients + "/relying-parties/list"
	p.BackendOidcRelyingPartiesDelete = p.BackendOidcClients + "/relying-parties/delete"
	p.BackendOidcRelyingPartiesRegenerate = p.BackendOidcClients + "/relying-parties/regenerate"

	p.BackendTerminal = p.BackendApi + "/terminal"

	p.BackendMaintenance = p.BackendApi + "/maintenance"
	p.BackendMaintenanceConfigsRead = p.BackendMaintenance + "/read-configs"
	p.BackendMaintenanceConfigsSave = p.BackendMaintenance + "/save-configs"
	p.BackendMaintenanceRetentionPolicyRead = p.BackendMaintenance + "/read-retention-policy"
	p.BackendMaintenanceRetentionPolicySave = p.BackendMaintenance + "/save-retention-policy"
	p.BackendMaintenanceTriggerMaintenanceJob = p.BackendMaintenance + "/trigger-maintenance-job"

	p.BackendIsDatabaseAvailable = p.BackendApi + "/is-database-available"

	return p
}()

var Metadata = struct {
	Openness string
	Homepage string
}{
	Openness: "Openness",
	Homepage: "Homepage",
}
