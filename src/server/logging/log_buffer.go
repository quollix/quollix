package logging

import "github.com/quollix/common/quollix/api"

type LogBuffer struct {
	active []api.LogRecord
	batch  []api.LogRecord
}

func NewLogBuffer(maximumBatch int) *LogBuffer {
	return &LogBuffer{
		active: make([]api.LogRecord, 0, maximumBatch),
		batch:  make([]api.LogRecord, 0, maximumBatch),
	}
}

func (b *LogBuffer) Append(record api.LogRecord) bool {
	if b.Full() {
		return false
	}
	b.active = append(b.active, record)
	return true
}

func (b *LogBuffer) Full() bool {
	return len(b.active) == cap(b.active)
}

func (b *LogBuffer) Batch() []api.LogRecord {
	if len(b.batch) == 0 {
		b.batch, b.active = b.active, b.batch
	}
	return b.batch
}

func (b *LogBuffer) ReleaseBatch() {
	clear(b.batch)
	b.batch = b.batch[:0]
}
