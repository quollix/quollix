package logging

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/quollix/common/assert"
	"github.com/quollix/common/quollix/api"
)

func TestParse_RecoversPanicAndContinues(t *testing.T) {
	parser := &LogLineInterpreterImpl{parser: newLogParser(func(line string) (map[string]any, bool) {
		if line == "broken" {
			panic("unexpected parser failure")
		}
		return map[string]any{"message": line, "level": "info"}, true
	}, time.RFC3339Nano)}
	context := LogRecordContext{ReadTimestamp: receivedAt, Maintainer: "owner", App: "app", Service: "service", Stream: "stderr"}
	assert.Equal(t, &api.LogRecord{
		ReadTimestamp: receivedAt, Message: "broken", Level: "unknown",
		Maintainer: "owner", App: "app", Service: "service", Stream: "stderr",
	}, parser.Parse("broken", context))
	record := parser.Parse("ready", context)
	assert.True(t, record.Parsed)
	assert.Equal(t, "ready", record.Message)
	assert.Equal(t, "info", record.Level)
}

func TestInterpret_UsesOnlyConfiguredServiceParser(t *testing.T) {
	for _, service := range []string{"grist", "unknown", "memcached"} {
		interpreter := NewLogLineInterpreter(service)
		for _, raw := range []string{`{"msg":"hello"}`, `{"message":"hello"}`, `level=info msg=hello`, "plain text"} {
			record := interpreter.Parse(raw, LogRecordContext{})
			if service == "grist" && strings.HasPrefix(raw, "{") {
				assert.True(t, record.Parsed)
				assert.Equal(t, "hello", record.Message)
			} else {
				assert.False(t, record.Parsed)
				assert.Equal(t, raw, record.Message)
			}
		}
	}
}

func TestInterpret_JSONMessageAliases(t *testing.T) {
	for _, service := range []string{"sampleapp", "grist", "grafana", "rabbitmq", "vikunja", "victorialogs", "vector"} {
		for _, tc := range []struct {
			line, message string
			attributes    map[string]any
		}{
			{`{"message":"ready"}`, "ready", nil},
			{`{"msg":"ready"}`, "ready", nil},
			{`{"message":"primary","msg":"secondary"}`, "primary", map[string]any{"msg": "secondary"}},
			{`{"message":"","msg":"secondary"}`, "", map[string]any{"msg": "secondary"}},
		} {
			t.Run(service+"/"+tc.line, func(t *testing.T) {
				record := NewLogLineInterpreter(service).Parse(tc.line, LogRecordContext{})
				assert.True(t, record.Parsed)
				assert.Equal(t, tc.message, record.Message)
				assert.Equal(t, tc.attributes, record.Attributes)
			})
		}
	}
}

