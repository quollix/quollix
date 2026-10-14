//go:build integration

package integration

import (
	"bytes"
	"context"
	"time"

	"os/exec"
	"testing"

	"github.com/moby/moby/api/pkg/stdcopy"
	"github.com/moby/moby/client"

	"server/logging"

	"github.com/quollix/common/assert"
	u "github.com/quollix/common/utils"
)

const (
	testLogLine  = "grafana-test-line"
	sampleLogger = "samplelogger"
)

func TestOpenContainerLogStream(t *testing.T) {
	startSampleLoggerContainer(t)
	defer removeSampleLoggerContainer(t)
	assertFirstLogLine(t)
}

func startSampleLoggerContainer(t *testing.T) {
	entrypoint := "echo " + testLogLine

	args := []string{
		"run", "-d",
		"--name", sampleLogger,
	}
	args = append(args, "busybox:1.37.0", "sh", "-c", entrypoint)

	cmd := exec.Command("docker", args...)
	assert.Nil(t, cmd.Run())
}

func removeSampleLoggerContainer(t *testing.T) {
	err := exec.Command("docker", "rm", "-f", sampleLogger).Run()
	assert.Nil(t, err)
}

func assertFirstLogLine(t *testing.T) {
	dockerClient, err := client.New(client.FromEnv)
	assert.Nil(t, err)
	defer u.Close(dockerClient)
	logClient := &logging.DockerLogsClientImpl{Client: dockerClient}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	r, err := logClient.OpenContainerLogStream(ctx, sampleLogger)
	assert.Nil(t, err)
	defer u.Close(r)

	assert.True(t, r.Multiplexed)
	var stdout, stderr bytes.Buffer
	_, err = stdcopy.StdCopy(&stdout, &stderr, r)
	assert.Nil(t, err)
	assert.Equal(t, testLogLine+"\n", stdout.String())
	assert.Equal(t, "", stderr.String())
}
