package logging

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/quollix/common/assert"
	"github.com/quollix/common/quollix/api"
)

var receivedAt = time.Date(2026, 10, 3, 12, 0, 0, 0, time.UTC)

func TestParse_MovesCommonFieldsOutOfAttributes(t *testing.T) {
	context := LogRecordContext{ReadTimestamp: receivedAt, Maintainer: "owner", App: "app", Service: "service", Stream: "stderr"}
	raw := `{"@timestamp":"2026-10-03T10:30:14.695462+01:00","log.level":"notice","message":"hello","app":"logged app","count":2,"nested":{"enabled":true}}`
	parser := &LogLineInterpreterImpl{parser: fieldLogParser(parseJSON, logSchema{"@timestamp", "log.level", "message", time.RFC3339Nano})}
	record := parser.Parse(raw, context)
	var fields map[string]any
	assert.Nil(t, json.Unmarshal([]byte(raw), &fields))
	delete(fields, "@timestamp")
	delete(fields, "log.level")
	delete(fields, "message")
	assert.Equal(t, &api.LogRecord{
		ReadTimestamp: receivedAt, LogTimestamp: testTimestamp(t, "2026-10-03T09:30:14.695462Z"),
		Level: "info", Message: "hello", Attributes: fields,
		Maintainer: "owner", App: "app", Service: "service", Stream: "stderr", Parsed: true,
	}, record)
	data, err := json.Marshal(record)
	assert.Nil(t, err)
	assert.False(t, strings.Contains(string(data), "original_line"))
}

func TestParse_MapsServiceLevels(t *testing.T) {
	for _, testCase := range []struct {
		levels   string
		expected string
	}{
		{"trace t trace1 trace2 trace3 trace4 trace5 trace6 trace7 trace8 finer finest silly", "trace"},
		{"debug d debug1 debug2 debug3 debug4 debug5 fine config verbose http .", "debug"},
		{"info i notice note log * -", "info"},
		{"warn warning w #", "warn"},
		{"error err e severe", "error"},
		{"fatal f panic critical crit alert emergency emerg", "fatal"},
		{"i-dont-exist", "unknown"},
	} {
		for _, level := range strings.Fields(testCase.levels) {
			for _, sourceLevel := range []string{level, strings.ToUpper(level)} {
				t.Run(sourceLevel, func(t *testing.T) {
					encoded, err := json.Marshal(map[string]any{"level": sourceLevel})
					assert.Nil(t, err)
					record := NewLogLineInterpreter("sampleapp").Parse(string(encoded), LogRecordContext{ReadTimestamp: receivedAt})
					assert.Equal(t, testCase.expected, record.Level)
					assert.Nil(t, record.Attributes)
				})
			}
		}
	}
}

func TestParse_UsesOnlyDeclaredFields(t *testing.T) {
	for _, raw := range []string{
		`{"timestamp":"2026-10-03 10:29:55","level":"debug","message":"","time":123,"t":false,"@timestamp":null,"log.level":{},"msg":"alternate message"}`,
		`{"time":"2026-10-03T10:29:55Z","log.level":"error","detail":"alternate message"}`,
	} {
		t.Run(raw, func(t *testing.T) {
			record := NewLogLineInterpreter("grist").Parse(raw, LogRecordContext{})
			assert.True(t, record.Parsed)
			assert.Equal(t, "", record.Message)
			var expected map[string]any
			assert.Nil(t, json.Unmarshal([]byte(raw), &expected))
			_, hasTimestamp := expected["timestamp"]
			delete(expected, "timestamp")
			delete(expected, "level")
			delete(expected, "message")
			assert.Equal(t, expected, record.Attributes)
			if hasTimestamp {
				assert.Equal(t, "debug", record.Level)
				assert.Equal(t, testTimestamp(t, "2026-10-03T10:29:55Z"), record.LogTimestamp)
			} else {
				assert.Equal(t, "unknown", record.Level)
				assert.Nil(t, record.LogTimestamp)
			}
		})
	}
}

func TestParse_DefaultsMissingLevel(t *testing.T) {
	for _, tc := range []struct {
		raw, message string
		parsed       bool
	}{
		{"plain text", "plain text", false},
		{`{"count":1}`, "", true},
	} {
		record := NewLogLineInterpreter("sampleapp").Parse(tc.raw, LogRecordContext{ReadTimestamp: receivedAt})
		assert.Equal(t, "unknown", record.Level)
		assert.Equal(t, tc.message, record.Message)
		assert.Equal(t, receivedAt, record.ReadTimestamp)
		assert.Nil(t, record.LogTimestamp)
		assert.Equal(t, tc.parsed, record.Parsed)
	}
}

func TestParse_MissingCommonFieldsRemainStructured(t *testing.T) {
	record := NewLogLineInterpreter("grist").Parse(`{"count":3}`, LogRecordContext{ReadTimestamp: receivedAt})
	assert.Equal(t, &api.LogRecord{
		ReadTimestamp: receivedAt, Level: "unknown", Parsed: true,
		Attributes: map[string]any{"count": float64(3)},
	}, record)
}

func TestParse_InvalidCommonFieldTypesFallBackToPlainText(t *testing.T) {
	for _, key := range []string{"@timestamp", "timestamp", "time", "t", "log.level", "level", "message", "msg"} {
		t.Run(key, func(t *testing.T) {
			for _, value := range []any{123, true, nil, []string{"value"}, map[string]any{"nested": "value"}} {
				encoded, err := json.Marshal(map[string]any{key: value})
				assert.Nil(t, err)
				raw := string(encoded)
				record := NewLogLineInterpreter("grist").Parse(raw, LogRecordContext{ReadTimestamp: receivedAt})
				if key == "timestamp" || key == "level" || key == "message" || key == "msg" {
					assert.Equal(t, &api.LogRecord{ReadTimestamp: receivedAt, Level: "unknown", Message: raw}, record)
				} else {
					assert.True(t, record.Parsed)
					var expected map[string]any
					assert.Nil(t, json.Unmarshal([]byte(raw), &expected))
					assert.Equal(t, expected, record.Attributes)
					assert.Equal(t, "unknown", record.Level)
				}
			}
		})
	}
}

func TestParse_AbsentDeclaredFieldsRemainOptional(t *testing.T) {
	for _, service := range []string{"grist", "wikijs", "rabbitmq", "vikunja", "sampleapp"} {
		t.Run(service, func(t *testing.T) {
			record := NewLogLineInterpreter(service).Parse(`{"count":3}`, LogRecordContext{})
			assert.True(t, record.Parsed)
			assert.Nil(t, record.LogTimestamp)
			assert.Equal(t, "unknown", record.Level)
			assert.Equal(t, "", record.Message)
		})
	}
}

func TestParse_ExplicitSchemaKeepsAliasFields(t *testing.T) {
	raw := `{"log.level":"warning","level":"debug","message":"","msg":"alternate message"}`
	parser := &LogLineInterpreterImpl{parser: fieldLogParser(parseJSON, logSchema{levelField: "log.level", messageField: "message"})}
	record := parser.Parse(raw, LogRecordContext{})
	assert.Equal(t, "warn", record.Level)
	assert.Equal(t, "", record.Message)
	assert.True(t, record.Parsed)
	var expected map[string]any
	assert.Nil(t, json.Unmarshal([]byte(raw), &expected))
	delete(expected, "log.level")
	delete(expected, "message")
	assert.Equal(t, expected, record.Attributes)
}
