package src

import (
	"os"
	"path/filepath"
	"slices"
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
	OidcProxyContainerName        = "quollix_oidc_proxy"
	OidcProviderUrl               = "https://quollix.oidc-provider.localhost"
	OidcClientUrl                 = "https://quollix.oidc-client.localhost"

	ProjectDir  = GetProjectDir()
	srcDir      = ProjectDir + "/src"
	ciRunnerDir = srcDir + "/ci-runner"
	ServerDir   = srcDir + "/server"

	serverTestsDir = ServerDir + "/tests"

	assetsDir             = ServerDir + "/assets"
	dockerDir             = assetsDir + "/docker"
	DockerPostgresTestDir = dockerDir + "/postgres/test"
	DockerOidcTestDir     = dockerDir + "/oidc"
	sampleAppDir          = dockerDir + "/sampleapp"

	DevProfile = "DEV"
)

func getDockerCommand(envVars []string) string {
	envVars = append([]string{"LOG_LEVEL=DEBUG"}, envVars...)
	publishedAppsMountVolumeFlag := ""
	mountFrontendFolderFromHost := ""
	mountCachedCertificatesFromHost := ""
	if hasEnvVar(envVars, "PROFILE", DevProfile) {
		publishedAppsMountVolumeFlag = getPublishedAppsMountVolumeFlag()
		mountFrontendFolderFromHost = "-v ./frontend/resources:/opt/server/frontend/resources:ro"
		mountCachedCertificatesFromHost = "-v ./cache:/opt/server/cache"
	}

	commandPrefix := []string{
		"docker run",
		"-d",
		"--rm",
		"--name " + BrandAppContainerName,
		"-p 127.0.0.1:80:80",
		"-p 127.0.0.1:443:443",
	}

	commandSuffix := []string{
		publishedAppsMountVolumeFlag,
		mountFrontendFolderFromHost,
		mountCachedCertificatesFromHost,
		"-v /tmp:/tmp",
		"-v /var/run/docker.sock:/var/run/docker.sock",
		dockerImage + "local",
	}

	command := append(commandPrefix, getDockerEnvFlags(envVars)...)
	command = append(command, commandSuffix...)
	return strings.Join(command, " ")
}

func hasEnvVar(envVars []string, name string, value string) bool {
	expectedEnvVar := name + "=" + value
	return slices.Contains(envVars, expectedEnvVar)
}

func getDockerEnvFlags(envVars []string) []string {
	envFlags := []string{}
	for _, envVar := range envVars {
		envFlags = append(envFlags, "-e "+envVar)
	}
	return envFlags
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
