package logging

import (
	"encoding/json"
	"strings"
)

func jsonLogParser(timestampField, layout string) logParser {
	return fieldLogParser(parseJSON, logSchema{timestampField, "level", "message", layout})
}

func parseJSON(line string) (map[string]any, bool) {
	var fields map[string]any
	if err := json.Unmarshal([]byte(line), &fields); err != nil || fields == nil {
		return nil, false
	}
	if _, exists := fields["message"]; !exists {
		if message, exists := fields["msg"]; exists {
			fields["message"] = message
			delete(fields, "msg")
		}
	}
	return fields, true
}

func parseVectorLog(line string) (map[string]any, bool) {
	fields, ok := parseJSON(line)
	if !ok {
		return nil, false
	}
	// Keep the first level (severity), since Vector can emit a second level field containing its logging configuration.
	decoder := json.NewDecoder(strings.NewReader(line))
	_, _ = decoder.Token()
	for decoder.More() {
		key, _ := decoder.Token()
		var value any
		_ = decoder.Decode(&value)
		if key == "level" {
			fields["level"] = value
			break
		}
	}
	return fields, true
}
