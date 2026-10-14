package logging

import (
	"context"
	"testing"

	"github.com/quollix/common/assert"
	"github.com/stretchr/testify/mock"
)

type containerMonitorTestObjects struct {
	Monitor *ContainerLogMonitorImpl
	Docker  *DockerLogsClientMock
	Manager *ContainerObserverManagerMock
	Sink    *LogSinkMock
}

func setupContainerMonitorTest() containerMonitorTestObjects {
	docker := &DockerLogsClientMock{}
	manager := &ContainerObserverManagerMock{}
	sink := &LogSinkMock{}
	return containerMonitorTestObjects{NewContainerLogMonitor(docker, manager, sink), docker, manager, sink}
}

func TestContainerLogMonitor_RunCycleFiltersContainersAndFlushes(t *testing.T) {
	objects := setupContainerMonitorTest()
	defer mock.AssertExpectationsForObjects(t, objects.Docker, objects.Manager, objects.Sink)
	objects.Docker.EXPECT().ListRunningContainers(mock.Anything).Return([]string{"quollix_grist_grist", "unrelated", "quollix_App_service"}, nil)
	objects.Manager.EXPECT().Synchronize(mock.Anything, []string{"quollix_grist_grist"})
	objects.Sink.EXPECT().Flush()
	objects.Monitor.RunCycle(context.Background())
}

func TestContainerLogMonitor_StartsOnceAndWaitsForShutdown(t *testing.T) {
	objects := setupContainerMonitorTest()
	defer mock.AssertExpectationsForObjects(t, objects.Docker, objects.Manager, objects.Sink)
	started := make(chan struct{})
	objects.Docker.EXPECT().ListRunningContainers(mock.Anything).Run(func(ctx context.Context) {
		close(started)
		<-ctx.Done()
	}).Return([]string{}, nil).Times(1)
	objects.Sink.EXPECT().Flush().Times(2)
	objects.Manager.EXPECT().StopAll().Times(1)
	objects.Monitor.StopMonitoring()
	objects.Monitor.StartMonitoring()
	objects.Monitor.StartMonitoring()
	<-started
	objects.Monitor.StopMonitoring()
	objects.Monitor.StopMonitoring()
	assert.Equal(t, true, objects.Monitor.cancelMonitoring == nil)
}
