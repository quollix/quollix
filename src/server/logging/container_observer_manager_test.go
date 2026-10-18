package logging

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/quollix/common/assert"
	"github.com/stretchr/testify/mock"
)

type observerManagerTestObjects struct {
	Manager *ContainerObserverManagerImpl
	Docker  *DockerLogsClientMock
	Reader  *LogStreamReaderMock
}

func setupObserverManagerTest() observerManagerTestObjects {
	docker := &DockerLogsClientMock{}
	reader := &LogStreamReaderMock{}
	return observerManagerTestObjects{NewContainerObserverManager(docker, reader), docker, reader}
}

func TestContainerObserverManager_ReusesObserverAndRestartsFinishedStream(t *testing.T) {
	objects := setupObserverManagerTest()
	defer mock.AssertExpectationsForObjects(t, objects.Docker, objects.Reader)
	finished := make(chan struct{})
	objects.Docker.EXPECT().OpenContainerLogStream(mock.Anything, "quollix_grist_grist").Return(&ContainerLogStream{ReadCloser: io.NopCloser(strings.NewReader("")), Multiplexed: true}, true, nil).Times(2)
	objects.Reader.EXPECT().ReadStream(mock.Anything).Run(func(stream *LogStreamContext) {
		assert.Equal(t, "quollix", stream.Maintainer)
		assert.Equal(t, "grist", stream.App)
		assert.Equal(t, "grist", stream.Service)
		assert.True(t, stream.Multiplexed)
		record := stream.Interpreter.Parse(`{"message":"ready"}`, LogRecordContext{})
		assert.True(t, record.Parsed)
		assert.Equal(t, "ready", record.Message)
		<-finished
	}).Return(nil).Times(2)
	names := []string{"quollix_grist_grist"}
	objects.Manager.Synchronize(context.Background(), names)
	first := objects.Manager.observers[names[0]]
	objects.Manager.Synchronize(context.Background(), names)
	assert.Equal(t, true, first == objects.Manager.observers[names[0]])
	close(finished)
	<-first.done
	objects.Manager.Synchronize(context.Background(), names)
	assert.Equal(t, false, first == objects.Manager.observers[names[0]])
	objects.Manager.StopAll()
	assert.Equal(t, 0, len(objects.Manager.observers))
}

func TestContainerObserverManager_CancelsRemovedContainersAndRemovesFinishedObservers(t *testing.T) {
	objects := setupObserverManagerTest()
	observerContext, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	objects.Manager.observers["quollix_grist_grist"] = &containerObserver{cancel: cancel, done: done}
	objects.Manager.Synchronize(context.Background(), nil)
	assert.Equal(t, context.Canceled, observerContext.Err())
	assert.Equal(t, 1, len(objects.Manager.observers))
	close(done)
	objects.Manager.Synchronize(context.Background(), nil)
	assert.Equal(t, 0, len(objects.Manager.observers))
}

func TestContainerObserverManager_StopAllCancelsAndWaitsForReaders(t *testing.T) {
	objects := setupObserverManagerTest()
	observerContext, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	objects.Manager.observers["quollix_grist_grist"] = &containerObserver{cancel: cancel, done: done}
	go func() { <-observerContext.Done(); close(done) }()
	objects.Manager.StopAll()
	assert.Equal(t, context.Canceled, observerContext.Err())
	assert.Equal(t, 0, len(objects.Manager.observers))
}
