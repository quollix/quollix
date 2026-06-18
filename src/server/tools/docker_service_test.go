//go:build integration

package tools

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/quollix/common/assert"
)

var (
	sampleMaintainer    = "samplemaintainer"
	sampleApp           = "sampleapp"
	sampleNetwork       = sampleMaintainer + "_" + sampleApp
	sampleContainerName = fmt.Sprintf("%s_%s_%s", sampleMaintainer, sampleApp, sampleApp)
	sampleImageTag      = "quollix-test-image:latest"

	dockerService = &DockerServiceImpl{}
)

func cleanup(t *testing.T) {
	err := exec.Command("docker", "rm", "-f", BrandAppContainerName).Run()
	assert.Nil(t, err)
	dockerService.RemoveNetwork(sampleMaintainer, sampleApp)
	err = exec.Command("docker", "rmi", "-f", sampleImageTag).Run()
	assert.Nil(t, err)
}

func TestDockerService_CreateAndRemoveNetwork(t *testing.T) {
	defer cleanup(t)
	assert.False(t, dockerNetworkExists(sampleNetwork))

	dockerService.CreateDockerNetwork(sampleMaintainer, sampleApp)
	assert.True(t, dockerNetworkExists(sampleNetwork))

	dockerService.RemoveNetwork(sampleMaintainer, sampleApp)
	assert.False(t, dockerNetworkExists(sampleNetwork))
}

func dockerNetworkExists(networkName string) bool {
	command := exec.Command("docker", "network", "inspect", networkName)
	return command.Run() == nil
}

type dockerContainerInspect struct {
	NetworkSettings struct {
		Networks map[string]dockerNetworkSettings `json:"Networks"`
	} `json:"NetworkSettings"`
}

type dockerNetworkSettings struct {
	Aliases []string `json:"Aliases"`
}

func TestDockerService_AttachBrandAppToNetwork(t *testing.T) {
	defer cleanup(t)
	host := "example.com"
	appNetworkAlias := sampleApp + "." + host
	hostNetworkAlias := BrandAppDomainPrefix + host

	err := exec.Command("docker", "rm", "-f", BrandAppContainerName).Run()
	assert.Nil(t, err)
	dockerService.RemoveNetwork(sampleMaintainer, sampleApp)

	dockerService.CreateDockerNetwork(sampleMaintainer, sampleApp)
	assert.True(t, dockerNetworkExists(sampleNetwork))

	_, err = exec.Command(
		"docker", "run", "-d",
		"--name", BrandAppContainerName,
		"alpine:latest",
		"sh", "-c", "while true; do sleep 60; done",
	).CombinedOutput()
	assert.Nil(t, err)

	dockerService.AttachBrandAppToNetwork(sampleMaintainer, sampleApp, host)

	containerInspectOutput, err := exec.Command("docker", "inspect", BrandAppContainerName).CombinedOutput()
	assert.Nil(t, err)

	var containerInspects []dockerContainerInspect
	assert.Nil(t, json.Unmarshal(containerInspectOutput, &containerInspects))
	assert.True(t, len(containerInspects) == 1)

	networkSettings, ok := containerInspects[0].NetworkSettings.Networks[sampleNetwork]
	assert.True(t, ok)

	assert.True(t, stringsSliceContains(networkSettings.Aliases, appNetworkAlias))
	assert.True(t, stringsSliceContains(networkSettings.Aliases, hostNetworkAlias))

	dockerService.DetachBrandAppFromNetwork(sampleMaintainer, sampleApp)

	containerInspectOutputAfterDisconnect, err := exec.Command("docker", "inspect", BrandAppContainerName).CombinedOutput()
	assert.Nil(t, err)

	var containerInspectsAfterDisconnect []dockerContainerInspect
	assert.Nil(t, json.Unmarshal(containerInspectOutputAfterDisconnect, &containerInspectsAfterDisconnect))
	assert.True(t, len(containerInspectsAfterDisconnect) == 1)

	_, ok = containerInspectsAfterDisconnect[0].NetworkSettings.Networks[sampleNetwork]
	assert.False(t, ok)
}

func stringsSliceContains(values []string, expected string) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
}

func TestDockerService_StartAndStopAppContainer(t *testing.T) {
	composeFileDirectory := t.TempDir()
	composeFilePath := filepath.Join(composeFileDirectory, "docker-compose.yml")
	composeYaml := fmt.Sprintf(`
services:
  %s:
    image: alpine:latest
    container_name: %s
    command: ["sh", "-c", "while true; do sleep 60; done"]
`, sampleApp, sampleContainerName)
	err := os.WriteFile(composeFilePath, []byte(composeYaml), 0o600)
	assert.Nil(t, err)
	err = exec.Command("docker", "rm", "-f", sampleContainerName).Run()
	assert.Nil(t, err)
	assert.Nil(t, dockerService.StartAppContainer(sampleMaintainer, sampleApp, composeFilePath, nil))
	assert.True(t, dockerContainerExists(sampleContainerName))
	dockerService.StopAppContainer(sampleMaintainer, sampleApp, composeFilePath)
	assert.False(t, dockerContainerExists(sampleContainerName))
}

func dockerContainerExists(containerName string) bool {
	_, err := exec.Command("docker", "inspect", containerName).CombinedOutput()
	return err == nil
}

func TestDockerService_StopAppContainers(t *testing.T) {
	err := exec.Command("docker", "rm", "-f", sampleContainerName).Run()
	assert.Nil(t, err)
	_, err = exec.Command(
		"docker", "run", "-d",
		"--name", sampleContainerName,
		"alpine:latest",
		"sh", "-c", "while true; do sleep 60; done",
	).CombinedOutput()
	assert.Nil(t, err)
	assert.True(t, dockerContainerExists(sampleContainerName))
	dockerService.StopAppContainers([]string{sampleContainerName})
	assert.False(t, dockerContainerExists(sampleContainerName))
	dockerService.StopAppContainers([]string{})
}

func TestDockerService_RemoveVolumes(t *testing.T) {
	volumeName := sampleMaintainer + "_" + sampleApp + "_volume"
	_ = exec.Command("docker", "volume", "rm", "-f", volumeName).Run()
	_, err := exec.Command("docker", "volume", "create", volumeName).CombinedOutput()
	assert.Nil(t, err)
	assert.True(t, dockerVolumeExists(volumeName))
	dockerService.RemoveVolumes([]string{volumeName})
	assert.False(t, dockerVolumeExists(volumeName))
	dockerService.RemoveVolumes([]string{})
}

func dockerVolumeExists(volumeName string) bool {
	command := exec.Command("docker", "volume", "inspect", volumeName)
	return command.Run() == nil
}

func TestDockerService_BuildDockerImage_AndDoesDockerImageExist(t *testing.T) {
	defer cleanup(t)
	buildContextDirectory := t.TempDir()

	dockerfilePath := filepath.Join(buildContextDirectory, "Dockerfile")
	err := os.WriteFile(dockerfilePath, []byte("FROM scratch\n"), 0o600)
	assert.Nil(t, err)

	imageExists, err := dockerService.DoesDockerImageExist(sampleImageTag)
	assert.Nil(t, err)
	assert.False(t, imageExists)

	err = dockerService.BuildDockerImage(sampleImageTag, dockerfilePath, false)
	assert.Nil(t, err)

	imageExists, err = dockerService.DoesDockerImageExist(sampleImageTag)
	assert.Nil(t, err)
	assert.True(t, imageExists)

	imageExists, err = dockerService.DoesDockerImageExist(sampleImageTag + "x")
	assert.Nil(t, err)
	assert.False(t, imageExists)
}
