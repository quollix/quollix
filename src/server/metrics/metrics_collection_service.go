package metrics

import (
	"github.com/quollix/common/quollix/api"
)

type MetricsCollectionService interface {
	CollectAndPush() error
}

type MetricsCollectionServiceImpl struct {
	Collector ConcurrentMetricsCollector
	Client    VictoriaMetricsClient
}

func NewMetricsCollectionService(collector ConcurrentMetricsCollector, client VictoriaMetricsClient) *MetricsCollectionServiceImpl {
	return &MetricsCollectionServiceImpl{Collector: collector, Client: client}
}

func (s *MetricsCollectionServiceImpl) CollectAndPush() error {
	records, err := s.Collector.Collect()
	if err != nil {
		return err
	}
	var batch []api.MetricSeries
	for _, record := range records {
		measurements := []struct {
			name  string
			value float64
		}{
			{ContainerCPUUsageSecondsMetric, record.Usage.CPUUsageSeconds},
			{ContainerMemoryUsageBytesMetric, float64(record.Usage.MemoryUsageBytes)},
			{ContainerMemoryLimitBytesMetric, float64(record.Usage.MemoryLimitBytes)},
			{ContainerNetworkReceivedBytesMetric, float64(record.Usage.NetworkReceivedBytes)},
			{ContainerNetworkSentBytesMetric, float64(record.Usage.NetworkSentBytes)},
			{ContainerDiskReadBytesMetric, float64(record.Usage.DiskReadBytes)},
			{ContainerDiskWrittenBytesMetric, float64(record.Usage.DiskWrittenBytes)},
		}
		for _, measurement := range measurements {
			batch = append(batch, api.MetricSeries{
				Metric: map[string]string{
					MetricNameLabel: measurement.name,
					MaintainerLabel: record.Maintainer,
					AppLabel:        record.App,
					ServiceLabel:    record.Service,
				},
				Values:     []float64{measurement.value},
				Timestamps: []int64{record.Timestamp.UnixMilli()},
			})
		}
	}
	if len(batch) == 0 {
		return nil
	}
	return s.Client.Push(batch)
}
