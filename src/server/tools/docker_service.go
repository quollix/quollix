package tools

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	u "github.com/quollix/common/utils"
)

var containerNamePattern = regexp.MustCompile(`^[a-z]{3,20}_[a-z]{3,20}_[a-z]{3,20}$`)

const (
	DockerHubRateLimitReachedErrorMessage      = "Docker Hub rate limit reached. Please try again later, consider using a Docker Hub account to increase your rate limit, or buy a Docker Hub subscription to remove rate limits."
	DockerImageUnsupportedPlatformErrorMessage = "This app cannot be started because one of its Docker images does not support this device's platform."
	DockerImageDownloadErrorMessage            = "A Docker image could not be downloaded because the container registry connection was interrupted. Please try again later and check the server's internet connection."
)

type ContainerIdentity struct {
	Maintainer string
	App        string
	Service    string
}

type DockerService interface {
	CreateDockerNetwork(networkMaintainer, networkApp string)
	RemoveNetwork(networkMaintainer, networkApp string)

	AttachBrandAppToNetwork(networkMaintainer, networkApp, host string)
	DetachBrandAppFromNetwork(networkMaintainer, networkApp string)

	StartAppContainer(maintainer, appName, composeFilePath string) error
	StopAppContainer(maintainer, appName, composeFilePath string)
	StopAppContainers(containerNames []string)
	RemoveVolumes(volumes []string)

	DoesDockerImageExist(dockerImageName string) (bool, error)
	BuildDockerImage(nameWithTag, dockerfilePath string, protectFromMaintenancePrune bool) error
}

type DockerServiceImpl struct {
	CommandRunner CommandRunner
}

func (d *DockerServiceImpl) StopAppContainers(containerNames []string) {
	if len(containerNames) == 0 {
		return
	}

	args := append([]string{"rm", "-f"}, containerNames...)
	cmd := exec.Command("docker", args...) // #nosec G204
	output, err := cmd.CombinedOutput()
	if err != nil {
		u.Logger.Error(err, "container_names", containerNames, "output", string(output))
	}
}

func (d *DockerServiceImpl) RemoveVolumes(volumes []string) {
	if len(volumes) == 0 {
		return
	}

	args := append([]string{"volume", "rm", "-f"}, volumes...)
	cmd := exec.Command("docker", args...) // #nosec G204
	output, err := cmd.CombinedOutput()
	if err != nil {
		u.Logger.Error(err, "volumes", volumes, "output", string(output))
	}
}

func (d *DockerServiceImpl) CreateDockerNetwork(networkMaintainer, networkApp string) {
	networkName := d.mergeNames(networkMaintainer, networkApp)
	cmd := exec.Command("docker", "network", "create", networkName) // #nosec G204 (CWE-78): Subprocess launched with a potential tainted input or cmd arguments
	output, err := cmd.CombinedOutput()
	if err != nil {
		if !strings.Contains(string(output), "already exists") {
			u.Logger.Error(err, "network", networkName, "output", string(output))
		}
	}
}

func (d *DockerServiceImpl) AttachBrandAppToNetwork(networkMaintainer, networkApp, host string) {
	networkName := d.mergeNames(networkMaintainer, networkApp)
	// We keep both aliases so the app can reach Quollix directly for OIDC and can also call itself by its app hostname when needed.
	appNetworkAlias := fmt.Sprintf("%s.%s", networkApp, host)
	hostNetworkAlias := fmt.Sprintf("%s%s", BrandAppDomainPrefix, host)
	cmd := exec.Command("docker", "network", "connect", networkName, "--alias", appNetworkAlias, "--alias", hostNetworkAlias, BrandAppContainerName) // #nosec G204 (CWE-78): Subprocess launched with a potential tainted input or cmd arguments
	output, err := cmd.CombinedOutput()
	if err != nil {
		if !strings.Contains(string(output), "already exists") {
			u.Logger.Error(err, "network", networkName, "container", BrandAppContainerName, "output", string(output))
		}
	}
}

func (d *DockerServiceImpl) DetachBrandAppFromNetwork(networkMaintainer, networkApp string) {
	networkName := d.mergeNames(networkMaintainer, networkApp)

	cmd := exec.Command("docker", "network", "disconnect", networkName, BrandAppContainerName) // #nosec G204 (CWE-78): Subprocess launched with a potential tainted input or cmd arguments
	output, err := cmd.CombinedOutput()
	if err != nil {
		if !strings.Contains(string(output), "not found") {
			u.Logger.Error(err, "network", networkName, "container", BrandAppContainerName, "output", string(output))
		}
	}
}

func (d *DockerServiceImpl) RemoveNetwork(networkMaintainer, networkApp string) {
	networkName := d.mergeNames(networkMaintainer, networkApp)
	if isProtectedDockerNetwork(networkName) {
		u.Logger.Info("skipping removal of protected Docker network", "network", networkName)
		return
	}

	cmd := exec.Command("docker", "network", "rm", networkName) // #nosec G204 (CWE-78): Subprocess launched with variable
	output, err := cmd.CombinedOutput()
	if err != nil {
		if !strings.Contains(string(output), "not found") {
			u.Logger.Error(err, "network", networkName, "output", string(output))
		}
	}
}

