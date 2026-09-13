package app_store

import (
	"bytes"
	"crypto/ed25519"

	"github.com/quollix/common/store"
	u "github.com/quollix/common/utils"
)

type VersionVerifier interface {
	Verify(version *store.Version) error
}

const (
	InvalidPackageSigningError       = "invalid package signing, app rejected for security reasons"
	SignatureVerificationFailedError = "signature verification failed"
)

type VersionVerifierImpl struct {
	MaintainerKeyResolver MaintainerKeyResolver
	VersionSigning        store.VersionSigningService
}

func (v *VersionVerifierImpl) Verify(version *store.Version) error {
	if version == nil {
		return u.Logger.NewError(InvalidPackageSigningError)
	}

	publicKey, err := v.MaintainerKeyResolver.ResolveTrustedPublicKey(version.Maintainer)
	if err != nil {
		return err
	}

	err = v.verifyWithPublicKey(publicKey, version)
	if err == nil {
		return nil
	}
	if version.Maintainer == u.OfficialMaintainer {
		return err
	}

	refreshedPublicKey, refreshErr := v.MaintainerKeyResolver.RefreshTrustedPublicKey(version.Maintainer)
	if refreshErr != nil {
		return refreshErr
	}
	return v.verifyWithPublicKey(refreshedPublicKey, version)
}

func (v *VersionVerifierImpl) verifyWithPublicKey(publicKey ed25519.PublicKey, version *store.Version) error {
	if !bytes.Equal(version.MaintainerPublicKeyRaw, publicKey) {
		return u.Logger.NewError(InvalidPackageSigningError, "reason", "public key mismatch")
	}

	isValid, err := v.VersionSigning.VerifyVersionSignature(publicKey, version)
	if err != nil {
		return u.Logger.NewError(InvalidPackageSigningError, "reason", err.Error())
	}
	if !isValid {
		return u.Logger.NewError(InvalidPackageSigningError, "reason", SignatureVerificationFailedError)
	}

	return nil
}
