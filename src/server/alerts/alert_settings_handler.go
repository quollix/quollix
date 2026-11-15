package alerts

import (
	"net/http"

	"github.com/quollix/common/quollix/api"
	u "github.com/quollix/common/utils"
	"github.com/quollix/common/validation"
)

type AlertSettingsHandler struct {
	Repository AlertSettingsRepository
	Service    AlertSettingsService
}

func (h *AlertSettingsHandler) Read(w http.ResponseWriter, _ *http.Request) {
	settings, err := h.Repository.Read()
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}
	u.SendJsonResponse(w, settings)
}

func (h *AlertSettingsHandler) Save(w http.ResponseWriter, r *http.Request) {
	settings, ok := validation.ReadBody[api.AlertSettings](w, r)
	if !ok {
		return
	}
	if err := h.Service.Save(settings); err != nil {
		u.WriteResponseError(w, u.MapOf(InvalidAlertSettingsError), err)
	}
}
