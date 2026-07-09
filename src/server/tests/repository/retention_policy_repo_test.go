//go:build integration

package repository

import (
	"server/maintenance/retention"
	"testing"

	"github.com/quollix/common/assert"
)

func TestRetentionPolicy(t *testing.T) {
	InitDeps()
	defer ConfigRepo.Wipe()

	isSet, err := RetentionRepo.IsRetentionPolicySet()
	assert.Nil(t, err)
	assert.False(t, isSet)

	_, err = RetentionRepo.GetRetentionPolicy()
	assert.NotNil(t, err)

	err = RetentionRepo.SetRetentionPolicy(retention.GetDefaultRetentionPolicy())
	assert.Nil(t, err)

	isSet, err = RetentionRepo.IsRetentionPolicySet()
	assert.Nil(t, err)
	assert.True(t, isSet)

	policy, err := RetentionRepo.GetRetentionPolicy()
	assert.Nil(t, err)
	assert.Equal(t, *retention.GetDefaultRetentionPolicy(), *policy)
}
