//go:build victoria

package victoria

import (
	"math"
	"testing"
	"time"

	"server/metrics"

	"github.com/quollix/common/assert"
	"github.com/quollix/common/quollix/api"
	u "github.com/quollix/common/utils"
)

func TestVictoriaAlertQueries_HostUsageQueries(t *testing.T) {
	clients := newVictoriaTestClients(t)
	now := time.Now()
	// VictoriaMetrics hides recently ingested samples for its default 30-second latency offset.
	cpuStart := now.Add(-4 * time.Minute)
	cpuEnd := now.Add(-40 * time.Second)
	assert.Nil(t, clients.Metrics.Push([]api.MetricSeries{
		{
			Metric:     map[string]string{metrics.MetricNameLabel: "host_cpu_seconds_total", "mode": "idle"},
			Values:     []float64{100, 140},
			Timestamps: []int64{cpuStart.UnixMilli(), cpuEnd.UnixMilli()},
		},
		{
			Metric:     map[string]string{metrics.MetricNameLabel: "host_memory_available_bytes"},
			Values:     []float64{25},
			Timestamps: []int64{cpuEnd.UnixMilli()},
		},
		{
			Metric:     map[string]string{metrics.MetricNameLabel: "host_memory_total_bytes"},
			Values:     []float64{100},
			Timestamps: []int64{cpuEnd.UnixMilli()},
		},
		{
			Metric:     map[string]string{metrics.MetricNameLabel: "host_filesystem_used_ratio"},
			Values:     []float64{0.42},
			Timestamps: []int64{cpuEnd.UnixMilli()},
		},
	}))

	assertHostUsage(t, clients.AlertQueries.GetHostCPUUsage, 80)
	assertHostUsage(t, clients.AlertQueries.GetHostMemoryUsage, 75)
	assertHostUsage(t, clients.AlertQueries.GetHostDiskUsage, 42)
}

func TestVictoriaAlertQueries_QuollixErrorCount(t *testing.T) {
	clients := newVictoriaTestClients(t)
	assert.Nil(t, clients.Logs.Push([]api.LogRecord{
		newQuollixLogRecord("error", "first error"),
		newQuollixLogRecord("fatal", "fatal error"),
		newQuollixLogRecord("warning", "warning is not counted"),
		{
			ReadTimestamp: time.Now().UTC(),
			Level:         "error",
			Message:       "error from another app is not counted",
			Maintainer:    "other",
			App:           "other",
			Service:       "other",
			Parsed:        true,
		},
	}))

	var errorCount int
	err := u.EventuallyWithTimeout(20*time.Second, 200*time.Millisecond, func() error {
		var err error
		errorCount, err = clients.AlertQueries.GetQuollixErrorCount()
		if err == nil && errorCount != 2 {
			return u.Logger.NewError("waiting for two Quollix error logs", "count", errorCount)
		}
		return err
	})
	assert.Nil(t, err)
	assert.Equal(t, 2, errorCount)
}

func assertHostUsage(t *testing.T, query func() (float64, error), expectedPercent float64) {
	var usage float64
	err := u.EventuallyWithTimeout(20*time.Second, 200*time.Millisecond, func() error {
		var err error
		usage, err = query()
		return err
	})
	assert.Nil(t, err)
	assert.True(t, math.Abs(usage-expectedPercent) < 0.05)
}

func newQuollixLogRecord(level, message string) api.LogRecord {
	return api.LogRecord{
		ReadTimestamp: time.Now().UTC(),
		Level:         level,
		Message:       message,
		Maintainer:    "quollix",
		App:           "quollix",
		Service:       "quollix",
		Parsed:        true,
	}
}
