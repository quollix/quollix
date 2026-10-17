package logging

import (
	"maps"
	"strings"
	"time"
)

var (
	apacheDiagnosticParser = newLogParser(parseApacheDiagnostic, "")
	s6RecordParser         = newLogParser(parseS6Log, "")
	rabbitmqJSONParser     = jsonLogParser("time", dateTimeOffsetLayout)
	rabbitmqTextParser     = newLogParser(parseRabbitmqTextLog, dateTimeOffsetLayout)
	apacheRecordParser     = fieldLogParser(parseApacheLog, logSchema{"timestamp", "log.level", "message", time.ANSIC})
	accessRecordParser     = newLogParser(parseHTTPAccessLog, accessTimestampLayout)
	nginxRecordParser      = newLogParser(parseNginxLog, slashTimestampLayout)
)

func parseProsodyFields(line string) (map[string]any, bool) {
	if fields, ok := parseProsodyLog(line); ok {
		return fields, true
	}
	if fields, ok := parseS6Log(line); ok {
		return fields, true
	}
	return parseComponentMessageLog(line)
}

func parseZulipFields(line string) (map[string]any, bool) {
	if fields, ok := parseTimestampLevelLog(line); ok {
		return fields, true
	}
	return parseLevelMessageLog(line)
}

func parseJitsiContainerLog(line string) (*parsedLogLine, bool) {
	if strings.HasPrefix(line, "s6-rc:") || strings.HasPrefix(line, "/package/admin/s6-overlay/libexec/preinit:") {
		return s6RecordParser(line)
	}
	return parseNginxContainerLog(line)
}

func parseRabbitmqLog(line string) (*parsedLogLine, bool) {
	if strings.HasPrefix(strings.TrimSpace(line), "{") {
		return rabbitmqJSONParser(line)
	}
	return rabbitmqTextParser(line)
}

func parseApacheContainerLog(line string) (*parsedLogLine, bool) {
	if strings.HasPrefix(line, "AH") {
		return apacheDiagnosticParser(line)
	}
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
		return parseS6Log(line)
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
