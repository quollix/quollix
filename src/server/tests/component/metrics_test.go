//go:build component

package component

import (
	"math"
	"testing"
	"time"

	"server/metrics"
	"server/tools"

	"github.com/quollix/common/assert"
	"github.com/quollix/common/quollix/api"
	"github.com/quollix/common/quollix/api_client"
	u "github.com/quollix/common/utils"
)

const sampleVolumeContent = "volume metrics sample"

func TestMonitoring_CollectsSampleAppMetrics(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()
	metricsClient := client.Metrics

	InstallGrafana(t, client)
	_, err := InstallSample(t, client, tools.SampleAppVersion1Name)
	assert.Nil(t, err)

	assert.Nil(t, StoreStringInSampleApp(GetAppClient(t, client), sampleVolumeContent))

	assert.Nil(t, metricsClient.Collect())
	samples := waitForSampleMetrics(t, metricsClient)
	assertSampleMetrics(t, samples)
}

func assertSampleMetrics(t *testing.T, samples []api.MetricSeries) {
	samplesByName := make(map[string]api.MetricSeries, len(samples))
	for _, sample := range samples {
		name := sample.Metric[metrics.MetricNameLabel]
		_, exists := samplesByName[name]
		assert.False(t, exists)
		samplesByName[name] = sample
	}
	assert.Equal(t, 8, len(samplesByName))

	volume, exists := samplesByName[metrics.VolumeUsedBytesMetric]
	assert.True(t, exists)
	assert.Equal(t, map[string]string{metrics.MetricNameLabel: metrics.VolumeUsedBytesMetric, metrics.MaintainerLabel: tools.SampleMaintainer, metrics.AppLabel: tools.SampleApp, metrics.VolumeLabel: "data"}, volume.Metric)
	assert.Equal(t, []float64{float64(len(sampleVolumeContent))}, volume.Values)
	assert.Equal(t, 1, len(volume.Timestamps))
	assert.True(t, volume.Timestamps[0] > 0)

	for _, name := range []string{
		metrics.ContainerCPUUsageSecondsMetric,
		metrics.ContainerDiskReadBytesMetric,
		metrics.ContainerDiskWrittenBytesMetric,
		metrics.ContainerMemoryLimitBytesMetric,
		metrics.ContainerMemoryUsageBytesMetric,
		metrics.ContainerNetworkReceivedBytesMetric,
		metrics.ContainerNetworkSentBytesMetric,
	} {
		sample, exists := samplesByName[name]
		assert.True(t, exists)
		assert.Equal(t, map[string]string{metrics.MetricNameLabel: name, metrics.MaintainerLabel: tools.SampleMaintainer, metrics.AppLabel: tools.SampleApp, metrics.ServiceLabel: tools.SampleApp}, sample.Metric)
		assert.Equal(t, 1, len(sample.Timestamps))
		assert.True(t, sample.Timestamps[0] > 0)
		assert.Equal(t, samplesByName[metrics.ContainerCPUUsageSecondsMetric].Timestamps, sample.Timestamps)
		assert.Equal(t, 1, len(sample.Values))
		value := sample.Values[0]
		assert.False(t, math.IsNaN(value) || math.IsInf(value, 0))
		assert.True(t, value >= 0)
	}
	assert.True(t, samplesByName[metrics.ContainerMemoryUsageBytesMetric].Values[0] > 0)
	assert.True(t, samplesByName[metrics.ContainerMemoryLimitBytesMetric].Values[0] > 0)
}

func waitForSampleMetrics(t *testing.T, client *api_client.MetricsClient) []api.MetricSeries {
	var samples []api.MetricSeries
	err := u.EventuallyWithTimeout(20*time.Second, 200*time.Millisecond, func() error {
		series, err := client.Read()
		if err != nil {
			return err
		}
		samples = nil
		for _, record := range series {
			if record.Metric[metrics.AppLabel] == tools.SampleApp {
				samples = append(samples, record)
			}
		}
		if len(samples) < 8 {
			return u.Logger.NewError("waiting for the collected metrics to become readable", "series", len(samples))
		}
		return nil
	})
	assert.Nil(t, err)
	return samples
}
