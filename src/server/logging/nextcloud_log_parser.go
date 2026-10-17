package logging

import (
	"encoding/json"
	"maps"
	"strings"
	"time"
)

var (
	nextcloudStartupParser = newLogParser(parseLevelMessageLog, "")
	nextcloudRecordParser  = fieldLogParser(parseNextcloudMessage, logSchema{"time", "level", "message", time.RFC3339Nano})
	nextcloudLevels        = map[float64]string{0: "debug", 1: "info", 2: "warn", 3: "error", 4: "fatal"}
	parseNextcloudEnvelope = grokParser(`\[nextcloud\]\[%{BRACKET:app}\]\[%{NONNEGINT:level:int}\] %{GREEDYDATA:payload}`)
)

func parseNextcloudLog(line string) (*parsedLogLine, bool) {
	parsed, ok := parseApacheContainerLog(line)
	if !ok {
		return nextcloudStartupParser(line)
	}
	if !strings.HasPrefix(parsed.Message, "[nextcloud]") {
		return parsed, true
	}
	application, ok := nextcloudRecordParser(parsed.Message)
	if !ok {
		return nil, false
	}
	maps.Copy(parsed.Attributes, application.Attributes)
	application.Attributes = parsed.Attributes
	return application, true
}

func parseNextcloudMessage(message string) (map[string]any, bool) {
	envelope, ok := parseNextcloudEnvelope(message)
	if !ok {
		return nil, false
	}
	payload, _ := envelope["payload"].(string)
	decoder := json.NewDecoder(strings.NewReader(payload))
	var fields map[string]any
	if err := decoder.Decode(&fields); err != nil || fields == nil {
		return nil, false
	}
	text, ok := fields["message"].(string)
	if !ok {
		return nil, false
	}
	if _, ok := fields["time"].(string); !ok {
		return nil, false
	}
	level, ok := fields["level"].(float64)
	normalizedLevel, known := nextcloudLevels[level]
	if !ok || !known || level != float64(envelope["level"].(int)) || fields["app"] != envelope["app"] {
		return nil, false
	}
	if suffix := payload[decoder.InputOffset():]; suffix != "" {
		referer, ok := strings.CutPrefix(suffix, ", referer: ")
		if !ok || referer == "" {
			return nil, false
		}
		fields["http.request.referrer"] = referer
	}
	fields["level"] = normalizedLevel
	if remoteAddr, ok := fields["remoteAddr"].(string); ok && remoteAddr != "" {
		fields["message"] = strings.TrimSuffix(text, " (Remote IP: "+remoteAddr+")")
	}
	return fields, true
}
