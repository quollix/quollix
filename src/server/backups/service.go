package backups

import (
	"server/apps_basic"
	"server/backup_server"
	"server/tools"

	u "github.com/quollix/common/utils"
)

type BackupService interface {
	CreateBackup(appId int, description string) error
	DeleteBackups(backupId []string) error
	RestoreBackup(backupRestoreRequest tools.BackupOperationRequest) (*tools.RestoredVersionInfo, error)
	ResolveSingleAppNameByBackupIds(backupIds []string) (string, error)

	ListBackupsOfApp(request tools.MaintainerAndApp) ([]tools.BackupInfo, error)
	ListAppsInBackupRepo() ([]tools.MaintainerAndApp, error)
}

type BackupServiceImpl struct {
	ResticService     backup_server.ResticService
	SnapshotPackager  SnapshotPackager
	OsWrapper         u.OsWrapper
	AppService        apps_basic.AppService
	AppRepo           apps_basic.AppRepository
	DatabaseConnector tools.DatabaseConnector
	DockerService     tools.DockerService

	BackupCreationAssembler    BackupCreationAssembler
	BackupQueryService         BackupQueryService
	RestoreFinalizer           RestoreFinalizer
	SshRepositoryConfigService backup_server.SshRepositoryService
	AppDetector                apps_basic.AppDetector
	SshRepository              backup_server.SshRepository
	ComposeExtractor           apps_basic.ComposeExtractorImpl
	DatabaseIndependentRuntime apps_basic.DatabaseIndependentRuntime
}

func (b *BackupServiceImpl) CreateBackup(appId int, description string) error {
	if err := b.SshRepositoryConfigService.EnsureBackupIsEnabled(); err != nil {
		return err
	}

	app, err := b.AppRepo.GetAppById(appId)
	if err != nil {
		return err
	}

	spec, err := b.DatabaseIndependentRuntime.CollectAppSpec(appId)
	if err != nil {
		return err
	}

	backupCreation, resticTags, meta := b.BackupCreationAssembler.BuildBackupCreation(app, description)

	repo, err := b.SshRepository.GetRemoteBackupRepository()
	if err != nil {
		return err
	}

	if err = b.DatabaseIndependentRuntime.StopApp(spec.App); err != nil {
		return err
	}

	prepared, err := b.SnapshotPackager.PrepareBackupFiles(backupCreation, meta)
	if err != nil {
		return err
	}
	defer b.removeDir(prepared.TempDir)

	if err = b.ResticService.CreateBackup(prepared.MountArgsForRestic, prepared.Volumes, resticTags, repo.EncryptionPassword); err != nil {
		return err
	}

	if b.AppDetector.IsOfficialDatabaseApp(app.AppName) {
		if err = b.DatabaseIndependentRuntime.StartApp(spec); err != nil {
			return err
		}
		return b.DatabaseConnector.Connect()
	} else if app.ShouldBeRunning {
		return b.DatabaseIndependentRuntime.StartApp(spec)
	}

	return nil
}

func (b *BackupServiceImpl) ListBackupsOfApp(request tools.MaintainerAndApp) ([]tools.BackupInfo, error) {
	if err := b.SshRepositoryConfigService.EnsureBackupIsEnabled(); err != nil {
		return nil, err
	}

	allBackups, err := b.ResticService.ListBackups()
	if err != nil {
		return nil, err
	}

	return b.BackupQueryService.FilterBackupsOfApp(allBackups, request), nil
}

func (b *BackupServiceImpl) DeleteBackups(backupId []string) error {
	if err := b.SshRepositoryConfigService.EnsureBackupIsEnabled(); err != nil {
		return err
	}
	return b.ResticService.DeleteBackup(backupId)
}

func (b *BackupServiceImpl) ResolveSingleAppNameByBackupIds(backupIds []string) (string, error) {
	allBackups, err := b.ResticService.ListBackups()
	if err != nil {
		return "", err
	}

	appNames := map[string]struct{}{}
	for _, backupId := range backupIds {
		found := false
		for _, backup := range allBackups {
			if backup.BackupId != backupId {
				continue
			}
			appNames[backup.AppName] = struct{}{}
			found = true
			break
		}
		if !found {
			return "", u.Logger.NewError("backup not found", "backup_id", backupId)
		}
	}

	if len(appNames) != 1 {
		return "", u.Logger.NewError("backups must belong to exactly one app")
	}

	for appName := range appNames {
		return appName, nil
	}
	return "", u.Logger.NewError("backups must belong to exactly one app")
}

func (b *BackupServiceImpl) RestoreBackup(request tools.BackupOperationRequest) (*tools.RestoredVersionInfo, error) {
	if err := b.SshRepositoryConfigService.EnsureBackupIsEnabled(); err != nil {
		return nil, err
	}
	repo, err := b.SshRepository.GetRemoteBackupRepository()
	if err != nil {
		return nil, err
	}

	tempDir, err := b.OsWrapper.GetTempDir()
	if err != nil {
		return nil, err
	}
	defer b.removeDir(tempDir)

	if err = b.ResticService.RestoreFiles(request.BackupId, tempDir, repo.EncryptionPassword); err != nil {
		return nil, err
	}

	snapshot, err := b.SnapshotPackager.ReadSnapshotFromDir(tempDir)
	if err != nil {
		return nil, err
	}

	backupInfo, err := b.ResticService.GetSnapshotInfo(request.BackupId, repo.EncryptionPassword)
	if err != nil {
		return nil, err
	}

	volumes, containerNames, err := b.ComposeExtractor.Extract(snapshot.DockerComposeYamlContent)
	if err != nil {
		return nil, err
	}

	b.DockerService.StopAppContainers(containerNames)
	b.DockerService.RemoveVolumes(volumes)

	if err = b.ResticService.RestoreVolumes(request.BackupId, snapshot.Volumes, repo.EncryptionPassword); err != nil {
		return nil, err
	}

	return b.RestoreFinalizer.FinalizeRestore(backupInfo, snapshot)
}

func (b *BackupServiceImpl) ListAppsInBackupRepo() ([]tools.MaintainerAndApp, error) {
	if err := b.SshRepositoryConfigService.EnsureBackupIsEnabled(); err != nil {
		return nil, err
	}
	allBackups, err := b.ResticService.ListBackups()
	if err != nil {
		return nil, err
	}
	return b.BackupQueryService.UniqueMaintainerAndAppPairs(allBackups), nil
}

func (b *BackupServiceImpl) removeDir(path string) {
	if err := b.OsWrapper.RemoveAll(path); err != nil {
		u.Logger.Error(err, "path", path)
	}
}
