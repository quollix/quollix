package backups

import "server/tools"

type BackupQueryService interface {
	FilterBackupsOfApp(allBackups []tools.BackupInfo, request tools.MaintainerAndApp) []tools.BackupInfo
	UniqueMaintainerAndAppPairs(allBackups []tools.BackupInfo) []tools.MaintainerAndApp
}

type BackupQueryServiceImpl struct{}

func (q *BackupQueryServiceImpl) FilterBackupsOfApp(allBackups []tools.BackupInfo, request tools.MaintainerAndApp) []tools.BackupInfo {
	var filtered []tools.BackupInfo
	for _, backup := range allBackups {
		if backup.Maintainer == request.Maintainer && backup.AppName == request.AppName {
			filtered = append(filtered, backup)
		}
	}
	return filtered
}

func (q *BackupQueryServiceImpl) UniqueMaintainerAndAppPairs(allBackups []tools.BackupInfo) []tools.MaintainerAndApp {
	var pairs []tools.MaintainerAndApp
	for _, backup := range allBackups {
		pairs = append(pairs, tools.MaintainerAndApp{
			Maintainer: backup.Maintainer,
			AppName:    backup.AppName,
		})
	}
	return findUniqueMaintainerAndAppNamePairs(pairs)
}

func findUniqueMaintainerAndAppNamePairs(apps []tools.MaintainerAndApp) []tools.MaintainerAndApp {
	seen := make(map[string]struct{})
	var result []tools.MaintainerAndApp

	for _, app := range apps {
		key := app.Maintainer + "|" + app.AppName
		if _, exists := seen[key]; !exists {
			seen[key] = struct{}{}
			result = append(result, tools.MaintainerAndApp{
				Maintainer: app.Maintainer,
				AppName:    app.AppName,
			})
		}
	}
	return result
}
