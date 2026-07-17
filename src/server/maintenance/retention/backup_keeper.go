package retention

import (
	"server/tools"

	api "github.com/quollix/common/quollix/api"
)

type BackupDeletionFinder interface {
	GetBackupsForRetention(backups []api.BackupInfo) ([]string, error)
}

type BackupDeletionFinderImpl struct {
	SelectionHelper     BackupRetentionSelector
	RetentionPolicyRepo RetentionPolicyRepository
}

func (b *BackupDeletionFinderImpl) GetBackupsForRetention(backups []api.BackupInfo) ([]string, error) {
	scheduledBackups, preUpdateBackups, backupIdsToPotentiallyRetent := b.splitUpBackups(backups)
	policy, err := b.RetentionPolicyRepo.GetRetentionPolicy()
	if err != nil {
		return nil, err
	}
	backupIdsToKeep := b.findBackupIdsToKeep(preUpdateBackups, policy, scheduledBackups)
	return b.SelectionHelper.FindBackupIdsToRetent(backupIdsToPotentiallyRetent, backupIdsToKeep), nil
}

func (b *BackupDeletionFinderImpl) findBackupIdsToKeep(preUpdateBackups []api.BackupInfo, policy *api.RetentionPolicy, scheduledBackups []api.BackupInfo) map[string]struct{} {
	preUpdateRetainedBackupIds := b.SelectionHelper.FindPreUpdateBackupsToRetain(preUpdateBackups, policy.KeepPreUpdate)
	dailyRetainedBackupIds := b.SelectionHelper.FindDailyBackupsToRetain(scheduledBackups, policy.KeepDaily)
	weeklyRetainedBackupIds := b.SelectionHelper.FindWeeklyBackupsToRetain(scheduledBackups, policy.KeepWeekly)
	monthlyRetainedBackupIds := b.SelectionHelper.FindMonthlyBackupsToRetain(scheduledBackups, policy.KeepMonthly)
	yearlyRetainedBackupIds := b.SelectionHelper.FindYearlyBackupsToRetain(scheduledBackups, policy.KeepYearly)

	backupIdsToKeep := b.SelectionHelper.MergeUniqueBackupIds(
		preUpdateRetainedBackupIds,
		dailyRetainedBackupIds,
		weeklyRetainedBackupIds,
		monthlyRetainedBackupIds,
		yearlyRetainedBackupIds,
	)
	return backupIdsToKeep
}

func (b *BackupDeletionFinderImpl) splitUpBackups(backups []api.BackupInfo) ([]api.BackupInfo, []api.BackupInfo, []string) {
	var nonManualBackups []api.BackupInfo
	for _, backup := range backups {
		if backup.Description == tools.ManualBackupDescription {
			continue
		}
		nonManualBackups = append(nonManualBackups, backup)
	}

	sortedNonManualBackups := b.SelectionHelper.CopyAndSortBackupsByNewestFirst(nonManualBackups)

	var scheduledBackups []api.BackupInfo
	var preUpdateBackups []api.BackupInfo
	var backupIdsToPotentiallyRetent []string

	for _, backup := range sortedNonManualBackups {
		backupIdsToPotentiallyRetent = append(backupIdsToPotentiallyRetent, backup.BackupId)
		if backup.Description == tools.ScheduledBackupDescription {
			scheduledBackups = append(scheduledBackups, backup)
		}
		if backup.Description == tools.PreUpdateBackupDescription {
			preUpdateBackups = append(preUpdateBackups, backup)
		}
	}
	return scheduledBackups, preUpdateBackups, backupIdsToPotentiallyRetent
}
