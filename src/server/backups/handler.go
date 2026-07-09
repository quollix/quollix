package backups

import (
	"fmt"
	"net/http"
	"server/apps_basic"
	"server/backup_server"
	"server/tools"
	"strconv"

	u "github.com/quollix/common/utils"
	"github.com/quollix/common/validation"
)

type BackupHandler struct {
	OperationRegistry          apps_basic.OperationRegistry
	BackupService              BackupService
	AppRepo                    apps_basic.AppRepository
	AppsHandler                *apps_basic.AppsHandler
	SshRepositoryConfigService backup_server.SshRepositoryService
}

func (b *BackupHandler) CreateBackupHandler(w http.ResponseWriter, r *http.Request) {
	appIdString, ok := validation.ReadBody[tools.NumberString](w, r)
	if !ok {
		return
	}

	appId, err := strconv.Atoi(appIdString.Value)
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}

	app, err := b.AppRepo.GetAppById(appId)
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}

	operation := fmt.Sprintf("backing up '%s'", app.AppName)
	handle, err := b.OperationRegistry.TryBlockAppOperation(app.AppName, operation)
	if err != nil {
		apps_basic.WriteConcurrentOperationError(w, operation, err)
		return
	}
	defer handle.Done()

	err = b.BackupService.CreateBackup(appId, tools.ManualBackupDescription)
	if err != nil {
		u.WriteResponseError(w, backup_server.BackupOperationExpectedErrors, err)
		return
	}
}

func (b *BackupHandler) ListBackupsHandler(w http.ResponseWriter, r *http.Request) {
	listRequest, ok := validation.ReadBody[tools.MaintainerAndApp](w, r)
	if !ok {
		return
	}

	backups, err := b.BackupService.ListBackupsOfApp(*listRequest)
	if err != nil {
		u.WriteResponseError(w, backup_server.BackupOperationExpectedErrors, err)
		return
	}
	u.SendJsonResponse(w, backups)
}

func (b *BackupHandler) RestoreBackupHandler(w http.ResponseWriter, r *http.Request) {
	backupRestoreRequest, ok := validation.ReadBody[tools.BackupOperationRequest](w, r)
	if !ok {
		return
	}

	if err := b.SshRepositoryConfigService.EnsureBackupIsEnabled(); err != nil {
		u.WriteResponseError(w, backup_server.BackupOperationExpectedErrors, err)
		return
	}

	appName, err := b.BackupService.ResolveSingleAppNameByBackupIds([]string{backupRestoreRequest.BackupId})
	if err != nil {
		u.WriteResponseError(w, backup_server.BackupOperationExpectedErrors, err)
		return
	}

	operation := fmt.Sprintf("restoring backup of '%s'", appName)
	handle, err := b.OperationRegistry.TryBlockAppOperation(appName, operation)
	if err != nil {
		apps_basic.WriteConcurrentOperationError(w, operation, err)
		return
	}
	defer handle.Done()

	_, err = b.BackupService.RestoreBackup(*backupRestoreRequest)
	if err != nil {
		u.WriteResponseError(w, backup_server.BackupOperationExpectedErrors, err)
		return
	}
}

func (b *BackupHandler) DeleteBackupHandler(w http.ResponseWriter, r *http.Request) {
	deleteBackupsRequest, ok := validation.ReadBody[tools.BackupsOperationRequest](w, r)
	if !ok {
		return
	}

	if err := b.SshRepositoryConfigService.EnsureBackupIsEnabled(); err != nil {
		u.WriteResponseError(w, backup_server.BackupOperationExpectedErrors, err)
		return
	}

	appName, err := b.BackupService.ResolveSingleAppNameByBackupIds(deleteBackupsRequest.BackupIds)
	if err != nil {
		u.WriteResponseError(w, backup_server.BackupOperationExpectedErrors, err)
		return
	}

	operation := fmt.Sprintf("deleting backups of '%s'", appName)
	handle, err := b.OperationRegistry.TryBlockAppOperation(appName, operation)
	if err != nil {
		apps_basic.WriteConcurrentOperationError(w, operation, err)
		return
	}
	defer handle.Done()

	err = b.BackupService.DeleteBackups(deleteBackupsRequest.BackupIds)
	if err != nil {
		u.WriteResponseError(w, backup_server.BackupOperationExpectedErrors, err)
		return
	}
}

func (b *BackupHandler) ListAppsOfBackupRepository(w http.ResponseWriter, r *http.Request) {
	apps, err := b.BackupService.ListAppsInBackupRepo()
	if err != nil {
		u.WriteResponseError(w, backup_server.BackupOperationExpectedErrors, err)
		return
	}

	u.SendJsonResponse(w, apps)
}