func TestInterpret_AcceptsServiceLevels(t *testing.T) {
	for _, testCase := range []struct {
		service string
		format  string
		levels  map[string]string
		mode    string
	}{
		{"Wiki.js", `{"level":"%s","message":"Initializing...","timestamp":"2026-10-03T09:29:08.589Z"}`, map[string]string{"error": "error", "warn": "warn", "info": "info", "http": "debug", "verbose": "debug", "debug": "debug", "silly": "trace"}, "json"},
		{"AdGuard", "2026/10/03 10:29:27.226036 [%s] starting adguard home", map[string]string{"debug": "debug", "info": "info", "warn": "warn", "warning": "warn", "error": "error", "fatal": "fatal"}, "adguard"},
		{"Forgejo", "2026/10/03 10:29:44 cmd/web.go:250:runWeb() [%s] Starting Forgejo", map[string]string{"T": "trace", "D": "debug", "I": "info", "W": "warn", "E": "error", "F": "fatal"}, "forgejo"},
		{"PostgreSQL", "2026-10-03 10:29:44.162 BST [1] %s: starting PostgreSQL", map[string]string{"DEBUG1": "debug", "DEBUG2": "debug", "DEBUG3": "debug", "DEBUG4": "debug", "DEBUG5": "debug", "INFO": "info", "NOTICE": "info", "WARNING": "warn", "ERROR": "error", "LOG": "info", "FATAL": "fatal", "PANIC": "fatal", "CUSTOM42": "unknown"}, "postgres"},
		{"HedgeDoc", "2026-10-03T09:30:55.672Z %s: All migrations performed successfully", map[string]string{"error": "error", "warn": "warn", "info": "info", "http": "debug", "verbose": "debug", "debug": "debug", "silly": "trace"}, "timestamp_level"},
		{"Jitsi", "JVB 2026-10-03 10:31:10.413 %s: [1] MainKt.main#82: Starting jitsi-videobridge", map[string]string{"FINEST": "trace", "FINER": "trace", "FINE": "debug", "CONFIG": "debug", "INFO": "info", "WARNING": "warn", "SEVERE": "error", "CUSTOM42": "unknown"}, "jitsi_component"},
		{"Prosody", "2026-10-03 10:31:09 startup             %s\tHello and welcome to Prosody", map[string]string{"debug": "debug", "info": "info", "warn": "warn", "error": "error", "CUSTOM42": "unknown"}, "prosody"},
		{"Nginx", "2026/10/03 10:31:09 [%s] 115#115: could not build optimal variables_hash", map[string]string{"debug": "debug", "info": "info", "notice": "info", "warn": "warn", "error": "error", "crit": "fatal", "alert": "fatal", "emerg": "fatal", "CUSTOM42": "unknown"}, "nginx"},
		{"Apache", "[Sat Oct 03 10:30:41.019843 2026] [mpm_prefork:%s] [pid 1:tid 1] resuming normal operations", map[string]string{"emerg": "fatal", "alert": "fatal", "crit": "fatal", "error": "error", "warn": "warn", "notice": "info", "info": "info", "debug": "debug", "trace1": "trace", "trace2": "trace", "trace3": "trace", "trace4": "trace", "trace5": "trace", "trace6": "trace", "trace7": "trace", "trace8": "trace", "CUSTOM42": "unknown"}, "apache"},
		{"Redis", "1:C 03 Oct 2026 10:30:40.487 %s Configuration loaded", map[string]string{".": "debug", "-": "info", "*": "info", "#": "warn"}, "redis"},
		{"MariaDB", "2026-10-03 10:30:41 0 [%s] Starting MariaDB", map[string]string{"Note": "info", "Warn": "warn", "Warning": "warn", "Error": "error", "ERROR": "error", "CUSTOM42": "unknown"}, "mariadb"},
		{"Vaultwarden", "[2026-10-03 10:31:17.421][start][%s] Rocket has launched", map[string]string{"TRACE": "trace", "DEBUG": "debug", "INFO": "info", "WARN": "warn", "ERROR": "error", "CUSTOM42": "unknown"}, "vaultwarden"},
	} {
		t.Run(testCase.service, func(t *testing.T) {
			interpreter := &LogLineInterpreterImpl{parser: logFormats[testCase.mode]}
			for level, normalizedLevel := range testCase.levels {
				t.Run(level, func(t *testing.T) {
					record := interpreter.Parse(fmt.Sprintf(testCase.format, level), LogRecordContext{})
					assert.True(t, record.Parsed)
					assert.Equal(t, normalizedLevel, record.Level)
				})
			}
		})
	}
}

func TestInterpret_FirstPlainTextLine(t *testing.T) {
	for _, line := range []string{"plain text line", "plain text with field=value", `level=info unrelated text`, `msg="unterminated`, `a="one"b=two`, `msg="invalid\q"`} {
		interpreter := NewLogLineInterpreter("grafana")
		record := interpreter.Parse(line, LogRecordContext{})
		assert.Nil(t, record.Attributes)
		assert.Equal(t, line, record.Message)
		assert.False(t, record.Parsed)
	}
}

