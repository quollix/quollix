package app_store

import (
	"crypto/ed25519"
	"testing"

	"github.com/quollix/common/assert"
	"github.com/quollix/common/store"
	u "github.com/quollix/common/utils"
)

type maintainerKeyResolverTestDependencies struct {
	resolver *MaintainerKeyResolverImpl
	client   *AppStoreClientLeanMock
	repo     *MaintainerRepositoryMock
}

func setupMaintainerKeyResolverTestDependencies(t *testing.T) maintainerKeyResolverTestDependencies {
	client := NewAppStoreClientLeanMock(t)
	repo := NewMaintainerRepositoryMock(t)
	return maintainerKeyResolverTestDependencies{
		resolver: &MaintainerKeyResolverImpl{
			AppStoreClientLean:          client,
			MaintainerRepository:        repo,
			OfficialMaintainerPublicKey: ed25519.PublicKey(u.GetLocalTestingPublicKeyRaw()),
		},
		client: client,
		repo:   repo,
	}
}

func TestMaintainerKeyResolverResolveTrustedPublicKey_OfficialMaintainerUsesTrustedAuthorizedKey(t *testing.T) {
	deps := setupMaintainerKeyResolverTestDependencies(t)

	publicKey, err := deps.resolver.ResolveTrustedPublicKey(u.OfficialMaintainer)

	assert.Nil(t, err)
	assert.Equal(t, ed25519.PublicKey(u.GetLocalTestingPublicKeyRaw()), publicKey)
}

func TestMaintainerKeyResolverResolveTrustedPublicKey_ThirdPartyReadsKnownKeyFromRepository(t *testing.T) {
	deps := setupMaintainerKeyResolverTestDependencies(t)
	deps.repo.EXPECT().GetPublicKey("samplemaintainer").Return(ed25519.PublicKey(u.GetOtherLocalTestingPublicKeyRaw()), true, nil)

	publicKey, err := deps.resolver.ResolveTrustedPublicKey("samplemaintainer")

	assert.Nil(t, err)
	assert.Equal(t, ed25519.PublicKey(u.GetOtherLocalTestingPublicKeyRaw()), publicKey)
}

func TestMaintainerKeyResolverRefreshTrustedPublicKey_InvalidKeySignatureReturnsSecurityError(t *testing.T) {
	deps := setupMaintainerKeyResolverTestDependencies(t)
	record, err := newSignedMaintainerPublicKeyRecord("samplemaintainer")
	assert.Nil(t, err)
	record.PublicKeySignature = []byte("invalid-signature")
	deps.client.EXPECT().GetMaintainerPublicKeyRecord("samplemaintainer").Return(record, nil)

	publicKey, err := deps.resolver.RefreshTrustedPublicKey("samplemaintainer")

	assert.Nil(t, publicKey)
	assert.NotNil(t, err)
	assert.Equal(t, SignatureVerificationFailedError, u.ExtractError(err))
}

func newSignedMaintainerPublicKeyRecord(maintainer string) (*store.MaintainerPublicKeyRecord, error) {
	officialPrivateKey, err := u.DecodeEd25519PrivateKeyOpenSSH([]byte(u.LocalTestingPrivateKeyOpenSSH), []byte(u.LocalTestingPrivateKeyPassphrase))
	if err != nil {
		return nil, err
	}
	maintainerPublicKey := u.GetOtherLocalTestingPublicKeyRaw()
	signature, err := store.SignMaintainerPublicKey(officialPrivateKey, maintainer, maintainerPublicKey)
	if err != nil {
		return nil, err
	}
	return &store.MaintainerPublicKeyRecord{
		Maintainer:         maintainer,
		PublicKeyRaw:       maintainerPublicKey,
		PublicKeySignature: signature,
	}, nil
}
