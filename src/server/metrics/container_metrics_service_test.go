package metrics

import (
	"testing"

	"server/tools"

	containertypes "github.com/moby/moby/api/types/container"
	"github.com/quollix/common/assert"
	"github.com/stretchr/testify/mock"
)

var metricsTestContainer = MetricsContainer{ID: "sample"}

type containerMetricsServiceTestObjects struct {
	Service *ContainerMetricsServiceImpl
	Client  *DockerStatsClientMock
}

func setupContainerMetricsServiceTest() containerMetricsServiceTestObjects {
	client := &DockerStatsClientMock{}
	return containerMetricsServiceTestObjects{Service: NewContainerMetricsService(client), Client: client}
}

func TestContainerMetricsService_FiltersNamesAndCollectsEachContainerOnlyOnce(t *testing.T) {
	objects := setupContainerMetricsServiceTest()
	defer mock.AssertExpectationsForObjects(t, objects.Client)
	running := []containertypes.Summary{
		{ID: "unrelated", Names: []string{"/unrelated"}},
		{ID: "invalid", Names: []string{"/quollix_App_service"}},
		{ID: "unnamed"},
		{ID: "sample", Names: []string{"/alias", "/quollix_sampleapp_web", "/quollix_sampleapp_other"}},
	}
	objects.Client.EXPECT().ListRunningContainers().Return(running, nil)
	matching, err := objects.Service.ListContainers()
	assert.Nil(t, err)
	assert.Equal(t, []MetricsContainer{{ID: "sample", Identity: tools.ContainerIdentity{Maintainer: "quollix", App: "sampleapp", Service: "web"}}}, matching)
}

func TestContainerMetricsService_CalculatesResourceUsage(t *testing.T) {
	objects := setupContainerMetricsServiceTest()
	defer mock.AssertExpectationsForObjects(t, objects.Client)
	stats := containertypes.StatsResponse{
		CPUStats:    containertypes.CPUStats{CPUUsage: containertypes.CPUUsage{TotalUsage: 1250000000}},
		MemoryStats: containertypes.MemoryStats{Usage: 1024, Limit: 8192, Stats: map[string]uint64{"cache": 256, "total_inactive_file": 128}},
		Networks: map[string]containertypes.NetworkStats{
			"eth0": {RxBytes: 10, TxBytes: 20},
			"eth1": {RxBytes: 30, TxBytes: 40},
		},
		BlkioStats: containertypes.BlkioStats{IoServiceBytesRecursive: []containertypes.BlkioStatEntry{
			{Major: 8, Minor: 0, Op: "Read", Value: 100},
			{Major: 8, Minor: 0, Op: "Write", Value: 200},
			{Major: 8, Minor: 16, Op: "read", Value: 300},
			{Major: 8, Minor: 16, Op: "write", Value: 400},
			{Op: "Total", Value: 1000},
			{Op: "Sync", Value: 1000},
		}},
	}
	objects.Client.EXPECT().ReadStats(metricsTestContainer.ID).Return(&stats, nil)
	sample, err := objects.Service.CollectContainer(metricsTestContainer)
	assert.Nil(t, err)
	assert.Equal(t, ResourceUsage{CPUUsageSeconds: 1.25, MemoryUsageBytes: 1024, MemoryLimitBytes: 8192, NetworkReceivedBytes: 40, NetworkSentBytes: 60, DiskReadBytes: 400, DiskWrittenBytes: 600}, sample.Usage)
}
