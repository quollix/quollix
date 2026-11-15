package alerts

import (
	"time"

	"github.com/quollix/common/quollix/api"
)

const (
	cpuUsageAlertName      = "Host CPU usage"
	memoryUsageAlertName   = "Host memory usage"
	diskUsageAlertName     = "Host disk usage"
	certificateAlertName   = "Certificate expiration"
	quollixErrorsAlertName = "Quollix errors"

	minimumHostUsagePercent        = 10
	maximumHostUsagePercent        = 100
	minimumHostUsagePendingMinutes = 0

	alertCheckInterval         = 5 * time.Minute
	certificateExpiryThreshold = 14 * 24 * time.Hour
	quollixErrorLookback       = 10 * time.Minute
)

const (
	defaultHostUsageTriggerPercent  = 90
	defaultHostUsageRecoveryPercent = 80
	defaultHostUsagePendingMinutes  = 30
)

func DefaultAlertSettings() *api.AlertSettings {
	return &api.AlertSettings{
		CPUUsageEnabled:          true,
		MemoryUsageEnabled:       true,
		DiskUsageEnabled:         true,
		CertificateExpiryEnabled: true,
		QuollixErrorsEnabled:     true,
		HostUsageTriggerPercent:  defaultHostUsageTriggerPercent,
		HostUsageRecoveryPercent: defaultHostUsageRecoveryPercent,
		HostUsagePendingMinutes:  defaultHostUsagePendingMinutes,
	}
}
