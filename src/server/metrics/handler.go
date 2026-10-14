package metrics

import (
	"errors"
	"net/http"

	"server/apps_basic"

	u "github.com/quollix/common/utils"
)

type Handler struct {
	ContainerMetricsMonitor apps_basic.ContainerMetricsMonitor
	VolumeMetricsMonitor    apps_basic.VolumeMetricsMonitor
	VictoriaMetricsClient   VictoriaMetricsClient
}

func (h *Handler) WipeMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if err := h.VictoriaMetricsClient.Wipe(); err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) CollectMetrics(w http.ResponseWriter, r *http.Request) {
	containerErr := h.ContainerMetricsMonitor.RunCycle()
	volumeErr := h.VolumeMetricsMonitor.RunCycle()
	if err := errors.Join(containerErr, volumeErr); err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) ReadMetrics(w http.ResponseWriter, r *http.Request) {
	series, err := h.VictoriaMetricsClient.Pull()
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}
	u.SendJsonResponse(w, series)
}
