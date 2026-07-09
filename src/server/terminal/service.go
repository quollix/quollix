package terminal

import (
	"context"
	"encoding/json"
	"io"

	u "github.com/quollix/common/utils"
)

type ExecSession interface {
	ShellName() string
	Read(outputBuffer []byte) (int, error)
	Write(inputBytes []byte) error
	Resize(cols int, rows int) error
	Close() error
}

type DockerTerminalService interface {
	RunTerminalSession(
		ctx context.Context,
		containerName string,
		readClientMessage func() (messageType int, messageBytes []byte, readError error),
		writeToClientBinary func(messageBytes []byte) error,
	) error
}

type DockerTerminalServiceImpl struct {
	DockerTerminalClient DockerTerminalClient
}

type TerminalClientMessage struct {
	Type string `json:"type"`
	Data string `json:"data"`
	Cols int    `json:"cols"`
	Rows int    `json:"rows"`
}

func (service *DockerTerminalServiceImpl) RunTerminalSession(
	ctx context.Context,
	containerName string,
	readClientMessage func() (messageType int, messageBytes []byte, readError error),
	writeToClientBinary func(messageBytes []byte) error,
) error {
	sessionContext, cancelSession := context.WithCancel(ctx)
	defer cancelSession()

	execSession, execError := service.DockerTerminalClient.StartExecShell(sessionContext, containerName)
	if execError != nil {
		return execError
	}
	defer u.Close(execSession)

	// Use CRLF (\r\n) instead of LF (\n) so the terminal cursor returns to column 0 before moving to the next line. Otherwise xterm keeps the previous column and the prompt appears indented.
	err := writeToClientBinary([]byte("Shell: " + execSession.ShellName() + "\r\n"))
	if err != nil {
		u.Logger.Error(err)
	}

	errorChannel := make(chan error, 2)

	sendError := func(runError error) {
		select {
		case errorChannel <- runError:
		default:
		}
		cancelSession()
	}

	go func() {
		outputBuffer := make([]byte, 32*1024)
		for {
			readCount, readError := execSession.Read(outputBuffer)
			if readCount > 0 {
				writeError := writeToClientBinary(outputBuffer[:readCount])
				if writeError != nil {
					sendError(writeError)
					return
				}
			}
			if readError != nil {
				if readError == io.EOF {
					sendError(nil)
					return
				}
				sendError(readError)
				return
			}
		}
	}()

	go func() {
		for {
			_, messageBytes, readError := readClientMessage()
			if readError != nil {
				sendError(readError)
				return
			}

			var terminalClientMessage TerminalClientMessage
			if json.Unmarshal(messageBytes, &terminalClientMessage) != nil {
				continue
			}

			if terminalClientMessage.Type == "resize" {
				if terminalClientMessage.Cols > 0 && terminalClientMessage.Rows > 0 {
					if resizeError := execSession.Resize(terminalClientMessage.Cols, terminalClientMessage.Rows); resizeError != nil {
						sendError(resizeError)
						return
					}
				}
				continue
			}

			if terminalClientMessage.Type == "input" {
				if writeError := execSession.Write([]byte(terminalClientMessage.Data)); writeError != nil {
					sendError(writeError)
					return
				}
			}
		}
	}()

	select {
	case <-sessionContext.Done():
		return sessionContext.Err()
	case runError := <-errorChannel:
		return runError
	}
}
