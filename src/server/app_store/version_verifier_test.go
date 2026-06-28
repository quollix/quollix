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

var testVersionVerifier = &VersionVerifierImpl{
	TrustedAuthorizedKey: u.LocalTestingPublicKeyOpenSSHBytes,
	VersionSigning:       testVersionSigningService,
}

func TestVersionVerifierVerify_HappyPath(t *testing.T) {
	version, err := newSignedTestingVersion()
	assert.Nil(t, err)

	err = testVersionVerifier.Verify(version)

	assert.Nil(t, err)
}

func TestVersionVerifierVerify_PublicKeyMismatchReturnsSecurityError(t *testing.T) {
	version, err := newSignedTestingVersion()
	assert.Nil(t, err)
	publicKey, _, keyErr := ed25519.GenerateKey(rand.Reader)
	assert.Nil(t, keyErr)
	version.MaintainerPublicKeyRaw = publicKey

	err = testVersionVerifier.Verify(version)

	assert.NotNil(t, err)
	assert.Equal(t, InvalidPackageSigningError, u.ExtractError(err))
}

func TestVersionVerifierVerify_MalformedTrustedAuthorizedKeyReturnsSecurityError(t *testing.T) {
	version, err := newSignedTestingVersion()
	assert.Nil(t, err)
	verifier := &VersionVerifierImpl{
		TrustedAuthorizedKey: []byte("not-a-valid-authorized-key"),
		VersionSigning:       testVersionSigningService,
	}

	err = verifier.Verify(version)

	assert.NotNil(t, err)
	assert.Equal(t, InvalidPackageSigningError, u.ExtractError(err))
}

func TestVersionVerifierVerify_TamperedSignatureReturnsSecurityError(t *testing.T) {
	version, err := newSignedTestingVersion()
	assert.Nil(t, err)
	version.Signature = []byte("tampered-signature")

	err = testVersionVerifier.Verify(version)

	assert.NotNil(t, err)
	assert.Equal(t, InvalidPackageSigningError, u.ExtractError(err))
}

func newSignedTestingVersion() (*store.Version, error) {
	privateKey, err := decodeTestingPrivateKeyForVerifierTests()
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

func decodeTestingPrivateKeyForVerifierTests() (ed25519.PrivateKey, error) {
	return u.DecodeEd25519PrivateKeyOpenSSH(u.GetLocalTestingPrivateKeyBytes())
}
