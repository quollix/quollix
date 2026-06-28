package oidc_provider

import (
	"net/http"
	"strconv"

	"server/tools"

	u "github.com/quollix/common/utils"
	"github.com/quollix/common/validation"
)

type OidcRelyingPartyHandler struct {
	ClientService OidcRelyingPartyService
	ClientRepo    OidcRelyingPartyRepository
}

func (h *OidcRelyingPartyHandler) CreateClient(w http.ResponseWriter, r *http.Request) {
	client, ok := validation.ReadBody[OidcRelyingPartyRequest](w, r)
	if !ok {
		return
	}
	if err := h.ClientService.CreateClient(client); err != nil {
		u.WriteResponseError(w, expectedOidcRelyingPartyErrors, err)
		return
	}
}

func (h *OidcRelyingPartyHandler) UpdateClient(w http.ResponseWriter, r *http.Request) {
	client, ok := validation.ReadBody[OidcRelyingPartyRequest](w, r)
	if !ok {
		return
	}
	if err := h.ClientService.UpdateClient(client); err != nil {
		u.WriteResponseError(w, expectedOidcRelyingPartyErrors, err)
		return
	}
}

func (h *OidcRelyingPartyHandler) ListClients(w http.ResponseWriter, _ *http.Request) {
	clients, err := h.ClientRepo.ListClients()
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}
	u.SendJsonResponse(w, clients)
}

func (h *OidcRelyingPartyHandler) DeleteClient(w http.ResponseWriter, r *http.Request) {
	clientIdString, ok := validation.ReadBody[tools.NumberString](w, r)
	if !ok {
		return
	}
	clientId, err := strconv.Atoi(clientIdString.Value)
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}
	if err := h.ClientRepo.DeleteClient(clientId); err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}
}

func (h *OidcRelyingPartyHandler) RegenerateClientCredentials(w http.ResponseWriter, r *http.Request) {
	clientIdString, ok := validation.ReadBody[tools.NumberString](w, r)
	if !ok {
		return
	}
	clientId, err := strconv.Atoi(clientIdString.Value)
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}
	if err := h.ClientService.RegenerateClientCredentials(clientId); err != nil {
		u.WriteResponseError(w, expectedOidcRelyingPartyErrors, err)
		return
	}
}

var expectedOidcRelyingPartyErrors = u.MapOf(
	OidcRelyingPartyNameAlreadyExistsError,
	OidcRelyingPartyClientIdAlreadyExistsError,
	OidcRelyingPartyRequiredFieldMissingError,
)
