package logging

import (
	"runtime/debug"

	"github.com/quollix/common/quollix/api"
	u "github.com/quollix/common/utils"
)

type LogLineInterpreter interface {
	Parse(line string, context LogRecordContext) *api.LogRecord
}

type LogLineInterpreterImpl struct {
	parser logParser
}

func NewLogLineInterpreter(service string) *LogLineInterpreterImpl {
	formatName := service
	if mapped, exists := serviceLogFormats[service]; exists {
		formatName = mapped
	}
	return &LogLineInterpreterImpl{parser: logFormats[formatName]}
}

func (interpreter *LogLineInterpreterImpl) Parse(line string, context LogRecordContext) (record *api.LogRecord) {
	defer func() {
		if recovered := recover(); recovered != nil {
			u.Logger.Error("log parser panicked", "panic", recovered, "line", line, "service", context.Service, "app", context.App, "maintainer", context.Maintainer, "stack", string(debug.Stack()))
			record = normalizeLogRecord(nil, line, context)
		}
	}()
	var parsed *parsedLogLine
	if interpreter.parser != nil {
		if result, ok := interpreter.parser(line); ok {
			parsed = result
		}
	}
	return normalizeLogRecord(parsed, line, context)
}
