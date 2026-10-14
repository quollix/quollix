package logging

import (
	"bufio"
	"bytes"
	"io"
	"strings"
	"time"

	"github.com/moby/moby/api/pkg/stdcopy"
	u "github.com/quollix/common/utils"
)

type LogStreamContext struct {
	Reader      io.ReadCloser
	Multiplexed bool
	Interpreter LogLineInterpreter
	Maintainer  string
	App         string
	Service     string
}

type LogStreamReader interface {
	ReadStream(ctx *LogStreamContext) error
}

type LogStreamReaderImpl struct {
	LogSink LogSink
}

func (r *LogStreamReaderImpl) ReadStream(ctx *LogStreamContext) error {
	defer u.Close(ctx.Reader)
	if ctx.Multiplexed {
		// Keeping stdout and stderr separate preserves each record's source channel and prevents interleaved Docker frames from mixing partial lines.
		stdout := &logLineWriter{send: func(line string) { r.sendLine(ctx, line, "stdout") }}
		stderr := &logLineWriter{send: func(line string) { r.sendLine(ctx, line, "stderr") }}
		_, err := stdcopy.StdCopy(stdout, stderr, ctx.Reader)
		stdout.flush()
		stderr.flush()
		if err != nil {
			return u.Logger.NewError(err.Error())
		}
		return nil
	}

	// TTY output has no channel information, so its stream stays unknown.
	scanner := bufio.NewScanner(ctx.Reader)
	for scanner.Scan() {
		r.sendLine(ctx, scanner.Text(), "")
	}
	if err := scanner.Err(); err != nil {
		return u.Logger.NewError(err.Error())
	}
	return nil
}

func (r *LogStreamReaderImpl) sendLine(ctx *LogStreamContext, line, stream string) {
	if strings.TrimSpace(line) == "" {
		return
	}
	readTimestamp := time.Now().UTC()
	record := ctx.Interpreter.Parse(line, LogRecordContext{
		Maintainer:    ctx.Maintainer,
		App:           ctx.App,
		Service:       ctx.Service,
		Stream:        stream,
		ReadTimestamp: readTimestamp,
	})
	r.LogSink.Send(*record)
}

type logLineWriter struct {
	pending []byte
	send    func(string)
}

func (w *logLineWriter) Write(data []byte) (int, error) {
	size := len(data)
	for len(data) > 0 {
		line, rest, complete := bytes.Cut(data, []byte{'\n'})
		if len(w.pending)+len(line) >= bufio.MaxScanTokenSize {
			w.pending = nil
			return size - len(data), bufio.ErrTooLong
		}
		if !complete {
			w.pending = append(w.pending, line...)
			break
		}
		if len(w.pending) > 0 {
			w.pending = append(w.pending, line...)
			w.flush()
		} else {
			w.send(string(bytes.TrimSuffix(line, []byte{'\r'})))
		}
		data = rest
	}
	return size, nil
}

func (w *logLineWriter) flush() {
	if len(w.pending) > 0 {
		w.send(string(bytes.TrimSuffix(w.pending, []byte{'\r'})))
		w.pending = w.pending[:0]
	}
}
