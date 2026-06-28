package backup_server

import (
	"fmt"
	"server/tools"
	"strings"

	u "github.com/quollix/common/utils"
)

var (
	WrongEncryptionPasswordErr                        = "The backup repository already exists, but the provided encryption password is incorrect. Please enter the correct password. If you have forgotten it, the only option is to purge the backup server." // #nosec G101 (CWE-798): Potential hardcoded credentials
	WrongEncryptionPasswordResticCommandOutputPattern = "wrong password or no key found"
)

type SshClient interface {
	GetKnownHosts(host, port string) (string, error)
	TestWhetherSshAccessWorks(repo *tools.SshConnectionRequest) error
	PrepareBackupServer(repo *tools.BackupServerConfigs) error
	PurgeBackupServer(repo *tools.SshConnectionRequest) error
}

type SshClientImpl struct {
	ResticContainerExecutor tools.ResticContainerExecutor
}

func (s *SshClientImpl) GetKnownHosts(host, port string) (string, error) {
	command := fmt.Sprintf("ssh-keyscan -p %s %s", port, host)
	output, err := s.ResticContainerExecutor.ExecuteSimple(command)
	if output == nil {
		return "", err
	}
	return output.Combined(), err
}

func (s *SshClientImpl) TestWhetherSshAccessWorks(repo *tools.SshConnectionRequest) error {
	mkdirCommand := fmt.Sprintf("mkdir -p %s", SshDirLocation)
	if _, err := s.ResticContainerExecutor.ExecuteSimple(mkdirCommand); err != nil {
		return err
	}

	writeKnownHostsCommand := fmt.Sprintf("printf '%s' > %s", repo.SshKnownHosts, KnownHostsFileLocation)
	if _, err := s.ResticContainerExecutor.ExecuteSimple(writeKnownHostsCommand); err != nil {
		return err
	}

	sshCommand := fmt.Sprintf(
		"sshpass -p %s ssh -p %s %s@%s -o UserKnownHostsFile=%s -o StrictHostKeyChecking=yes -o PreferredAuthentications=password -o PasswordAuthentication=yes -o BatchMode=no -o ConnectTimeout=1 exit",
		repo.SshPassword,
		repo.SshPort,
		repo.SshUser,
		repo.Host,
		KnownHostsFileLocation,
	)
	_, err := s.ResticContainerExecutor.ExecuteSimple(sshCommand)
	if err != nil {
		return err
	}
	return nil
}

func GetSampleRemoteRepo() *tools.BackupServerConfigs {
	return &tools.BackupServerConfigs{
		IsEnabled:          true,
		Host:               tools.TestSshServerHost,
		SshPort:            tools.TestSshServerPort,
		SshUser:            "sshadmin",
		SshPassword:        "sshpassword",
		SshKnownHosts:      "sample-value",
		EncryptionPassword: "restic-password",
	}
}

func (r *SshClientImpl) PrepareBackupServer(repo *tools.BackupServerConfigs) error {
	if err := r.prepareConfigFiles(repo); err != nil {
		return err
	}

	output, err := r.ResticContainerExecutor.ExecuteSimpleWithPassword("restic check", repo.EncryptionPassword)
	if err != nil {
		if output != nil && strings.Contains(output.Combined(), WrongEncryptionPasswordResticCommandOutputPattern) {
			return u.Logger.NewError(WrongEncryptionPasswordErr)
		}

		_, err = r.ResticContainerExecutor.ExecuteSimpleWithPassword("restic init", repo.EncryptionPassword)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *SshClientImpl) prepareConfigFiles(repo *tools.BackupServerConfigs) error {
	sshConnection := &tools.SshConnectionRequest{
		Host:          repo.Host,
		SshPort:       repo.SshPort,
		SshUser:       repo.SshUser,
		SshPassword:   repo.SshPassword,
		SshKnownHosts: repo.SshKnownHosts,
	}
	return r.prepareConfigFilesShared(sshConnection, repo.EncryptionPassword)
}

func (r *SshClientImpl) prepareConfigFilesShared(repo *tools.SshConnectionRequest, resticEncryptionPassword string) error {
	// Use rclone's SFTP backend for SSH-based repositories. SFTP is the SSH File
	// Transfer Protocol, not FTPS/FTP, so a regular SSH server is enough when it
	// exposes the SFTP subsystem. Modern scp clients use SFTP internally by default,
	// and rclone/restic expose SFTP as the supported reusable remote backend rather
	// than the legacy SCP/RCP protocol.
	rcloneSetupCmd := fmt.Sprintf(
		"rclone config create %s sftp host=%s user=%s pass=%s port=%s known_hosts_file=%s use_insecure_cipher=false",
		tools.SshConfigName,
		repo.Host,
		repo.SshUser,
		repo.SshPassword,
		repo.SshPort,
		KnownHostsFileLocation,
	)
	_, err := r.ResticContainerExecutor.ExecuteSimpleWithPassword(rcloneSetupCmd, resticEncryptionPassword)
	if err != nil {
		return err
	}

	knownHostsFileCreationCmd := fmt.Sprintf("printf '%s' > %s", repo.SshKnownHosts, KnownHostsFileLocation)
	_, err = r.ResticContainerExecutor.ExecuteSimpleWithPassword(knownHostsFileCreationCmd, resticEncryptionPassword)
	return err
}

func (r *SshClientImpl) PurgeBackupServer(repo *tools.SshConnectionRequest) error {
	if err := r.TestWhetherSshAccessWorks(repo); err != nil {
		return err
	}

	if err := r.prepareConfigFilesShared(repo, ""); err != nil {
		return err
	}
	purgeCmd := fmt.Sprintf("rclone purge %s:%s", tools.SshConfigName, tools.RelativeBackupRepoPathInResticContainer)
	_, err := r.ResticContainerExecutor.ExecuteSimple(purgeCmd)
	return err
}
