//go:build integration

package repository

import (
	"server/backup_server"
	"server/tools"
	"testing"

	"github.com/quollix/common/assert"
)

func TestSshRepo(t *testing.T) {
	InitDeps()
	defer ConfigRepo.Wipe()

	isConfigPresent, err := SshRepo.IsRemoteBackupConfigPresent()
	assert.Nil(t, err)
	assert.False(t, isConfigPresent)

	_, err = SshRepo.GetRemoteBackupRepository()
	assert.NotNil(t, err)

	sample := backup_server.GetSampleRemoteRepo()
	assert.Nil(t, SshRepo.SaveRemoteBackupRepository(sample))
	assertRepoState(t, sample)

	isConfigPresent, err = SshRepo.IsRemoteBackupConfigPresent()
	assert.Nil(t, err)
	assert.True(t, isConfigPresent)

	sample.IsEnabled = false
	assert.Nil(t, SshRepo.SaveRemoteBackupRepository(sample))
	assertRepoState(t, sample)
}

func assertRepoState(t *testing.T, expected *tools.BackupServerConfigs) {
	repo, err := SshRepo.GetRemoteBackupRepository()
	assert.Nil(t, err)
	assert.Equal(t, expected, repo)

	isEnabled, err := SshRepo.IsRemoteBackupEnabled()
	assert.Nil(t, err)
	assert.Equal(t, expected.IsEnabled, isEnabled)
}
