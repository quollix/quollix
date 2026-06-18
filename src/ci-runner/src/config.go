package src

import (
	"os"
	"path/filepath"
	"strings"
)

var (
	localBrandAppUrl              = "https://quollix.localhost"
	BrandAppContainerName         = "quollix_quollix_quollix"
	OfficialDatabaseContainerName = "quollix_postgres_postgres"
	OfficialDatabaseNetworkName   = "quollix_postgres"
	OfficialDatabaseVolumeName    = "quollix_postgres_data"
	ResticCleanupLabel            = "quollix.cleanup=restic"
	SampleAppContainerName        = "samplemaintainer_sampleapp_sampleapp"
	SampleAppVolumeName           = "samplemaintainer_sampleapp_data"
	DummyBackupServerContainer    = "dummy_backup_server"

	ProjectDir  = GetProjectDir()
	srcDir      = ProjectDir + "/src"
	ciRunnerDir = srcDir + "/ci-runner"
	ServerDir   = srcDir + "/server"

	serverTestsDir = ServerDir + "/tests"

	assetsDir             = ServerDir + "/assets"
	dockerDir             = assetsDir + "/docker"
	DockerPostgresTestDir = dockerDir + "/postgres/test"
	sampleAppDir          = dockerDir + "/sampleapp"

	DevProfile = "DEV"
)

func getDockerCommand(useDevProfile bool, imageName string) string {
	profileEnv := ""
	publishedAppsMountVolumeFlag := ""
	mountFrontendFolderFromHost := ""
	mountCachedCertificatesFromHost := ""
	if useDevProfile {
		profileEnv = "-e PROFILE=" + DevProfile
		publishedAppsMountVolumeFlag = getPublishedAppsMountVolumeFlag()
		mountFrontendFolderFromHost = "-v ./frontend/resources:/opt/server/frontend/resources:ro"
		mountCachedCertificatesFromHost = "-v ./cache:/opt/server/cache"
	}

	baseCommand := []string{
		"docker run",
		"-d",
		"--rm",
		"--name " + BrandAppContainerName,
		"-p 127.0.0.1:80:80",
		"-p 127.0.0.1:443:443",
		"-e LOG_LEVEL=DEBUG",
		profileEnv,
		publishedAppsMountVolumeFlag,
		mountFrontendFolderFromHost,
		mountCachedCertificatesFromHost,
		"-v /tmp:/tmp",
		"-v /var/run/docker.sock:/var/run/docker.sock",
		imageName,
	}

	return strings.Join(baseCommand, " ")
}

func getPublishedAppsMountVolumeFlag() string {
	publishedAppsDir := filepath.Join(GetWorkspaceDir(), "store", "src", "client", "assets", "published-apps")
	if _, err := os.Stat(publishedAppsDir); err != nil {
		Tr.Log.Info("Warning: no published apps dir found in store project, starting DEV container without local store apps mount: %s", publishedAppsDir)
		return ""
	}
	return "-v " + publishedAppsDir + ":/opt/server/assets/docker/published-apps:ro"
}

func GetProjectDir() string {
	ciRunnerDirectory, _ := os.Getwd()
	src := filepath.Dir(ciRunnerDirectory)
	return filepath.Dir(src)
}

func GetWorkspaceDir() string {
	return filepath.Dir(ProjectDir)
}
