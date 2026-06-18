package component

import (
	"fmt"
	"os/exec"
	"server/tools"
	"strings"
	"testing"

	"github.com/quollix/common/assert"
)

type DockerObject int

const (
	Network DockerObject = iota
	Volume
	Container
	ComposeStack
)

func ExpectDockerObject(t *testing.T, dockerObject DockerObject, shouldExist bool) {
	var object string
	var expectedSampleValue string
	switch dockerObject {
	case Network:
		object = "network"
		expectedSampleValue = tools.SampleAppDockerNetwork
	case Volume:
		object = "volume"
		expectedSampleValue = tools.SampleAppDockerVolume
	case Container:
		object = "container"
		containerName := fmt.Sprintf("%s_%s_%s", tools.SampleMaintainer, tools.SampleApp, tools.SampleApp)
		expectedSampleValue = containerName
	case ComposeStack:
		object = "compose"
		expectedSampleValue = tools.SampleMaintainer + "_" + tools.SampleApp
	}

	var dockerObjectExistenceCheckCommand string
	if dockerObject == ComposeStack {
		cmd := exec.Command("docker", "compose", "-p", expectedSampleValue, "ps", "--quiet") // #nosec G204 (CWE-78): Subprocess launched with variable; acceptable for testing purposes
		output, err := cmd.Output()
		assert.Nil(t, err)
		trimmedOutput := strings.TrimSpace(string(output))
		if shouldExist {
			assert.NotEqual(t, "", trimmedOutput)
		} else {
			assert.Equal(t, "", trimmedOutput)
		}
	} else {
		dockerObjectExistenceCheckCommand = fmt.Sprintf("docker %s inspect %s", object, expectedSampleValue)
		err := exec.Command("/bin/sh", "-c", dockerObjectExistenceCheckCommand).Run() // #nosec G204 (CWE-78): Subprocess launched with variable; acceptable for testing purposes
		if shouldExist {
			assert.Nil(t, err)
		} else {
			assert.NotNil(t, err)
		}
	}

}
