package tools

import (
	"slices"
	"strings"
)

var (
	RelativeBackupRepoPathInResticContainer = "backups"
	AbsoluteBackupRepoPathInResticContainer = "/" + RelativeBackupRepoPathInResticContainer
	BackupDockerVolumeName                  = "backups"
	ResticContainerRootDirVolume            = "quollix_restic_root"

	SshConfigName = "myssh"
)

type ResticContainerExecutor interface {
	// The encryption password must be passed explicitly because during backup or restore of the database, it is temporarily shut down and the password can no longer be retrieved from it. Therefore, the password is kept in memory by the backup/restore logic and injected for these operations.
	Execute(command []string, appVolumes, resticTags []string, resticEncryptionPassword, mountVolume string) (*CommandOutput, error)
}

type ResticContainerExecutorImpl struct {
	CommandRunner CommandRunner
}

func (r *ResticContainerExecutorImpl) Execute(command []string, appVolumes, resticTags []string, resticEncryptionPassword, mountVolume string) (*CommandOutput, error) {
	args := r.buildDockerArgs(command, appVolumes, resticTags, resticEncryptionPassword, mountVolume)
	return r.CommandRunner.RunCommand("docker", args...)
}

func (r *ResticContainerExecutorImpl) buildDockerArgs(command []string, appVolumes, resticTags []string, resticEncryptionPassword, mountVolume string) []string {
	args := []string{"run", "--rm"}
	args = append(
		args,
		"--label", ResticCleanupLabel,
		"--network", OfficialDatabaseAppNetworkName,
	)

	args = append(args, strings.Fields(mountVolume)...)
	args = append(args, "-v", BackupDockerVolumeName+":"+AbsoluteBackupRepoPathInResticContainer)
	for _, volume := range appVolumes {
		args = append(args, "-v", volume+":/source/"+volume)
	}
	args = append(
		args,
		"-e", "RESTIC_REPOSITORY=rclone:"+SshConfigName+":"+RelativeBackupRepoPathInResticContainer,
		"-e", "RESTIC_PASSWORD="+resticEncryptionPassword,
		"--entrypoint", "",
		"-v", ResticContainerRootDirVolume+":/root",
		ResticImageName,
	)

	commandWithTags := slices.Clone(command)
	for _, tag := range resticTags {
		commandWithTags = append(commandWithTags, "--tag", tag)
	}

	return append(args, commandWithTags...)
}