func isProtectedDockerNetwork(networkName string) bool {
	// The official Postgres network must survive even while the Postgres app is stopped for backup/restore. In component tests the dummy SSH backup server is attached to this network, so Docker refuses to remove it and the issue is masked. In production the network can be empty at that moment, so removing it breaks Docker DNS for quollix_postgres_postgres and Quollix loses DB access.
	return networkName == OfficialDatabaseAppNetworkName
}

func (d *DockerServiceImpl) StartAppContainer(maintainer, appName, composeFilePath string) error {
	u.Logger.Info("Starting app", MaintainerField, maintainer, AppField, appName, "compose_file_path", composeFilePath)
	output, err := d.CommandRunner.RunCommand(
		"docker", "compose",
		"-p", d.mergeNames(maintainer, appName),
		"-f", composeFilePath,
		"up", "-d",
	)
	if err = dockerComposeStartError(output.Combined(), err); err != nil {
		d.StopAppContainer(maintainer, appName, composeFilePath)
		return u.Logger.AddContext(err, "maintainer", maintainer, "app_name", appName, "compose_file_path", composeFilePath)
	}
	return nil
}

func dockerComposeStartError(output string, err error) error {
	if isDockerHubRateLimitError(output) {
		return u.Logger.NewError(DockerHubRateLimitReachedErrorMessage, "output", output)
	}
	if isDockerImageUnsupportedPlatformError(output) {
		return u.Logger.NewError(DockerImageUnsupportedPlatformErrorMessage, "output", output)
	}
	if isDockerImageDownloadError(output) {
		return u.Logger.NewError(DockerImageDownloadErrorMessage, "output", output)
	}
	if err != nil && strings.Contains(output, "is already in use by container") {
		return nil
	}
	return err
}

func isDockerHubRateLimitError(output string) bool {
	outputLower := strings.ToLower(output)
	return strings.Contains(outputLower, "toomanyrequests:") ||
		strings.Contains(outputLower, "you have reached your pull rate limit") ||
		strings.Contains(outputLower, "429 too many requests") ||
		strings.Contains(outputLower, "hap429") ||
		strings.Contains(outputLower, "increase-rate-limit")
}

func isDockerImageUnsupportedPlatformError(output string) bool {
	outputLower := strings.ToLower(output)
	return strings.Contains(outputLower, "no matching manifest for linux/") ||
		strings.Contains(outputLower, "no match for platform in manifest") ||
		(strings.Contains(outputLower, "manifest list entries") && strings.Contains(outputLower, "linux/"))
}

func isDockerImageDownloadError(output string) bool {
	outputLower := strings.ToLower(output)
	return strings.Contains(outputLower, "connection reset by peer") ||
		strings.Contains(outputLower, "connection refused") ||
		strings.Contains(outputLower, "i/o timeout") ||
		strings.Contains(outputLower, "tls handshake timeout") ||
		strings.Contains(outputLower, "context deadline exceeded")
}

func (d *DockerServiceImpl) StopAppContainer(maintainer, appName, composeFilePath string) {
	u.Logger.Info("Stopping app", MaintainerField, maintainer, AppField, appName, "compose_file_path", composeFilePath)
	_, err := d.CommandRunner.RunCommand(
		"docker", "compose",
		"-p", d.mergeNames(maintainer, appName),
		"-f", composeFilePath,
		"down",
	)
	if err != nil {
		u.Logger.Error(err)
	}
}

func (d *DockerServiceImpl) mergeNames(maintainer, app string) string {
	return maintainer + "_" + app
}

func (d *DockerServiceImpl) DoesDockerImageExist(name string) (bool, error) {
	cmd := exec.Command("docker", "images", "-q", name) // #nosec G204 (CWE-78): Subprocess launched with a potential tainted input or cmd arguments
	output, err := cmd.Output()
	if err != nil {
		return false, u.Logger.NewError(err.Error(), "image_name", name, "output", string(output))
	}
	return strings.TrimSpace(string(output)) != "", nil
}

func (d *DockerServiceImpl) BuildDockerImage(nameWithTag, dockerfilePath string, protectFromMaintenancePrune bool) error {
	args := []string{"build", "-t", nameWithTag}
	if protectFromMaintenancePrune {
		args = append(args, "--label", ResticImageMaintenanceKeepLabel)
	}
	args = append(args, "-f", dockerfilePath, ".")
	cmd := exec.Command("docker", args...) // #nosec G204 (CWE-78): Subprocess launched with a potential tainted input or cmd arguments

	output, err := cmd.CombinedOutput()
	if err != nil {
		return u.Logger.NewError(
			err.Error(),
			"image_name", nameWithTag,
			"dockerfile_path", dockerfilePath,
			"output", string(output),
		)
	}

	return nil
}

func ParseContainerName(name string) (*ContainerIdentity, bool) {
	name = strings.TrimPrefix(name, "/")
	if !containerNamePattern.MatchString(name) {
		return nil, false
	}
	parts := strings.Split(name, "_")
	return &ContainerIdentity{Maintainer: parts[0], App: parts[1], Service: parts[2]}, true
}
