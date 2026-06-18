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
	src.BuildLocalSampleAppDockerImageIfNotPresent()
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
	src.DeployCmd.PersistentFlags().BoolP(src.KeepDbFlagName, "k", false, "keep the existing database container and volume")
	src.DeployCmd.AddCommand(
		src.DeployDevCmd,
		src.DeployProdCmd,
		src.DeployDbCmd,
	)
	src.ConfigureTestCmd()
	src.RootCmd.AddCommand(
		src.TestCmd,
		src.ReleaseCmd,
		src.DeployCmd,
		ci.NewCommonCmd(Tr, src.ProjectDir+"/src", "quollix"),
	)
}
