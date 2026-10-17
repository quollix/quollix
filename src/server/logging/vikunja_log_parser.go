package logging

import (
	"strings"
	"time"
)

var (
	vikunjaJSONParser   = jsonLogParser("time", time.RFC3339Nano)
	vikunjaLogfmtParser = fieldLogParser(parseVikunjaLogfmt, logSchema{"time", "level", "msg", time.RFC3339Nano})
	xormRecordParser    = newLogParser(grokParser(`\[(?P<component>xorm)\] \[%{LEVEL:level}\]\s+%{STAMP:timestamp} %{GREEDYDATA:message}`), slashTimestampLayout)
)

func parseVikunjaLog(line string) (*parsedLogLine, bool) {
	if strings.HasPrefix(strings.TrimSpace(line), "{") {
		return vikunjaJSONParser(line)
	}
	if strings.HasPrefix(line, "[xorm]") {
		return xormRecordParser(line)
	}
	return vikunjaLogfmtParser(line)
}

func parseVikunjaLogfmt(line string) (map[string]any, bool) {
	message, fields, ok := splitLogfmtSuffix(line)
	if !ok || message != "" || fields["time"] == nil || fields["level"] == nil || fields["msg"] == nil {
		return nil, false
	}
	return fields, true
}
