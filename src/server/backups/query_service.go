package backups

import "github.com/quollix/common/quollix/api"

type BackupQueryService interface {
	FilterBackupsOfApp(allBackups []api.BackupInfo, request api.MaintainerAndApp) []api.BackupInfo
	UniqueMaintainerAndAppPairs(allBackups []api.BackupInfo) []api.MaintainerAndApp
}

type BackupQueryServiceImpl struct{}

func (q *BackupQueryServiceImpl) FilterBackupsOfApp(allBackups []api.BackupInfo, request api.MaintainerAndApp) []api.BackupInfo {
	var filtered []api.BackupInfo
	for _, backup := range allBackups {
		if backup.Maintainer == request.Maintainer && backup.AppName == request.AppName {
			filtered = append(filtered, backup)
		}
	}
	return filtered
}

func (q *BackupQueryServiceImpl) UniqueMaintainerAndAppPairs(allBackups []api.BackupInfo) []api.MaintainerAndApp {
	var pairs []api.MaintainerAndApp
	for _, backup := range allBackups {
		pairs = append(pairs, api.MaintainerAndApp{
			Maintainer: backup.Maintainer,
			AppName:    backup.AppName,
		})
	}
	return findUniqueMaintainerAndAppNamePairs(pairs)
}

func findUniqueMaintainerAndAppNamePairs(apps []api.MaintainerAndApp) []api.MaintainerAndApp {
	seen := make(map[string]struct{})
	var result []api.MaintainerAndApp

	for _, app := range apps {
		key := app.Maintainer + "|" + app.AppName
		if _, exists := seen[key]; !exists {
			seen[key] = struct{}{}
			result = append(result, api.MaintainerAndApp{
				Maintainer: app.Maintainer,
				AppName:    app.AppName,
			})
		}
	}
	return result
}
