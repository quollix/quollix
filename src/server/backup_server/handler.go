package backup_server

import (
	"net/http"
	"server/apps_basic"
	"server/tools"

	u "github.com/quollix/common/utils"
	"github.com/quollix/common/validation"
)

var (
	CantBackupAppWithoutVolumes   = "can't backup app without volumes"
	BackupOperationExpectedErrors = u.MapOf(ErrBackupRepoNotConfigured, CantBackupAppWithoutVolumes)
	expectedSaveSshSettingsErrors = u.MapOf(WrongEncryptionPasswordErr)
)

type SshHandler struct {
	OperationRegistry          apps_basic.OperationRegistry
	SshRepositoryConfigService SshRepositoryService
	SshClient                  SshClient
}

func (s *SshHandler) SaveSshSettingsHandler(w http.ResponseWriter, r *http.Request) {
	remoteRepo, ok := validation.ReadBody[tools.BackupServerConfigs](w, r)
	if !ok {
		return
	}

	handle := s.OperationRegistry.RegisterOperation("preparing backup server")
	defer handle.Done()

	err := error(nil)
	err = s.SshRepositoryConfigService.SetRemoteBackupRepository(remoteRepo)
	if err != nil {
		u.WriteResponseError(w, expectedSaveSshSettingsErrors, err)
		return
	}
}

func (s *SshHandler) ReadSshSettingsHandler(w http.ResponseWriter, r *http.Request) {
	remoteRepo, err := s.SshRepositoryConfigService.GetRemoteBackupRepository()
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}
	u.SendJsonResponse(w, remoteRepo)
}

func (s *SshHandler) ResetSshSettingsHandler(w http.ResponseWriter, r *http.Request) {
	err := s.SshRepositoryConfigService.SetRemoteBackupRepository(getEmptyBackupServerConfigs())
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}
}

func getEmptyBackupServerConfigs() *tools.BackupServerConfigs {
	return &tools.BackupServerConfigs{
		IsEnabled: false,
	}
}

type KnownHostsRequest struct {
	Host string `json:"host" validate:"remote_host"`
	Port string `json:"port" validate:"number"`
}

func (s *SshHandler) GetKnownHostsHandler(w http.ResponseWriter, r *http.Request) {
	repo, ok := validation.ReadBody[KnownHostsRequest](w, r)
	if !ok {
		return
	}

	handle := s.OperationRegistry.RegisterOperation("getting known hosts from SSH host")
	defer handle.Done()

	knownHost, err := s.SshClient.GetKnownHosts(repo.Host, repo.Port)
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}

	u.SendJsonResponse(w, tools.SingleString{Value: knownHost})
}

func (s *SshHandler) TestSshAccessHandler(w http.ResponseWriter, r *http.Request) {
	request, ok := validation.ReadBody[tools.SshConnectionRequest](w, r)
	if !ok {
		return
	}

	err := s.SshClient.TestWhetherSshAccessWorks(request)
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}
}

func (b *SshHandler) PurgeBackupServerHandler(w http.ResponseWriter, r *http.Request) {
	request, ok := validation.ReadBody[tools.SshConnectionRequest](w, r)
	if !ok {
		return
	}

	handle := b.OperationRegistry.RegisterOperation("purging backup server")
	defer handle.Done()

	err := error(nil)
	if err = b.SshClient.PurgeBackupServer(request); err != nil {
		u.WriteResponseError(w, BackupOperationExpectedErrors, err)
		return
	}
}
