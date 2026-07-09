package tools

import (
	"fmt"
	"testing"

	"github.com/quollix/common/assert"
)

func TestExecuteInResticContainer(t *testing.T) {
	mountVolume := "-v /mnt/sample:/mnt/sample"
	expectedCommand := fmt.Sprintf(`docker run --rm --label %s --network %s %s -v %s:%s -v sample-vol-1:/source/sample-vol-1 -v sample-vol-2:/source/sample-vol-2 -e RESTIC_REPOSITORY=rclone:%s:backups -e RESTIC_PASSWORD=restic-encryption-password --entrypoint "" -v %s:/root %s sh -c "some-restic-command --tag tag1 --tag tag2 "`, ResticCleanupLabel, OfficialDatabaseAppNetworkName, mountVolume, BackupDockerVolumeName, AbsoluteBackupRepoPathInResticContainer, SshConfigName, ResticContainerRootDirVolume, ResticImageName)

	assertExecuteInResticContainer(
		t,
		"some-restic-command",
		[]string{"sample-vol-1", "sample-vol-2"},
		[]string{"tag1", "tag2"},
		"restic-encryption-password",
		"-v /mnt/sample:/mnt/sample",
		expectedCommand,
	)
}

func TestExecuteInResticContainer_NilInputs(t *testing.T) {
	expectedCommand := fmt.Sprintf(`docker run --rm --label %s --network %s %s -v %s:%s -e RESTIC_REPOSITORY=rclone:%s:backups -e RESTIC_PASSWORD= --entrypoint "" -v %s:/root %s sh -c "some-restic-command "`, ResticCleanupLabel, OfficialDatabaseAppNetworkName, "", BackupDockerVolumeName, AbsoluteBackupRepoPathInResticContainer, SshConfigName, ResticContainerRootDirVolume, ResticImageName)

	assertExecuteInResticContainer(t, "some-restic-command", nil, nil, "", "", expectedCommand)
}

func assertExecuteInResticContainer(
	t *testing.T,
	command string,
	appVolumes, resticTags []string,
	resticEncryptionPassword, mountVolume string,
	expectedCommand string,
) {
	commandRunnerMock := &CommandRunnerMock{}
	resticContainerAgent := ResticContainerExecutorImpl{
		CommandRunner: commandRunnerMock,
	}

	commandRunnerMock.EXPECT().RunCommand(expectedCommand).Return(&CommandOutput{Stdout: "output"}, nil)

	output, err := resticContainerAgent.Execute(command, appVolumes, resticTags, resticEncryptionPassword, mountVolume)
	assert.Nil(t, err)
	assert.Equal(t, "output", output.Combined())

	commandRunnerMock.AssertExpectations(t)
}
