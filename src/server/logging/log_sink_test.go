package logging

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/quollix/common/assert"
	"github.com/quollix/common/quollix/api"
	"github.com/stretchr/testify/mock"
)

type logSinkTestObjects struct {
	Sink   *AsyncLogSinkImpl
	Client *VictoriaLogsClientMock
}

func newLogSinkTestObjects(t *testing.T, maximumBatch int) logSinkTestObjects {
	client := NewVictoriaLogsClientMock(t)
	return logSinkTestObjects{
		Sink:   NewAsyncLogSink(client, maximumBatch),
		Client: client,
	}
}

func TestAsyncLogSink_BuffersRecordsBelowBatchSize(t *testing.T) {
	objects := newLogSinkTestObjects(t, 10)
	record := api.LogRecord{Message: "first"}

	objects.Sink.Send(record)

	assert.Equal(t, uint64(1), objects.Sink.acceptedRecords)
}

func TestAsyncLogSink_FlushesWhenBatchSizeIsReached(t *testing.T) {
	objects := newLogSinkTestObjects(t, 2)
	records := []api.LogRecord{{Message: "first"}, {Message: "second"}}
	pushed := make(chan struct{})
	objects.Client.EXPECT().Push(mock.Anything).Run(func(batch []api.LogRecord) { assert.Equal(t, records, batch); close(pushed) }).Return(nil).Times(1)

	for _, record := range records {
		objects.Sink.Send(record)
	}
	waitForLogSink(t, pushed)
	objects.Sink.Flush()
}

func TestAsyncLogSink_AcceptsRecordsDuringPush(t *testing.T) {
	objects := newLogSinkTestObjects(t, 2)
	started := make(chan struct{})
	release := make(chan struct{})
	var releaseOnce sync.Once
	defer releaseOnce.Do(func() { close(release) })
	first := []api.LogRecord{{Message: "first"}, {Message: "second"}}
	second := []api.LogRecord{{Message: "third"}, {Message: "fourth"}}
	objects.Client.EXPECT().Push(mock.Anything).RunAndReturn(func(batch []api.LogRecord) error {
		assert.Equal(t, first, batch)
		close(started)
		<-release
		assert.Equal(t, first, batch)
		return nil
	}).Times(1)
	objects.Client.EXPECT().Push(mock.Anything).RunAndReturn(func(batch []api.LogRecord) error {
		assert.Equal(t, second, batch)
		return nil
	}).Times(1)

	objects.Sink.Send(first[0])
	objects.Sink.Send(first[1])
	waitForLogSink(t, started)
	added := make(chan struct{})
	go func() {
		objects.Sink.Send(second[0])
		objects.Sink.Send(second[1])
		objects.Sink.Send(api.LogRecord{Message: "discarded"})
		close(added)
	}()
	waitForLogSink(t, added)
	releaseOnce.Do(func() { close(release) })
	objects.Sink.Flush()
}

func TestAsyncLogSink_RetainsFailedBatchBeforeNewRecords(t *testing.T) {
	objects := newLogSinkTestObjects(t, 10)
	first := []api.LogRecord{{Message: "first"}}
	second := []api.LogRecord{{Message: "second"}}
	objects.Client.EXPECT().Push(mock.Anything).Run(func(batch []api.LogRecord) { assert.Equal(t, first, batch) }).Return(errors.New("VictoriaLogs unavailable")).Times(2)
	objects.Client.EXPECT().Push(mock.Anything).Run(func(batch []api.LogRecord) { assert.Equal(t, first, batch) }).Return(nil).Times(1)
	objects.Client.EXPECT().Push(mock.Anything).Run(func(batch []api.LogRecord) { assert.Equal(t, second, batch) }).Return(nil).Times(1)

	objects.Sink.Send(first[0])
	objects.Sink.Flush()
	objects.Sink.Send(second[0])
	objects.Sink.Flush()

	objects.Sink.Flush()
	objects.Sink.Flush()
}

func TestAsyncLogSink_ConcurrentFlushesWaitForOnePush(t *testing.T) {
	objects := newLogSinkTestObjects(t, 10)
	started := make(chan struct{})
	release := make(chan struct{})
	var releaseOnce sync.Once
	defer releaseOnce.Do(func() { close(release) })
	objects.Client.EXPECT().Push(mock.Anything).Run(func([]api.LogRecord) { close(started); <-release }).Return(nil).Times(1)
	objects.Sink.Send(api.LogRecord{Message: "first"})
	var flushes sync.WaitGroup
	for range 5 {
		flushes.Go(objects.Sink.Flush)
	}
	waitForLogSink(t, started)
	finished := make(chan struct{})
	go func() { flushes.Wait(); close(finished) }()
	select {
	case <-finished:
		t.Fatal("flush returned while the push was blocked")
	default:
	}
	releaseOnce.Do(func() { close(release) })
	waitForLogSink(t, finished)
}

func TestAsyncLogSink_FlushDoesNotWaitForLaterRecords(t *testing.T) {
	objects := newLogSinkTestObjects(t, 2)
	started := make(chan struct{})
	releaseFirst := make(chan struct{})
	releaseSecond := make(chan struct{})
	var firstOnce, secondOnce sync.Once
	defer firstOnce.Do(func() { close(releaseFirst) })
	defer secondOnce.Do(func() { close(releaseSecond) })
	objects.Client.EXPECT().Push(mock.Anything).Run(func([]api.LogRecord) { close(started); <-releaseFirst }).Return(nil).Times(1)
	objects.Client.EXPECT().Push(mock.Anything).Run(func([]api.LogRecord) { <-releaseSecond }).Return(nil).Times(1)
	objects.Sink.Send(api.LogRecord{Message: "first"})
	flushed := make(chan struct{})
	go func() { objects.Sink.Flush(); close(flushed) }()
	waitForLogSink(t, started)

	objects.Sink.Send(api.LogRecord{Message: "second"})
	objects.Sink.Send(api.LogRecord{Message: "third"})
	firstOnce.Do(func() { close(releaseFirst) })
	waitForLogSink(t, flushed)

	secondOnce.Do(func() { close(releaseSecond) })
	objects.Sink.Flush()
}

func waitForLogSink(t *testing.T, done <-chan struct{}) {
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for log sink")
	}
}
