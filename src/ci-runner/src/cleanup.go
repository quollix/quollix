package src

import "strings"

var projectCleanupContainers = []string{
	BrandAppContainerName,
	OfficialDatabaseContainerName,
	SampleAppContainerName,
	DummyBackupServerContainer,
	"quollix_provider_db",
	"quollix_client_db",
	"quollix_provider_server",
	"quollix_client_server",
	OidcProxyContainerName,
}

var projectCleanupVolumes = []string{
	OfficialDatabaseVolumeName,
	SampleAppVolumeName,
	"quollix_provider_data",
	"quollix_client_data",
	"quollix_proxy_data",
	"quollix_proxy_config",
}

var projectCleanupNetworks = []string{
	OfficialDatabaseNetworkName,
	"quollix_oidc",
}

func GetRunningContainers() []string {
	output := Tr.Cmd().AllowFail().Run("docker ps -a --format '{{.Names}}'").Output()
	return strings.Fields(output)
}

func getContainersWithLabel(label string) []string {
	output := Tr.Cmd().AllowFail().Run("docker ps -aq --filter label=%s", label).Output()
	return strings.Fields(output)
}

func CustomCleanup() {
	for _, container := range projectCleanupContainers {
		Tr.Cmd().AllowFail().Run("docker rm -f %s", container)
	}
	for _, container := range getContainersWithLabel(ResticCleanupLabel) {
		Tr.Cmd().AllowFail().Run("docker rm -f %s", container)
	}

	for _, volume := range projectCleanupVolumes {
		Tr.Cmd().AllowFail().Run("docker volume rm -f %s", volume)
	}

	for _, network := range projectCleanupNetworks {
		Tr.Cmd().AllowFail().Run("docker network rm %s", network)
	}
}
