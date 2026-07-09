package frontend

import (
	"net/http"
	"server/backups"
	"server/tools"
	"sort"
	"sync"

	u "github.com/quollix/common/utils"
	"github.com/quollix/common/validation"
)

type BackupsPageLoaderHandler struct {
	BackupService backups.BackupService

	mutex      sync.Mutex   `wire:"-"`
	isRunning  bool         `wire:"-"`
	maintainer string       `wire:"-"`
	appName    string       `wire:"-"`
	backups    []BackupsDto `wire:"-"`
}

func (h *BackupsPageLoaderHandler) StartLoading(request tools.MaintainerAndApp, force bool) {
	h.mutex.Lock()
	if h.isRunning {
		h.mutex.Unlock()
		return
	}

	isCurrentRequest := h.maintainer == request.Maintainer && h.appName == request.AppName
	if !force && isCurrentRequest && h.backups != nil {
		h.mutex.Unlock()
		return
	}

	h.isRunning = true
	h.maintainer = request.Maintainer
	h.appName = request.AppName
	h.backups = nil
	h.mutex.Unlock()

	go h.load(request)
}

func (h *BackupsPageLoaderHandler) Read(w http.ResponseWriter, r *http.Request) {
	request := tools.MaintainerAndApp{
		Maintainer: r.URL.Query().Get("maintainer"),
		AppName:    r.URL.Query().Get("app"),
	}
	if err := validation.Validate("maintainer", validation.FieldDefault, request.Maintainer); err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}
	if err := validation.Validate("app", validation.FieldDefault, request.AppName); err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}

	h.StartLoading(request, r.URL.Query().Get("reload") == "true")
	u.SendJsonResponse(w, h.ReadLoad())
}

func (h *BackupsPageLoaderHandler) ReadLoad() BackupsPageLoadResponse {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	backups := make([]BackupsDto, len(h.backups))
	copy(backups, h.backups)

	return BackupsPageLoadResponse{
		IsRunning: h.isRunning,
		Backups:   backups,
	}
}

func (h *BackupsPageLoaderHandler) load(request tools.MaintainerAndApp) {
	backupsOfApp, err := h.BackupService.ListBackupsOfApp(request)
	if err != nil {
		u.Logger.Error(err)
		backupsOfApp = nil
	}

	sort.Slice(backupsOfApp, func(i, j int) bool {
		return backupsOfApp[i].BackupCreationTimestamp.After(backupsOfApp[j].BackupCreationTimestamp)
	})

	backups := make([]BackupsDto, 0, len(backupsOfApp))
	for _, backup := range backupsOfApp {
		backups = append(backups, BackupsDto{
			BackupId:                      backup.BackupId,
			VersionName:                   backup.VersionName,
			Description:                   backup.Description,
			BackupCreationDate:            backup.BackupCreationTimestamp.Format(tools.PrettyFrontendTimeLayoutWithDay),
			CreatedWithApplicationVersion: backup.ApplicationVersion,
		})
	}

	h.mutex.Lock()
	h.backups = backups
	h.isRunning = false
	h.mutex.Unlock()
}
