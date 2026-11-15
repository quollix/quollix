package alerts

import (
	"testing"

	"github.com/quollix/common/assert"
)

type alertAgentTestDependencies struct {
	Agent              *AlertAgentImpl
	Checks             *AlertChecksMock
	SettingsRepository *AlertSettingsRepositoryMock
	StateRepository    *StateRepositoryMock
}

func setupAlertAgentTestDependencies(t *testing.T) *alertAgentTestDependencies {
	checks := NewAlertChecksMock(t)
	settingsRepository := NewAlertSettingsRepositoryMock(t)
	stateRepository := NewStateRepositoryMock(t)
	agent := &AlertAgentImpl{
		AlertChecks:        checks,
		SettingsRepository: settingsRepository,
		StateRepository:    stateRepository,
	}

	return &alertAgentTestDependencies{
		Agent:              agent,
		Checks:             checks,
		SettingsRepository: settingsRepository,
		StateRepository:    stateRepository,
	}
}

func TestAlertAgent_CheckAlertsRunsEnabledChecksAndContinuesAfterErrors(t *testing.T) {
	dependencies := setupAlertAgentTestDependencies(t)
	settings := DefaultAlertSettings()
	state := DefaultRuntimeState()
	dependencies.SettingsRepository.EXPECT().Read().Return(settings, nil)
	dependencies.StateRepository.EXPECT().Read().Return(state, nil)
	dependencies.Checks.EXPECT().CheckHostCPUUsage(settings, state).Return(nil)
	dependencies.Checks.EXPECT().CheckHostMemoryUsage(settings, state).Return(nil)
	dependencies.Checks.EXPECT().CheckHostDiskUsage(settings, state).Return(nil)
	dependencies.Checks.EXPECT().CheckCertificateExpiry(state).Return(nil)
	dependencies.Checks.EXPECT().CheckQuollixErrors(state).Return(nil)
	dependencies.StateRepository.EXPECT().Save(state).Return(nil)

	dependencies.Agent.CheckAlerts()
	assert.Equal(t, AlertStatusHealthy, state.CPUUsage.Status)
}
