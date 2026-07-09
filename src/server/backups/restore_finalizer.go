package backups

import (
	"server/apps_basic"
	"server/backup_server"
	"server/tools"
)

type RestoreFinalizer interface {
	FinalizeRestore(info *tools.BackupInfo, snapshot *AppSnapshot) (*tools.RestoredVersionInfo, error)
}

type RestoreFinalizerImpl struct {
	ResticService     backup_server.ResticService
	AppService        apps_basic.AppService
	AppRepo           apps_basic.AppRepository
	DatabaseConnector tools.DatabaseConnector
	AppDetector       apps_basic.AppDetector
}

func (f *RestoreFinalizerImpl) FinalizeRestore(info *tools.BackupInfo, snapshot *AppSnapshot) (*tools.RestoredVersionInfo, error) {
	restoredVersionInfo := &tools.RestoredVersionInfo{
		Maintainer:     info.Maintainer,
		AppName:        info.AppName,
		VersionName:    info.VersionName,
		VersionContent: snapshot.DockerComposeYamlContent,
	}

	app := apps_basic.NewRepoApp(
		info.Maintainer,
		info.AppName,
		info.VersionName,
		snapshot.Meta.AccessPolicy,
		snapshot.Meta.Port,
		snapshot.Meta.ClientId,
		snapshot.Meta.ClientSecret,
		snapshot.Meta.VersionCreationTimestamp,
		snapshot.DockerComposeYamlContent,
		true,
		snapshot.Meta.AutomaticUpdatesEnabled,
		snapshot.Meta.AutomaticBackupsEnabled,
	)

	if f.AppDetector.IsOfficialDatabaseApp(app.AppName) {
		if err := f.DatabaseConnector.StartDatabaseAndConnect(); err != nil {
			return nil, err
		}
		if err := f.AppService.UpsertAppInDatabase(app); err != nil {
			return nil, err
		}
		return restoredVersionInfo, nil
	}

	if err := f.AppService.UpsertAppInDatabase(app); err != nil {
		return nil, err
	}

	updatedApp, err := f.AppRepo.GetAppByName(info.AppName)
	if err != nil {
		return nil, err
	}

	if err = f.AppService.StartApp(updatedApp.AppId); err != nil {
		return nil, err
	}

	return restoredVersionInfo, nil
}
