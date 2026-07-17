package backup_server

import (
	api "github.com/quollix/common/quollix/api"
	u "github.com/quollix/common/utils"
)

const ErrBackupRepoNotConfigured = "backup is not enabled yet, please do this in the settings"

type SshRepositoryService interface {
	IsBackupEnabled() (bool, error)
	SetRemoteBackupRepository(backupRepo *api.BackupServerConfigs) error
	GetRemoteBackupRepository() (*api.BackupServerConfigs, error)
	EnsureBackupIsEnabled() error
}

type SshRepositoryServiceImpl struct {
	SshClient     SshClient
	SshRepository SshRepository
}

func (s *SshRepositoryServiceImpl) IsBackupEnabled() (bool, error) {
	return s.SshRepository.IsRemoteBackupEnabled()
}

func (s *SshRepositoryServiceImpl) GetRemoteBackupRepository() (*api.BackupServerConfigs, error) {
	return s.SshRepository.GetRemoteBackupRepository()
}

func (s *SshRepositoryServiceImpl) SetRemoteBackupRepository(backupRepo *api.BackupServerConfigs) error {
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
