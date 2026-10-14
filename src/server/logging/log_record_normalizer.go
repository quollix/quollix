package logging

import (
	"strings"
	"time"

	"github.com/quollix/common/quollix/api"
)

type LogRecordContext struct {
	ReadTimestamp time.Time
	Maintainer    string
	App           string
	Service       string
	Stream        string
}

func normalizeLogRecord(line *parsedLogLine, original string, context LogRecordContext) *api.LogRecord {
	record := &api.LogRecord{
		ReadTimestamp: context.ReadTimestamp.UTC(),
		Level:         "unknown",
		Message:       original,
		Maintainer:    context.Maintainer,
		App:           context.App,
		Service:       context.Service,
		Stream:        context.Stream,
	}
	if line == nil {
		return record
	}
	record.LogTimestamp = line.Timestamp
	record.Level = normalizeLogLevel(line.Level)
	record.Message = line.Message
	if len(line.Attributes) > 0 {
		record.Attributes = line.Attributes
	}

	record.Parsed = true
	return record
}

func normalizeLogLevel(level string) string {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "trace", "t", "trace1", "trace2", "trace3", "trace4", "trace5", "trace6", "trace7", "trace8", "finer", "finest", "silly":
		return "trace"
	case "debug", "d", "debug1", "debug2", "debug3", "debug4", "debug5", "fine", "config", "verbose", "http", ".":
		return "debug"
	case "info", "i", "notice", "note", "log", "*", "-":
		return "info"
	case "warn", "warning", "w", "#":
		return "warn"
	case "error", "err", "e", "severe":
		return "error"
	case "fatal", "f", "panic", "critical", "crit", "alert", "emergency", "emerg":
		return "fatal"
	default:
		return "unknown"
	}
}
