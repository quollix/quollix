package backup_server

import (
	"path/filepath"
	"strings"

	"server/tools"

	api "github.com/quollix/common/quollix/api"
	u "github.com/quollix/common/utils"
)

const (
	MetaYmlFileName   = "meta.yml"
	DockerComposeYaml = "docker-compose.yml"
)

var (
	SshDirLocation         = "/root/.ssh"
	KnownHostsFileLocation = filepath.Join(SshDirLocation, "known_hosts")
)

type ResticService interface {
	ListBackups() ([]api.BackupInfo, error)
	DeleteBackup(backupId []string) error
	CreateBackup(sourceMount string, volumes []string, tags []string, encryptionPassword string) error
	RestoreFiles(backupId string, targetDir string, encryptionPassword string) error
	RestoreVolumes(backupId string, volumes []string, encryptionPassword string) error
	GetSnapshotInfo(backupId string, encryptionPassword string) (*api.BackupInfo, error)
}

type ResticServiceImpl struct {
	ResticContainerExecutor tools.ResticContainerExecutor
	ResticSnapshotsParser   ResticSnapshotsParser
	SshRepository           SshRepository
	SshClient               SshClient
}

func (r *ResticServiceImpl) ListBackups() ([]api.BackupInfo, error) {
	repo, err := r.SshRepository.GetRemoteBackupRepository()
	if err != nil {
		return nil, err
	}

	command := []string{"restic", "snapshots", "--json"}
	output, err := r.ResticContainerExecutor.Execute(command, nil, nil, repo.EncryptionPassword, "")
	if err != nil {
		return nil, err
	}
	logResticStderr(output.Stderr)
	return r.ResticSnapshotsParser.Parse(output.Stdout)
}

func (r *ResticServiceImpl) DeleteBackup(backupIds []string) error {
	repo, err := r.SshRepository.GetRemoteBackupRepository()
	if err != nil {
		return err
	}
	// The forget command only removes the snapshot reference; it does not immediately reclaim storage. We therefore add the --prune flag to physically remove unreferenced data from the repository.
	command := append([]string{"restic", "forget", "--prune"}, backupIds...)
	_, err = r.ResticContainerExecutor.Execute(command, nil, nil, repo.EncryptionPassword, "")
	return err
}

func (r *ResticServiceImpl) CreateBackup(sourceMount string, volumes []string, tags []string, encryptionPassword string) error {
	_, err := r.ResticContainerExecutor.Execute([]string{"restic", "backup", "/source"}, volumes, tags, encryptionPassword, sourceMount)
	return err
}

func (r *ResticServiceImpl) RestoreFiles(backupId string, targetDir string, encryptionPassword string) error {
	command := []string{"restic", "restore", backupId, "--target", "/", "--include", MetaYmlFileName, "--include", DockerComposeYaml}
	mount := "-v " + targetDir + ":/source "
	_, err := r.ResticContainerExecutor.Execute(command, nil, nil, encryptionPassword, mount)
	return err
}

func (r *ResticServiceImpl) RestoreVolumes(backupId string, volumes []string, encryptionPassword string) error {
	_, err := r.ResticContainerExecutor.Execute([]string{"restic", "restore", backupId, "--target", "/"}, volumes, nil, encryptionPassword, "")
	return err
}

func (r *ResticServiceImpl) GetSnapshotInfo(backupId string, encryptionPassword string) (*api.BackupInfo, error) {
	command := []string{"restic", "snapshots", "--json", backupId}
	output, err := r.ResticContainerExecutor.Execute(command, nil, nil, encryptionPassword, "")
	if err != nil {
		return nil, err
	}
	logResticStderr(output.Stderr)
	backupInfos, err := r.ResticSnapshotsParser.Parse(output.Stdout)
	if err != nil {
		return nil, err
	}
	if len(backupInfos) != 1 {
		return nil, u.Logger.NewError("expected exactly one backup info", tools.BackupNumberField, len(backupInfos), tools.BackupIdField, backupId)
	}
	return &backupInfos[0], nil
}

func logResticStderr(stderr string) {
	stderr = strings.TrimSpace(stderr)
	if stderr != "" {
		u.Logger.Info("restic wrote to stderr", "stderr", stderr)
	}
}
