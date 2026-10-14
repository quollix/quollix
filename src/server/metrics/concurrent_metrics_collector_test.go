package metrics

import (
	"testing"

	"github.com/quollix/common/assert"
	u "github.com/quollix/common/utils"
	"github.com/stretchr/testify/mock"
)

type concurrentMetricsCollectorTestObjects struct {
	Collector *ConcurrentMetricsCollectorImpl
	Service   *ContainerMetricsServiceMock
}

func setupConcurrentMetricsCollectorTest() concurrentMetricsCollectorTestObjects {
	service := &ContainerMetricsServiceMock{}
	return concurrentMetricsCollectorTestObjects{Collector: NewConcurrentMetricsCollector(service), Service: service}
}

func TestConcurrentMetricsCollector_ExcludesFailedContainerSamples(t *testing.T) {
	objects := setupConcurrentMetricsCollectorTest()
	defer mock.AssertExpectationsForObjects(t, objects.Service)
	failed := MetricsContainer{ID: "failed"}
	healthy := MetricsContainer{ID: "healthy"}
	healthySample := ContainerMetricsSample{Service: "healthy", Usage: ResourceUsage{MemoryUsageBytes: 1024}}
	objects.Service.EXPECT().ListContainers().Return([]MetricsContainer{failed, healthy}, nil)
	objects.Service.EXPECT().CollectContainer(failed).Return(nil, u.Logger.NewError("container disappeared"))
	objects.Service.EXPECT().CollectContainer(healthy).Return(&healthySample, nil)

	samples, err := objects.Collector.Collect()
	assert.Nil(t, err)
	assert.Equal(t, []ContainerMetricsSample{healthySample}, samples)
}

func TestConcurrentMetricsCollector_CollectsAllContainerSamples(t *testing.T) {
	objects := setupConcurrentMetricsCollectorTest()
	defer mock.AssertExpectationsForObjects(t, objects.Service)
	web := MetricsContainer{ID: "web"}
	database := MetricsContainer{ID: "database"}
	webSample := ContainerMetricsSample{Service: "web", Usage: ResourceUsage{MemoryUsageBytes: 1024}}
	databaseSample := ContainerMetricsSample{Service: "database", Usage: ResourceUsage{MemoryUsageBytes: 2048}}
	objects.Service.EXPECT().ListContainers().Return([]MetricsContainer{web, database}, nil)
	objects.Service.EXPECT().CollectContainer(web).Return(&webSample, nil)
	objects.Service.EXPECT().CollectContainer(database).Return(&databaseSample, nil)

	samples, err := objects.Collector.Collect()
	assert.Nil(t, err)
	assert.Equal(t, []ContainerMetricsSample{webSample, databaseSample}, samples)
}
