package observability

import (
	"net/http"

	"server/apps_basic"
	"server/configs"
	"server/tools"

	"github.com/quollix/common/quollix/api"
	u "github.com/quollix/common/utils"
	"github.com/quollix/common/validation"
)

type SettingsHandler struct {
	Repository        configs.ObservabilityRepository
	Service           SettingsService
	OperationRegistry apps_basic.OperationRegistry
}

func (h *SettingsHandler) Read(w http.ResponseWriter, r *http.Request) {
	settings, err := h.Repository.Read()
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}
	u.SendJsonResponse(w, settings)
}

func (h *SettingsHandler) Save(w http.ResponseWriter, r *http.Request) {
	settings, ok := validation.ReadBody[api.ObservabilitySettings](w, r)
	if !ok {
		return
	}
	operation := "saving observability settings"
	handle, err := h.OperationRegistry.TryBlockAppOperation(tools.GrafanaAppName, operation)
	if err != nil {
		apps_basic.WriteConcurrentOperationError(w, operation, err)
		return
	}
	defer handle.Done()
	if err := h.Service.Save(settings); err != nil {
		u.WriteResponseError(w, u.MapOf(InvalidSettingsError), err)
	}
}
