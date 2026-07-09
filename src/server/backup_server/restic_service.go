package backup_server

import (
	"fmt"
	"path/filepath"
	"server/tools"
	"strings"

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
	ListBackups() ([]tools.BackupInfo, error)
	DeleteBackup(backupId []string) error
	CreateBackup(sourceMount string, volumes []string, tags []string, encryptionPassword string) error
	RestoreFiles(backupId string, targetDir string, encryptionPassword string) error
	RestoreVolumes(backupId string, volumes []string, encryptionPassword string) error
	GetSnapshotInfo(backupId string, encryptionPassword string) (*tools.BackupInfo, error)
}

type ResticServiceImpl struct {
	ResticContainerExecutor tools.ResticContainerExecutor
	ResticSnapshotsParser   ResticSnapshotsParser
	SshRepository           SshRepository
	SshClient               SshClient
}

func (r *ResticServiceImpl) ListBackups() ([]tools.BackupInfo, error) {
	repo, err := r.SshRepository.GetRemoteBackupRepository()
	if err != nil {
		return nil, err
	}

	output, err := r.ResticContainerExecutor.Execute("restic snapshots --json", nil, nil, repo.EncryptionPassword, "")
	if err != nil {
		return nil, err
	}
	logResticStderr("restic snapshots --json", output.Stderr)
	return r.ResticSnapshotsParser.Parse(output.Stdout)
}

func (r *ResticServiceImpl) DeleteBackup(backupIds []string) error {
	repo, err := r.SshRepository.GetRemoteBackupRepository()
	if err != nil {
		return err
	}
	// The forget command only removes the snapshot reference; it does not immediately reclaim storage. We therefore add the --prune flag to physically remove unreferenced data from the repository.
	backupPruneCommand := "restic forget --prune " + strings.Join(backupIds, " ")
	_, err = r.ResticContainerExecutor.ExecuteSimpleWithPassword(backupPruneCommand, repo.EncryptionPassword)
	return err
}

func (r *ResticServiceImpl) CreateBackup(sourceMount string, volumes []string, tags []string, encryptionPassword string) error {
	_, err := r.ResticContainerExecutor.Execute("restic backup /source", volumes, tags, encryptionPassword, sourceMount)
	return err
}

func (r *ResticServiceImpl) RestoreFiles(backupId string, targetDir string, encryptionPassword string) error {
	cmd := fmt.Sprintf("restic restore %s --target / --include '%s' --include '%s'", backupId, MetaYmlFileName, DockerComposeYaml)
	mount := "-v " + targetDir + ":/source "
	_, err := r.ResticContainerExecutor.Execute(cmd, nil, nil, encryptionPassword, mount)
	return err
}

func (r *ResticServiceImpl) RestoreVolumes(backupId string, volumes []string, encryptionPassword string) error {
	_, err := r.ResticContainerExecutor.Execute("restic restore "+backupId+" --target /", volumes, nil, encryptionPassword, "")
	return err
}

func (r *ResticServiceImpl) GetSnapshotInfo(backupId string, encryptionPassword string) (*tools.BackupInfo, error) {
	command := "restic snapshots --json " + backupId
	output, err := r.ResticContainerExecutor.Execute(command, nil, nil, encryptionPassword, "")
	if err != nil {
		return nil, err
	}
	logResticStderr(command, output.Stderr)
	backupInfos, err := r.ResticSnapshotsParser.Parse(output.Stdout)
	if err != nil {
		return nil, err
	}
	if len(backupInfos) != 1 {
		return nil, u.Logger.NewError("expected exactly one backup info", tools.BackupNumberField, len(backupInfos), tools.BackupIdField, backupId)
	}
	return &backupInfos[0], nil
}

func logResticStderr(command, stderr string) {
	stderr = strings.TrimSpace(stderr)
	if stderr != "" {
		u.Logger.Info("restic wrote to stderr", "command", command, "stderr", stderr)
	}
}
