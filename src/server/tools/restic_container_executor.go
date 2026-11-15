package tools

import (
	"slices"
	"strings"
)

var (
	LegacyRemoteBackupFolderPath            = "./backups"
	RelativeBackupRepoPathInResticContainer = "backups"
	AbsoluteBackupRepoPathInResticContainer = "/" + RelativeBackupRepoPathInResticContainer
	BackupDockerVolumeName                  = "backups"
	ResticContainerRootDirVolume            = "quollix_restic_root"

	SshConfigName = "myssh"
)

type ResticContainerExecutor interface {
	// The encryption password must be passed explicitly because during backup or restore of the database, it is temporarily shut down and the password can no longer be retrieved from it. Therefore, the password is kept in memory by the backup/restore logic and injected for these operations.
	Execute(command []string, appVolumes, resticTags []string, resticEncryptionPassword, mountVolume, backupFolderPath string) (*CommandOutput, error)
}

type ResticContainerExecutorImpl struct {
	CommandRunner CommandRunner
}

func (r *ResticContainerExecutorImpl) Execute(command []string, appVolumes, resticTags []string, resticEncryptionPassword, mountVolume, backupFolderPath string) (*CommandOutput, error) {
	args := r.buildDockerArgs(command, appVolumes, resticTags, resticEncryptionPassword, mountVolume, backupFolderPath)
	return r.CommandRunner.RunCommand("docker", args...)
}

func (r *ResticContainerExecutorImpl) buildDockerArgs(command []string, appVolumes, resticTags []string, resticEncryptionPassword, mountVolume, backupFolderPath string) []string {
	// The 'restic' docker image is built locally, so never pull an unexpected image if that build had failed.
	args := []string{"run", "--rm", "--pull=never"}
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
	if backupFolderPath != "" {
		args = append(args, "-e", "RESTIC_REPOSITORY=rclone:"+SshConfigName+":"+backupFolderPath)
	}
	args = append(
		args,
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
