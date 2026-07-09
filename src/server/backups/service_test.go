package backups

import (
	"server/backup_server"
	"server/tools"
	"testing"

	"github.com/quollix/common/assert"
	u "github.com/quollix/common/utils"
)

func TestBackupServiceImpl_ResolveSingleAppNameByBackupIds_SingleApp(t *testing.T) {
	resticService := backup_server.NewResticServiceMock(t)
	service := &BackupServiceImpl{
		ResticService: resticService,
	}

	resticService.EXPECT().ListBackups().Return([]tools.BackupInfo{
		{BackupId: "backup-1", AppName: "xwiki"},
		{BackupId: "backup-2", AppName: "xwiki"},
		{BackupId: "backup-3", AppName: "vaultwarden"},
	}, nil)

	appName, err := service.ResolveSingleAppNameByBackupIds([]string{"backup-1", "backup-2"})
	assert.Nil(t, err)
	assert.Equal(t, "xwiki", appName)
}

func TestBackupServiceImpl_ResolveSingleAppNameByBackupIds_BackupNotFound(t *testing.T) {
	resticService := backup_server.NewResticServiceMock(t)
	service := &BackupServiceImpl{
		ResticService: resticService,
	}

	resticService.EXPECT().ListBackups().Return([]tools.BackupInfo{
		{BackupId: "backup-1", AppName: "xwiki"},
	}, nil)

	_, err := service.ResolveSingleAppNameByBackupIds([]string{"missing-backup"})
	assert.Equal(t, "backup not found", u.ExtractError(err))
}

func TestBackupServiceImpl_ResolveSingleAppNameByBackupIds_MultipleAppsReturnsError(t *testing.T) {
	resticService := backup_server.NewResticServiceMock(t)
	service := &BackupServiceImpl{
		ResticService: resticService,
	}

	resticService.EXPECT().ListBackups().Return([]tools.BackupInfo{
		{BackupId: "backup-1", AppName: "xwiki"},
		{BackupId: "backup-2", AppName: "vaultwarden"},
	}, nil)

	_, err := service.ResolveSingleAppNameByBackupIds([]string{"backup-1", "backup-2"})
	assert.Equal(t, "backups must belong to exactly one app", u.ExtractError(err))
}
