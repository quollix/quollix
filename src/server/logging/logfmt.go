package logging

import (
	"encoding/json"
	"strings"
)

func logfmtKey(value string) bool {
	if value == "" {
		return false
	}
	for i := range len(value) {
		ch := value[i]
		isLetter := ch >= 'A' && ch <= 'Z' || ch >= 'a' && ch <= 'z'
		isAllowedSuffix := i > 0 && (ch >= '0' && ch <= '9' || ch == '.' || ch == '-')
		if ch != '_' && !isLetter && !isAllowedSuffix {
			return false
		}
	}
	return true
}

func splitLogfmtSuffix(line string) (string, map[string]any, bool) {
	line = strings.TrimSpace(line)
	start := 0
	for start < len(line) {
		end := strings.IndexAny(line[start:], " \t")
		if end < 0 {
			end = len(line) - start
		}
		key, _, ok := strings.Cut(line[start:start+end], "=")
		if ok && logfmtKey(key) {
			break
		}
		start += end
		for start < len(line) && (line[start] == ' ' || line[start] == '\t') {
			start++
		}
	}
	if start == len(line) {
		return line, nil, true
	}
	message := strings.TrimSpace(line[:start])
	fields := make(map[string]any)
	rest := line[start:]
	for rest != "" {
		end := strings.IndexAny(rest, "= \t")
		if end < 0 || rest[end] != '=' || !logfmtKey(rest[:end]) {
			return "", nil, false
		}
		key := rest[:end]
		rest = rest[end+1:]
		var value string
		if strings.HasPrefix(rest, `"`) {
			var ok bool
			value, rest, ok = decodeLogfmtQuoted(rest)
			if !ok || rest != "" && rest[0] != ' ' && rest[0] != '\t' {
				return "", nil, false
			}
		} else {
			end = 0
			for end < len(rest) && rest[end] != ' ' && rest[end] != '\t' {
				if rest[end] == '=' || rest[end] == '"' || rest[end] < ' ' {
					return "", nil, false
				}
				end++
			}
			value, rest = rest[:end], rest[end:]
		}
		fields[key] = value
		rest = strings.TrimLeft(rest, " \t")
	}
	return message, fields, true
}

func decodeLogfmtQuoted(line string) (string, string, bool) {
	value, rest, ok := quotedLogValue(line)
	if !ok {
		return "", "", false
	}
	for _, ch := range value {
		if ch == '\\' || ch < ' ' {
			var decoded string
			err := json.Unmarshal([]byte(line[:len(line)-len(rest)]), &decoded)
			return decoded, rest, err == nil
		}
	}
	return value, rest, true
}

func quotedLogValue(line string) (value, rest string, ok bool) {
	if !strings.HasPrefix(line, `"`) {
		return "", line, false
	}
	for i := 1; i < len(line); i++ {
		switch line[i] {
		case '\\':
			i++
		case '"':
			return line[1:i], line[i+1:], true
		}
	}
	return "", line, false
}
