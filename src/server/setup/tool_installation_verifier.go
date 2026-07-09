package setup

import (
	"os/exec"
	"server/tools"
	"strings"

	u "github.com/quollix/common/utils"
)

type CliToolInstallationVerifier struct{}

func (c *CliToolInstallationVerifier) Verify() error {
	cliTools := []string{
		"docker version",
		"docker compose version",
	}

	for _, fullCmd := range cliTools {
		parts := strings.Split(fullCmd, " ")
		toolName := parts[0]
		cmdArgs := parts[1:]

		err := crashIfToolIsNotInstalled(toolName, cmdArgs)
		if err != nil {
			return err
		}
	}
	u.Logger.Info("all required CLI tools are installed")
	return nil
}

func crashIfToolIsNotInstalled(toolName string, args []string) error {
	cmd := exec.Command(toolName, args...) // #nosec G204: tool verification intentionally executes trusted CLI binaries with structured args
	output, err := cmd.CombinedOutput()
	if err != nil {
		argsAsSingleString := strings.Join(args, " ")
		return u.Logger.NewError(
			"tried command but CLI tool seems not to be installed properly",
			tools.CommandField, toolName,
			tools.CommandArgsField, argsAsSingleString,
			"executory_error", err.Error(),
			"output", string(output),
		)
	}
	return nil
}
