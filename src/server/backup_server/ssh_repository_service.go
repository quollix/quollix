package backup_server

import (
	"server/tools"

	u "github.com/quollix/common/utils"
)

const ErrBackupRepoNotConfigured = "backup is not enabled yet, please do this in the settings"

type SshRepositoryService interface {
	IsBackupEnabled() (bool, error)
	SetRemoteBackupRepository(backupRepo *tools.BackupServerConfigs) error
	GetRemoteBackupRepository() (*tools.BackupServerConfigs, error)
	EnsureBackupIsEnabled() error
}

type SshRepositoryServiceImpl struct {
	SshClient     SshClient
	SshRepository SshRepository
}

func (s *SshRepositoryServiceImpl) IsBackupEnabled() (bool, error) {
	return s.SshRepository.IsRemoteBackupEnabled()
}

func (s *SshRepositoryServiceImpl) GetRemoteBackupRepository() (*tools.BackupServerConfigs, error) {
	return s.SshRepository.GetRemoteBackupRepository()
}

func (s *SshRepositoryServiceImpl) SetRemoteBackupRepository(backupRepo *tools.BackupServerConfigs) error {
	if backupRepo.IsEnabled {
		if err := s.SshClient.TestWhetherSshAccessWorks(backupRepo.ConvertToSshConnectionTestRequest()); err != nil {
			return err
		}
		if err := s.SshClient.PrepareBackupServer(backupRepo); err != nil {
			return err
		}
	}
	return s.SshRepository.SaveRemoteBackupRepository(backupRepo)
}

func (s *SshRepositoryServiceImpl) EnsureBackupIsEnabled() error {
	enabled, err := s.SshRepository.IsRemoteBackupEnabled()
	if err != nil {
		return err
	}
	if !enabled {
		return u.Logger.NewError(ErrBackupRepoNotConfigured)
	}
	return nil
}
