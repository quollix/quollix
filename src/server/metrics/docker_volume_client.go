package metrics

import (
	"context"
	"time"

	"github.com/moby/moby/api/types/volume"
	"github.com/moby/moby/client"
	u "github.com/quollix/common/utils"
)

type DockerVolumeClient interface {
	ReadUsage() ([]volume.Volume, error)
}

type DockerVolumeUsageAPI interface {
	DiskUsage(context.Context, client.DiskUsageOptions) (client.DiskUsageResult, error)
}

type DockerVolumeClientImpl struct {
	Client         DockerVolumeUsageAPI
	RequestTimeout time.Duration
}

func NewDockerVolumeClient(client DockerVolumeUsageAPI) *DockerVolumeClientImpl {
	return &DockerVolumeClientImpl{Client: client, RequestTimeout: time.Hour}
}

func (c *DockerVolumeClientImpl) ReadUsage() ([]volume.Volume, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.RequestTimeout)
	defer cancel()
	result, err := c.Client.DiskUsage(ctx, client.DiskUsageOptions{Volumes: true, Verbose: true})
	if err != nil {
		return nil, u.Logger.NewError(err.Error())
	}
	return result.Volumes.Items, nil
}
