package alerts

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"server/logging"
	"server/metrics"

	u "github.com/quollix/common/utils"
)

type VictoriaAlertQueries interface {
	GetHostCPUUsage() (float64, error)
	GetHostMemoryUsage() (float64, error)
	GetHostDiskUsage() (float64, error)
	GetQuollixErrorCount() (int, error)
}

type VictoriaAlertQueriesImpl struct {
	victoriaMetrics metrics.VictoriaMetricsClient
	victoriaLogs    logging.VictoriaLogsClient
}

func NewVictoriaAlertQueries(victoriaMetrics metrics.VictoriaMetricsClient, victoriaLogs logging.VictoriaLogsClient) *VictoriaAlertQueriesImpl {
	return &VictoriaAlertQueriesImpl{victoriaMetrics: victoriaMetrics, victoriaLogs: victoriaLogs}
}

func (s *VictoriaAlertQueriesImpl) GetHostCPUUsage() (float64, error) {
	query := `100 * (1 - avg(rate(host_cpu_seconds_total{mode="idle"}[5m])))`
	return s.queryUsagePercent(query)
}

func (s *VictoriaAlertQueriesImpl) GetHostMemoryUsage() (float64, error) {
	return s.queryUsagePercent(`100 * (1 - host_memory_available_bytes / host_memory_total_bytes)`)
}

func (s *VictoriaAlertQueriesImpl) GetHostDiskUsage() (float64, error) {
	return s.queryUsagePercent(`100 * host_filesystem_used_ratio{mountpoint="/host"}`)
}

func (s *VictoriaAlertQueriesImpl) GetQuollixErrorCount() (int, error) {
	query := fmt.Sprintf(`_time:%s maintainer:quollix app:quollix service:quollix level:in("error", "fatal") | stats count() as errors`, time.Duration(quollixErrorLookbackMinutes)*time.Minute)
	rows, err := s.victoriaLogs.Query(query)
	if err != nil {
		return 0, err
	}
	if len(rows) == 0 {
		return 0, nil
	}
	var count int
	if err := json.Unmarshal(rows[0]["errors"], &count); err != nil {
		var encoded string
		if stringErr := json.Unmarshal(rows[0]["errors"], &encoded); stringErr != nil {
			return 0, u.Logger.NewError(err.Error())
		}
		parsedCount, parseErr := strconv.Atoi(encoded)
		if parseErr != nil {
			return 0, u.Logger.NewError(parseErr.Error())
		}
		count = parsedCount
	}
	return count, nil
}

func (s *VictoriaAlertQueriesImpl) queryUsagePercent(query string) (float64, error) {
	result, err := s.victoriaMetrics.Query(query)
	if err != nil {
		return 0, err
	}
	if result.Status != "success" {
		return 0, u.Logger.NewError("VictoriaMetrics query failed", "status", result.Status)
	}
	if len(result.Data.Result) != 1 {
		return 0, u.Logger.NewError("VictoriaMetrics query did not return exactly one usage percentage", "result_count", len(result.Data.Result))
	}
	sample := result.Data.Result[0]
	if len(sample.Value) != 2 {
		return 0, u.Logger.NewError("VictoriaMetrics query returned an invalid sample")
	}
	var encodedValue string
	if err := json.Unmarshal(sample.Value[1], &encodedValue); err != nil {
		return 0, u.Logger.NewError(err.Error())
	}
	value, err := strconv.ParseFloat(encodedValue, 64)
	if err != nil {
		return 0, u.Logger.NewError(err.Error())
	}
	return value, nil
}
