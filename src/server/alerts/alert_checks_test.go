package alerts

import (
	"testing"
	"time"

	"server/certificates"
	"server/tools"

	"github.com/quollix/common/assert"
	"github.com/stretchr/testify/mock"
)

var alertChecksTestNow = time.Date(2026, 11, 7, 12, 0, 0, 0, time.UTC)

type alertChecksTestDependencies struct {
	Checks             *AlertChecksImpl
	Queries            *VictoriaAlertQueriesMock
	CertificateService *certificates.CertificateServiceMock
	Notifier           *NotifierMock
	OsWrapper          *tools.CommonOsWrapperMock
}

func setupAlertChecksTestDependencies(t *testing.T) *alertChecksTestDependencies {
	queries := NewVictoriaAlertQueriesMock(t)
	certificateService := certificates.NewCertificateServiceMock(t)
	notifier := NewNotifierMock(t)
	osWrapper := tools.NewCommonOsWrapperMock(t)
	checks := &AlertChecksImpl{
		Queries:            queries,
		CertificateService: certificateService,
		Notifier:           notifier,
		OsWrapper:          osWrapper,
	}

	return &alertChecksTestDependencies{
		Checks:             checks,
		Queries:            queries,
		CertificateService: certificateService,
		Notifier:           notifier,
		OsWrapper:          osWrapper,
	}
}

func TestAlertChecks_HostUsagePendingFireAndRecovery(t *testing.T) {
	dependencies := setupAlertChecksTestDependencies(t)
	settings := DefaultAlertSettings()
	pendingPeriod := time.Duration(settings.HostUsagePendingMinutes) * time.Minute
	state := DefaultRuntimeState()
	dependencies.Queries.EXPECT().GetHostCPUUsage().Return(95, nil).Once()
	dependencies.OsWrapper.EXPECT().Now().Return(alertChecksTestNow).Once()
	assert.Nil(t, dependencies.Checks.CheckHostCPUUsage(settings, state))
	assert.Equal(t, AlertRuntimeState{Status: AlertStatusPending, PendingSince: alertChecksTestNow}, state.CPUUsage)

	dependencies.Queries.EXPECT().GetHostCPUUsage().Return(95, nil).Once()
	dependencies.OsWrapper.EXPECT().Now().Return(alertChecksTestNow.Add(pendingPeriod + time.Minute)).Once()
	dependencies.Notifier.EXPECT().Send(mock.MatchedBy(func(notification Notification) bool {
		return notification.AlertName == cpuUsageAlertName && notification.Firing == true
	})).Return(nil)
	assert.Nil(t, dependencies.Checks.CheckHostCPUUsage(settings, state))
	assert.Equal(t, AlertRuntimeState{Status: AlertStatusFiring}, state.CPUUsage)

	dependencies.Queries.EXPECT().GetHostCPUUsage().Return(85, nil).Once()
	dependencies.OsWrapper.EXPECT().Now().Return(alertChecksTestNow.Add(pendingPeriod + 2*time.Minute)).Once()
	assert.Nil(t, dependencies.Checks.CheckHostCPUUsage(settings, state))
	assert.Equal(t, AlertRuntimeState{Status: AlertStatusFiring}, state.CPUUsage)

	dependencies.Queries.EXPECT().GetHostCPUUsage().Return(79, nil).Once()
	dependencies.OsWrapper.EXPECT().Now().Return(alertChecksTestNow.Add(pendingPeriod + 3*time.Minute)).Once()
	dependencies.Notifier.EXPECT().Send(mock.MatchedBy(func(notification Notification) bool {
		return notification.AlertName == cpuUsageAlertName && notification.Firing == false
	})).Return(nil)
	assert.Nil(t, dependencies.Checks.CheckHostCPUUsage(settings, state))
	assert.Equal(t, AlertRuntimeState{Status: AlertStatusHealthy}, state.CPUUsage)
}

