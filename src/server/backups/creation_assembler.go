package backups

import (
	"server/apps_basic"
	"server/backup_server"
	"server/tools"
	"time"
)

type BackupCreationAssembler interface {
	BuildBackupCreation(app *apps_basic.RepoApp, description string) (BackupCreationDto, []string, *MetaData)
}

type BackupCreationAssemblerImpl struct{}

func (a *BackupCreationAssemblerImpl) BuildBackupCreation(app *apps_basic.RepoApp, description string) (BackupCreationDto, []string, *MetaData) {
	backupCreation := BackupCreationDto{
		Maintainer:               app.Maintainer,
		AppName:                  app.AppName,
		VersionName:              app.VersionName,
		VersionCreationTimestamp: app.VersionCreationTimestamp.Format(time.RFC3339),
		Description:              description,
		VersionContent:           app.VersionContent,
	}

	resticTags := []string{
		backup_server.MaintainerResticTag + "=" + backupCreation.Maintainer,
		backup_server.AppResticTag + "=" + backupCreation.AppName,
		backup_server.VersionResticTag + "=" + backupCreation.VersionName,
		backup_server.DescriptionResticTag + "=" + backupCreation.Description,
		backup_server.ApplicationVersionResticTag + "=" + tools.ApplicationVersion,
	}

	meta := NewMetaData(app.ClientId, app.ClientSecret, app.AccessPolicy, app.Port, app.VersionCreationTimestamp, app.AutomaticUpdatesEnabled, app.AutomaticBackupsEnabled)
	return backupCreation, resticTags, meta
}
