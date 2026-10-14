package logging

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/quollix/common/assert"
	"github.com/quollix/common/quollix/api"
)

func TestDecodeLogRecord_DecodesAttributesAndStringFields(t *testing.T) {
	logTimestamp := receivedAt.Add(-time.Hour)
	cases := []struct {
		name     string
		fields   map[string]any
		expected api.LogRecord
	}{
		{
			name: "flattened attributes retain dotted names and string values",
			fields: map[string]any{
				"_time":                                receivedAt.Format(time.RFC3339Nano),
				"log_timestamp":                        logTimestamp.Format(time.RFC3339Nano),
				"_msg":                                 "hello",
				"level":                                "info",
				"parsed":                               "true",
				"attributes.level":                     "ERROR",
				"attributes.n":                         "00123",
				"attributes.nested.enabled":            "true",
				"attributes.http.response.status_code": "200",
				"attributes.items":                     `[1,"two"]`,
			},
			expected: api.LogRecord{
				ReadTimestamp: receivedAt,
				LogTimestamp:  &logTimestamp,
				Message:       "hello",
				Level:         "info",
				Parsed:        true,
				Attributes: map[string]any{
					"level":                     "ERROR",
					"n":                         "00123",
					"nested.enabled":            "true",
					"http.response.status_code": "200",
					"items":                     `[1,"two"]`,
				},
			},
		},
		{
			name: "preserved extras and string fields",
			fields: map[string]any{
				"_time":         receivedAt.Format(time.RFC3339Nano),
				"log_timestamp": logTimestamp.Format(time.RFC3339Nano),
				"_msg":          "hello",
				"level":         "info",
				"maintainer":    "owner",
				"app":           "application",
				"service":       "service",
				"parsed":        "true",
				"extra":         `{"n":42,"nested":{"enabled":true},"items":[1,"two"]}`,
			},
			expected: api.LogRecord{
				ReadTimestamp: receivedAt,
				LogTimestamp:  &logTimestamp,
				Message:       "hello",
				Level:         "info",
				Maintainer:    "owner",
				App:           "application",
				Service:       "service",
				Parsed:        true,
				Attributes: map[string]any{
					"n":      float64(42),
					"nested": map[string]any{"enabled": true},
					"items":  []any{float64(1), "two"},
				},
			},
		},
		{
			name: "plain text without log timestamp or extras",
			fields: map[string]any{
				"_time":  receivedAt.Format(time.RFC3339Nano),
				"_msg":   "plain text",
				"level":  "info",
				"parsed": "false",
			},
			expected: api.LogRecord{
				ReadTimestamp: receivedAt,
				Message:       "plain text",
				Level:         "info",
			},
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			data, err := json.Marshal(testCase.fields)
			assert.Nil(t, err)
			var row map[string]json.RawMessage
			assert.Nil(t, json.Unmarshal(data, &row))
			record, err := decodeLogRecord(row)
			assert.Nil(t, err)
			assert.Equal(t, testCase.expected, record)
		})
	}
}

func TestDecodeLogRecord_RejectsMalformedPreservedFields(t *testing.T) {
	for _, testCase := range []struct {
		field string
		value string
	}{
		{"extra", `"not JSON"`},
		{"parsed", `"not a boolean"`},
	} {
		t.Run(testCase.field, func(t *testing.T) {
			_, err := decodeLogRecord(map[string]json.RawMessage{
				testCase.field: json.RawMessage(testCase.value),
			})
			assert.NotNil(t, err)
		})
	}
}
