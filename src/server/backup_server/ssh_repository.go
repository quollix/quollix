package backup_server

import (
	"server/configs"
	"server/tools"

	u "github.com/quollix/common/utils"
)

type SshRepository interface {
	IsRemoteBackupEnabled() (bool, error)
	SaveRemoteBackupRepository(backupRepo *tools.BackupServerConfigs) error
	GetRemoteBackupRepository() (*tools.BackupServerConfigs, error)
	IsRemoteBackupConfigPresent() (bool, error)
}

type SshRepositoryImpl struct {
	DbProvider tools.DatabaseConnector
}

func (r *SshRepositoryImpl) IsRemoteBackupEnabled() (bool, error) {
	repo, err := r.GetRemoteBackupRepository()
	if err != nil {
		return false, err
	}
	return repo.IsEnabled, nil
}

func (r *SshRepositoryImpl) SaveRemoteBackupRepository(remoteBackupRepository *tools.BackupServerConfigs) error {
	isEnabledValue := "false"
	if remoteBackupRepository.IsEnabled {
		isEnabledValue = "true"
	}

	_, err := r.DbProvider.GetDB().Exec(`
INSERT INTO configs (key, value)
VALUES
	($1, $2),
	($3, $4),
	($5, $6),
	($7, $8),
	($9, $10),
	($11, $12),
	($13, $14)
ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value
`,
		configs.ConfigKeys.BackupEnabled, isEnabledValue,
		configs.ConfigKeys.BackupServerHost, remoteBackupRepository.Host,
		configs.ConfigKeys.BackupServerPort, remoteBackupRepository.SshPort,
		configs.ConfigKeys.BackupServerUser, remoteBackupRepository.SshUser,
		configs.ConfigKeys.BackupServerPassword, remoteBackupRepository.SshPassword,
		configs.ConfigKeys.BackupServerKnownHosts, remoteBackupRepository.SshKnownHosts,
		configs.ConfigKeys.BackupEncryptionPassword, remoteBackupRepository.EncryptionPassword,
	)
	if err != nil {
		return u.Logger.NewError(err.Error())
	}
	return nil
}

func (r *SshRepositoryImpl) GetRemoteBackupRepository() (*tools.BackupServerConfigs, error) {
	var repo tools.BackupServerConfigs
	var isEnabledString string

	err := r.DbProvider.GetDB().QueryRow(`
SELECT
	(SELECT value FROM configs WHERE key = $1),
	(SELECT value FROM configs WHERE key = $2),
	(SELECT value FROM configs WHERE key = $3),
	(SELECT value FROM configs WHERE key = $4),
	(SELECT value FROM configs WHERE key = $5),
	(SELECT value FROM configs WHERE key = $6),
	(SELECT value FROM configs WHERE key = $7);
`,
		configs.ConfigKeys.BackupEnabled,
		configs.ConfigKeys.BackupServerHost,
		configs.ConfigKeys.BackupServerPort,
		configs.ConfigKeys.BackupServerUser,
		configs.ConfigKeys.BackupServerPassword,
		configs.ConfigKeys.BackupServerKnownHosts,
		configs.ConfigKeys.BackupEncryptionPassword,
	).Scan(
		&isEnabledString,
		&repo.Host,
		&repo.SshPort,
		&repo.SshUser,
		&repo.SshPassword,
		&repo.SshKnownHosts,
		&repo.EncryptionPassword,
	)

	if err != nil {
		return nil, err
	}

	repo.IsEnabled = isEnabledString == "true"
	return &repo, nil
}

func (r *SshRepositoryImpl) IsRemoteBackupConfigPresent() (bool, error) {
	var count int

	err := r.DbProvider.GetDB().QueryRow(`
SELECT COUNT(*) FROM configs
WHERE key IN ($1,$2,$3,$4,$5,$6,$7);
`,
		configs.ConfigKeys.BackupEnabled,
		configs.ConfigKeys.BackupServerHost,
		configs.ConfigKeys.BackupServerPort,
		configs.ConfigKeys.BackupServerUser,
		configs.ConfigKeys.BackupServerPassword,
		configs.ConfigKeys.BackupServerKnownHosts,
		configs.ConfigKeys.BackupEncryptionPassword,
	).Scan(&count)

	if err != nil {
		return false, err
	}

	return count == 7, nil
}

func GetEmptyRemoteRepoConfigs() *tools.BackupServerConfigs {
	return &tools.BackupServerConfigs{
		IsEnabled:          false,
		Host:               "",
		SshPort:            "",
		SshUser:            "",
		SshPassword:        "",
		EncryptionPassword: "",
	}
}
