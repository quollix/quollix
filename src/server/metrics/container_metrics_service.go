package metrics

import (
	"strings"
	"time"

	"server/tools"

	containertypes "github.com/moby/moby/api/types/container"
)

type MetricsContainer struct {
	ID       string
	Identity tools.ContainerIdentity
}

type ContainerMetricsService interface {
	ListContainers() ([]MetricsContainer, error)
	CollectContainer(MetricsContainer) (*ContainerMetricsSample, error)
}

type ContainerMetricsServiceImpl struct {
	DockerStatsClient DockerStatsClient
}

func NewContainerMetricsService(client DockerStatsClient) *ContainerMetricsServiceImpl {
	return &ContainerMetricsServiceImpl{DockerStatsClient: client}
}

func (s *ContainerMetricsServiceImpl) ListContainers() ([]MetricsContainer, error) {
	running, err := s.DockerStatsClient.ListRunningContainers()
	if err != nil {
		return nil, err
	}
	var matching []MetricsContainer
	for _, container := range running {
		for _, name := range container.Names {
			if identity, valid := tools.ParseContainerName(name); valid {
				matching = append(matching, MetricsContainer{ID: container.ID, Identity: *identity})
				break
			}
		}
	}
	return matching, nil
}

func (s *ContainerMetricsServiceImpl) CollectContainer(container MetricsContainer) (*ContainerMetricsSample, error) {
	stats, err := s.DockerStatsClient.ReadStats(container.ID)
	if err != nil {
		return nil, err
	}
	return &ContainerMetricsSample{
		Timestamp:  time.Now().UTC(),
		Maintainer: container.Identity.Maintainer,
		App:        container.Identity.App,
		Service:    container.Identity.Service,
		Usage:      calculateResourceUsage(stats),
	}, nil
}

func calculateResourceUsage(stats *containertypes.StatsResponse) ResourceUsage {
	usage := ResourceUsage{
		CPUUsageSeconds:  float64(stats.CPUStats.CPUUsage.TotalUsage) / float64(time.Second),
		MemoryUsageBytes: stats.MemoryStats.Usage,
		MemoryLimitBytes: stats.MemoryStats.Limit,
	}
	for _, network := range stats.Networks {
		usage.NetworkReceivedBytes += network.RxBytes
		usage.NetworkSentBytes += network.TxBytes
	}
	for _, entry := range stats.BlkioStats.IoServiceBytesRecursive {
		switch strings.ToLower(entry.Op) {
		case "read":
			usage.DiskReadBytes += entry.Value
		case "write":
			usage.DiskWrittenBytes += entry.Value
		}
	}
	return usage
}
