package backup_server

import (
	"fmt"
	"strings"

	"server/tools"

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
	output, err := s.ResticContainerExecutor.Execute([]string{"ssh-keyscan", "-p", port, host}, nil, nil, "", "")
	if output == nil {
		return "", err
	}
	return output.Combined(), err
}

func (s *SshClientImpl) TestWhetherSshAccessWorks(repo *tools.SshConnectionRequest) error {
	if _, err := s.ResticContainerExecutor.Execute([]string{"mkdir", "-p", SshDirLocation}, nil, nil, "", ""); err != nil {
		return err
	}

	if _, err := s.ResticContainerExecutor.Execute(writeKnownHostsCommand(repo.SshKnownHosts), nil, nil, "", ""); err != nil {
		return err
	}

	sshCommand := []string{
		"sshpass", "-p", repo.SshPassword,
		"ssh", "-p", repo.SshPort, repo.SshUser + "@" + repo.Host,
		"-o", "UserKnownHostsFile=" + KnownHostsFileLocation,
		"-o", "StrictHostKeyChecking=yes",
		"-o", "PreferredAuthentications=password",
		"-o", "PasswordAuthentication=yes",
		"-o", "BatchMode=no",
		"-o", "ConnectTimeout=1",
		"exit",
	}
	_, err := s.ResticContainerExecutor.Execute(sshCommand, nil, nil, "", "")
	if err != nil {
		return err
	}
	return nil
}

func writeKnownHostsCommand(knownHosts string) []string {
	return []string{
		"sh",
		"-c",
		fmt.Sprintf(`printf '%%s' "$1" > %s`, KnownHostsFileLocation),
		"sh",
		knownHosts,
	}
}

func rcloneConfigCreateCommand(repo *tools.SshConnectionRequest) []string {
	return []string{
		"rclone",
		"config",
		"create",
		tools.SshConfigName,
		"sftp",
		"host=" + repo.Host,
		"user=" + repo.SshUser,
		"pass=" + repo.SshPassword,
		"port=" + repo.SshPort,
		"known_hosts_file=" + KnownHostsFileLocation,
		"use_insecure_cipher=false",
	}
}

func rclonePurgeCommand() []string {
	return []string{
		"rclone",
		"purge",
		fmt.Sprintf("%s:%s", tools.SshConfigName, tools.RelativeBackupRepoPathInResticContainer),
	}
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

	output, err := r.ResticContainerExecutor.Execute([]string{"restic", "check"}, nil, nil, repo.EncryptionPassword, "")
	if err != nil {
		if output != nil && strings.Contains(output.Combined(), WrongEncryptionPasswordResticCommandOutputPattern) {
			return u.Logger.NewError(WrongEncryptionPasswordErr)
		}

		_, err = r.ResticContainerExecutor.Execute([]string{"restic", "init"}, nil, nil, repo.EncryptionPassword, "")
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
	_, err := r.ResticContainerExecutor.Execute(rcloneConfigCreateCommand(repo), nil, nil, resticEncryptionPassword, "")
	if err != nil {
		return err
	}

	_, err = r.ResticContainerExecutor.Execute(writeKnownHostsCommand(repo.SshKnownHosts), nil, nil, resticEncryptionPassword, "")
	return err
}

func (r *SshClientImpl) PurgeBackupServer(repo *tools.SshConnectionRequest) error {
	if err := r.TestWhetherSshAccessWorks(repo); err != nil {
		return err
	}

	if err := r.prepareConfigFilesShared(repo, ""); err != nil {
		return err
	}
	_, err := r.ResticContainerExecutor.Execute(rclonePurgeCommand(), nil, nil, "", "")
	return err
}
