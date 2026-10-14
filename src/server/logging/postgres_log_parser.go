package logging

import (
	"maps"
	"strings"
	"time"
)

var (
	postgresRecordParser    = newLogParser(parsePostgresLog, time.DateTime)
	parsePostgresHeader     = grokParser(`%{STAMP:timestamp} %{NOTSPACE:timezone} \[%{NONNEGINT:pid:int}\] %{LEVEL:level}:\s+%{GREEDYDATA:message}`)
	parsePostgresCheckpoint = grokParser(
		`checkpoint complete: wrote %{NONNEGINT:checkpoint.buffers_written:int} buffers \(%{DECIMAL:checkpoint.buffers_percent:float}%\); ` +
			`%{NONNEGINT:checkpoint.wal_files_added:int} WAL file\(s\) added, %{NONNEGINT:checkpoint.wal_files_removed:int} removed, %{NONNEGINT:checkpoint.wal_files_recycled:int} recycled; ` +
			`write=%{DECIMAL:checkpoint.write_seconds:float} s, sync=%{DECIMAL:checkpoint.sync_seconds:float} s, total=%{DECIMAL:checkpoint.total_seconds:float} s; ` +
			`sync files=%{NONNEGINT:checkpoint.sync_files:int}, longest=%{DECIMAL:checkpoint.longest_sync_seconds:float} s, average=%{DECIMAL:checkpoint.average_sync_seconds:float} s; ` +
			`distance=%{NONNEGINT:checkpoint.distance_kb:int} kB, estimate=%{NONNEGINT:checkpoint.estimate_kb:int} kB; ` +
			`lsn=%{LSN:checkpoint.lsn}, redo lsn=%{LSN:checkpoint.redo_lsn}`)
	parsePostgresWorker = grokParser(`background worker "%{DATA:worker.name}" \(PID %{NONNEGINT:worker.pid:int}\) exited with exit code %{NONNEGINT:worker.exit_code:int}`)
)

func parsePostgresRecord(line string) (*parsedLogLine, bool) {
	parsed, ok := postgresRecordParser(line)
	if !ok {
		return nil, false
	}
	switch parsed.Attributes["timezone"] {
	case "BST":
		timestamp := parsed.Timestamp.Add(-time.Hour)
		parsed.Timestamp = &timestamp
	case "UTC", "GMT":
	default:
		return nil, false
	}
	delete(parsed.Attributes, "timezone")
	return parsed, true
}

func parsePostgresLog(line string) (map[string]any, bool) {
	fields, ok := parsePostgresHeader(line)
	if !ok {
		return nil, false
	}
	message, _ := fields["message"].(string)
	maps.Copy(fields, parsePostgresMessage(message))
	return fields, true
}

func parsePostgresMessage(message string) map[string]any {
	switch {
	case strings.HasPrefix(message, "checkpoint complete: "):
		if fields, ok := parsePostgresCheckpoint(message); ok {
			fields["message"] = "checkpoint complete"
			return fields
		}
	case strings.HasPrefix(message, "checkpoint starting: "):
		reason := strings.TrimPrefix(message, "checkpoint starting: ")
		if strings.TrimSpace(reason) != "" {
			return map[string]any{"message": "checkpoint starting", "checkpoint.reason": reason}
		}
	case strings.HasPrefix(message, `background worker "`):
		if fields, ok := parsePostgresWorker(message); ok && fields["worker.name"] != nil {
			fields["message"] = "background worker exited"
			return fields
		}
	}
	return nil
}