func TestAlertChecks_HostUsageReturnsToHealthyWhenHighUsageDoesNotPersist(t *testing.T) {
	dependencies := setupAlertChecksTestDependencies(t)
	settings := DefaultAlertSettings()
	state := DefaultRuntimeState()
	dependencies.Queries.EXPECT().GetHostMemoryUsage().Return(95, nil).Once()
	dependencies.Queries.EXPECT().GetHostMemoryUsage().Return(85, nil).Once()
	dependencies.OsWrapper.EXPECT().Now().Return(alertChecksTestNow)

	assert.Nil(t, dependencies.Checks.CheckHostMemoryUsage(settings, state))
	assert.Nil(t, dependencies.Checks.CheckHostMemoryUsage(settings, state))
	assert.Equal(t, AlertRuntimeState{Status: AlertStatusHealthy}, state.MemoryUsage)
}

func TestAlertChecks_HostUsageFiresImmediatelyWhenPendingTimeIsZero(t *testing.T) {
	dependencies := setupAlertChecksTestDependencies(t)
	settings := DefaultAlertSettings()
	settings.HostUsagePendingMinutes = 0
	state := DefaultRuntimeState()
	dependencies.Queries.EXPECT().GetHostCPUUsage().Return(95, nil)
	dependencies.OsWrapper.EXPECT().Now().Return(alertChecksTestNow)
	dependencies.Notifier.EXPECT().Send(mock.MatchedBy(func(notification Notification) bool {
		return notification.AlertName == cpuUsageAlertName && notification.Firing
	})).Return(nil)

	assert.Nil(t, dependencies.Checks.CheckHostCPUUsage(settings, state))
	assert.Equal(t, AlertRuntimeState{Status: AlertStatusFiring}, state.CPUUsage)
}

func TestAlertChecks_CertificateExpiryFiresAndRecovers(t *testing.T) {
	dependencies := setupAlertChecksTestDependencies(t)
	state := DefaultRuntimeState()
	expirationDate := alertChecksTestNow.Add(certificateExpiryThreshold - time.Hour)
	dependencies.CertificateService.EXPECT().GetCurrentCertificateExpirationDate().Return(expirationDate, nil).Once()
	dependencies.OsWrapper.EXPECT().Now().Return(alertChecksTestNow)
	dependencies.Notifier.EXPECT().Send(mock.MatchedBy(func(notification Notification) bool {
		return notification.AlertName == certificateAlertName && notification.Firing
	})).Return(nil)
	assert.Nil(t, dependencies.Checks.CheckCertificateExpiry(state))
	assert.Equal(t, AlertStatusFiring, state.CertificateExpiry.Status)

	dependencies.CertificateService.EXPECT().GetCurrentCertificateExpirationDate().Return(alertChecksTestNow.Add(certificateExpiryThreshold+30*24*time.Hour), nil).Once()
	dependencies.OsWrapper.EXPECT().Now().Return(alertChecksTestNow.Add(2 * time.Hour))
	dependencies.Notifier.EXPECT().Send(mock.MatchedBy(func(notification Notification) bool {
		return notification.AlertName == certificateAlertName && notification.Firing == false
	})).Return(nil)
	assert.Nil(t, dependencies.Checks.CheckCertificateExpiry(state))
	assert.Equal(t, AlertStatusHealthy, state.CertificateExpiry.Status)
}

func TestAlertChecks_QuollixErrorsFireOnlyWhenErrorsAreFound(t *testing.T) {
	dependencies := setupAlertChecksTestDependencies(t)
	state := DefaultRuntimeState()
	dependencies.Queries.EXPECT().GetQuollixErrorCount().Return(2, nil).Once()
	dependencies.Notifier.EXPECT().Send(mock.MatchedBy(func(notification Notification) bool {
		return notification.AlertName == quollixErrorsAlertName && notification.Firing
	})).Return(nil)
	assert.Nil(t, dependencies.Checks.CheckQuollixErrors(state))
	assert.Equal(t, AlertStatusFiring, state.QuollixErrors.Status)

	dependencies.Queries.EXPECT().GetQuollixErrorCount().Return(0, nil).Once()
	dependencies.Notifier.EXPECT().Send(mock.MatchedBy(func(notification Notification) bool {
		return notification.AlertName == quollixErrorsAlertName && !notification.Firing
	})).Return(nil)
	assert.Nil(t, dependencies.Checks.CheckQuollixErrors(state))
	assert.Equal(t, AlertStatusHealthy, state.QuollixErrors.Status)
}
