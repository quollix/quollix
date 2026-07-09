//go:build special_passwords

package special_passwords

import (
	"strings"
	"testing"

	"server/backup_server"
	"server/tests/component"
	"server/tools"

	"github.com/quollix/common/assert"
)

const loosePolicyPassword = "Aa09!@#$%^&*()_-+=.,:;/?[]{}|~<>"

func TestBackupServerWithLoosePolicyPasswords(t *testing.T) {
	client := component.GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	_, err := component.InstallSample(t, client, "2.0")
	assert.Nil(t, err)

	repo := backup_server.GetSampleRemoteRepo()
	repo.SshPassword = loosePolicyPassword
	repo.EncryptionPassword = loosePolicyPassword

	knownHosts, err := client.Settings.GetKnownHosts(repo)
	assert.Nil(t, err)
	assert.True(t, strings.Contains(knownHosts, "["+repo.Host+"]:"+repo.SshPort))

	repo.SshKnownHosts = knownHosts
	assert.Nil(t, client.Settings.SaveSshConfigs(repo))
	assert.Nil(t, component.CreateSampleBackup(t, client))

	backups, err := client.Backups.ListByApp(tools.SampleMaintainer, tools.SampleApp)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(backups))
}
