package src

import (
	u "github.com/quollix/common/utils"
	"github.com/quollix/taskrunner"
)

var (
	sampleAppContainerName = "sampleapp"
	sampleAppImageName     = sampleAppContainerName + ":local"
	dockerImage            = "quollix/quollix:"
	Tr                     = taskrunner.GetTaskRunner()
)

func BuildLocalDockerImage() {
	Tr.Cmd().Dir(ProjectDir).AllowFail().Run("docker rmi -f %s", dockerImage+"local")
	Tr.Cmd().Dir(ServerDir).Run("docker build --target local -t %s -f %s/Dockerfile.quollix .", dockerImage+"local", dockerDir)
}

func ReleaseDockerImage(tag string) {
	image := dockerImage + tag
	Tr.Cmd().Dir(ServerDir).Run(
		"docker buildx build --platform linux/amd64,linux/arm64 --push --target release -t %s -f %s/Dockerfile.quollix .",
		image,
		dockerDir,
	)
	Tr.Log.Info("Published Docker image: %s", image)
}

func imageExistsLocally(image string) bool {
	exists, err := (&u.DockerCliWrapperImpl{}).ImageExists(image)
	if err != nil {
		Tr.Log.Error("Error executing docker command: %v", err)
		return false
	}
	return exists
}

func BuildLocalSampleAppDockerImageIfNotPresent() {
	if imageExistsLocally(sampleAppImageName) {
		Tr.Log.Info("Docker image already exists, skipping its build: " + sampleAppImageName)
	} else {
		Tr.Log.Info("building Sample Docker image: " + sampleAppImageName)
		Tr.Cmd().Dir(sampleAppDir).Env("CGO_ENABLED", "0").Env("GOOS", "linux").Env("GOARCH", "amd64").Run("go build")
		Tr.Cmd().Dir(sampleAppDir).Run("docker build -t %s .", sampleAppImageName)
	}
}
