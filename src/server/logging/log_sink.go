package logging

import (
	"sync"
	"time"

	"github.com/quollix/common/quollix/api"
	u "github.com/quollix/common/utils"
)

const (
	logDeliveryFailureGracePeriod = 30 * time.Minute
	logDeliveryErrorLogInterval   = 15 * time.Minute
)

type LogSink interface {
	Send(r api.LogRecord)
	Flush()
}

type AsyncLogSinkImpl struct {
	client              VictoriaLogsClient
	mutex               sync.Mutex
	buffer              *LogBuffer
	acceptedRecords     uint64
	sentRecords         uint64
	inFlight            *flushAttempt
	deliveryFailedAt    time.Time
	lastDeliveryErrorAt time.Time
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
	if err == nil {
		s.buffer.ReleaseBatch()
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.logDeliveryResult(err)
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

// Brief delivery interruptions are expected during backups and container restarts. Report transitions at info level and rate-limit persistent failures.
func (s *AsyncLogSinkImpl) logDeliveryResult(err error) {
	now := time.Now()
	if err == nil {
		if !s.deliveryFailedAt.IsZero() {
			u.Logger.Info("Log delivery resumed", "outage_duration", now.Sub(s.deliveryFailedAt).String())
			s.deliveryFailedAt = time.Time{}
			s.lastDeliveryErrorAt = time.Time{}
		}
		return
	}

	if s.deliveryFailedAt.IsZero() {
		s.deliveryFailedAt = now
		u.Logger.Info("Log delivery temporarily unavailable; will retry", "error", err.Error())
		return
	}

	if now.Sub(s.deliveryFailedAt) < logDeliveryFailureGracePeriod {
		return
	}
	if !s.lastDeliveryErrorAt.IsZero() && now.Sub(s.lastDeliveryErrorAt) < logDeliveryErrorLogInterval {
		return
	}

	s.lastDeliveryErrorAt = now
	u.Logger.Error(err, "outage_duration", now.Sub(s.deliveryFailedAt).String())
}
