package app_store

import (
	"net/http"

	api "github.com/quollix/common/quollix/api"
	u "github.com/quollix/common/utils"
	"github.com/quollix/common/validation"
)

type MaintainerHandler struct {
	MaintainerRepository MaintainerRepository
}

func (h *MaintainerHandler) ListMaintainersHandler(w http.ResponseWriter, r *http.Request) {
	maintainerKeys, err := h.MaintainerRepository.ListPublicKeys()
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}

	dtos := make([]api.MaintainerPublicKeyDto, 0, len(maintainerKeys))
	for _, maintainerKey := range maintainerKeys {
		publicKeyOpenSSH, err := MaintainerPublicKeyOpenSSH(maintainerKey.PublicKeyRaw)
		if err != nil {
			u.WriteResponseError(w, nil, err)
			return
		}
		dtos = append(dtos, api.MaintainerPublicKeyDto{
			Name:          maintainerKey.Name,
			PublicKey:     publicKeyOpenSSH,
			Fingerprint:   maintainerKey.PublicKeyFingerprint,
			LastUpdatedAt: maintainerKey.LastUpdatedAt,
		})
	}

	u.SendJsonResponse(w, dtos)
}

func (h *MaintainerHandler) AddMaintainerHandler(w http.ResponseWriter, r *http.Request) {
	request, ok := validation.ReadBody[api.MaintainerPublicKeyCreateRequest](w, r)
	if !ok {
		return
	}

	publicKey, err := u.DecodeAuthorizedEd25519PublicKey([]byte(request.PublicKey))
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}

	err = h.MaintainerRepository.UpsertPublicKey(request.Name, publicKey)
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}
}

func (h *MaintainerHandler) DeleteMaintainerHandler(w http.ResponseWriter, r *http.Request) {
	request, ok := validation.ReadBody[api.MaintainerPublicKeyDeleteRequest](w, r)
	if !ok {
		return
	}

	err := h.MaintainerRepository.DeletePublicKey(request.Name)
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}
}
