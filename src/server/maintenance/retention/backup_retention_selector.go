package retention

import (
	"fmt"
	"server/tools"
	"sort"
)

type BackupRetentionSelector interface {
	FindPreUpdateBackupsToRetain(backups []tools.BackupInfo, keepPreUpdate int) []string

	FindDailyBackupsToRetain(backups []tools.BackupInfo, keepDaily int) []string
	FindWeeklyBackupsToRetain(backups []tools.BackupInfo, keepWeekly int) []string
	FindMonthlyBackupsToRetain(backups []tools.BackupInfo, keepMonthly int) []string
	FindYearlyBackupsToRetain(backups []tools.BackupInfo, keepYearly int) []string

	CopyAndSortBackupsByNewestFirst(backups []tools.BackupInfo) []tools.BackupInfo
	ExtractBackupIds(backups []tools.BackupInfo) []string

	MergeUniqueBackupIds(backupIdLists ...[]string) map[string]struct{}
	FindBackupIdsToRetent(backupIds []string, retainedBackupIds map[string]struct{}) []string
}

type BackupRetentionSelectorImpl struct{}

func (b BackupRetentionSelectorImpl) FindPreUpdateBackupsToRetain(sortedPreUpdateBackups []tools.BackupInfo, keepPreUpdate int) []string {
	var retainedBackupIds []string
	for backupIndex, backup := range sortedPreUpdateBackups {
		if backupIndex >= keepPreUpdate {
			break
		}
		retainedBackupIds = append(retainedBackupIds, backup.BackupId)
	}
	return retainedBackupIds
}

func (s *BackupRetentionSelectorImpl) FindDailyBackupsToRetain(backups []tools.BackupInfo, keepDaily int) []string {
	return s.retainNewestIdsByStringBucket(backups, keepDaily, func(backup tools.BackupInfo) string {
		return backup.BackupCreationTimestamp.Format("2006-01-02")
	})
}

func (s *BackupRetentionSelectorImpl) FindWeeklyBackupsToRetain(backups []tools.BackupInfo, keepWeekly int) []string {
	return s.retainNewestIdsByStringBucket(backups, keepWeekly, func(backup tools.BackupInfo) string {
		year, week := backup.BackupCreationTimestamp.ISOWeek()
		return fmt.Sprintf("%04d-%02d", year, week)
	})
}

func (s *BackupRetentionSelectorImpl) FindMonthlyBackupsToRetain(backups []tools.BackupInfo, keepMonthly int) []string {
	return s.retainNewestIdsByStringBucket(backups, keepMonthly, func(backup tools.BackupInfo) string {
		return backup.BackupCreationTimestamp.Format("2006-01")
	})
}

func (s *BackupRetentionSelectorImpl) FindYearlyBackupsToRetain(backups []tools.BackupInfo, keepYearly int) []string {
	return s.retainNewestIdsByStringBucket(backups, keepYearly, func(backup tools.BackupInfo) string {
		return backup.BackupCreationTimestamp.Format("2006")
	})
}

func (s *BackupRetentionSelectorImpl) retainNewestIdsByStringBucket(backups []tools.BackupInfo, keepCount int, bucketKeyFunc func(tools.BackupInfo) string) []string {
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

func (s *BackupRetentionSelectorImpl) CopyAndSortBackupsByNewestFirst(backups []tools.BackupInfo) []tools.BackupInfo {
	sortedBackups := make([]tools.BackupInfo, 0, len(backups))
	sortedBackups = append(sortedBackups, backups...)
	sort.Slice(sortedBackups, func(firstIndex, secondIndex int) bool {
		return sortedBackups[firstIndex].BackupCreationTimestamp.After(sortedBackups[secondIndex].BackupCreationTimestamp)
	})
	return sortedBackups
}

func (s *BackupRetentionSelectorImpl) ExtractBackupIds(backups []tools.BackupInfo) []string {
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
