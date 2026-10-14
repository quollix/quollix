package metrics

import (
	"testing"

	"github.com/moby/moby/api/types/volume"
	"github.com/quollix/common/assert"
	"github.com/quollix/common/quollix/api"
	"github.com/stretchr/testify/mock"
)

type volumeMetricsCollectionServiceTestObjects struct {
	Service       *VolumeMetricsCollectionServiceImpl
	DockerClient  *DockerVolumeClientMock
	MetricsClient *VictoriaMetricsClientMock
}

func setupVolumeMetricsCollectionServiceTest() volumeMetricsCollectionServiceTestObjects {
	dockerClient := &DockerVolumeClientMock{}
	metricsClient := &VictoriaMetricsClientMock{}
	return volumeMetricsCollectionServiceTestObjects{Service: NewVolumeMetricsCollectionService(dockerClient, metricsClient), DockerClient: dockerClient, MetricsClient: metricsClient}
}

func TestVolumeMetricsCollectionService_StoresUsage(t *testing.T) {
	for _, test := range []struct {
		name string
		size int64
	}{
		{name: "zero", size: 0},
		{name: "positive", size: 4096},
	} {
		t.Run(test.name, func(t *testing.T) {
			objects := setupVolumeMetricsCollectionServiceTest()
			defer mock.AssertExpectationsForObjects(t, objects.DockerClient, objects.MetricsClient)
			objects.DockerClient.EXPECT().ReadUsage().Return([]volume.Volume{
				{Name: "quollix_sample_data", UsageData: &volume.UsageData{Size: test.size}},
			}, nil)
			var batch []api.MetricSeries
			objects.MetricsClient.EXPECT().Push(mock.Anything).Run(func(series []api.MetricSeries) { batch = series }).Return(nil)

			assert.Nil(t, objects.Service.CollectAndPush())
			assert.Equal(t, 1, len(batch))
			assert.Equal(t, map[string]string{MetricNameLabel: VolumeUsedBytesMetric, MaintainerLabel: "quollix", AppLabel: "sample", VolumeLabel: "data"}, batch[0].Metric)
			assert.Equal(t, []float64{float64(test.size)}, batch[0].Values)
			assert.Equal(t, 1, len(batch[0].Timestamps))
			assert.True(t, batch[0].Timestamps[0] > 0)
		})
	}
}

func TestVolumeMetricsCollectionService_SkipsUnavailableUsageAndContinues(t *testing.T) {
	objects := setupVolumeMetricsCollectionServiceTest()
	defer mock.AssertExpectationsForObjects(t, objects.DockerClient, objects.MetricsClient)
	objects.DockerClient.EXPECT().ReadUsage().Return([]volume.Volume{
		{Name: "quollix_sample_missing"},
		{Name: "quollix_sample_unknown", UsageData: &volume.UsageData{Size: -1}},
		{Name: "quollix_sample_data", UsageData: &volume.UsageData{Size: 4096}},
	}, nil)
	var batch []api.MetricSeries
	objects.MetricsClient.EXPECT().Push(mock.Anything).Run(func(series []api.MetricSeries) { batch = series }).Return(nil)

	assert.Nil(t, objects.Service.CollectAndPush())
	assert.Equal(t, 1, len(batch))
	assert.Equal(t, "data", batch[0].Metric[VolumeLabel])
	assert.Equal(t, []float64{4096}, batch[0].Values)
}
