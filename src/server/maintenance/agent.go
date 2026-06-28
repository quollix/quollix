package maintenance

import (
	"server/apps_basic"
	"server/backup_server"
	"server/configs"
	"server/oidc_client"
	"server/users"

	"runtime/debug"
	"server/oidc_provider"
	"server/tools"
	"time"

	u "github.com/quollix/common/utils"
)

type MaintenanceAgent interface {
	StartMaintenanceAgentDaemon()
	RunMaintenanceJob()
}

const dockerImagePruneCommand = "docker image prune -af --filter label!=" + tools.ResticImageMaintenanceKeepLabel

type MaintenanceAgentImpl struct {
	Repository          configs.MaintenanceRepository
	ServiceHelper       AgentHelper
	AppRepo             apps_basic.AppRepository
	GlobalConfig        *tools.GlobalConfig
	OperationRegistry   apps_basic.OperationRegistry
	OsWrapper           u.OsWrapper
	CommandRunner       tools.CommandRunner
	OidcProviderCache   oidc_provider.OidcCache
	OidcLoginStateCache oidc_client.OidcLoginStateCache
	SessionRepo         users.SessionRepository
	ResticImage         backup_server.ResticDockerImageService
}

func (s *MaintenanceAgentImpl) StartMaintenanceAgentDaemon() {
	if !s.GlobalConfig.ShouldRunMaintenanceAgent {
		u.Logger.Info("maintenance agent is disabled, not starting maintenance agent daemon")
		return
	}

	go func() {
		for {
			s.tryRunningMaintenanceJobOnce()
			time.Sleep(1 * time.Minute)
		}
	}()
}

func (s *MaintenanceAgentImpl) tryRunningMaintenanceJobOnce() {
	defer func() {
		if recoveredValue := recover(); recoveredValue != nil {
			u.Logger.Error("maintenance job panicked", "panic_cause", recoveredValue, "stack", string(debug.Stack()))
		}
	}()
	s.considerRunningMaintenanceJob()
}

func (s *MaintenanceAgentImpl) considerRunningMaintenanceJob() {
	config, err := s.Repository.GetMaintenanceConfig()
	if err != nil {
		u.Logger.Error(err)
		return
	}

	now := s.OsWrapper.Now()
	if now.Before(config.NextMaintenanceAt) {
		return
	}

	newNextMaintenanceAt, err := s.ServiceHelper.CalculateNextMaintenanceAtUtc(now, config.IanaTimezone, config.MaintenanceWindowStartHour)
	if err != nil {
		u.Logger.Error(err)
		return
	}
	config.NextMaintenanceAt = *newNextMaintenanceAt
	if err := s.Repository.SetMaintenanceConfig(config); err != nil {
		u.Logger.Error(err)
		return
	}

	s.RunMaintenanceJob()
}

func (s *MaintenanceAgentImpl) RunMaintenanceJob() {
	u.Logger.Info("running maintenance job")
	s.ServiceHelper.UpdateAllApps()
	if err := s.ResticImage.UpdateResticDockerImage(); err != nil {
		u.Logger.Error(err)
	}
	s.ServiceHelper.CreateBackupsForAllAppsIfConfigured()
	s.ServiceHelper.RetentBackupsForAllAppsIfConfigured()
	s.OidcProviderCache.Cleanup()
	s.OidcLoginStateCache.CleanupExpiredLoginStates()
	if err := s.SessionRepo.DeleteExpiredSessions(); err != nil {
		u.Logger.Error(err)
	}
	if s.GlobalConfig.PruneDockerSystemDuringMaintenance {
		// Never prune Docker networks here. Pruning networks can accidentally delete the Postgres app network while Postgres is temporarily stopped for backup or restore, which can make Quollix lose database connectivity.
		if _, err := s.CommandRunner.RunCommand(dockerImagePruneCommand); err != nil {
			u.Logger.Error(err)
		}
	}
}
