package logging

import (
	"testing"

	"github.com/quollix/common/assert"
	"github.com/quollix/common/quollix/api"
)

func TestInterpret_VictoriaJSON(t *testing.T) {
	const line = `{"ts":"2026-10-08T10:28:24.299Z","level":"info","caller":"VictoriaLogs/lib/logstorage/storage.go:1043","msg":"finished processing delete task {\"task_id\":\"1791455303665086669\",\"tenant_ids\":[{\"account_id\":0,\"project_id\":0}],\"filter\":\"*\",\"start_time\":\"2026-10-08T10:28:23.665086669Z\"} in 0.007 seconds"}`
	for _, service := range []string{"victorialogs", "victoriametrics"} {
		t.Run(service, func(t *testing.T) {
			record := NewLogLineInterpreter(service).Parse(line, LogRecordContext{ReadTimestamp: receivedAt})
			assert.Equal(t, &api.LogRecord{
				ReadTimestamp: receivedAt,
				LogTimestamp:  testTimestamp(t, "2026-10-08T10:28:24.299Z"),
				Level:         "info",
				Message:       `finished processing delete task {"task_id":"1791455303665086669","tenant_ids":[{"account_id":0,"project_id":0}],"filter":"*","start_time":"2026-10-08T10:28:23.665086669Z"} in 0.007 seconds`,
				Attributes:    map[string]any{"caller": "VictoriaLogs/lib/logstorage/storage.go:1043"},
				Parsed:        true,
			}, record)
		})
	}
}

func TestInterpret_VictoriaMalformedJSON(t *testing.T) {
	for _, service := range []string{"victorialogs", "victoriametrics"} {
		for _, line := range []string{
			`{"ts":"2026-10-08T10:28:24.299Z","msg":`,
			`{"ts":"invalid","level":"info","msg":"ready"}`,
		} {
			t.Run(service+line, func(t *testing.T) {
				record := NewLogLineInterpreter(service).Parse(line, LogRecordContext{ReadTimestamp: receivedAt})
				assert.Equal(t, &api.LogRecord{ReadTimestamp: receivedAt, Level: "unknown", Message: line}, record)
			})
		}
	}
}
