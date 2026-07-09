package maintenance

import (
	"server/apps_advanced"
	"server/apps_basic"
	"server/backup_server"
	"server/backups"
	"server/maintenance/retention"
	"server/tools"
	"testing"
	"time"

	"github.com/quollix/common/assert"
)

func TestCalculateNextMaintenanceAtUtc_HappyPath(t *testing.T) {
	mockRandomProvider := NewRandomProviderMock(t)
	maintenanceServiceHelper := &AgentHelperImpl{
		RandomProvider: mockRandomProvider,
	}
	defer mockRandomProvider.AssertExpectations(t)

	mockRandomProvider.EXPECT().Intn(50).Return(15)
	mockRandomProvider.EXPECT().Intn(60).Return(10)

	nowUtc := time.Date(2026, 2, 5, 12, 0, 0, 0, time.UTC)

	nextMaintenanceAtUtc, err := maintenanceServiceHelper.CalculateNextMaintenanceAtUtc(nowUtc, "UTC", 2)
	assert.Nil(t, err)

	expected := time.Date(2026, 2, 6, 2, 15, 10, 0, time.UTC)
	assert.Equal(t, expected, *nextMaintenanceAtUtc)
	assert.Equal(t, time.UTC, nextMaintenanceAtUtc.Location())
}

type agentHelperTestObjects struct {
	Helper               *AgentHelperImpl
	AppRepo              *apps_basic.AppRepositoryMock
	AppDetector          *apps_basic.AppDetectorMock
	BackupService        *backups.BackupServiceMock
	SshRepositoryService *backup_server.SshRepositoryServiceMock
	BackupDeletionFinder *retention.BackupDeletionFinderMock
	AppsServiceAdvanced  *apps_advanced.AppsServiceAdvancedMock
	OperationRegistry    *apps_basic.OperationRegistryImpl
}

func newAgentHelperTestObjects(t *testing.T) agentHelperTestObjects {
	appRepo := apps_basic.NewAppRepositoryMock(t)
	appDetector := apps_basic.NewAppDetectorMock(t)
	backupService := backups.NewBackupServiceMock(t)
	sshRepositoryService := backup_server.NewSshRepositoryServiceMock(t)
	backupDeletionFinder := retention.NewBackupDeletionFinderMock(t)
	appUpdaterService := apps_advanced.NewAppsServiceAdvancedMock(t)
	operationRegistry := &apps_basic.OperationRegistryImpl{}

	helper := &AgentHelperImpl{
		AppRepo:              appRepo,
		AppDetector:          appDetector,
		BackupService:        backupService,
		SshRepositoryService: sshRepositoryService,
		BackupDeletionFinder: backupDeletionFinder,
		AppsServiceAdvanced:  appUpdaterService,
		OperationRegistry:    operationRegistry,
	}

	return agentHelperTestObjects{
		Helper:               helper,
		AppRepo:              appRepo,
		AppDetector:          appDetector,
		BackupService:        backupService,
		SshRepositoryService: sshRepositoryService,
		BackupDeletionFinder: backupDeletionFinder,
		AppsServiceAdvanced:  appUpdaterService,
		OperationRegistry:    operationRegistry,
	}
}

func assertAgentHelperExpectations(t *testing.T, testObjects agentHelperTestObjects) {
	testObjects.AppRepo.AssertExpectations(t)
	testObjects.AppDetector.AssertExpectations(t)
	testObjects.BackupService.AssertExpectations(t)
	testObjects.SshRepositoryService.AssertExpectations(t)
	testObjects.BackupDeletionFinder.AssertExpectations(t)
	testObjects.AppsServiceAdvanced.AssertExpectations(t)
}

