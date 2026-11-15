package src

import (
	u "github.com/quollix/common/utils"
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "ci-runner",
	Short: "local build, test, and deployment runner for Quollix",
	Long:  `Local build, test, and deployment runner for Quollix.`,
	Run: func(cmd *cobra.Command, args []string) {
		u.ShowHelpCommand(cmd)
	},
}

var ReleaseCmd = &cobra.Command{
	Use:   "release <tag>",
	Short: "build and publish the production Docker image",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ReleaseDockerImage(args[0])
	},
}

var BuildCmd = &cobra.Command{
	Use:   "build",
	Short: "build commands",
	Run: func(cmd *cobra.Command, args []string) {
		u.ShowHelpCommand(cmd)
	},
}

var BuildLocalCmd = &cobra.Command{
	Use:   "local-image",
	Short: "build the local Docker image",
	Run: func(cmd *cobra.Command, args []string) {
		BuildLocalDockerImage()
	},
}

var TestCmd = &cobra.Command{
	Use:   "test",
	Short: "test commands",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		BuildLocalSampleAppDockerImageIfNotPresent()
	},
	Run: func(cmd *cobra.Command, args []string) {
		u.ShowHelpCommand(cmd)
	},
}

var TestAllCmd = &cobra.Command{
	Use:   "all",
	Short: "run all tests",
	Run: func(cmd *cobra.Command, args []string) {
		TestAll()
	},
}

var TestUnitCmd = &cobra.Command{
	Use:   "unit",
	Short: "run unit tests",
	Run: func(cmd *cobra.Command, args []string) {
		TestUnit()
	},
}

var TestIntegrationCmd = &cobra.Command{
	Use:   "integration",
	Short: "run integration tests",
	Run: func(cmd *cobra.Command, args []string) {
		TestIntegration()
	},
}

var TestVictoriaCmd = &cobra.Command{
	Use:              "victoria",
	Short:            "run VictoriaMetrics and VictoriaLogs integration tests",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {},
	Run: func(cmd *cobra.Command, args []string) {
		TestVictoria(getBoolFlag(cmd, VictoriaKeepStackFlagName))
	},
}

var TestComponentCmd = &cobra.Command{
	Use:   "component",
	Short: "run component tests against the dev container",
	Run: func(cmd *cobra.Command, args []string) {
		TestComponentWithDevProfile()
	},
}

var TestBehindProxyCmd = &cobra.Command{
	Use:   "behind-proxy",
	Short: "run component tests against a behind-proxy HTTP mode container",
	Run: func(cmd *cobra.Command, args []string) {
		TestBehindProxyHttpMode()
	},
}

var TestProdCmd = &cobra.Command{
	Use:   "prod",
	Short: "run component tests against the PROD container",
	Run: func(cmd *cobra.Command, args []string) {
		TestComponentWithProdProfile()
	},
}

var TestQuollixRestartCmd = &cobra.Command{
	Use:   "quollix-restart",
	Short: "verify running apps remain accessible after Quollix restarts",
	Run: func(cmd *cobra.Command, args []string) {
		TestQuollixRestart()
	},
}

var TestInitialAdminPasswordCmd = &cobra.Command{
	Use:   "initial-admin-password",
	Short: "verify generated initial admin password login",
	Run: func(cmd *cobra.Command, args []string) {
		TestInitialAdminPassword()
	},
}

var TestSpecialPasswordsCmd = &cobra.Command{
	Use:   "special-passwords",
	Short: "run component tests with loose-policy backup passwords",
	Run: func(cmd *cobra.Command, args []string) {
		TestSpecialPasswords()
	},
}

var TestFrontendCmd = &cobra.Command{
	Use:   "frontend",
	Short: "run frontend tests",
	Run: func(cmd *cobra.Command, args []string) {
		keepSetup := getBoolFlag(cmd, FrontendKeepSetupFlagName)
		testFilter := getStringFlag(cmd, FrontendTestFilterFlagName)
		withGui := getBoolFlag(cmd, FrontendWithGuiFlagName)
		TestFrontend(keepSetup, testFilter, withGui)
	},
}

var TestOidcCmd = &cobra.Command{
	Use:   "oidc",
	Short: "run OIDC two-instance tests",
	Run: func(cmd *cobra.Command, args []string) {
		keepDatabases := getBoolFlag(cmd, OidcKeepDatabasesFlagName)
		keepContainers := getBoolFlag(cmd, OidcKeepContainersFlagName)
		TestOidc(keepDatabases, keepContainers)
	},
}

var TestReleaseCmd = &cobra.Command{
	Use:   "release",
	Short: "run release tests",
	Run: func(cmd *cobra.Command, args []string) {
		TestRelease()
	},
}

