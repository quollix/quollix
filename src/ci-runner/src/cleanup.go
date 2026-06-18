package src

import "strings"

var projectCleanupContainers = []string{
	BrandAppContainerName,
	OfficialDatabaseContainerName,
	SampleAppContainerName,
	DummyBackupServerContainer,
}

var projectCleanupVolumes = []string{
	OfficialDatabaseVolumeName,
	SampleAppVolumeName,
}

var projectCleanupNetworks = []string{
	OfficialDatabaseNetworkName,
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
