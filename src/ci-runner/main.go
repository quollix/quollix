package main

import (
	"ci-runner/src"

	"github.com/quollix/common/ci"
	"github.com/quollix/taskrunner"
	"github.com/spf13/cobra"
)

var Tr = taskrunner.GetTaskRunner()

func main() {
	Tr.EnableAbortForKeystrokeControlPlusC()
	Tr.Config.DefaultEnvironmentVariables = []string{"LOG_LEVEL=DEBUG"}
	Tr.Config.CleanupFunc = src.CustomCleanup
	src.Tr = Tr

	src.RootCmd.SetHelpCommand(&cobra.Command{Hidden: true})
	src.RootCmd.Root().CompletionOptions.DisableDefaultCmd = true
	buildSubCommands()

	if err := src.RootCmd.Execute(); err != nil {
		Tr.Log.Error("Error during execution: %s", err.Error())
		Tr.ExitWithError()
	}
}

func buildSubCommands() {
	src.DeployDevCmd.Flags().BoolP(src.KeepDbFlagName, "k", false, "keep the existing database container and volume")
	src.DeployProdCmd.Flags().BoolP(src.KeepDbFlagName, "k", false, "keep the existing database container and volume")
	src.DeployDevCmd.Flags().BoolP(src.DisableInitialAdminEnvFlagName, src.DisableInitialAdminEnvFlagShort, false, "do not inject INITIAL_ADMIN_NAME and INITIAL_ADMIN_PASSWORD")
	src.DeployProdCmd.Flags().BoolP(src.DisableInitialAdminEnvFlagName, src.DisableInitialAdminEnvFlagShort, false, "do not inject INITIAL_ADMIN_NAME and INITIAL_ADMIN_PASSWORD")
	src.DeployCmd.AddCommand(
		src.DeployDevCmd,
		src.DeployProdCmd,
		src.DeployDbCmd,
		src.DeployOidcCmd,
	)
	src.DeployOidcCmd.Flags().BoolP(src.OidcKeepDatabasesFlagName, src.OidcKeepDatabasesFlagShort, false, "keep OIDC database volumes")
	src.DeployOidcCmd.Flags().BoolP(src.OidcKeepContainersFlagName, src.OidcKeepContainersFlagShort, false, "keep OIDC containers and database volumes")
	src.DeployOidcCmd.Flags().BoolP(src.OidcSetupFlagName, src.OidcSetupFlagShort, false, "reset and configure OIDC provider and client after deployment")
	src.ConfigureTestCmd()
	src.RootCmd.AddCommand(
		src.TestCmd,
		src.ReleaseCmd,
		src.DeployCmd,
		ci.NewCommonCmd(Tr, src.ProjectDir+"/src", "quollix"),
	)
}
