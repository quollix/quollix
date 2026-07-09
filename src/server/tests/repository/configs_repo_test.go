//go:build integration

package repository

import (
	"testing"

	"github.com/quollix/common/assert"
)

func TestConfigsSetGetAndUpdate(t *testing.T) {
	InitDeps()
	defer ConfigRepo.Wipe()
	key := "test-key-configs"
	assert.Nil(t, ConfigRepo.SetConfig(key, "value-1"))

	val, err := ConfigRepo.GetConfig(key)
	assert.Nil(t, err)
	assert.Equal(t, "value-1", val)

	assert.Nil(t, ConfigRepo.SetConfig(key, "value-2"))
	val, err = ConfigRepo.GetConfig(key)
	assert.Nil(t, err)
	assert.Equal(t, "value-2", val)
}

func TestConfigsDeleteKeyAndIsKeySet(t *testing.T) {
	InitDeps()
	defer ConfigRepo.Wipe()
	key := "test-key-configs-delete"
	assert.Nil(t, ConfigRepo.DeleteConfig(key))

	isSet, err := ConfigRepo.IsConfigSet(key)
	assert.Nil(t, err)
	assert.False(t, isSet)

	assert.Nil(t, ConfigRepo.SetConfig(key, "some-value"))

	isSet, err = ConfigRepo.IsConfigSet(key)
	assert.Nil(t, err)
	assert.True(t, isSet)

	assert.Nil(t, ConfigRepo.DeleteConfig(key))
	_, err = ConfigRepo.GetConfig(key)
	assert.NotNil(t, err)

	isSet, err = ConfigRepo.IsConfigSet(key)
	assert.Nil(t, err)
	assert.False(t, isSet)
}