func TestUpdateAllApps(t *testing.T) {
	testObjects := newAgentHelperTestObjects(t)
	defer assertAgentHelperExpectations(t, testObjects)

	systemApp := apps_basic.RepoApp{AppId: 1, AppName: "system-app", AutomaticUpdatesEnabled: true}
	updatesDisabledApp := apps_basic.RepoApp{AppId: 2, AppName: "app-no-updates", AutomaticUpdatesEnabled: false}
	updatesEnabledApp := apps_basic.RepoApp{AppId: 3, AppName: "app-updates", AutomaticUpdatesEnabled: true}

	testObjects.AppRepo.EXPECT().ListApps().Return([]apps_basic.RepoApp{systemApp, updatesDisabledApp, updatesEnabledApp}, nil)
	testObjects.AppDetector.EXPECT().IsSystemApp(systemApp.AppName).Return(true)
	testObjects.AppDetector.EXPECT().IsSystemApp(updatesDisabledApp.AppName).Return(false)
	testObjects.AppDetector.EXPECT().IsSystemApp(updatesEnabledApp.AppName).Return(false)

	testObjects.AppsServiceAdvanced.EXPECT().UpdateAppViaAppStore(updatesEnabledApp.AppId).Return(nil)

	testObjects.Helper.UpdateAllApps()
}

func TestCreateBackups_BackupsEnabled(t *testing.T) {
	testObjects := newAgentHelperTestObjects(t)
	defer assertAgentHelperExpectations(t, testObjects)

	testObjects.SshRepositoryService.EXPECT().IsBackupEnabled().Return(true, nil)

	backupsDisabledApp := apps_basic.RepoApp{AppId: 1, AppName: "app-no-backups", AutomaticBackupsEnabled: false}
	backupsEnabledApp := apps_basic.RepoApp{AppId: 2, AppName: "app-backups", AutomaticBackupsEnabled: true}

	testObjects.AppRepo.EXPECT().ListApps().Return([]apps_basic.RepoApp{backupsDisabledApp, backupsEnabledApp}, nil)

	testObjects.BackupService.EXPECT().CreateBackup(backupsEnabledApp.AppId, tools.ScheduledBackupDescription).Return(nil)

	testObjects.Helper.CreateBackupsForAllAppsIfConfigured()
}

func TestCreateBackups_BackupsDisabled(t *testing.T) {
	testObjects := newAgentHelperTestObjects(t)
	defer assertAgentHelperExpectations(t, testObjects)

	testObjects.SshRepositoryService.EXPECT().IsBackupEnabled().Return(false, nil)

	testObjects.Helper.CreateBackupsForAllAppsIfConfigured()
}
func TestRetentBackups_BackupsEnabled(t *testing.T) {
	testObjects := newAgentHelperTestObjects(t)
	defer assertAgentHelperExpectations(t, testObjects)

	testObjects.SshRepositoryService.EXPECT().IsBackupEnabled().Return(true, nil)

	appOne := tools.MaintainerAndApp{Maintainer: "maintainer-1", AppName: "app-1"}
	backupInfo := []tools.BackupInfo{
		{BackupId: "123", Maintainer: "maintainer-1", AppName: "app-1"},
	}

	testObjects.BackupService.EXPECT().ListAppsInBackupRepo().Return([]tools.MaintainerAndApp{appOne}, nil)

	testObjects.BackupService.EXPECT().ListBackupsOfApp(appOne).Return(backupInfo, nil)
	testObjects.BackupDeletionFinder.EXPECT().GetBackupsForRetention(backupInfo).Return([]string{"123"}, nil)
	testObjects.BackupService.EXPECT().DeleteBackups([]string{"123"}).Return(nil)

	testObjects.Helper.RetentBackupsForAllAppsIfConfigured()
}

func TestRetentBackups_BackupsDisabled(t *testing.T) {
	testObjects := newAgentHelperTestObjects(t)
	defer assertAgentHelperExpectations(t, testObjects)

	testObjects.SshRepositoryService.EXPECT().IsBackupEnabled().Return(false, nil)

	testObjects.Helper.RetentBackupsForAllAppsIfConfigured()
	assert.Nil(t, nil)
}
