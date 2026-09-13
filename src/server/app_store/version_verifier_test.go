package app_store

import (
	"crypto/ed25519"
	"crypto/rand"
	"testing"
	"time"

	"github.com/quollix/common/assert"
	"github.com/quollix/common/store"
	u "github.com/quollix/common/utils"
)

var testVersionSigningService = &store.VersionSigningServiceImpl{
	Codec:       &store.VersionSigningCodecImpl{},
	BytesSigner: &u.BytesSignerImpl{},
}

type versionVerifierTestDependencies struct {
	verifier *VersionVerifierImpl
	resolver *MaintainerKeyResolverMock
}

func setupVersionVerifierTestDependencies(t *testing.T) versionVerifierTestDependencies {
	resolver := NewMaintainerKeyResolverMock(t)
	return versionVerifierTestDependencies{
		verifier: &VersionVerifierImpl{
			MaintainerKeyResolver: resolver,
			VersionSigning:        testVersionSigningService,
		},
		resolver: resolver,
	}
}

func TestVersionVerifierVerify_NilVersionReturnsSecurityError(t *testing.T) {
	deps := setupVersionVerifierTestDependencies(t)

	err := deps.verifier.Verify(nil)

	assert.NotNil(t, err)
	assert.Equal(t, InvalidPackageSigningError, u.ExtractError(err))
}

func TestVersionVerifierVerify_OfficialPublicKeyMismatchReturnsSecurityErrorWithoutRefresh(t *testing.T) {
	deps := setupVersionVerifierTestDependencies(t)
	version, err := newSignedTestingVersion([]byte(u.LocalTestingPrivateKeyOpenSSH), []byte(u.LocalTestingPrivateKeyPassphrase))
	assert.Nil(t, err)
	version.Maintainer = u.OfficialMaintainer
	publicKey, _, keyErr := ed25519.GenerateKey(rand.Reader)
	assert.Nil(t, keyErr)
	version.MaintainerPublicKeyRaw = publicKey
	deps.resolver.EXPECT().ResolveTrustedPublicKey(version.Maintainer).Return(ed25519.PublicKey(u.GetLocalTestingPublicKeyRaw()), nil)

	err = deps.verifier.Verify(version)

	assert.NotNil(t, err)
	assert.Equal(t, InvalidPackageSigningError, u.ExtractError(err))
}

func TestVersionVerifierVerify_TamperedSignatureReturnsSecurityError(t *testing.T) {
	deps := setupVersionVerifierTestDependencies(t)
	version, err := newSignedTestingVersion([]byte(u.LocalTestingPrivateKeyOpenSSH), []byte(u.LocalTestingPrivateKeyPassphrase))
	assert.Nil(t, err)
	version.Signature = []byte("tampered-signature")
	version.Maintainer = u.OfficialMaintainer
	deps.resolver.EXPECT().ResolveTrustedPublicKey(version.Maintainer).Return(ed25519.PublicKey(u.GetLocalTestingPublicKeyRaw()), nil)

	err = deps.verifier.Verify(version)

	assert.NotNil(t, err)
	assert.Equal(t, InvalidPackageSigningError, u.ExtractError(err))
}

func newSignedTestingVersion(privateKeyBytes []byte, passphrase []byte) (*store.Version, error) {
	privateKey, err := u.DecodeEd25519PrivateKeyOpenSSH(privateKeyBytes, passphrase)
	if err != nil {
		return nil, err
	}
	version := &store.Version{
		Maintainer:               "samplemaintainer",
		AppName:                  "sampleapp",
		VersionName:              "1.0",
		Content:                  []byte("services:\n  app:\n    image: test\n"),
		VersionCreationTimestamp: time.Date(2021, time.January, 1, 1, 0, 0, 0, time.UTC),
		MaintainerPublicKeyRaw:   privateKey.Public().(ed25519.PublicKey),
	}
	signature, err := testVersionSigningService.SignVersion(privateKey, version)
	if err != nil {
		return nil, err
	}
	version.Signature = signature
	return version, nil
}
