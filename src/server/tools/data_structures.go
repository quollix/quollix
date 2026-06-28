package tools

import "time"

var (
	PreUpdateBackupDescription = "pre-update"
	ManualBackupDescription    = "manual"
	ScheduledBackupDescription = "scheduled"
)

type SshConnectionRequest struct {
	Host          string `json:"host" validate:"remote_host"`
	SshPort       string `json:"port" validate:"number"`
	SshUser       string `json:"user" validate:"default"`
	SshPassword   string `json:"password" validate:"loose"`
	SshKnownHosts string `json:"known_hosts" validate:"known_hosts"`
}

type BackupServerConfigs struct {
	IsEnabled          bool   `json:"is_enabled"`
	Host               string `json:"host" validate:"remote_host"`
	SshPort            string `json:"port" validate:"number"`
	SshUser            string `json:"user" validate:"default"`
	SshPassword        string `json:"password" validate:"loose"`
	SshKnownHosts      string `json:"known_hosts" validate:"known_hosts"`
	EncryptionPassword string `json:"encryption_password" validate:"password"`
}

func (r *BackupServerConfigs) ConvertToSshConnectionTestRequest() *SshConnectionRequest {
	return &SshConnectionRequest{
		Host:          r.Host,
		SshPort:       r.SshPort,
		SshUser:       r.SshUser,
		SshPassword:   r.SshPassword,
		SshKnownHosts: r.SshKnownHosts,
	}
}

type RestoredVersionInfo struct {
	Maintainer     string
	AppName        string
	VersionName    string
	VersionContent []byte
}

type BackupInfo struct {
	BackupId                string
	Maintainer              string
	AppName                 string
	VersionName             string
	Description             string
	ApplicationVersion      string
	BackupCreationTimestamp time.Time
}

type User struct {
	Id                             int
	Username                       string
	Email                          string
	HashedPassword                 string
	HashedCookieValue              string
	CookieExpirationDate           time.Time
	IsAdmin                        bool
	IsEnabled                      bool
	SetPasswordToken               string
	SetPasswordTokenExpirationDate time.Time
	CreationDate                   time.Time
}

func (u *User) IsPasswordSet() bool {
	return u.HashedPassword != ""
}

type MaintainerAndApp struct {
	Maintainer string `json:"maintainer" validate:"default"`
	AppName    string `json:"app_name" validate:"default"`
}

type SingleBool struct {
	Value bool `json:"value"`
}

type SingleString struct {
	Value string `json:"value"`
}

type BackupOperationRequest struct {
	BackupId string `json:"backup_id" validate:"restic_backup_id"`
}

type BackupsOperationRequest struct {
	BackupIds []string `json:"backup_ids" validate:"restic_backup_id"`
}

type NumberString struct {
	Value string `json:"value" validate:"number"`
}

type ChangeOwnPasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"password"`
	NewPassword     string `json:"new_password" validate:"password"`
}

type SetOwnPasswordRequest struct {
	NewPassword string `json:"new_password" validate:"password"`
}

type BaseDomainString struct {
	Value string `json:"value" validate:"host"`
}

type KnownHostsString struct {
	Value string `json:"value" validate:"known_hosts"`
}

type DefaultString struct {
	Value string `json:"value" validate:"default"`
}

var Policies = struct {
	PublicAccessPolicy          string
	AuthenticatedAccessPolicy   string
	GroupRestrictedAccessPolicy string
	AdminOnlyAccessPolicy       string
}{
	PublicAccessPolicy:          "public",
	AuthenticatedAccessPolicy:   "authenticated",
	GroupRestrictedAccessPolicy: "group_restricted",
	AdminOnlyAccessPolicy:       "admin_only",
}

type BinaryFile struct {
	FileName string `json:"file_name" validate:"file_name"`
	Content  []byte `json:"content"`
}