func TestInterpret_UsesOnlyDeclaredTimestampLayout(t *testing.T) {
	for _, tc := range []struct{ service, line, timestamp string }{
		{"grist", `{"timestamp":"2026-10-03 10:29:55.285","message":"ready"}`, "2026-10-03T10:29:55.285Z"},
		{"wikijs", `{"timestamp":"2026-10-03T10:29:55.285Z","message":"ready"}`, "2026-10-03T10:29:55.285Z"},
		{"vikunja", `{"time":"2026-10-03T10:29:55.285Z","msg":"ready"}`, "2026-10-03T10:29:55.285Z"},
		{"sampleapp", `{"timestamp":"2026-10-03T10:29:55.285Z","message":"ready"}`, "2026-10-03T10:29:55.285Z"},
		{"hedgedoc", "2025-01-23T01:23:23.123456789+01:00 INFO ready", "2025-01-23T00:23:23.123456789Z"},
		{"victorialogs", `{"ts":"2025-01-23T01:23:23.123Z","level":"info","msg":"ready"}`, "2025-01-23T01:23:23.123Z"},
		{"victoriametrics", `{"ts":"2025-01-23T01:23:23.123Z","level":"info","msg":"ready"}`, "2025-01-23T01:23:23.123Z"},
		{"hedgedoc", "2025-01-23 01:23:23 INFO ready", "2025-01-23T01:23:23Z"},
		{"hedgedoc", "2025-01-23T01:23:23 INFO ready", "2025-01-23T01:23:23Z"},
		{"zulip", "2025-01-23 01:23:23,123 INFO ready", "2025-01-23T01:23:23.123Z"},
	} {
		t.Run(tc.service, func(t *testing.T) {
			record := NewLogLineInterpreter(tc.service).Parse(tc.line, LogRecordContext{})
			assert.True(t, record.Parsed)
			assert.Equal(t, testTimestamp(t, tc.timestamp), record.LogTimestamp)
			assert.Equal(t, "ready", record.Message)
		})
	}
	for _, tc := range []struct{ service, line string }{
		{"grist", `{"timestamp":"2026-10-03T10:29:55Z"}`},
		{"wikijs", `{"timestamp":"2026-10-03 10:29:55"}`},
		{"rabbitmq", `{"time":"2026-10-03T10:29:55Z","level":"info","msg":"ready"}`},
		{"vikunja", `{"time":"2026-10-03 10:29:55","level":"info","msg":"ready"}`},
		{"zulip", "2025-01-23T01:23:23Z INFO ready"},
	} {
		t.Run(tc.service+tc.line, func(t *testing.T) {
			record := NewLogLineInterpreter(tc.service).Parse(tc.line, LogRecordContext{})
			assert.False(t, record.Parsed)
			assert.Equal(t, tc.line, record.Message)
			assert.Nil(t, record.Attributes)
			assert.Nil(t, record.LogTimestamp)
		})
	}
}

func testTimestamp(t *testing.T, value string) *time.Time {
	timestamp, err := time.Parse(time.RFC3339Nano, value)
	assert.Nil(t, err)
	return &timestamp
}

func TestInterpret_PostgresTimezones(t *testing.T) {
	for _, tc := range []struct{ zone, timestamp string }{
		{"BST", "2026-10-03T09:29:44.176Z"},
		{"UTC", "2026-10-03T10:29:44.176Z"},
		{"GMT", "2026-10-03T10:29:44.176Z"},
	} {
		record := NewLogLineInterpreter("postgres").Parse("2026-10-03 10:29:44.176 "+tc.zone+" [1] LOG:  ready", LogRecordContext{})
		assert.True(t, record.Parsed)
		assert.Equal(t, testTimestamp(t, tc.timestamp), record.LogTimestamp)
	}
}

func TestServiceLogFormats_ReferenceExistingFormats(t *testing.T) {
	for service, format := range serviceLogFormats {
		t.Run(service, func(t *testing.T) {
			if format != "" {
				assert.NotNil(t, logFormats[format])
			}
		})
	}
}

func TestInterpret_JitsiComponentsAcceptS6Logs(t *testing.T) {
	for _, service := range []string{"jvb", "jicofo", "prosody"} {
		t.Run(service, func(t *testing.T) {
			record := NewLogLineInterpreter(service).Parse("s6-rc: info: service legacy-services successfully started", LogRecordContext{ReadTimestamp: receivedAt})
			assert.Equal(t, &api.LogRecord{
				ReadTimestamp: receivedAt, Level: "info",
				Message: "service legacy-services successfully started", Parsed: true,
			}, record)
		})
	}
}

