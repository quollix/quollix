package retention

import (
	"server/tools"
	"testing"
	"time"

	"github.com/quollix/common/assert"
)

var backupRetentionSelector = BackupRetentionSelectorImpl{}

func TestFindDailyBackupsToRetain(t *testing.T) {
	backups := []tools.BackupInfo{
		{BackupId: "d1a", BackupCreationTimestamp: time.Date(2023, 12, 31, 12, 0, 0, 0, time.UTC)},
		{BackupId: "d1b", BackupCreationTimestamp: time.Date(2023, 12, 31, 11, 0, 0, 0, time.UTC)},
		{BackupId: "d2", BackupCreationTimestamp: time.Date(2023, 12, 30, 12, 0, 0, 0, time.UTC)},
		{BackupId: "d3", BackupCreationTimestamp: time.Date(2023, 12, 29, 12, 0, 0, 0, time.UTC)},
	}
	sortedBackups := backupRetentionSelector.CopyAndSortBackupsByNewestFirst(backups)

	retainedBackupIds := backupRetentionSelector.FindDailyBackupsToRetain(sortedBackups, 2)

	assert.Equal(t, []string{"d1a", "d2"}, retainedBackupIds)
}

func TestFindWeeklyBackupsToRetain(t *testing.T) {
	backups := []tools.BackupInfo{
		{BackupId: "w1a", BackupCreationTimestamp: time.Date(2023, 12, 31, 12, 0, 0, 0, time.UTC)},
		{BackupId: "w1b", BackupCreationTimestamp: time.Date(2023, 12, 31, 11, 0, 0, 0, time.UTC)},
		{BackupId: "w2", BackupCreationTimestamp: time.Date(2023, 12, 24, 12, 0, 0, 0, time.UTC)},
		{BackupId: "w3", BackupCreationTimestamp: time.Date(2023, 12, 17, 12, 0, 0, 0, time.UTC)},
		{BackupId: "w4", BackupCreationTimestamp: time.Date(2023, 12, 10, 12, 0, 0, 0, time.UTC)},
	}
	sortedBackups := backupRetentionSelector.CopyAndSortBackupsByNewestFirst(backups)

	retainedBackupIds := backupRetentionSelector.FindWeeklyBackupsToRetain(sortedBackups, 3)

	assert.Equal(t, []string{"w1a", "w2", "w3"}, retainedBackupIds)
}

func TestFindMonthlyBackupsToRetain(t *testing.T) {
	backups := []tools.BackupInfo{
		{BackupId: "mDecA", BackupCreationTimestamp: time.Date(2023, 12, 31, 12, 0, 0, 0, time.UTC)},
		{BackupId: "mDecB", BackupCreationTimestamp: time.Date(2023, 12, 30, 12, 0, 0, 0, time.UTC)},
		{BackupId: "mNov", BackupCreationTimestamp: time.Date(2023, 11, 30, 12, 0, 0, 0, time.UTC)},
		{BackupId: "mOct", BackupCreationTimestamp: time.Date(2023, 10, 31, 12, 0, 0, 0, time.UTC)},
		{BackupId: "mSep", BackupCreationTimestamp: time.Date(2023, 9, 30, 12, 0, 0, 0, time.UTC)},
	}
	sortedBackups := backupRetentionSelector.CopyAndSortBackupsByNewestFirst(backups)

	retainedBackupIds := backupRetentionSelector.FindMonthlyBackupsToRetain(sortedBackups, 2)

	assert.Equal(t, []string{"mDecA", "mNov"}, retainedBackupIds)
}

func TestFindYearlyBackupsToRetain(t *testing.T) {
	backups := []tools.BackupInfo{
		{BackupId: "y2023", BackupCreationTimestamp: time.Date(2023, 12, 31, 12, 0, 0, 0, time.UTC)},
		{BackupId: "y2022", BackupCreationTimestamp: time.Date(2022, 10, 30, 12, 0, 0, 0, time.UTC)},
		{BackupId: "y2021a", BackupCreationTimestamp: time.Date(2021, 6, 15, 12, 0, 0, 0, time.UTC)},
		{BackupId: "y2021b", BackupCreationTimestamp: time.Date(2021, 6, 14, 12, 0, 0, 0, time.UTC)},
	}
	sortedBackups := backupRetentionSelector.CopyAndSortBackupsByNewestFirst(backups)

	retainedBackupIds := backupRetentionSelector.FindYearlyBackupsToRetain(sortedBackups, 3)

	assert.Equal(t, []string{"y2023", "y2022", "y2021a"}, retainedBackupIds)
}

func TestCopyAndSortBackupsByNewestFirst(t *testing.T) {
	oldest := tools.BackupInfo{BackupId: "oldest", BackupCreationTimestamp: time.Date(2026, 2, 1, 10, 0, 0, 0, time.UTC)}
	newest := tools.BackupInfo{BackupId: "newest", BackupCreationTimestamp: time.Date(2026, 2, 5, 10, 0, 0, 0, time.UTC)}
	inputBackups := []tools.BackupInfo{oldest, newest}

	sortedBackups := backupRetentionSelector.CopyAndSortBackupsByNewestFirst(inputBackups)

	assert.Equal(t, []tools.BackupInfo{newest, oldest}, sortedBackups)
}

func TestMergeUniqueBackupIds(t *testing.T) {
	listOne := []string{"a", "b"}
	listTwo := []string{"b", "c"}

	merged := backupRetentionSelector.MergeUniqueBackupIds(listOne, listTwo)

	_, hasA := merged["a"]
	_, hasB := merged["b"]
	_, hasC := merged["c"]

	assert.Equal(t, 3, len(merged))
	assert.True(t, hasA)
	assert.True(t, hasB)
	assert.True(t, hasC)
}

func TestFindBackupIdsNotInIdSet(t *testing.T) {
	backupIds := []string{"keep-1", "delete-1"}
	retainedBackupIds := map[string]struct{}{
		"keep-1": {},
	}

	candidatesForDeletion := backupRetentionSelector.FindBackupIdsToRetent(backupIds, retainedBackupIds)

	assert.Equal(t, []string{"delete-1"}, candidatesForDeletion)
}

func TestFindPreUpdateBackupsToRetain(t *testing.T) {
	backups := []tools.BackupInfo{
		{BackupId: "p1a", Description: tools.PreUpdateBackupDescription, BackupCreationTimestamp: time.Date(2023, 12, 31, 12, 0, 0, 0, time.UTC)},
		{BackupId: "p1b", Description: tools.PreUpdateBackupDescription, BackupCreationTimestamp: time.Date(2023, 12, 31, 11, 0, 0, 0, time.UTC)},
		{BackupId: "p2", Description: tools.PreUpdateBackupDescription, BackupCreationTimestamp: time.Date(2023, 12, 30, 12, 0, 0, 0, time.UTC)},
	}
	sortedBackups := backupRetentionSelector.CopyAndSortBackupsByNewestFirst(backups)

	retainedBackupIds := backupRetentionSelector.FindPreUpdateBackupsToRetain(sortedBackups, 2)

	assert.Equal(t, []string{"p1a", "p1b"}, retainedBackupIds)
}
