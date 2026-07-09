package tools

import (
	"testing"

	"github.com/quollix/common/assert"
)

func TestExecuteInResticContainer(t *testing.T) {
	expectedArgs := []string{
		"run", "--rm",
		"--label", ResticCleanupLabel,
		"--network", OfficialDatabaseAppNetworkName,
		"-v", "/mnt/sample:/mnt/sample",
		"-v", BackupDockerVolumeName + ":" + AbsoluteBackupRepoPathInResticContainer,
		"-v", "sample-vol-1:/source/sample-vol-1",
		"-v", "sample-vol-2:/source/sample-vol-2",
		"-e", "RESTIC_REPOSITORY=rclone:" + SshConfigName + ":backups",
		"-e", "RESTIC_PASSWORD=restic-encryption-password",
		"--entrypoint", "",
		"-v", ResticContainerRootDirVolume + ":/root",
		ResticImageName,
		"some-restic-command",
		"--tag", "tag1",
		"--tag", "tag2",
	}

	assertExecuteInResticContainer(
		t,
		[]string{"some-restic-command"},
		[]string{"sample-vol-1", "sample-vol-2"},
		[]string{"tag1", "tag2"},
		"restic-encryption-password",
		"-v /mnt/sample:/mnt/sample",
		expectedArgs,
	)
}

func TestExecuteInResticContainer_NilInputs(t *testing.T) {
	expectedArgs := []string{
		"run", "--rm",
		"--label", ResticCleanupLabel,
		"--network", OfficialDatabaseAppNetworkName,
		"-v", BackupDockerVolumeName + ":" + AbsoluteBackupRepoPathInResticContainer,
		"-e", "RESTIC_REPOSITORY=rclone:" + SshConfigName + ":backups",
		"-e", "RESTIC_PASSWORD=",
		"--entrypoint", "",
		"-v", ResticContainerRootDirVolume + ":/root",
		ResticImageName,
		"some-restic-command",
	}

	assertExecuteInResticContainer(t, []string{"some-restic-command"}, nil, nil, "", "", expectedArgs)
}

func assertExecuteInResticContainer(
	t *testing.T,
	command []string,
	appVolumes, resticTags []string,
	resticEncryptionPassword, mountVolume string,
	expectedArgs []string,
) {
	commandRunnerMock := &CommandRunnerMock{}
	resticContainerAgent := ResticContainerExecutorImpl{
		CommandRunner: commandRunnerMock,
	}

	commandRunnerMock.EXPECT().RunCommand("docker", expectedArgs).Return(&CommandOutput{Stdout: "output"}, nil)

	output, err := resticContainerAgent.Execute(command, appVolumes, resticTags, resticEncryptionPassword, mountVolume)
	assert.Nil(t, err)
	assert.Equal(t, "output", output.Combined())

	commandRunnerMock.AssertExpectations(t)
}
