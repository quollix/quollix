package terminal

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/moby/moby/client"
	u "github.com/quollix/common/utils"
)

type DockerTerminalClient interface {
	StartExecShell(ctx context.Context, containerName string) (ExecSession, error)
}

type DockerTerminalClientImpl struct {
	DockerClient *client.Client
}

func (dockerTerminalClient *DockerTerminalClientImpl) StartExecShell(ctx context.Context, containerName string) (ExecSession, error) {
	shellCommandCandidates := defaultShellCommandCandidates()

	userCandidates := []string{"0", ""}

	var lastError error
	for _, userCandidate := range userCandidates {
		selectedShellCommand, selectError := dockerTerminalClient.selectWorkingShellCommand(ctx, containerName, userCandidate, shellCommandCandidates)
		if selectError != nil {
			lastError = selectError
			continue
		}

		execSession, startError := dockerTerminalClient.startInteractiveShell(ctx, containerName, userCandidate, selectedShellCommand)
		if startError != nil {
			lastError = startError
			continue
		}

		return execSession, nil
	}

	if lastError == nil {
		lastError = errors.New("no user candidates attempted")
	}
	return nil, lastError
}

func defaultShellCommandCandidates() [][]string {
	return [][]string{
		{"/bin/bash"},
		{"/usr/bin/bash"},
		{"bash"},
		{"/bin/sh"},
		{"sh"},
		{"/bin/ash"},
		{"ash"},
		{"busybox", "sh"},
	}
}

func (dockerTerminalClient *DockerTerminalClientImpl) selectWorkingShellCommand(ctx context.Context, containerName string, execUser string, shellCommandCandidates [][]string) ([]string, error) {
	var lastError error

	for _, shellCommand := range shellCommandCandidates {
		exitCode, probeError := dockerTerminalClient.probeShellExitCode(ctx, containerName, execUser, shellCommand)
		if probeError != nil {
			lastError = probeError
			continue
		}
		if exitCode != 0 {
			lastError = errors.New("probe failed for shell: " + strings.Join(shellCommand, " "))
			continue
		}
		return shellCommand, nil
	}

	if lastError == nil {
		lastError = errors.New("no shell candidates attempted")
	}
	return nil, lastError
}

func (dockerTerminalClient *DockerTerminalClientImpl) probeShellExitCode(ctx context.Context, containerName string, execUser string, shellCommand []string) (int, error) {
	probeCommand := buildProbeCommand(shellCommand)

	probeCreateResult, probeCreateError := dockerTerminalClient.DockerClient.ExecCreate(ctx, containerName, client.ExecCreateOptions{
		User:         execUser,
		AttachStdout: true,
		AttachStderr: true,
		TTY:          false,
		Cmd:          probeCommand,
	})
	if probeCreateError != nil {
		return -1, probeCreateError
	}

	_, probeStartError := dockerTerminalClient.DockerClient.ExecStart(ctx, probeCreateResult.ID, client.ExecStartOptions{
		TTY: false,
	})
	if probeStartError != nil {
		return -1, probeStartError
	}

	return dockerTerminalClient.waitExecExitCode(ctx, probeCreateResult.ID)
}

func buildProbeCommand(shellCommand []string) []string {
	if len(shellCommand) == 2 && shellCommand[0] == "busybox" && shellCommand[1] == "sh" {
		return []string{"busybox", "sh", "-c", "exit 0"}
	}
	if len(shellCommand) == 1 {
		return []string{shellCommand[0], "-c", "exit 0"}
	}
	return append(append([]string{}, shellCommand...), "-c", "exit 0")
}

func (dockerTerminalClient *DockerTerminalClientImpl) waitExecExitCode(ctx context.Context, execId string) (int, error) {
	for {
		inspectResult, inspectError := dockerTerminalClient.DockerClient.ExecInspect(ctx, execId, client.ExecInspectOptions{})
		if inspectError != nil {
			return -1, inspectError
		}
		if !inspectResult.Running {
			return inspectResult.ExitCode, nil
		}
		select {
		case <-ctx.Done():
			return -1, ctx.Err()
		case <-time.After(25 * time.Millisecond):
		}
	}
}

func (dockerTerminalClient *DockerTerminalClientImpl) startInteractiveShell(ctx context.Context, containerName string, execUser string, shellCommand []string) (ExecSession, error) {
	execCreateResult, execCreateError := dockerTerminalClient.DockerClient.ExecCreate(ctx, containerName, client.ExecCreateOptions{
		User:         execUser,
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		TTY:          true,
		Cmd:          shellCommand,
	})
	if execCreateError != nil {
		u.Logger.Error("terminal exec create failed", "containerName", containerName, "shellCommand", strings.Join(shellCommand, " "), "error", execCreateError.Error())
		return nil, execCreateError
	}

	execAttachResult, execAttachError := dockerTerminalClient.DockerClient.ExecAttach(ctx, execCreateResult.ID, client.ExecAttachOptions{
		TTY: true,
	})
	if execAttachError != nil {
		u.Logger.Error("terminal exec attach failed", "containerName", containerName, "shellCommand", strings.Join(shellCommand, " "), "error", execAttachError.Error())
		return nil, execAttachError
	}

	_, execStartError := dockerTerminalClient.DockerClient.ExecStart(ctx, execCreateResult.ID, client.ExecStartOptions{
		TTY: true,
	})
	if execStartError != nil {
		execAttachResult.Close()
		u.Logger.Error("terminal exec start failed", "containerName", containerName, "shellCommand", strings.Join(shellCommand, " "), "error", execStartError.Error())
		return nil, execStartError
	}

	return &dockerExecSession{
		dockerClient:     dockerTerminalClient.DockerClient,
		execId:           execCreateResult.ID,
		hijackedResponse: execAttachResult.HijackedResponse,
		shellName:        strings.Join(shellCommand, " "),
	}, nil
}

type dockerExecSession struct {
	dockerClient     *client.Client
	execId           string
	hijackedResponse client.HijackedResponse
	shellName        string
}

func (dockerExecSession *dockerExecSession) ShellName() string {
	return dockerExecSession.shellName
}

func (dockerExecSession *dockerExecSession) Read(outputBuffer []byte) (int, error) {
	return dockerExecSession.hijackedResponse.Reader.Read(outputBuffer)
}

func (dockerExecSession *dockerExecSession) Write(inputBytes []byte) error {
	_, writeError := dockerExecSession.hijackedResponse.Conn.Write(inputBytes)
	return writeError
}

func (dockerExecSession *dockerExecSession) Close() error {
	dockerExecSession.hijackedResponse.Close()
	return nil
}

func (dockerExecSession *dockerExecSession) Resize(cols int, rows int) error {
	width, height, ok := clampTerminalSize(cols, rows)
	if !ok {
		return nil
	}

	_, resizeError := dockerExecSession.dockerClient.ExecResize(
		context.Background(),
		dockerExecSession.execId,
		client.ExecResizeOptions{
			Width:  width,
			Height: height,
		},
	)
	return resizeError
}

// direct conversion without clamping would lead to potential security vulnerability: G115 (CWE-190): integer overflow conversion int -> uint
func clampTerminalSize(cols int, rows int) (uint, uint, bool) {
	if cols <= 0 || rows <= 0 {
		return 0, 0, false
	}
	if cols > 1000 {
		cols = 1000
	}
	if rows > 500 {
		rows = 500
	}
	return uint(cols), uint(rows), true // #nosec
}
