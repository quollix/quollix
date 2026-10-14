package metrics

import (
	"context"
	"encoding/json"
	"time"

	containertypes "github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
	u "github.com/quollix/common/utils"
)

type DockerStatsClient interface {
	ListRunningContainers() ([]containertypes.Summary, error)
	ReadStats(string) (*containertypes.StatsResponse, error)
}

type DockerContainerStatsAPI interface {
	ContainerList(context.Context, client.ContainerListOptions) (client.ContainerListResult, error)
	ContainerStats(context.Context, string, client.ContainerStatsOptions) (client.ContainerStatsResult, error)
}

type DockerStatsClientImpl struct {
	Client         DockerContainerStatsAPI
	RequestTimeout time.Duration
}

func NewDockerStatsClient(client DockerContainerStatsAPI) *DockerStatsClientImpl {
	return &DockerStatsClientImpl{Client: client, RequestTimeout: 10 * time.Second}
}

func (d *DockerStatsClientImpl) ListRunningContainers() ([]containertypes.Summary, error) {
	requestContext, cancel := context.WithTimeout(context.Background(), d.RequestTimeout)
	defer cancel()
	result, err := d.Client.ContainerList(requestContext, client.ContainerListOptions{})
	if err != nil {
		return nil, u.Logger.NewError(err.Error())
	}
	return result.Items, nil
}

func (d *DockerStatsClientImpl) ReadStats(containerID string) (*containertypes.StatsResponse, error) {
	requestContext, cancel := context.WithTimeout(context.Background(), d.RequestTimeout)
	defer cancel()
	result, err := d.Client.ContainerStats(requestContext, containerID, client.ContainerStatsOptions{Stream: false})
	if err != nil {
		return nil, u.Logger.NewError(err.Error())
	}
	defer u.Close(result.Body)
	var stats containertypes.StatsResponse
	if err := json.NewDecoder(result.Body).Decode(&stats); err != nil {
		return nil, u.Logger.NewError(err.Error())
	}
	return &stats, nil
}
