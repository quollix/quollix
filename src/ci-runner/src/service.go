package src

import (
	"strings"
	"time"

	u "github.com/quollix/common/utils"
)

const (
	ComponentBuildTag   = "component"
	ProdProfileBuildTag = "prod_profile"
	BehindProxyBuildTag = "behind_proxy"
	IntegrationBuildTag = "integration"
	FrontendBuildTag    = "frontend"
	OidcBuildTag        = "oidc"
	ReleaseBuildTag     = "release"
	SpecialPasswordsTag = "special_passwords"

	specialPolicyPassword = "Aa09!@#$%^&*()_-+=.,:;/?[]{}|~<>"

	KeepDbFlagName                  = "keep-db"
	DisableInitialAdminEnvFlagName  = "disable-initial-admin-env"
	DisableInitialAdminEnvFlagShort = "a"
	FrontendKeepSetupFlagName       = "keep-setup"
	FrontendKeepSetupFlagShort      = "k"
	FrontendTestFilterFlagName      = "test-filter"
	FrontendTestFilterFlagShort     = "f"
	FrontendWithGuiFlagName         = "with-gui"
	OidcKeepDatabasesFlagName       = "keep-databases"
	OidcKeepDatabasesFlagShort      = "d"
	OidcKeepContainersFlagName      = "keep-containers"
	OidcKeepContainersFlagShort     = "c"
	OidcSetupFlagName               = "setup"
	OidcSetupFlagShort              = "s"
)

func TestUnit() {
	Tr.Log.TaskDescription("Running unit tests")
	BuildAllGoModules()
	TestCIRunnerUnit()
	newTestCommand(ServerDir).Run(Tr)
}

func BuildAllGoModules() {
	Tr.Log.TaskDescription("Building all Go modules")
	u.BuildWholeGoProject(Tr, ServerDir)
	u.BuildWholeGoProject(Tr, ciRunnerDir)
	u.BuildWholeGoProject(Tr, sampleAppDir)
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
	DeployLocalContainer(false, containerEnv(false, false))
	testCommand := u.GoTest(serverTestsDir)
	testCommand.Tag(ReleaseBuildTag)
	testCommand.Run(Tr)
}

func DeployLocalDatabase() {
	Tr.Cmd().AllowFail().Run("docker rm -f %s", OfficialDatabaseContainerName)
	Tr.Cmd().AllowFail().Run("docker network create %s", OfficialDatabaseNetworkName)
	Tr.Cmd().Dir(DockerPostgresTestDir).Run("docker compose up -d postgres")
}

func DeployLocalContainer(keepDb bool, envVars []string) {
	deployLocalContainer(keepDb, envVars, "")
}

func deployLocalContainer(keepDb bool, envVars []string, sshPassword string) {
	Tr.Cmd().AllowFail().Run("docker rm -f %s", BrandAppContainerName)
	if !keepDb {
		Tr.Cmd().AllowFail().Run("docker rm -f %s", OfficialDatabaseContainerName)
		Tr.Cmd().AllowFail().Run("docker volume rm -f %s", OfficialDatabaseVolumeName)
	}
	StartSshTestContainer(sshPassword)
	BuildLocalDockerImage()
	Tr.Cmd().Dir(ServerDir).Run("%s", getDockerCommand(envVars))
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
	TestInitialAdminPassword()
	TestBehindProxyHttpMode()
	TestOidc(false, false)
	TestSpecialPasswords()
	TestFrontend(false, "", false)
}

func TestFrontend(keepSetup bool, testFilter string, headful bool) {
	Tr.Log.TaskDescription("Running frontend tests")
	BuildLocalSampleAppDockerImageIfNotPresent()
	if keepSetup {
		Tr.Config.CleanupFunc = nil
		Tr.Log.Info("Keep-setup enabled, reusing frontend environment when possible")
		if !isContainerRunning(BrandAppContainerName) {
			DeployLocalContainer(false, containerEnv(true, false))
		}
	} else {
		defer Tr.Cleanup()
		defer StopSshTestContainer()
		DeployLocalContainer(false, containerEnv(true, false))
	}

	runFrontendTests(testFilter, headful)
}

func TestOidc(keepDatabases bool, keepContainers bool) {
	Tr.Log.TaskDescription("Running OIDC two-instance tests")
	if keepContainers {
		Tr.Config.CleanupFunc = nil
		Tr.Log.Info("Keep-containers enabled, leaving OIDC test environment running")
		DeployOidcTestEnvironmentIfNeeded()
	} else {
		defer StopOidcTestEnvironment(keepDatabases)
		deployOidcTestEnvironment(keepDatabases)
	}

	Tr.Log.Info("OIDC test environment is ready")
	runServerTests(OidcBuildTag)
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
	BuildLocalSampleAppDockerImageIfNotPresent()
	DeployLocalContainer(false, containerEnv(false, false))
	runServerTests(ProdProfileBuildTag)
}

