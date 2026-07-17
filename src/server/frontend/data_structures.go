package frontend

import (
	api "github.com/quollix/common/quollix/api"
	"github.com/quollix/common/store"
	u "github.com/quollix/common/utils"
)

type AppSsoPageContent struct {
	Apps []api.AppDto
}

type ProvidersPageContent struct {
	AuthProviders []api.OidcAuthProviderDto
}

type OidcClientsPageContent struct {
	Clients []api.OidcRelyingPartyDto
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
	Apps []api.AppDto
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
	Apps            []api.MaintainerAndApp
}

type BackedUpAppsPageLoadResponse struct {
	IsRunning bool                   `json:"is_running"`
	Apps      []api.MaintainerAndApp `json:"apps"`
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
	IsInstalled                    bool
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
	Apps            []api.AppDto
	IsBackupEnabled bool
}

type MaintenanceWindowOption struct {
	Value int
	Label string
}

type SettingsPageContent struct {
	BackupServer             *api.BackupServerConfigs
	MaintenanceConfig        *api.MaintenanceConfig
	RetentionPolicy          *api.RetentionPolicy
	MaintenanceWindowOptions []MaintenanceWindowOption
	IanaTimezoneOptions      []string
	NextMaintenanceAt        string
}

type UserEditPage struct {
	UserId string
	User   *api.User
}

type MaintenancePage struct {
	Apps []api.AppDto
}

type AccountPageData struct {
	Username      string
	Email         string
	Role          string
	IsPasswordSet bool
}
