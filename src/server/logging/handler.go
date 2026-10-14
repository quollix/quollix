package logging

import (
	"net/http"

	u "github.com/quollix/common/utils"
)

type Handler struct {
	LogSink            LogSink
	VictoriaLogsClient VictoriaLogsClient
}

func (h *Handler) ReadLogs(w http.ResponseWriter, r *http.Request) {
	records, err := h.VictoriaLogsClient.Pull()
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}
	u.SendJsonResponse(w, records)
}

func (h *Handler) FlushLogs(w http.ResponseWriter, r *http.Request) {
	h.LogSink.Flush()
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) WipeLogs(w http.ResponseWriter, r *http.Request) {
	h.LogSink.Flush()
	if err := h.VictoriaLogsClient.Wipe(); err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}
