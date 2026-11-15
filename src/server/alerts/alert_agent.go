package alerts

import (
	"github.com/quollix/common/quollix/api"
	u "github.com/quollix/common/utils"
)

type AlertAgent interface {
	StartAlertAgentDaemon()
}

type AlertAgentImpl struct {
	AlertChecks        AlertChecks
	SettingsRepository AlertSettingsRepository
	StateRepository    StateRepository
	OsWrapper          u.OsWrapper
}

func (a *AlertAgentImpl) StartAlertAgentDaemon() {
	go a.run()
}

func (a *AlertAgentImpl) run() {
	for {
		// Wait before the first check and between checks, giving other services time to start after a server restart.
		a.OsWrapper.Sleep(alertCheckInterval)
		a.CheckAlerts()
	}
}

func (a *AlertAgentImpl) CheckAlerts() {
	settings, err := a.SettingsRepository.Read()
	if err != nil {
		u.Logger.Error(err)
		return
	}
	state, err := a.StateRepository.Read()
	if err != nil {
		u.Logger.Error(err)
		return
	}
	if settings.CPUUsageEnabled {
		a.runHostUsageAlertCheck(a.AlertChecks.CheckHostCPUUsage, settings, state)
	}
	if settings.MemoryUsageEnabled {
		a.runHostUsageAlertCheck(a.AlertChecks.CheckHostMemoryUsage, settings, state)
	}
	if settings.DiskUsageEnabled {
		a.runHostUsageAlertCheck(a.AlertChecks.CheckHostDiskUsage, settings, state)
	}
	if settings.CertificateExpiryEnabled {
		a.runAlertCheck(a.AlertChecks.CheckCertificateExpiry, state)
	}
	if settings.QuollixErrorsEnabled {
		a.runAlertCheck(a.AlertChecks.CheckQuollixErrors, state)
	}
	if err := a.StateRepository.Save(state); err != nil {
		u.Logger.Error(err)
	}
}

func (a *AlertAgentImpl) runHostUsageAlertCheck(check func(*api.AlertSettings, *RuntimeState) error, settings *api.AlertSettings, state *RuntimeState) {
	if err := check(settings, state); err != nil {
		u.Logger.Error(err)
	}
}

func (a *AlertAgentImpl) runAlertCheck(check func(*RuntimeState) error, state *RuntimeState) {
	if err := check(state); err != nil {
		u.Logger.Error(err)
	}
}