func TestComponentWithDevProfile() {
	Tr.Log.TaskDescription("Running component tests against the dev container")
	defer Tr.Cleanup()
	BuildLocalSampleAppDockerImageIfNotPresent()
	DeployLocalContainer(false, containerEnv(true, false))
	runServerTests(ComponentBuildTag)
}

func TestBehindProxyHttpMode() {
	Tr.Log.TaskDescription("Running behind-proxy HTTP mode component tests")
	defer Tr.Cleanup()
	BuildLocalSampleAppDockerImageIfNotPresent()
	DeployLocalContainer(false, containerEnv(true, false,
		"REDIRECT_HTTP_TO_HTTPS=false",
		"APP_FORWARDED_PROTO=http",
	))
	runServerTests(BehindProxyBuildTag)
}

func TestSpecialPasswords() {
	Tr.Log.TaskDescription("Running special-password component tests")
	defer Tr.Cleanup()
	BuildLocalSampleAppDockerImageIfNotPresent()
	deployLocalContainer(false, containerEnv(true, false), specialPolicyPassword)
	runServerTests(SpecialPasswordsTag)
}

func containerEnv(useDevProfile bool, disableInitialAdminEnv bool, extraEnv ...string) []string {
	envVars := []string{}

	if useDevProfile {
		envVars = append(envVars, "PROFILE="+DevProfile)
	}

	if !disableInitialAdminEnv {
		envVars = append(envVars, "INITIAL_ADMIN_NAME=admin", "INITIAL_ADMIN_PASSWORD=password")
	}

	return append(envVars, extraEnv...)
}

func runFrontendTests(testFilter string, headful bool) {
	testCommand := newTestCommand(serverTestsDir, FrontendBuildTag)
	runFrontendTestCommand(testCommand, testFilter, headful)
}

func runFrontendTestCommand(testCommand *u.GoTestCommand, testFilter string, headful bool) {
	if headful {
		testCommand.Env("HEADFUL", "true")
	}

	if testFilter != "" {
		Tr.Log.Info("Running frontend tests matching '%s'", testFilter)
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

func StartSshTestContainer(sshPassword string) {
	Tr.Cmd().AllowFail().Run("docker network create quollix_postgres")
	cmd := Tr.Cmd().Dir(dockerDir)
	if sshPassword != "" {
		cmd.Env("DUMMY_SSH_PASSWORD", sshPassword)
	}
	cmd.Run("docker compose -f docker-compose.dummy-ssh.yml up -d")
}

func StopSshTestContainer() {
	Tr.Cmd().Dir(dockerDir).AllowFail().Run("docker compose -f docker-compose.dummy-ssh.yml down")
}

func DeployOidcTestEnvironment(keepDatabases bool) {
	deployOidcTestEnvironment(keepDatabases)
}

func DeployAndSetupOidcTestEnvironment(keepDatabases bool, keepContainers bool) {
	if keepContainers {
		DeployOidcTestEnvironmentIfNeeded()
	} else {
		deployOidcTestEnvironment(keepDatabases)
	}
	ConfigureOidcTestEnvironment()
}

func deployOidcTestEnvironment(keepDatabases bool) {
	Tr.Log.TaskDescription("Running OIDC two-instance environment")
	BuildLocalDockerImage()
	StopOidcTestEnvironment(keepDatabases)
	Tr.Cmd().Dir(DockerOidcTestDir).Run("docker compose up -d")
	waitForOidcTestEnvironment()
}

func DeployOidcTestEnvironmentIfNeeded() {
	if !isContainerRunning(OidcProxyContainerName) {
		deployOidcTestEnvironment(true)
		return
	}

	Tr.Log.Info("OIDC test environment is already running")
	waitForOidcTestEnvironment()
}

func StopOidcTestEnvironment(keepDatabases bool) {
	if keepDatabases {
		Tr.Cmd().Dir(DockerOidcTestDir).AllowFail().Run("docker compose down --remove-orphans")
		return
	}
	Tr.Cmd().Dir(DockerOidcTestDir).AllowFail().Run("docker compose down --volumes --remove-orphans")
}

func waitForOidcTestEnvironment() {
	Tr.WaitForWebPageToBeReady(OidcProviderUrl)
	Tr.WaitForWebPageToBeReady(OidcClientUrl)
}
