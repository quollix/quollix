package tools

import (
	"fmt"
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
	Execute(command string, appVolumes, resticTags []string, resticEncryptionPassword, mountVolume string) (*CommandOutput, error)
	ExecuteSimple(command string) (*CommandOutput, error)
	ExecuteSimpleWithPassword(command, resticEncryptionPassword string) (*CommandOutput, error)
}

type ResticContainerExecutorImpl struct {
	CommandRunner CommandRunner
}

func (r *ResticContainerExecutorImpl) ExecuteSimpleWithPassword(command, resticEncryptionPassword string) (*CommandOutput, error) {
	return r.Execute(command, nil, nil, resticEncryptionPassword, "")
}

func (r *ResticContainerExecutorImpl) Execute(command string, appVolumes, resticTags []string, resticEncryptionPassword, mountVolume string) (*CommandOutput, error) {
	wholeCommand := r.buildCommand(command, appVolumes, resticTags, resticEncryptionPassword, mountVolume)
	return r.CommandRunner.RunCommand(wholeCommand)
}

func (r *ResticContainerExecutorImpl) buildCommand(command string, appVolumes, resticTags []string, resticEncryptionPassword, mountVolume string) string {
	envFlags := fmt.Sprintf("-e RESTIC_REPOSITORY=rclone:%s:%s -e RESTIC_PASSWORD=%s ", SshConfigName, RelativeBackupRepoPathInResticContainer, resticEncryptionPassword)
	resticTagsFlags := ""
	for _, tag := range resticTags {
		resticTagsFlags += `--tag ` + tag + ` `
	}
	volumeFlags := ""
	for _, volume := range appVolumes {
		volumeFlags += `-v ` + volume + `:/source/` + volume + ` `
	}

	// the "--network" is only needed for testing during development as the dummy SSH container is running in the same network
	wholeCommand := fmt.Sprintf(`docker run --rm --label %s --network %s %s -v %s:%s %s%s--entrypoint "" -v %s:/root %s sh -c "%s %s"`, ResticCleanupLabel, OfficialDatabaseAppNetworkName, mountVolume, BackupDockerVolumeName, AbsoluteBackupRepoPathInResticContainer, volumeFlags, envFlags, ResticContainerRootDirVolume, ResticImageName, command, resticTagsFlags)
	return wholeCommand
}

func (r *ResticContainerExecutorImpl) ExecuteSimple(command string) (*CommandOutput, error) {
	return r.Execute(command, nil, nil, "", "")
}
