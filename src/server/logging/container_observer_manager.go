package logging

import (
	"context"

	"server/tools"

	u "github.com/quollix/common/utils"
)

type ContainerObserverManager interface {
	Synchronize(context.Context, []string)
	StopAll()
}

type containerObserver struct {
	cancel context.CancelFunc
	done   chan struct{}
}

type ContainerObserverManagerImpl struct {
	DockerLogsClient DockerLogsClient
	LogStreamReader  LogStreamReader
	observers        map[string]*containerObserver
}

func NewContainerObserverManager(client DockerLogsClient, reader LogStreamReader) *ContainerObserverManagerImpl {
	return &ContainerObserverManagerImpl{DockerLogsClient: client, LogStreamReader: reader, observers: make(map[string]*containerObserver)}
}

func (m *ContainerObserverManagerImpl) Synchronize(ctx context.Context, names []string) {
	wanted := make(map[string]bool, len(names))
	for _, name := range names {
		wanted[name] = true
	}
	for name, observer := range m.observers {
		if !wanted[name] {
			observer.cancel()
		}
		select {
		case <-observer.done:
			delete(m.observers, name)
		default:
		}
	}
	for _, name := range names {
		if ctx.Err() != nil {
			return
		}
		if _, exists := m.observers[name]; exists {
			continue
		}
		identity, valid := tools.ParseContainerName(name)
		if !valid {
			continue
		}
		observerContext, cancel := context.WithCancel(ctx)
		stream, available, err := m.DockerLogsClient.OpenContainerLogStream(observerContext, name)
		if err != nil {
			cancel()
			if ctx.Err() == nil {
				u.Logger.Error(err)
			}
			continue
		}
		if !available {
			cancel()
			if ctx.Err() == nil {
				u.Logger.Info("skipping log collection because container is no longer available", "container", name)
			}
			continue
		}
		observer := &containerObserver{cancel: cancel, done: make(chan struct{})}
		m.observers[name] = observer
		streamContext := &LogStreamContext{Reader: stream, Multiplexed: stream.Multiplexed, Interpreter: NewLogLineInterpreter(identity.Service), Maintainer: identity.Maintainer, App: identity.App, Service: identity.Service}
		go func() {
			defer close(observer.done)
			defer cancel()
			if err := m.LogStreamReader.ReadStream(streamContext); err != nil && observerContext.Err() == nil {
				u.Logger.Error(err, "container", name)
			}
		}()
	}
}

func (m *ContainerObserverManagerImpl) StopAll() {
	for _, observer := range m.observers {
		observer.cancel()
	}
	for name, observer := range m.observers {
		<-observer.done
		delete(m.observers, name)
	}
}