func ConfigureTestCmd() {
	TestVictoriaCmd.Flags().BoolP(VictoriaKeepStackFlagName, VictoriaKeepStackFlagShort, false, "keep the Victoria test stack running after the tests")
	TestFrontendCmd.Flags().BoolP(FrontendKeepSetupFlagName, FrontendKeepSetupFlagShort, false, "reuse frontend setup and skip cleanup")
	TestFrontendCmd.Flags().StringP(FrontendTestFilterFlagName, FrontendTestFilterFlagShort, "", "run only frontend tests matching regex")
	TestFrontendCmd.Flags().BoolP(FrontendWithGuiFlagName, "g", false, "run frontend tests with visible browser window")
	TestOidcCmd.Flags().BoolP(OidcKeepDatabasesFlagName, OidcKeepDatabasesFlagShort, false, "keep OIDC database volumes after the test")
	TestOidcCmd.Flags().BoolP(OidcKeepContainersFlagName, OidcKeepContainersFlagShort, false, "keep OIDC containers and database volumes after the test")
	TestCmd.AddCommand(
		TestAllCmd,
		TestComponentCmd,
		TestBehindProxyCmd,
		TestProdCmd,
		TestQuollixRestartCmd,
		TestInitialAdminPasswordCmd,
		TestSpecialPasswordsCmd,
		TestIntegrationCmd,
		TestVictoriaCmd,
		TestUnitCmd,
		TestFrontendCmd,
		TestOidcCmd,
		TestReleaseCmd,
	)
}

var DeployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "deploy commands",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		BuildLocalSampleAppDockerImageIfNotPresent()
	},
	Run: func(cmd *cobra.Command, args []string) {
		u.ShowHelpCommand(cmd)
	},
}

var DeployDbCmd = &cobra.Command{
	Use:   "db",
	Short: "deploy local postgres",
	Run: func(cmd *cobra.Command, args []string) {
		Tr.Log.TaskDescription("Running local postgres")
		DeployLocalDatabase()
	},
}

var DeployDevCmd = &cobra.Command{
	Use:   "dev",
	Short: "deploy local quollix container with DEV profile",
	Run: func(cmd *cobra.Command, args []string) {
		Tr.Log.Info("Running local DEV container")
		keepDb := getBoolFlag(cmd, KeepDbFlagName)
		disableInitialAdminEnv := getBoolFlag(cmd, DisableInitialAdminEnvFlagName)
		shouldConfigureDevTestDomain := getBoolFlag(cmd, TestDomainFlagName)
		DeployLocalContainer(keepDb, containerEnv(true, disableInitialAdminEnv))
		if shouldConfigureDevTestDomain {
			if err := ConfigureDevTestDomain(); err != nil {
				Tr.Log.Error("Failed to configure test domain: %v", err)
				Tr.ExitWithError()
				return
			}
			Tr.Log.Info("Base domain: test.quollix.org; open https://quollix.test.quollix.org")
		}
	},
}

var DeployProdCmd = &cobra.Command{
	Use:   "prod",
	Short: "deploy local quollix container with PROD profile",
	Run: func(cmd *cobra.Command, args []string) {
		Tr.Log.Info("Running local PROD container")
		keepDb := getBoolFlag(cmd, KeepDbFlagName)
		disableInitialAdminEnv := getBoolFlag(cmd, DisableInitialAdminEnvFlagName)
		DeployLocalContainer(keepDb, containerEnv(false, disableInitialAdminEnv))
	},
}

var DeployOidcCmd = &cobra.Command{
	Use:   "oidc",
	Short: "deploy local OIDC two-instance environment",
	Run: func(cmd *cobra.Command, args []string) {
		keepDatabases := getBoolFlag(cmd, OidcKeepDatabasesFlagName)
		keepContainers := getBoolFlag(cmd, OidcKeepContainersFlagName)
		setup := getBoolFlag(cmd, OidcSetupFlagName)
		if setup {
			DeployAndSetupOidcTestEnvironment(keepDatabases, keepContainers)
			return
		}
		if keepContainers {
			DeployOidcTestEnvironmentIfNeeded()
			return
		}
		DeployOidcTestEnvironment(keepDatabases)
	},
}

func getBoolFlag(cmd *cobra.Command, name string) bool {
	value, err := cmd.Flags().GetBool(name)
	exitOnFlagError(err)
	return value
}

func getStringFlag(cmd *cobra.Command, name string) string {
	value, err := cmd.Flags().GetString(name)
	exitOnFlagError(err)
	return value
}

func exitOnFlagError(err error) {
	if err == nil {
		return
	}
	Tr.Log.Error("%v", err)
	Tr.ExitWithError()
}
