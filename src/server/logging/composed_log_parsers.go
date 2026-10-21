package logging

import (
	"fmt"
	"maps"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	rabbitmqDurationPattern = regexp.MustCompile(`^(?:(\d+)D, )?(\d+)H, (\d+)M, (\d+)s$`)
	apacheDiagnosticParser  = newLogParser(parseApacheDiagnostic, "")
	s6RecordParser          = newLogParser(parseS6Log, "")
	rabbitmqJSONParser      = jsonLogParser("time", dateTimeOffsetLayout)
	rabbitmqTextParser      = newLogParser(parseRabbitmqTextLog, dateTimeOffsetLayout)
	apacheRecordParser      = fieldLogParser(parseApacheLog, logSchema{"timestamp", "log.level", "message", time.ANSIC})
	accessRecordParser      = newLogParser(parseHTTPAccessLog, accessTimestampLayout)
	nginxRecordParser       = newLogParser(parseNginxLog, slashTimestampLayout)
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
	parsed, ok := rabbitmqTextParser(line)
	if !ok {
		return nil, false
	}
	if fields, ok := parseRabbitClose(parsed.Message); ok {
		if duration, exists := fields["connection.duration"].(string); exists {
			delete(fields, "connection.duration")
			if nanoseconds, ok := parseRabbitmqDuration(duration); ok {
				fields["event.duration"] = nanoseconds
			}
		}
		maps.Copy(parsed.Attributes, fields)
	}
	return parsed, true
}

func parseRabbitmqDuration(value string) (int64, bool) {
	parts := rabbitmqDurationPattern.FindStringSubmatch(value)
	if parts == nil {
		return 0, false
	}
	days, hours, minutes := 0, 0, 0
	var err error
	if parts[1] != "" {
		days, err = strconv.Atoi(parts[1])
		if err != nil {
			return 0, false
		}
	}
	if hours, err = strconv.Atoi(parts[2]); err != nil {
		return 0, false
	}
	if minutes, err = strconv.Atoi(parts[3]); err != nil {
		return 0, false
	}
	seconds, err := strconv.Atoi(parts[4])
	if err != nil {
		return 0, false
	}
	duration, err := time.ParseDuration(fmt.Sprintf("%dh%dm%ds", days*24+hours, minutes, seconds))
	if err != nil {
		return 0, false
	}
	return int64(duration), true
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
