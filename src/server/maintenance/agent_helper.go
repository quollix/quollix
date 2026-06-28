package maintenance

import (
	"math/rand"
	"server/apps_advanced"
	"server/apps_basic"
	"server/backup_server"
	"server/backups"
	"server/maintenance/retention"
	"server/tools"
	"time"

	u "github.com/quollix/common/utils"
)

type AgentHelper interface {
	CalculateNextMaintenanceAtUtc(now time.Time, ianaTimeZone string, maintenanceWindowStartHour int) (*time.Time, error)
	UpdateAllApps()
	CreateBackupsForAllAppsIfConfigured()
	RetentBackupsForAllAppsIfConfigured()
}
type AgentHelperImpl struct {
	RandomProvider       RandomProvider
	AppRepo              apps_basic.AppRepository
	AppDetector          apps_basic.AppDetector
	BackupService        backups.BackupService
	SshRepositoryService backup_server.SshRepositoryService
	BackupDeletionFinder retention.BackupDeletionFinder
	OperationRegistry    apps_basic.OperationRegistry
	AppsServiceAdvanced  apps_advanced.AppsServiceAdvanced
}

func (m *AgentHelperImpl) UpdateAllApps() {
	repoApps, err := m.AppRepo.ListApps()
	if err != nil {
		u.Logger.Error(err)
		return
	}

	for _, repoApp := range repoApps {
		if m.AppDetector.IsSystemApp(repoApp.AppName) {
			continue
		}
		if !repoApp.AutomaticUpdatesEnabled {
			continue
		}

		handle, err := m.OperationRegistry.TryBlockAppOperation(repoApp.AppName, "running maintenance job - updating app: "+repoApp.AppName)
		if err != nil {
			u.Logger.Error(err)
			return
		}

		func() {
			defer handle.Done()

			if err := m.AppsServiceAdvanced.UpdateAppViaAppStore(repoApp.AppId); err != nil {
				u.Logger.Error(err)
				return
			}
		}()
	}
}

func (m *AgentHelperImpl) CreateBackupsForAllAppsIfConfigured() {
	isEnabled, err := m.SshRepositoryService.IsBackupEnabled()
	if err != nil {
		u.Logger.Error(err)
		return
	}
	if !isEnabled {
		u.Logger.Info("Skipping creation of backups for all apps because backup is not enabled")
		return
	}

	repoApps, err := m.AppRepo.ListApps()
	if err != nil {
		u.Logger.Error(err)
		return
	}

	for _, repoApp := range repoApps {
		if !repoApp.AutomaticBackupsEnabled {
			continue
		}

		handle, err := m.OperationRegistry.TryBlockAppOperation(repoApp.AppName, "running maintenance job - creating app backup of: "+repoApp.AppName)
		if err != nil {
			u.Logger.Error(err)
			return
		}

		func() {
			defer handle.Done()

			if err := m.BackupService.CreateBackup(repoApp.AppId, tools.ScheduledBackupDescription); err != nil {
				u.Logger.Error(err)
				return
			}
		}()
	}
}

func (m *AgentHelperImpl) RetentBackupsForAllAppsIfConfigured() {
	isEnabled, err := m.SshRepositoryService.IsBackupEnabled()
	if err != nil {
		u.Logger.Error(err)
		return
	}
	if !isEnabled {
		u.Logger.Info("Skipping creation of backups for all apps because backup is not enabled")
		return
	}

	backedUpApps, err := m.BackupService.ListAppsInBackupRepo()
	if err != nil {
		u.Logger.Error(err)
		return
	}

	for _, backedUpApp := range backedUpApps {
		handle, err := m.OperationRegistry.TryBlockAppOperation(backedUpApp.AppName, "running maintenance job - retent backups of app: "+backedUpApp.AppName)
		if err != nil {
			u.Logger.Error(err)
			return
		}

		func() {
			defer handle.Done()

			maintainerAndApp := tools.MaintainerAndApp{
				Maintainer: backedUpApp.Maintainer,
				AppName:    backedUpApp.AppName,
			}

			backupInfo, err := m.BackupService.ListBackupsOfApp(maintainerAndApp)
			if err != nil {
				u.Logger.Error(err)
				return
			}

			backupIdsToDelete, err := m.BackupDeletionFinder.GetBackupsForRetention(backupInfo)
			if err != nil {
				u.Logger.Error(err)
				return
			}

			if err := m.BackupService.DeleteBackups(backupIdsToDelete); err != nil {
				u.Logger.Error(err)
				return
			}
		}()
	}
}

func (m *AgentHelperImpl) CalculateNextMaintenanceAtUtc(now time.Time, ianaTimeZone string, maintenanceWindowStartHour int) (*time.Time, error) {
	location, err := time.LoadLocation(ianaTimeZone)
	if err != nil {
		return nil, u.Logger.NewError(err.Error())
	}

	localNow := now.In(location)

	nextLocal := time.Date(
		localNow.Year(),
		localNow.Month(),
		localNow.Day()+1,
		maintenanceWindowStartHour,
		0,
		0,
		0,
		location,
	)

	// If "now" is close to midnight in the user's timezone, adding a full 0–59 minute jitter could push the calculated time into the following day (skipping one execution). Limiting minutes to 0–49 avoids crossing the day boundary.
	randomMinute := m.RandomProvider.Intn(50)
	randomSecond := m.RandomProvider.Intn(60)

	nextLocal = nextLocal.Add(
		time.Minute*time.Duration(randomMinute) +
			time.Second*time.Duration(randomSecond),
	)

	nextUtc := nextLocal.UTC()
	return &nextUtc, nil
}

type RandomProvider interface {
	Intn(max int) int
}

type RandomProviderImpl struct{}

func (d *RandomProviderImpl) Intn(max int) int {
	return rand.Intn(max) // #nosec G404 (CWE-338): Use of weak random number generator (math/rand or math/rand/v2 instead of crypto/rand); reasoning: no security relevant use case here
}
