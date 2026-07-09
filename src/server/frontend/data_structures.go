package frontend

import (
	"server/apps_basic"
	"server/configs"
	"server/maintenance/retention"
	"server/oidc_client"
	"server/oidc_provider"
	"server/tools"

	"github.com/quollix/common/store"
	u "github.com/quollix/common/utils"
)

type AppSsoPageContent struct {
	Apps []apps_basic.AppDto
}

type ProvidersPageContent struct {
	AuthProviders []oidc_client.OidcAuthProviderDto
}

type OidcClientsPageContent struct {
	Clients []oidc_provider.OidcRelyingPartyDto
}

type SignInPageContent struct {
	OidcAuthProviders []SignInOidcProviderDto
}

type SignInOidcProviderDto struct {
	Id   int
	Name string
}

type EmailPageContent struct {
	EmailConfig                *u.EmailConfig
	ExposeRealEmailInOidcToken bool
	InvitationEmailTemplate    string
}

type TerminalAppsPageContent struct {
	Apps []apps_basic.AppDto
}

type TerminalServicesPageContent struct {
	Maintainer   string
	AppName      string
	ServiceNames []string
}

type TerminalViewPageContent struct {
	Maintainer  string
	AppName     string
	ServiceName string
}

type GroupDTO struct {
	Id   string
	Name string
}

type GroupsPageContent struct {
	Groups []GroupDTO
}

type MemberDto struct {
	Id   string
	Name string
}

type GroupMembersPageContent struct {
	In        []MemberDto
	NotIn     []MemberDto
	GroupId   string
	GroupName string
}

type GroupAppsPageContent struct {
	AccessGrantedApps    []string
	AccessNotGrantedApps []string
	GroupId              string
	GroupName            string
}

type SetPasswordPageContent struct {
	Username string
}

type BackupsPageContent struct {
	Maintainer string
	AppName    string
	IsLoading  bool
	Backups    []BackupsDto
}

type BackedUpAppsPageContent struct {
	IsBackupEnabled bool
	IsLoading       bool
	Apps            []tools.MaintainerAndApp
}

type BackedUpAppsPageLoadResponse struct {
	IsRunning bool                     `json:"is_running"`
	Apps      []tools.MaintainerAndApp `json:"apps"`
}

type BackupsPageLoadResponse struct {
	IsRunning bool         `json:"is_running"`
	Backups   []BackupsDto `json:"backups"`
}

type BackupsDto struct {
	BackupId                      string `json:"backup_id"`
	VersionName                   string `json:"version_name"`
	Description                   string `json:"description"`
	BackupCreationDate            string `json:"backup_creation_date"`
	CreatedWithApplicationVersion string `json:"created_with_application_version"`
}

type VersionsPageContent struct {
	Maintainer string
	App        string
	Versions   []store.LeanVersionDto
}

type StorePageContent struct {
	MaintainerSearchTerm string
	AppSearchTerm        string
	ShowUnofficialApps   bool
	ShowUnofficialToggle bool
	Apps                 []StoreAppDto
}

type StoreAppDto struct {
	Maintainer                     string
	AppName                        string
	LatestVersionName              string
	LatestVersionCreationTimestamp string
}

type UsersPageContent struct {
	Users          []UserFrontendDto
	IsEmailEnabled bool
}

type UserFrontendDto struct {
	Id                             int
	Username                       string
	Email                          string
	IsAdmin                        bool
	IsEnabled                      bool
	SetPasswordLink                string
	SetPasswordTokenExpirationDate string
	CreatedAt                      string
}

type AppsPageContent struct {
	Apps            []apps_basic.AppDto
	IsBackupEnabled bool
}

type MaintenanceWindowOption struct {
	Value int
	Label string
}

type SettingsPageContent struct {
	BackupServer             *tools.BackupServerConfigs
	MaintenanceConfig        *configs.MaintenanceConfig
	RetentionPolicy          *retention.RetentionPolicy
	MaintenanceWindowOptions []MaintenanceWindowOption
	IanaTimezoneOptions      []string
	NextMaintenanceAt        string
}

type UserEditPage struct {
	UserId string
	User   *tools.User
}

type MaintenancePage struct {
	Apps []apps_basic.AppDto
}

type AccountPageData struct {
	Username      string
	Email         string
	Role          string
	IsPasswordSet bool
}
