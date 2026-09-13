package app_store

import (
	"crypto/ed25519"

	"github.com/quollix/common/store"
	u "github.com/quollix/common/utils"
)

type MaintainerKeyResolver interface {
	ResolveTrustedPublicKey(maintainer string) (ed25519.PublicKey, error)
	RefreshTrustedPublicKey(maintainer string) (ed25519.PublicKey, error)
}

type MaintainerKeyResolverImpl struct {
	AppStoreClientLean          AppStoreClientLean
	MaintainerRepository        MaintainerRepository
	OfficialMaintainerPublicKey ed25519.PublicKey
}

func NewMaintainerKeyResolver(
	appStoreClientLean AppStoreClientLean,
	maintainerRepository MaintainerRepository,
	officialMaintainerPublicKey ed25519.PublicKey,
) MaintainerKeyResolver {
	return &MaintainerKeyResolverImpl{
		AppStoreClientLean:          appStoreClientLean,
		MaintainerRepository:        maintainerRepository,
		OfficialMaintainerPublicKey: officialMaintainerPublicKey,
	}
}

func (r *MaintainerKeyResolverImpl) ResolveTrustedPublicKey(maintainer string) (ed25519.PublicKey, error) {
	if maintainer == u.OfficialMaintainer {
		return r.OfficialMaintainerPublicKey, nil
	}

	publicKey, exists, err := r.MaintainerRepository.GetPublicKey(maintainer)
	if err != nil {
		return nil, err
	}
	if exists {
		return publicKey, nil
	}

	return r.RefreshTrustedPublicKey(maintainer)
}

func (r *MaintainerKeyResolverImpl) RefreshTrustedPublicKey(maintainer string) (ed25519.PublicKey, error) {
	if maintainer == u.OfficialMaintainer {
		return r.OfficialMaintainerPublicKey, nil
	}

	record, err := r.AppStoreClientLean.GetMaintainerPublicKeyRecord(maintainer)
	if err != nil {
		return nil, err
	}

	ok, err := store.VerifyMaintainerPublicKeySignature(r.OfficialMaintainerPublicKey, record)
	if err != nil {
		return nil, u.Logger.NewError(SignatureVerificationFailedError, "reason", err.Error())
	}
	if !ok {
		return nil, u.Logger.NewError(SignatureVerificationFailedError)
	}

	publicKey := ed25519.PublicKey(record.PublicKeyRaw)
	if err = r.MaintainerRepository.UpsertPublicKey(maintainer, publicKey); err != nil {
		return nil, err
	}
	return publicKey, nil
}
