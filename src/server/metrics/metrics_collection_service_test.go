package metrics

import (
	"testing"
	"time"

	"github.com/quollix/common/assert"
	"github.com/quollix/common/quollix/api"
	"github.com/stretchr/testify/mock"
)

var collectionTestTimestamp = time.Date(2026, 10, 9, 12, 0, 0, 123000000, time.UTC)

type metricsCollectionServiceTestObjects struct {
	Service   *MetricsCollectionServiceImpl
	Collector *ConcurrentMetricsCollectorMock
	Client    *VictoriaMetricsClientMock
}

func setupMetricsCollectionServiceTest() metricsCollectionServiceTestObjects {
	collector := &ConcurrentMetricsCollectorMock{}
	client := &VictoriaMetricsClientMock{}
	return metricsCollectionServiceTestObjects{Service: NewMetricsCollectionService(collector, client), Collector: collector, Client: client}
}

func TestMetricsCollectionService_ConvertsSamplesToNumericSeries(t *testing.T) {
	objects := setupMetricsCollectionServiceTest()
	defer mock.AssertExpectationsForObjects(t, objects.Collector, objects.Client)
	sample := ContainerMetricsSample{
		Timestamp:  collectionTestTimestamp,
		Maintainer: "quollix",
		App:        "sampleapp",
		Service:    "web",
		Usage:      ResourceUsage{CPUUsageSeconds: 123.5, MemoryUsageBytes: 1024, MemoryLimitBytes: 8192, NetworkReceivedBytes: 10, NetworkSentBytes: 20, DiskReadBytes: 30, DiskWrittenBytes: 40},
	}
	expectedSeries := []api.MetricSeries{
		series(sample, ContainerCPUUsageSecondsMetric, 123.5),
		series(sample, ContainerMemoryUsageBytesMetric, 1024),
		series(sample, ContainerMemoryLimitBytesMetric, 8192),
		series(sample, ContainerNetworkReceivedBytesMetric, 10),
		series(sample, ContainerNetworkSentBytesMetric, 20),
		series(sample, ContainerDiskReadBytesMetric, 30),
		series(sample, ContainerDiskWrittenBytesMetric, 40),
	}
	objects.Collector.EXPECT().Collect().Return([]ContainerMetricsSample{sample}, nil)
	objects.Client.EXPECT().Push(expectedSeries).Return(nil)
	assert.Nil(t, objects.Service.CollectAndPush())
}

func series(sample ContainerMetricsSample, name string, value float64) api.MetricSeries {
	return api.MetricSeries{
		Metric: map[string]string{
			MetricNameLabel: name,
			MaintainerLabel: sample.Maintainer,
			AppLabel:        sample.App,
			ServiceLabel:    sample.Service,
		},
		Values:     []float64{value},
		Timestamps: []int64{sample.Timestamp.UnixMilli()},
	}
}

func TestMetricsCollectionService_SkipsEmptyCollection(t *testing.T) {
	objects := setupMetricsCollectionServiceTest()
	defer mock.AssertExpectationsForObjects(t, objects.Collector, objects.Client)
	objects.Collector.EXPECT().Collect().Return(nil, nil)
	assert.Nil(t, objects.Service.CollectAndPush())
}
