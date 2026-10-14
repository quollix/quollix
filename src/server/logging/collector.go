package logging

import (
	"context"
	"sync"
	"time"

	"server/tools"

	u "github.com/quollix/common/utils"
)

type ContainerLogMonitorImpl struct {
	DockerLogsClient DockerLogsClient
	ObserverManager  ContainerObserverManager
	LogSink          LogSink
	Interval         time.Duration
	lifecycleMutex   sync.Mutex
	cancelMonitoring context.CancelFunc
	monitoringDone   chan struct{}
}

func NewContainerLogMonitor(client DockerLogsClient, manager ContainerObserverManager, sink LogSink) *ContainerLogMonitorImpl {
	return &ContainerLogMonitorImpl{DockerLogsClient: client, ObserverManager: manager, LogSink: sink, Interval: 5 * time.Second}
}

func (c *ContainerLogMonitorImpl) StartMonitoring() {
	c.lifecycleMutex.Lock()
	defer c.lifecycleMutex.Unlock()
	if c.cancelMonitoring != nil {
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	c.cancelMonitoring = cancel
	c.monitoringDone = make(chan struct{})
	go func() {
		defer close(c.monitoringDone)
		c.run(ctx)
	}()
}

func (c *ContainerLogMonitorImpl) StopMonitoring() {
	c.lifecycleMutex.Lock()
	defer c.lifecycleMutex.Unlock()
	if c.cancelMonitoring == nil {
		return
	}
	c.cancelMonitoring()
	<-c.monitoringDone
	c.cancelMonitoring = nil
	c.monitoringDone = nil
}

func (c *ContainerLogMonitorImpl) run(ctx context.Context) {
	defer func() {
		c.ObserverManager.StopAll()
		c.LogSink.Flush()
	}()
	for {
		c.RunCycle(ctx)
		select {
		case <-ctx.Done():
			return
		case <-time.After(c.Interval):
		}
	}
}

func (c *ContainerLogMonitorImpl) RunCycle(ctx context.Context) {
	defer c.LogSink.Flush()
	names, err := c.DockerLogsClient.ListRunningContainers(ctx)
	if err != nil {
		if ctx.Err() == nil {
			u.Logger.Error(err)
		}
		return
	}
	var wanted []string
	for _, name := range names {
		if _, matches := tools.ParseContainerName(name); matches {
			wanted = append(wanted, name)
		}
	}
	if ctx.Err() == nil {
		c.ObserverManager.Synchronize(ctx, wanted)
	}
}
