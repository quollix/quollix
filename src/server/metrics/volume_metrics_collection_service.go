package metrics

import (
	"regexp"
	"strings"
	"time"

	"github.com/moby/moby/api/types/volume"
	"github.com/quollix/common/quollix/api"
	u "github.com/quollix/common/utils"
)

var volumeNamePattern = regexp.MustCompile(`^[a-z]{3,20}_[a-z]{3,20}_[a-z]{3,20}$`)

type VolumeMetricsCollectionService interface {
	CollectAndPush() error
}

type VolumeMetricsCollectionServiceImpl struct {
	DockerClient  DockerVolumeClient
	MetricsClient VictoriaMetricsClient
}

func NewVolumeMetricsCollectionService(dockerClient DockerVolumeClient, metricsClient VictoriaMetricsClient) *VolumeMetricsCollectionServiceImpl {
	return &VolumeMetricsCollectionServiceImpl{DockerClient: dockerClient, MetricsClient: metricsClient}
}

func (s *VolumeMetricsCollectionServiceImpl) CollectAndPush() error {
	volumes, err := s.DockerClient.ReadUsage()
	if err != nil {
		return err
	}
	collectedAt := time.Now().UTC()
	var batch []api.MetricSeries
	for _, volume := range volumes {
		if !volumeNamePattern.MatchString(volume.Name) {
			continue
		}
		sample, err := collectVolumeSample(volume, collectedAt)
		if err != nil {
			u.Logger.Warn(err, "volume", volume.Name)
			continue
		}
		batch = append(batch, api.MetricSeries{
			Metric:     map[string]string{MetricNameLabel: VolumeUsedBytesMetric, MaintainerLabel: sample.Maintainer, AppLabel: sample.App, VolumeLabel: sample.Volume},
			Values:     []float64{float64(sample.UsedBytes)},
			Timestamps: []int64{sample.Timestamp.UnixMilli()},
		})
	}
	if len(batch) == 0 {
		return nil
	}
	return s.MetricsClient.Push(batch)
}

func collectVolumeSample(volume volume.Volume, timestamp time.Time) (*VolumeMetricsSample, error) {
	if volume.UsageData == nil || volume.UsageData.Size < 0 {
		return nil, u.Logger.NewError("volume usage is unavailable")
	}
	parts := strings.SplitN(volume.Name, "_", 3)
	return &VolumeMetricsSample{Timestamp: timestamp, Maintainer: parts[0], App: parts[1], Volume: parts[2], UsedBytes: uint64(volume.UsageData.Size)}, nil
}
