package alerts

import (
	"fmt"
	"time"

	"server/certificates"

	"github.com/quollix/common/quollix/api"
	u "github.com/quollix/common/utils"
)

type AlertChecks interface {
	CheckHostCPUUsage(*api.AlertSettings, *RuntimeState) error
	CheckHostMemoryUsage(*api.AlertSettings, *RuntimeState) error
	CheckHostDiskUsage(*api.AlertSettings, *RuntimeState) error
	CheckCertificateExpiry(*RuntimeState) error
	CheckQuollixErrors(*RuntimeState) error
}

type AlertChecksImpl struct {
	Queries            VictoriaAlertQueries
	CertificateService certificates.CertificateService
	Notifier           Notifier
	OsWrapper          u.OsWrapper
}

func (c *AlertChecksImpl) CheckHostCPUUsage(settings *api.AlertSettings, state *RuntimeState) error {
	return c.checkHostUsage(cpuUsageAlertName, c.Queries.GetHostCPUUsage, &state.CPUUsage, settings)
}

func (c *AlertChecksImpl) CheckHostMemoryUsage(settings *api.AlertSettings, state *RuntimeState) error {
	return c.checkHostUsage(memoryUsageAlertName, c.Queries.GetHostMemoryUsage, &state.MemoryUsage, settings)
}

func (c *AlertChecksImpl) CheckHostDiskUsage(settings *api.AlertSettings, state *RuntimeState) error {
	return c.checkHostUsage(diskUsageAlertName, c.Queries.GetHostDiskUsage, &state.DiskUsage, settings)
}

func (c *AlertChecksImpl) CheckCertificateExpiry(state *RuntimeState) error {
	expirationDate, err := c.CertificateService.GetCurrentCertificateExpirationDate()
	if err != nil {
		return err
	}
	now := c.OsWrapper.Now().UTC()
	remaining := expirationDate.Sub(now)
	target := AlertStatusHealthy
	if remaining <= certificateExpiryThreshold {
		target = AlertStatusFiring
	}
	details := fmt.Sprintf(
		"Certificate expiration date: %s\nTime until expiration: %s",
		expirationDate.UTC().Format(time.RFC3339),
		remaining.Round(time.Second),
	)
	if target == AlertStatusFiring {
		details += "\nUpload a renewed certificate or generate one via DNS-01 in Settings."
	}
	return c.changeAlertState(certificateAlertName, &state.CertificateExpiry, target, details)
}

func (c *AlertChecksImpl) CheckQuollixErrors(state *RuntimeState) error {
	count, err := c.Queries.GetQuollixErrorCount()
	if err != nil {
		return err
	}
	target := AlertStatusHealthy
	if count > 0 {
		target = AlertStatusFiring
	}
	details := fmt.Sprintf("Quollix logged %d error or fatal events in the last %d minutes.", count, quollixErrorLookbackMinutes)
	if target == AlertStatusFiring {
		details += " Review Quollix logs in the Grafana app."
	}
	return c.changeAlertState(quollixErrorsAlertName, &state.QuollixErrors, target, details)
}

func (c *AlertChecksImpl) checkHostUsage(alertName string, query func() (float64, error), alertState *AlertRuntimeState, settings *api.AlertSettings) error {
	usagePercent, err := query()
	if err != nil {
		return err
	}

	now := c.OsWrapper.Now().UTC()
	if usagePercent > float64(settings.HostUsageTriggerPercent) {
		switch alertState.Status {
		case AlertStatusHealthy:
			alertState.PendingSince = now
			if settings.HostUsagePendingMinutes == 0 {
				return c.changeAlertState(alertName, alertState, AlertStatusFiring, hostUsageAlertDetails(settings, usagePercent, 0))
			}
			alertState.Status = AlertStatusPending
			u.Logger.Info("alert condition pending", "alert", alertName, "usage_percent", usagePercent, "pending_period_minutes", settings.HostUsagePendingMinutes)
			return nil
		case AlertStatusPending:
			elapsedMinutes := int(now.Sub(alertState.PendingSince) / time.Minute)
			if elapsedMinutes >= settings.HostUsagePendingMinutes {
				return c.changeAlertState(alertName, alertState, AlertStatusFiring, hostUsageAlertDetails(settings, usagePercent, elapsedMinutes))
			}
		}
		return nil
	}

	switch alertState.Status {
	case AlertStatusPending:
		alertState.Status = AlertStatusHealthy
		alertState.PendingSince = time.Time{}
		return nil
	case AlertStatusFiring:
		if usagePercent < float64(settings.HostUsageRecoveryPercent) {
			details := fmt.Sprintf("Host usage has recovered below %d%%: %.1f%%", settings.HostUsageRecoveryPercent, usagePercent)
			return c.changeAlertState(alertName, alertState, AlertStatusHealthy, details)
		}
	}
	return nil
}

func hostUsageAlertDetails(settings *api.AlertSettings, usagePercent float64, elapsedMinutes int) string {
	details := fmt.Sprintf("Host usage is above %d%%: %.1f%%.", settings.HostUsageTriggerPercent, usagePercent)
	if elapsedMinutes > 0 {
		details += fmt.Sprintf(" It has remained above the trigger for at least %d minutes.", elapsedMinutes)
	}
	return details + " Review host metrics in the Grafana app."
}

func (c *AlertChecksImpl) changeAlertState(alertName string, alertState *AlertRuntimeState, target AlertStatus, details string) error {
	if alertState.Status == target {
		return nil
	}
	if target == AlertStatusFiring || alertState.Status == AlertStatusFiring {
		if err := c.Notifier.Send(Notification{
			AlertName: alertName,
			Firing:    target == AlertStatusFiring,
			Details:   details,
		}); err != nil {
			return err
		}
	}
	alertState.Status = target
	alertState.PendingSince = time.Time{} // reset pending time
	return nil
}
