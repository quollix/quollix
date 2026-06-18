package src

import (
	"strings"
	"time"

	u "github.com/quollix/common/utils"
)

const (
	ComponentBuildTag   = "component"
	ProdProfileBuildTag = "prod_profile"
	IntegrationBuildTag = "integration"
	AcceptanceBuildTag  = "acceptance"
	ReleaseBuildTag     = "release"

	KeepDbFlagName                = "keep-db"
	AcceptanceKeepSetupFlagName   = "keep-setup"
	AcceptanceKeepSetupFlagShort  = "k"
	AcceptanceTestFilterFlagName  = "test-filter"
	AcceptanceTestFilterFlagShort = "f"
	AcceptanceWithGuiFlagName     = "with-gui"
)

func TestUnit() {
	Tr.Log.TaskDescription("Running unit tests")
	TestCIRunnerUnit()
	newTestCommand(ServerDir).Run(Tr)
}

func TestCIRunnerUnit() {
	Tr.Log.TaskDescription("Running ci-runner unit tests")
	Tr.Cmd().Dir(ciRunnerDir).Run("go test ./...")
}

func TestIntegration() {
	Tr.Log.TaskDescription("Running integration tests")
	Tr.Cleanup()
	defer Tr.Cleanup()
	DeployLocalDatabase()

	runServerTests(IntegrationBuildTag)
}

func TestRelease() {
	Tr.Log.TaskDescription("Running release tests")
	defer Tr.Cleanup()
	DeployLocalContainer(false, false)
	testCommand := u.GoTest(serverTestsDir)
	testCommand.Tag(ReleaseBuildTag)
	testCommand.Run(Tr)
}

func DeployLocalDatabase() {
	Tr.Cmd().AllowFail().Run("docker rm -f %s", OfficialDatabaseContainerName)
	Tr.Cmd().AllowFail().Run("docker network create %s", OfficialDatabaseNetworkName)
	Tr.Cmd().Dir(DockerPostgresTestDir).Run("docker compose up -d postgres")
}

func DeployLocalContainer(useDevProfile bool, keepDb bool) {
	Tr.Cmd().AllowFail().Run("docker rm -f %s", BrandAppContainerName)
	if !keepDb {
		Tr.Cmd().AllowFail().Run("docker rm -f %s", OfficialDatabaseContainerName)
		Tr.Cmd().AllowFail().Run("docker volume rm -f %s", OfficialDatabaseVolumeName)
	}
	StartSshTestContainer()
	BuildLocalDockerImage()
	Tr.Cmd().Dir(ServerDir).Run("%s", getDockerCommand(useDevProfile, dockerImage+"local"))
	time.Sleep(500 * time.Millisecond)
	Tr.Cmd().AsDaemon("docker-logs").Run("docker logs -f %s", BrandAppContainerName)
	Tr.WaitForWebPageToBeReady(localBrandAppUrl)
}

func TestAll() {
	Tr.Log.Info("Running all CI tests")
	TestUnit()
	TestIntegration()
	TestComponentWithDevProfile()
	TestComponentWithProdProfile()
	TestAcceptance(false, "", false)
}

func TestAcceptance(keepSetup bool, testFilter string, headful bool) {
	Tr.Log.TaskDescription("Running acceptance tests")
	BuildLocalSampleAppDockerImageIfNotPresent()
	if keepSetup {
		Tr.Config.CleanupFunc = nil
		Tr.Log.Info("Keep-setup enabled, reusing acceptance environment when possible")
		if !isContainerRunning(BrandAppContainerName) {
			DeployLocalContainer(true, false)
		}
	} else {
		defer Tr.Cleanup()
		defer StopSshTestContainer()
		DeployLocalContainer(true, false)
	}

	runAcceptanceTests(testFilter, headful)
}

func isContainerRunning(containerName string) bool {
	containers := GetRunningContainers()
	for _, container := range containers {
		if container == containerName {
			return true
		}
	}
	return false
}

func TestComponentWithProdProfile() {
	Tr.Log.TaskDescription("Running component tests against the PROD container")
	defer Tr.Cleanup()
	DeployLocalContainer(false, false)
	runServerTests(ProdProfileBuildTag)
}

func TestComponentWithDevProfile() {
	Tr.Log.TaskDescription("Running component tests against the dev container")
	defer Tr.Cleanup()
	DeployLocalContainer(true, false)
	runServerTests(ComponentBuildTag)
}

func runAcceptanceTests(testFilter string, headful bool) {
	testCommand := newTestCommand(serverTestsDir, AcceptanceBuildTag)
	runAcceptanceTestCommand(testCommand, testFilter, headful)
}

func runAcceptanceTestCommand(testCommand *u.GoTestCommand, testFilter string, headful bool) {
	if headful {
		testCommand.Env("HEADFUL", "true")
	}

	if testFilter != "" {
		Tr.Log.Info("Running acceptance tests matching '%s'", testFilter)
		testCommand.Filter(testFilter)
	}

	testCommand.Run(Tr)
}

func runServerTests(extraTags ...string) {
	newTestCommand(serverTestsDir, extraTags...).Run(Tr)
}

func newTestCommand(dir string, extraTags ...string) *u.GoTestCommand {
	testCommand := u.GoTest(dir)
	if len(extraTags) > 0 {
		testCommand.Tag(strings.Join(extraTags, ","))
	}
	return testCommand
}

func StartSshTestContainer() {
	Tr.Cmd().AllowFail().Run("docker network create quollix_postgres")
	Tr.Cmd().Dir(dockerDir).Run("docker compose -f docker-compose.dummy-ssh.yml up -d")
}

func StopSshTestContainer() {
	Tr.Cmd().Dir(dockerDir).AllowFail().Run("docker compose -f docker-compose.dummy-ssh.yml down")
}
