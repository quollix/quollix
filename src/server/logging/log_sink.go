package logging

import (
	"sync"

	"github.com/quollix/common/quollix/api"
	u "github.com/quollix/common/utils"
)

type LogSink interface {
	Send(r api.LogRecord)
	Flush()
}

type AsyncLogSinkImpl struct {
	client          VictoriaLogsClient
	mutex           sync.Mutex
	buffer          *LogBuffer
	acceptedRecords uint64
	sentRecords     uint64
	inFlight        *flushAttempt
}

type flushAttempt struct {
	done      chan struct{}
	succeeded bool
}

func NewAsyncLogSink(client VictoriaLogsClient, maximumBatch int) *AsyncLogSinkImpl {
	return &AsyncLogSinkImpl{
		client: client,
		buffer: NewLogBuffer(maximumBatch),
	}
}

func (s *AsyncLogSinkImpl) Send(rec api.LogRecord) {
	s.mutex.Lock()
	if !s.buffer.Append(rec) {
		s.mutex.Unlock()
		return
	}
	s.acceptedRecords++
	if s.buffer.Full() {
		s.ensurePushStarted()
	}
	s.mutex.Unlock()
}

func (s *AsyncLogSinkImpl) Flush() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	target := s.acceptedRecords
	for s.sentRecords < target {
		attempt := s.ensurePushStarted()
		s.mutex.Unlock()
		<-attempt.done
		s.mutex.Lock()
		if !attempt.succeeded {
			return
		}
	}
}

func (s *AsyncLogSinkImpl) ensurePushStarted() *flushAttempt {
	if s.inFlight != nil {
		return s.inFlight
	}
	batch := s.buffer.Batch()
	attempt := &flushAttempt{done: make(chan struct{})}
	s.inFlight = attempt
	go s.pushBatch(attempt, batch)
	return attempt
}

func (s *AsyncLogSinkImpl) pushBatch(attempt *flushAttempt, batch []api.LogRecord) {
	err := s.client.Push(batch)
	if err != nil {
		u.Logger.Error(err)
	} else {
		s.buffer.ReleaseBatch()
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.inFlight = nil
	if err == nil {
		s.sentRecords += uint64(len(batch))
		attempt.succeeded = true
	}
	close(attempt.done)
	if err == nil && s.buffer.Full() {
		s.ensurePushStarted()
	}
}
