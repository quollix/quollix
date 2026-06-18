package app_store

import (
	"bytes"

	"github.com/quollix/common/store"
	u "github.com/quollix/common/utils"
)

type VersionVerifier interface {
	Verify(version *store.Version) error
}

const InvalidPackageSigningError = "invalid package signing, app rejected for security reasons"

type VersionVerifierImpl struct {
	TrustedAuthorizedKey []byte
	VersionSigning       store.VersionSigningService
}

func (v *VersionVerifierImpl) Verify(version *store.Version) error {
	if version == nil {
		return u.Logger.NewError(InvalidPackageSigningError, "reason", "version must not be nil")
	}

	publicKey, err := u.DecodeAuthorizedEd25519PublicKey(v.TrustedAuthorizedKey)
	if err != nil {
		return u.Logger.NewError(InvalidPackageSigningError, "reason", err.Error())
	}
	if !bytes.Equal(version.MaintainerPublicKeyRaw, publicKey) {
		return u.Logger.NewError(InvalidPackageSigningError, "reason", "public key mismatch")
	}

	isValid, err := v.VersionSigning.VerifyVersionSignature(publicKey, version)
	if err != nil {
		return u.Logger.NewError(InvalidPackageSigningError, "reason", err.Error())
	}
	if !isValid {
		return u.Logger.NewError(InvalidPackageSigningError, "reason", "signature verification failed")
	}

	return nil
}
