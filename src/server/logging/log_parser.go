package logging

import (
	"strings"
	"time"
)

const (
	dateTimeOffsetLayout    = time.DateTime + "Z07:00"
	isoLocalTimestampLayout = time.DateOnly + "T" + time.TimeOnly
	slashTimestampLayout    = "2006/01/02 15:04:05"
	accessTimestampLayout   = "02/Jan/2006:15:04:05 -0700"
	redisTimestampLayout    = "2 Jan 2006 15:04:05"
)

// Every registered parser returns the same fields, including for nested logs.
type logParser func(string) (*parsedLogLine, bool)

type logSchema struct {
	timestampField, levelField, messageField string
	timestampLayout                          string // Empty selects the layout from the timestamp's shape.
}

type parsedLogLine struct {
	Timestamp      *time.Time
	Level, Message string
	Attributes     map[string]any
}

func newLogParser(extract func(string) (map[string]any, bool), layout string) logParser {
	return fieldLogParser(extract, logSchema{"timestamp", "level", "message", layout})
}

func fieldLogParser(extract func(string) (map[string]any, bool), schema logSchema) logParser {
	return func(line string) (*parsedLogLine, bool) {
		fields, ok := extract(line)
		if !ok {
			return nil, false
		}
		level, levelOK := logFieldString(fields, schema.levelField)
		message, messageOK := logFieldString(fields, schema.messageField)
		timestamp, timestampOK := parseLogTimestamp(fields, schema)
		if !levelOK || !messageOK || !timestampOK {
			return nil, false
		}
		for _, key := range []string{schema.timestampField, schema.levelField, schema.messageField} {
			if key != "" {
				delete(fields, key)
			}
		}
		return &parsedLogLine{Timestamp: timestamp, Level: level, Message: message, Attributes: fields}, true
	}
}

func logFieldString(fields map[string]any, key string) (string, bool) {
	if key == "" {
		return "", true
	}
	value, exists := fields[key]
	if !exists {
		return "", true
	}
	text, ok := value.(string)
	return text, ok
}

func parseLogTimestamp(fields map[string]any, schema logSchema) (*time.Time, bool) {
	if schema.timestampField == "" {
		return nil, true
	}
	value, exists := fields[schema.timestampField]
	if !exists {
		return nil, true
	}
	raw, ok := value.(string)
	if !ok {
		return nil, false
	}
	layout := schema.timestampLayout
	if layout == "" {
		layout = timestampLayout(raw)
	}
	timestamp, err := time.ParseInLocation(layout, raw, time.UTC)
	if err != nil {
		return nil, false
	}
	timestamp = timestamp.UTC()
	return &timestamp, true
}

func timestampLayout(raw string) string {
	hasZone := len(raw) > 19 && strings.ContainsAny(raw[19:], "Z+-")
	if len(raw) > 10 && raw[10] == 'T' {
		if hasZone {
			return time.RFC3339
		}
		return isoLocalTimestampLayout
	}
	if hasZone {
		return dateTimeOffsetLayout
	}
	return time.DateTime
}
