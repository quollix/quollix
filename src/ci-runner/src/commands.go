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

var TestCmd = &cobra.Command{
	Use:   "test",
	Short: "test commands",
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

var TestComponentCmd = &cobra.Command{
	Use:   "component",
	Short: "run component tests against the dev container",
	Run: func(cmd *cobra.Command, args []string) {
		TestComponentWithDevProfile()
	},
}

var TestProdCmd = &cobra.Command{
	Use:   "prod",
	Short: "run component tests against the PROD container",
	Run: func(cmd *cobra.Command, args []string) {
		TestComponentWithProdProfile()
	},
}

var TestAcceptanceCmd = &cobra.Command{
	Use:   "acceptance",
	Short: "run acceptance tests",
	Run: func(cmd *cobra.Command, args []string) {
		keepSetup := getBoolFlag(cmd, AcceptanceKeepSetupFlagName)
		testFilter := getStringFlag(cmd, AcceptanceTestFilterFlagName)
		withGui := getBoolFlag(cmd, AcceptanceWithGuiFlagName)
		TestAcceptance(keepSetup, testFilter, withGui)
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
	TestAcceptanceCmd.Flags().BoolP(AcceptanceKeepSetupFlagName, AcceptanceKeepSetupFlagShort, false, "reuse acceptance setup and skip cleanup")
	TestAcceptanceCmd.Flags().StringP(AcceptanceTestFilterFlagName, AcceptanceTestFilterFlagShort, "", "run only acceptance tests matching regex")
	TestAcceptanceCmd.Flags().BoolP(AcceptanceWithGuiFlagName, "g", false, "run acceptance tests with visible browser window")
	TestCmd.AddCommand(
		TestAllCmd,
		TestComponentCmd,
		TestProdCmd,
		TestIntegrationCmd,
		TestUnitCmd,
		TestAcceptanceCmd,
		TestReleaseCmd,
	)
}

var DeployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "deploy commands",
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
		DeployLocalContainer(true, keepDb)
	},
}

var DeployProdCmd = &cobra.Command{
	Use:   "prod",
	Short: "deploy local quollix container with PROD profile",
	Run: func(cmd *cobra.Command, args []string) {
		Tr.Log.Info("Running local PROD container")
		keepDb := getBoolFlag(cmd, KeepDbFlagName)
		DeployLocalContainer(false, keepDb)
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
