//go:build integration

package repository

import (
	"crypto/ed25519"
	"testing"

	"github.com/quollix/common/assert"
	u "github.com/quollix/common/utils"
)

func TestMaintainerRepository_PublicKeyLifecycle(t *testing.T) {
	InitDeps()
	defer MaintainerRepo.Wipe()

	maintainer := "samplemaintainer"
	initialPublicKey := ed25519.PublicKey(u.GetOtherLocalTestingPublicKeyRaw())
	updatedPublicKey := ed25519.PublicKey(u.GetLocalTestingPublicKeyRaw())

	publicKeys, err := MaintainerRepo.ListPublicKeys()
	assert.Nil(t, err)
	assert.Equal(t, 0, len(publicKeys))

	publicKey, exists, err := MaintainerRepo.GetPublicKey(maintainer)
	assert.Nil(t, err)
	assert.False(t, exists)
	assert.Nil(t, publicKey)

	assert.Nil(t, MaintainerRepo.UpsertPublicKey(maintainer, initialPublicKey))
	assertMaintainerPublicKey(t, maintainer, initialPublicKey, u.OtherLocalTestingPublicKeyFingerprintSHA256)

	assert.Nil(t, MaintainerRepo.UpsertPublicKey(maintainer, updatedPublicKey))
	assertMaintainerPublicKey(t, maintainer, updatedPublicKey, u.LocalTestingPublicKeyFingerprintSHA256)

	assert.Nil(t, MaintainerRepo.DeletePublicKey(maintainer))
	publicKeys, err = MaintainerRepo.ListPublicKeys()
	assert.Nil(t, err)
	assert.Equal(t, 0, len(publicKeys))

	publicKey, exists, err = MaintainerRepo.GetPublicKey(maintainer)
	assert.Nil(t, err)
	assert.False(t, exists)
	assert.Nil(t, publicKey)
}

func assertMaintainerPublicKey(t *testing.T, maintainer string, expectedPublicKey ed25519.PublicKey, expectedFingerprint string) {
	actualPublicKey, exists, err := MaintainerRepo.GetPublicKey(maintainer)
	assert.Nil(t, err)
	assert.True(t, exists)
	assert.Equal(t, expectedPublicKey, actualPublicKey)

	publicKeys, err := MaintainerRepo.ListPublicKeys()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(publicKeys))
	assert.Equal(t, maintainer, publicKeys[0].Name)
	assert.Equal(t, expectedPublicKey, publicKeys[0].PublicKeyRaw)
	assert.Equal(t, expectedFingerprint, publicKeys[0].PublicKeyFingerprint)
	assert.False(t, publicKeys[0].LastUpdatedAt.IsZero())
}
