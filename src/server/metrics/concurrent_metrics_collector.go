package metrics

import (
	"sync"

	u "github.com/quollix/common/utils"
)

type ConcurrentMetricsCollector interface {
	Collect() ([]ContainerMetricsSample, error)
}

type ConcurrentMetricsCollectorImpl struct {
	ContainerMetricsService ContainerMetricsService
}

func NewConcurrentMetricsCollector(service ContainerMetricsService) *ConcurrentMetricsCollectorImpl {
	return &ConcurrentMetricsCollectorImpl{ContainerMetricsService: service}
}

func (c *ConcurrentMetricsCollectorImpl) Collect() ([]ContainerMetricsSample, error) {
	containers, err := c.ContainerMetricsService.ListContainers()
	if err != nil {
		return nil, err
	}
	results := make([]*ContainerMetricsSample, len(containers))
	var collection sync.WaitGroup
	for index, container := range containers {
		collection.Go(func() {
			result, err := c.ContainerMetricsService.CollectContainer(container)
			if err != nil {
				u.Logger.Warn(err, "container_id", container.ID)
				return
			}
			results[index] = result
		})
	}
	collection.Wait()

	var samples []ContainerMetricsSample
	for _, sample := range results {
		if sample != nil {
			samples = append(samples, *sample)
		}
	}
	return samples, nil
}
