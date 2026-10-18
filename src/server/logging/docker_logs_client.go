package logging

import (
	"context"
	"io"
	"strings"
	"time"

	"github.com/containerd/errdefs"
	"github.com/moby/moby/client"
	u "github.com/quollix/common/utils"
)

const containerLogLookback = 30 * time.Second

type DockerLogsClient interface {
	ListRunningContainers(ctx context.Context) ([]string, error)
	OpenContainerLogStream(ctx context.Context, name string) (*ContainerLogStream, bool, error)
}

type DockerLogsClientImpl struct {
	Client *client.Client
}

func (d *DockerLogsClientImpl) ListRunningContainers(ctx context.Context) ([]string, error) {
	requestContext, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	result, err := d.Client.ContainerList(requestContext, client.ContainerListOptions{})
	if err != nil {
		return nil, u.Logger.NewError(err.Error())
	}
	var names []string
	for _, container := range result.Items {
		for _, name := range container.Names {
			names = append(names, strings.TrimPrefix(name, "/"))
		}
	}
	return names, nil
}

func (d *DockerLogsClientImpl) OpenContainerLogStream(ctx context.Context, containerName string) (*ContainerLogStream, bool, error) {
	inspectContext, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	result, err := d.Client.ContainerInspect(inspectContext, containerName, client.ContainerInspectOptions{})
	if err != nil {
		if errdefs.IsNotFound(err) {
			return nil, false, nil
		}
		return nil, false, u.Logger.NewError(err.Error())
	}
	streamContext, cancelStream := context.WithCancel(ctx)
	connectionTimer := time.AfterFunc(10*time.Second, cancelStream)
	// Replay recent output to cover discovery delays and brief disconnects; reconnects may duplicate logs.
	logs, err := d.Client.ContainerLogs(streamContext, containerName, client.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
		Tail:       "all",
		Since:      time.Now().Add(-containerLogLookback).UTC().Format(time.RFC3339Nano),
	})
	connectionTimer.Stop()
	if err != nil {
		cancelStream()
		// Maintenance can remove a container between discovery, inspection and opening its logs.
		if errdefs.IsNotFound(err) || strings.Contains(err.Error(), "can not get logs from container which is dead or marked for removal") {
			return nil, false, nil
		}
		return nil, false, u.Logger.NewError(err.Error())
	}
	return &ContainerLogStream{
		ReadCloser:  logs,
		Multiplexed: result.Container.Config == nil || !result.Container.Config.Tty,
		cancel:      cancelStream,
	}, true, nil
}

type ContainerLogStream struct {
	io.ReadCloser
	Multiplexed bool
	cancel      context.CancelFunc
}

func (s *ContainerLogStream) Close() error {
	s.cancel()
	return s.ReadCloser.Close()
}
