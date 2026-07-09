package tools

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"strings"

	u "github.com/quollix/common/utils"
)

type CommandRunner interface {
	RunCommand(command string) (*CommandOutput, error)
}

type CommandOutput struct {
	Stdout string
	Stderr string
}

func (c CommandOutput) Combined() string {
	return c.Stdout + c.Stderr
}

type CommandRunnerImpl struct {
	Config *GlobalConfig
}

func (c *CommandRunnerImpl) RunCommand(command string) (*CommandOutput, error) {
	cmd := exec.Command("sh", "-c", command) // #nosec G204: this helper is explicitly designed to run internal shell command strings
	var stdoutBuffer bytes.Buffer
	var stderrBuffer bytes.Buffer
	if c.Config.PrintCommandOutput {
		cmd.Stdout = io.MultiWriter(os.Stdout, &stdoutBuffer)
		cmd.Stderr = io.MultiWriter(os.Stderr, &stderrBuffer)
	} else {
		cmd.Stdout = &stdoutBuffer
		cmd.Stderr = &stderrBuffer
	}
	err := cmd.Run()
	output := &CommandOutput{
		Stdout: stdoutBuffer.String(),
		Stderr: stderrBuffer.String(),
	}
	if err != nil {
		return output, u.Logger.NewError(err.Error(), "stdout", strings.TrimSpace(output.Stdout), "stderr", strings.TrimSpace(output.Stderr))
	}
	return output, nil
}
