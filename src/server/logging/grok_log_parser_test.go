package logging

import (
	"maps"
	"testing"

	"github.com/quollix/common/assert"
)

func TestGrokParser_InvalidPattern(t *testing.T) {
	for _, pattern := range []string{`%{UNKNOWN_PATTERN}`, `(`} {
		t.Run(pattern, func(t *testing.T) {
			parse := grokParser(pattern)
			fields, ok := parse("original log line")
			assert.False(t, ok)
			assert.Nil(t, fields)
		})
	}
}

func FuzzInterpret(f *testing.F) {
	services := maps.Clone(serviceLogFormats)
	for name := range logFormats {
		if _, exists := services[name]; !exists {
			services[name] = name
		}
	}
	for _, line := range []string{
		"", "[", "plain text", `msg="escaped \\"`,
		`2026-10-03 10:30:41 0 [Note] Starting MariaDB`,
		`2026-10-03 10:30:40+01:00 [Note] [Entrypoint]: ready`,
		`[Sat Oct 03 10:30:41 2026] [core:info] [pid 12] ready`,
		`127.0.0.1 - - [03/Oct/2026:10:30:41 +0000] "GET / HTTP/1.1" 200 12 "-" "-"`,
		`JVB 2026-10-03 10:31:14.002 WARNING: [35] [a=b] Source: ready`,
		`2026/10/03 10:29:27.226036 [info] ready a="two words"`,
		`{"timestamp":"2026-10-03T10:30:41Z","level":"info","message":"ready"}`,
	} {
		f.Add(line)
	}
	f.Fuzz(func(t *testing.T, line string) {
		for service := range services {
			record := NewLogLineInterpreter(service).Parse(line, LogRecordContext{ReadTimestamp: receivedAt, Service: service})
			assert.Equal(t, receivedAt, record.ReadTimestamp)
			assert.Equal(t, service, record.Service)
			if !record.Parsed {
				assert.Equal(t, line, record.Message)
				assert.Equal(t, "unknown", record.Level)
				assert.Nil(t, record.Attributes)
				assert.Nil(t, record.LogTimestamp)
			}
		}
	})
}

func BenchmarkInterpret(b *testing.B) {
	for _, tc := range []struct{ service, line string }{
		{"mariadb", `2026-10-03 10:30:41 0 [Note] Starting MariaDB`},
		{"nextcloud", `[Sat Oct 03 10:30:41.019843 2026] [mpm_prefork:notice] [pid 1:tid 1] AH00163: Apache configured`},
		{"wordpress", `127.0.0.1 - - [03/Oct/2026:10:30:41 +0000] "GET / HTTP/1.1" 200 12 "-" "client"`},
		{"adguard", `2026/10/03 10:29:27.226036 [info] dnsproxy: ready a=b c="two words"`},
		{"jvb", `JVB 2026-10-03 10:31:14.002 WARNING: [35] [a=b] Source: ready`},
	} {
		b.Run(tc.service, func(b *testing.B) {
			interpreter := NewLogLineInterpreter(tc.service)
			b.ReportAllocs()
			for b.Loop() {
				interpreter.Parse(tc.line, LogRecordContext{ReadTimestamp: receivedAt})
			}
		})
	}
}
