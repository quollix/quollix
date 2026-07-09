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

func (s *SettingsHandler) ReadBaseDomainHandler(w http.ResponseWriter, r *http.Request) {
	baseDomain, err := s.ConfigsService.GetBaseDomain()
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}
	u.SendJsonResponse(w, baseDomain)
}

func (s *SettingsHandler) SaveBaseDomainHandler(w http.ResponseWriter, r *http.Request) {
	baseDomainString, ok := validation.ReadBody[tools.BaseDomainString](w, r)
	if !ok {
		return
	}

	err := s.ConfigsService.SetBaseDomain(baseDomainString.Value)
	if err != nil {
		u.WriteResponseError(w, nil, err, tools.BaseDomainField, baseDomainString.Value)
		return
	}
}
