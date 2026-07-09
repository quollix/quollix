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
	AutomaticUpdatesEnabled  bool      `yaml:"automatic_updates_enabled"`
	AutomaticBackupsEnabled  bool      `yaml:"automatic_backups_enabled"`
}
