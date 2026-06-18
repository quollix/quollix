package configs

import (
	"net/http"
	"server/tools"

	u "github.com/quollix/common/utils"
	"github.com/quollix/common/validation"
)

type Settings struct {
	Data []KeyAndValue `json:"data"`
}

type KeyAndValue struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type SettingsHandler struct {
	ConfigsService        ConfigsService
	TimezoneProvider      tools.TimezoneProvider
	MaintenanceRepository MaintenanceRepository
}

func (s *SettingsHandler) ReadHostHandler(w http.ResponseWriter, r *http.Request) {
	host, err := s.ConfigsService.GetServerHost()
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}
	u.SendJsonResponse(w, host)
}

func (s *SettingsHandler) SaveHostHandler(w http.ResponseWriter, r *http.Request) {
	hostString, ok := validation.ReadBody[tools.HostString](w, r)
	if !ok {
		return
	}

	err := s.ConfigsService.SetServerHost(hostString.Value)
	if err != nil {
		u.WriteResponseError(w, nil, err, tools.RequestHostField, hostString.Value)
		return
	}
}
