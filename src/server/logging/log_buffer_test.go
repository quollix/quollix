package logging

import (
	"testing"

	"github.com/quollix/common/assert"
	"github.com/quollix/common/quollix/api"
)

func TestLogBuffer_BoundsEachBufferAndPreservesBatchUntilReleased(t *testing.T) {
	buffer := NewLogBuffer(2)
	first := []api.LogRecord{{Message: "first"}, {Message: "second"}}
	second := []api.LogRecord{{Message: "third"}, {Message: "fourth"}}

	assert.Equal(t, 2, cap(buffer.active))
	assert.True(t, !buffer.Full())
	for _, record := range first {
		assert.True(t, buffer.Append(record))
	}
	assert.True(t, buffer.Full())
	assert.True(t, !buffer.Append(api.LogRecord{Message: "discarded"}))
	batch := buffer.Batch()
	assert.Equal(t, first, batch)
	assert.True(t, !buffer.Full())

	for _, record := range second {
		assert.True(t, buffer.Append(record))
	}
	assert.True(t, buffer.Full())
	assert.True(t, !buffer.Append(api.LogRecord{Message: "also discarded"}))
	assert.Equal(t, first, buffer.Batch())
	assert.True(t, &batch[0] == &buffer.Batch()[0])

	buffer.ReleaseBatch()
	assert.Equal(t, second, buffer.Batch())
}

func TestLogBuffer_ClearsReferencesAndReusesBothAllocations(t *testing.T) {
	buffer := NewLogBuffer(2)
	firstAllocation := &buffer.active[:cap(buffer.active)][0]
	secondAllocation := &buffer.batch[:cap(buffer.batch)][0]

	for _, allocation := range []*api.LogRecord{firstAllocation, secondAllocation, firstAllocation, secondAllocation} {
		assert.True(t, buffer.Append(api.LogRecord{Message: "record", Attributes: map[string]any{"key": "value"}}))
		batch := buffer.Batch()
		assert.True(t, allocation == &batch[0])
		assert.Equal(t, 2, cap(batch))
		buffer.ReleaseBatch()
		assert.Equal(t, []api.LogRecord{{}, {}}, batch[:cap(batch)])
	}
	assert.Equal(t, 0, len(buffer.Batch()))
}
