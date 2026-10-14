package logging

import (
	"maps"
	"strings"
	"time"
)

var (
	apacheRecordParser = fieldLogParser(parseApacheLog, logSchema{"timestamp", "log.level", "message", time.ANSIC})
	accessRecordParser = newLogParser(parseHTTPAccessLog, accessTimestampLayout)
	nginxRecordParser  = newLogParser(parseNginxLog, slashTimestampLayout)
)

func parseApacheContainerLog(line string) (*parsedLogLine, bool) {
	if strings.HasPrefix(line, "[") {
		return apacheRecordParser(line)
	}
	return accessRecordParser(line)
}

func parseNginxContainerLog(line string) (*parsedLogLine, bool) {
	if len(line) > 4 && line[4] == '/' {
		return nginxRecordParser(line)
	}
	return accessRecordParser(line)
}

func parseJitsiFields(line string) (map[string]any, bool) {
	fields, ok := parseJitsiLog(line)
	if !ok {
		return nil, false
	}
	if value, exists := fields["context"].(string); exists {
		message, context, ok := splitLogfmtSuffix(value)
		if !ok || message != "" {
			return nil, false
		}
		delete(fields, "context")
		if len(context) > 0 {
			fields["context"] = context
		}
	}
	return fields, true
}

func parseAdguardFields(line string) (map[string]any, bool) {
	fields, ok := parseAdguardLog(line)
	if !ok {
		return nil, false
	}
	raw, _ := fields["message"].(string)
	message, context, ok := splitLogfmtSuffix(raw)
	if !ok {
		return nil, false
	}
	if len(context) > 0 {
		fields["context"] = context
	}
	if component, body, ok := strings.Cut(message, ": "); ok && component != "" && !strings.ContainsAny(component, " \t") {
		fields["component"] = component
		message = body
	}
	fields["message"] = message
	return fields, true
}

func parseHedgedocLog(line string) (map[string]any, bool) {
	fields, ok := parseTimestampLevelLog(line)
	if !ok {
		return nil, false
	}
	message, _ := fields["message"].(string)
	access, ok := parseHTTPAccessLog(message)
	if ok {
		if _, err := time.Parse(accessTimestampLayout, access["timestamp"].(string)); err == nil {
			// Keep the outer timestamp, which includes millisecond precision.
			delete(access, "timestamp")
			maps.Copy(fields, access)
		}
	}
	return fields, true
}
