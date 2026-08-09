package backups

import "time"

type BackupCreationDto struct {
	Maintainer               string
	AppName                  string
	VersionName              string
	VersionCreationTimestamp string
	Description              string
	VersionContent           []byte
}

type MetaData struct {
	AccessPolicy             string    `yaml:"access_policy"`
	Port                     string    `yaml:"port"`
	VersionCreationTimestamp time.Time `yaml:"version_creation_timestamp"`
	ClientId                 string    `yaml:"client_id"`
	ClientSecret             string    `yaml:"client_secret"`
	// AppSecret is deprecated legacy APP_SECRET backup compatibility. New backups should persist purpose-specific SECRET_* values.
	AppSecret               string            `yaml:"app_secret"`
	AutomaticUpdatesEnabled bool              `yaml:"automatic_updates_enabled"`
	AutomaticBackupsEnabled bool              `yaml:"automatic_backups_enabled"`
	Secrets                 map[string]string `yaml:"secrets,omitempty"`
}
