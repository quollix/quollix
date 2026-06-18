package retention

import (
	"server/tools"
	"testing"

	"github.com/quollix/common/assert"
)

func TestScheduledBackupsHappyPath(t *testing.T) {
	selectionHelper := NewBackupRetentionSelectorMock(t)
	retentionPolicyRepo := NewRetentionPolicyRepositoryMock(t)
	backupDeletionFinder := &BackupDeletionFinderImpl{
		SelectionHelper:     selectionHelper,
		RetentionPolicyRepo: retentionPolicyRepo,
	}

	inputBackups := []tools.BackupInfo{
		{BackupId: "scheduled-daily", Description: tools.ScheduledBackupDescription},
		{BackupId: "scheduled-weekly", Description: tools.ScheduledBackupDescription},
		{BackupId: "scheduled-monthly", Description: tools.ScheduledBackupDescription},
		{BackupId: "scheduled-yearly", Description: tools.ScheduledBackupDescription},
		{BackupId: "preupdate-1", Description: tools.PreUpdateBackupDescription},
		{BackupId: "manual-1", Description: tools.ManualBackupDescription},
	}

	nonManualBackups := []tools.BackupInfo{
		{BackupId: "scheduled-daily", Description: tools.ScheduledBackupDescription},
		{BackupId: "scheduled-weekly", Description: tools.ScheduledBackupDescription},
		{BackupId: "scheduled-monthly", Description: tools.ScheduledBackupDescription},
		{BackupId: "scheduled-yearly", Description: tools.ScheduledBackupDescription},
		{BackupId: "preupdate-1", Description: tools.PreUpdateBackupDescription},
	}

	sortedNonManualBackups := []tools.BackupInfo{
		{BackupId: "preupdate-1", Description: tools.PreUpdateBackupDescription},
		{BackupId: "scheduled-daily", Description: tools.ScheduledBackupDescription},
		{BackupId: "scheduled-weekly", Description: tools.ScheduledBackupDescription},
		{BackupId: "scheduled-monthly", Description: tools.ScheduledBackupDescription},
		{BackupId: "scheduled-yearly", Description: tools.ScheduledBackupDescription},
	}

	backupIdsToPotentiallyRetent := []string{
		"preupdate-1",
		"scheduled-daily",
		"scheduled-weekly",
		"scheduled-monthly",
		"scheduled-yearly",
	}

	scheduledBackups := []tools.BackupInfo{
		{BackupId: "scheduled-daily", Description: tools.ScheduledBackupDescription},
		{BackupId: "scheduled-weekly", Description: tools.ScheduledBackupDescription},
		{BackupId: "scheduled-monthly", Description: tools.ScheduledBackupDescription},
		{BackupId: "scheduled-yearly", Description: tools.ScheduledBackupDescription},
	}

	preUpdateBackups := []tools.BackupInfo{
		{BackupId: "preupdate-1", Description: tools.PreUpdateBackupDescription},
	}

	policy := GetDefaultRetentionPolicy()

	preUpdateRetainedBackupIds := []string{"preupdate-1"}
	dailyRetainedBackupIds := []string{"scheduled-daily"}
	weeklyRetainedBackupIds := []string{"scheduled-weekly"}
	monthlyRetainedBackupIds := []string{"scheduled-monthly"}
	yearlyRetainedBackupIds := []string{"scheduled-yearly"}

	backupIdsToKeep := map[string]struct{}{
		"preupdate-1":       {},
		"scheduled-daily":   {},
		"scheduled-weekly":  {},
		"scheduled-monthly": {},
		"scheduled-yearly":  {},
	}

	expectedBackupsToDelete := []string{}

	retentionPolicyRepo.EXPECT().GetRetentionPolicy().Return(policy, nil)
	selectionHelper.EXPECT().CopyAndSortBackupsByNewestFirst(nonManualBackups).Return(sortedNonManualBackups)

	selectionHelper.EXPECT().FindPreUpdateBackupsToRetain(preUpdateBackups, policy.KeepPreUpdate).Return(preUpdateRetainedBackupIds)
	selectionHelper.EXPECT().FindDailyBackupsToRetain(scheduledBackups, policy.KeepDaily).Return(dailyRetainedBackupIds)
	selectionHelper.EXPECT().FindWeeklyBackupsToRetain(scheduledBackups, policy.KeepWeekly).Return(weeklyRetainedBackupIds)
	selectionHelper.EXPECT().FindMonthlyBackupsToRetain(scheduledBackups, policy.KeepMonthly).Return(monthlyRetainedBackupIds)
	selectionHelper.EXPECT().FindYearlyBackupsToRetain(scheduledBackups, policy.KeepYearly).Return(yearlyRetainedBackupIds)

	selectionHelper.EXPECT().
		MergeUniqueBackupIds([][]string{
			preUpdateRetainedBackupIds,
			dailyRetainedBackupIds,
			weeklyRetainedBackupIds,
			monthlyRetainedBackupIds,
			yearlyRetainedBackupIds,
		}).
		Return(backupIdsToKeep)

	selectionHelper.EXPECT().FindBackupIdsToRetent(backupIdsToPotentiallyRetent, backupIdsToKeep).Return(expectedBackupsToDelete)

	actualBackupsToDelete, err := backupDeletionFinder.GetBackupsForRetention(inputBackups)
	assert.Nil(t, err)
	assert.Equal(t, expectedBackupsToDelete, actualBackupsToDelete)
}
