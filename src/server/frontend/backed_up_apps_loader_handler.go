package frontend

import (
	"net/http"
	"server/backups"
	"server/tools"
	"sort"
	"sync"

	u "github.com/quollix/common/utils"
)

type BackedUpAppsLoaderHandler struct {
	BackupService backups.BackupService

	mutex     sync.Mutex               `wire:"-"`
	isRunning bool                     `wire:"-"`
	apps      []tools.MaintainerAndApp `wire:"-"`
}

func (h *BackedUpAppsLoaderHandler) StartLoading() {
	h.mutex.Lock()
	if h.isRunning {
		h.mutex.Unlock()
		return
	}
	h.isRunning = true
	h.apps = nil
	h.mutex.Unlock()

	go h.load()
}

func (h *BackedUpAppsLoaderHandler) Read(w http.ResponseWriter, r *http.Request) {
	u.SendJsonResponse(w, h.ReadLoad())
}

func (h *BackedUpAppsLoaderHandler) ReadLoad() BackedUpAppsPageLoadResponse {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	apps := make([]tools.MaintainerAndApp, len(h.apps))
	copy(apps, h.apps)

	return BackedUpAppsPageLoadResponse{
		IsRunning: h.isRunning,
		Apps:      apps,
	}
}

func (h *BackedUpAppsLoaderHandler) load() {
	apps, err := h.BackupService.ListAppsInBackupRepo()
	if err != nil {
		u.Logger.Error(err)
		apps = nil
	}

	sort.Slice(apps, func(i, j int) bool {
		if apps[i].Maintainer == apps[j].Maintainer {
			return apps[i].AppName < apps[j].AppName
		}
		return apps[i].Maintainer < apps[j].Maintainer
	})

	h.mutex.Lock()
	h.apps = apps
	h.isRunning = false
	h.mutex.Unlock()
}
