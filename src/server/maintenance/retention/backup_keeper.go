package retention

import (
	"server/tools"
)

type BackupDeletionFinder interface {
	GetBackupsForRetention(backups []tools.BackupInfo) ([]string, error)
}

type BackupDeletionFinderImpl struct {
	SelectionHelper     BackupRetentionSelector
	RetentionPolicyRepo RetentionPolicyRepository
}

type RetentionPolicy struct {
	KeepPreUpdate int `json:"keep_pre_update"`
	KeepDaily     int `json:"keep_daily"`
	KeepWeekly    int `json:"keep_weekly"`
	KeepMonthly   int `json:"keep_monthly"`
	KeepYearly    int `json:"keep_yearly"`
}

func (b *BackupDeletionFinderImpl) GetBackupsForRetention(backups []tools.BackupInfo) ([]string, error) {
	scheduledBackups, preUpdateBackups, backupIdsToPotentiallyRetent := b.splitUpBackups(backups)
	policy, err := b.RetentionPolicyRepo.GetRetentionPolicy()
	if err != nil {
		return nil, err
	}
	backupIdsToKeep := b.findBackupIdsToKeep(preUpdateBackups, policy, scheduledBackups)
	return b.SelectionHelper.FindBackupIdsToRetent(backupIdsToPotentiallyRetent, backupIdsToKeep), nil
}

func (b *BackupDeletionFinderImpl) findBackupIdsToKeep(preUpdateBackups []tools.BackupInfo, policy *RetentionPolicy, scheduledBackups []tools.BackupInfo) map[string]struct{} {
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

func (b *BackupDeletionFinderImpl) splitUpBackups(backups []tools.BackupInfo) ([]tools.BackupInfo, []tools.BackupInfo, []string) {
	var nonManualBackups []tools.BackupInfo
	for _, backup := range backups {
		if backup.Description == tools.ManualBackupDescription {
			continue
		}
		nonManualBackups = append(nonManualBackups, backup)
	}

	sortedNonManualBackups := b.SelectionHelper.CopyAndSortBackupsByNewestFirst(nonManualBackups)

	var scheduledBackups []tools.BackupInfo
	var preUpdateBackups []tools.BackupInfo
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
