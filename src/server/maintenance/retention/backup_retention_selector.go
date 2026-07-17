package retention

import (
	"fmt"
	"sort"

	api "github.com/quollix/common/quollix/api"
)

type BackupRetentionSelector interface {
	FindPreUpdateBackupsToRetain(backups []api.BackupInfo, keepPreUpdate int) []string

	FindDailyBackupsToRetain(backups []api.BackupInfo, keepDaily int) []string
	FindWeeklyBackupsToRetain(backups []api.BackupInfo, keepWeekly int) []string
	FindMonthlyBackupsToRetain(backups []api.BackupInfo, keepMonthly int) []string
	FindYearlyBackupsToRetain(backups []api.BackupInfo, keepYearly int) []string

	CopyAndSortBackupsByNewestFirst(backups []api.BackupInfo) []api.BackupInfo
	ExtractBackupIds(backups []api.BackupInfo) []string

	MergeUniqueBackupIds(backupIdLists ...[]string) map[string]struct{}
	FindBackupIdsToRetent(backupIds []string, retainedBackupIds map[string]struct{}) []string
}

type BackupRetentionSelectorImpl struct{}

func (b BackupRetentionSelectorImpl) FindPreUpdateBackupsToRetain(sortedPreUpdateBackups []api.BackupInfo, keepPreUpdate int) []string {
	var retainedBackupIds []string
	for backupIndex, backup := range sortedPreUpdateBackups {
		if backupIndex >= keepPreUpdate {
			break
		}
		retainedBackupIds = append(retainedBackupIds, backup.BackupId)
	}
	return retainedBackupIds
}

func (s *BackupRetentionSelectorImpl) FindDailyBackupsToRetain(backups []api.BackupInfo, keepDaily int) []string {
	return s.retainNewestIdsByStringBucket(backups, keepDaily, func(backup api.BackupInfo) string {
		return backup.BackupCreationTimestamp.Format("2006-01-02")
	})
}

func (s *BackupRetentionSelectorImpl) FindWeeklyBackupsToRetain(backups []api.BackupInfo, keepWeekly int) []string {
	return s.retainNewestIdsByStringBucket(backups, keepWeekly, func(backup api.BackupInfo) string {
		year, week := backup.BackupCreationTimestamp.ISOWeek()
		return fmt.Sprintf("%04d-%02d", year, week)
	})
}

func (s *BackupRetentionSelectorImpl) FindMonthlyBackupsToRetain(backups []api.BackupInfo, keepMonthly int) []string {
	return s.retainNewestIdsByStringBucket(backups, keepMonthly, func(backup api.BackupInfo) string {
		return backup.BackupCreationTimestamp.Format("2006-01")
	})
}

func (s *BackupRetentionSelectorImpl) FindYearlyBackupsToRetain(backups []api.BackupInfo, keepYearly int) []string {
	return s.retainNewestIdsByStringBucket(backups, keepYearly, func(backup api.BackupInfo) string {
		return backup.BackupCreationTimestamp.Format("2006")
	})
}

func (s *BackupRetentionSelectorImpl) retainNewestIdsByStringBucket(backups []api.BackupInfo, keepCount int, bucketKeyFunc func(api.BackupInfo) string) []string {
	if keepCount <= 0 {
		return nil
	}
	seenBucketKeys := map[string]bool{}
	retainedBackupIds := make([]string, 0, keepCount)
	for _, backup := range backups {
		bucketKey := bucketKeyFunc(backup)
		if seenBucketKeys[bucketKey] {
			continue
		}
		seenBucketKeys[bucketKey] = true
		retainedBackupIds = append(retainedBackupIds, backup.BackupId)
		if len(retainedBackupIds) >= keepCount {
			break
		}
	}
	return retainedBackupIds
}

func (s *BackupRetentionSelectorImpl) CopyAndSortBackupsByNewestFirst(backups []api.BackupInfo) []api.BackupInfo {
	sortedBackups := make([]api.BackupInfo, 0, len(backups))
	sortedBackups = append(sortedBackups, backups...)
	sort.Slice(sortedBackups, func(firstIndex, secondIndex int) bool {
		return sortedBackups[firstIndex].BackupCreationTimestamp.After(sortedBackups[secondIndex].BackupCreationTimestamp)
	})
	return sortedBackups
}

func (s *BackupRetentionSelectorImpl) ExtractBackupIds(backups []api.BackupInfo) []string {
	backupIds := make([]string, 0, len(backups))
	for _, backup := range backups {
		backupIds = append(backupIds, backup.BackupId)
	}
	return backupIds
}

func (s *BackupRetentionSelectorImpl) MergeUniqueBackupIds(backupIdLists ...[]string) map[string]struct{} {
	retainedBackupIds := map[string]struct{}{}
	for _, backupIdList := range backupIdLists {
		for _, backupId := range backupIdList {
			retainedBackupIds[backupId] = struct{}{}
		}
	}
	return retainedBackupIds
}

func (s *BackupRetentionSelectorImpl) FindBackupIdsToRetent(backupIds []string, retainedBackupIds map[string]struct{}) []string {
	var candidatesForDeletion []string
	for _, backupId := range backupIds {
		if _, exists := retainedBackupIds[backupId]; !exists {
			candidatesForDeletion = append(candidatesForDeletion, backupId)
		}
	}
	return candidatesForDeletion
}