func TestInterpret_KnownServiceVariants(t *testing.T) {
	for _, tc := range []struct{ service, line string }{
		{"nextcloud", `127.0.0.1 - - [03/Oct/2026:10:30:41 +0000] "GET / HTTP/1.1" 200 12 "-" "-"`},
		{"jitsi", `127.0.0.1 - - [03/Oct/2026:10:30:41 +0000] "GET / HTTP/1.1" 200 12 "-" "-"`},
		{"wordpress", `[Sat Oct 03 10:30:41 2026] [core:info] [pid 12] ready`},
		{"wordpress", `127.0.0.1 - - [03/Oct/2026:10:30:41 +0000] "GET / HTTP/1.1" 200 12`},
		{"mariadb", `2026-10-03 10:30:40+01:00 [Note] [Entrypoint]: Starting MariaDB`},
	} {
		t.Run(tc.service+tc.line, func(t *testing.T) {
			record := NewLogLineInterpreter(tc.service).Parse(tc.line, LogRecordContext{ReadTimestamp: receivedAt})
			assert.True(t, record.Parsed)
			assert.NotNil(t, record.LogTimestamp)
			assert.Equal(t, receivedAt, record.ReadTimestamp)
			switch tc.service {
			case "mariadb":
				assert.Equal(t, "Starting MariaDB", record.Message)
				assert.Equal(t, "Entrypoint", record.Attributes["component"])
			default:
				if strings.HasPrefix(tc.line, "[") {
					assert.Equal(t, "ready", record.Message)
					assert.Equal(t, "info", record.Level)
				} else {
					assert.Equal(t, 200, record.Attributes["http.response.status_code"])
					assert.Equal(t, "/", record.Attributes["url.original"])
				}
			}
		})
	}
}

func TestInterpret_HedgedocMalformedAccessLog(t *testing.T) {
	for _, message := range []string{
		`127.0.0.1 - - [08/Oct/2026:09:30:29 +0000] "GET /_health HTTP/1.1" 200 invalid "-" "healthcheck"`,
		`127.0.0.1 - - [31/Feb/2026:09:30:29 +0000] "GET /_health HTTP/1.1" 200 14 "-" "healthcheck"`,
	} {
		t.Run(message, func(t *testing.T) {
			record := NewLogLineInterpreter("hedgedoc").Parse("2026-10-08T09:30:29.967Z info: \t"+message, LogRecordContext{})
			assert.True(t, record.Parsed)
			assert.Equal(t, testTimestamp(t, "2026-10-08T09:30:29.967Z"), record.LogTimestamp)
			assert.Equal(t, "info", record.Level)
			assert.Equal(t, message, record.Message)
			assert.Nil(t, record.Attributes)
		})
	}
}

func TestInterpret_TimestampLayouts(t *testing.T) {
	for _, tc := range []struct{ line, timestamp string }{
		{"2025-01-23 01:23:23.123 INFO something happened", "2025-01-23T01:23:23.123Z"},
		{"2025-01-23T01:23:23.123 INFO something happened", "2025-01-23T01:23:23.123Z"},
		{"2025-01-23 01:23:23 INFO something happened", "2025-01-23T01:23:23Z"},
		{"2025-01-23T01:23:23 INFO something happened", "2025-01-23T01:23:23Z"},
		{"2025-01-23T01:23:23.123456789+01:00 INFO something happened", "2025-01-23T00:23:23.123456789Z"},
		{"2025-01-23 01:23:23.123456789-02:00 INFO something happened", "2025-01-23T03:23:23.123456789Z"},
		{"2025-01-23 01:23:23,123456789 INFO something happened", "2025-01-23T01:23:23.123456789Z"},
	} {
		record := NewLogLineInterpreter("hedgedoc").Parse(tc.line, LogRecordContext{})
		assert.True(t, record.Parsed)
		assert.Equal(t, testTimestamp(t, tc.timestamp), record.LogTimestamp)
		assert.Equal(t, "info", record.Level)
		assert.Equal(t, "something happened", record.Message)
	}
}
