package setup

import (
	"net/http"
	"os"
	"os/exec"
	"server/configs"
	"server/tools"
	"sync"

	u "github.com/quollix/common/utils"
)

type TestStateResetHandler struct {
	DirectoryProvider          tools.DirectoryProvider
	DatabaseSnapshotRepository u.DatabaseSnapshotRepository
	ConfigsService             configs.ConfigsService
}

func (d *TestStateResetHandler) ResetTestStateHandler(w http.ResponseWriter, r *http.Request) {
	d.resetTestState()
}

func (d *TestStateResetHandler) resetTestState() {
	var waitGroup sync.WaitGroup
	waitGroup.Add(3)

	go func() {
		defer waitGroup.Done()
		err := d.DatabaseSnapshotRepository.ResetDatabaseToSnapshot()
		if err != nil {
			u.Logger.Error(err.Error())
		}
	}()

	go func() {
		defer waitGroup.Done()
		err := runDockerCmd(d.DirectoryProvider.GetDockerDir(), "exec", tools.TestSshServerHost, "rm", "-rf", tools.TestSshServerBackupsDirectory)
		if err != nil {
			u.Logger.Error(err.Error())
		}
	}()

	go func() {
		defer waitGroup.Done()
		err := runDockerCmd(".", "rm", "-f", tools.SampleAppContainerName)
		if err != nil {
			u.Logger.Error(err.Error())
		}
		err = runDockerCmd(".", "volume", "rm", "-f", tools.SampleAppDockerVolume)
		if err != nil {
			u.Logger.Error(err.Error())
		}
	}()

	waitGroup.Wait()
	d.ConfigsService.ResetBaseDomainCacheToLocalhost()
}

func runDockerCmd(dir string, args ...string) error {
	cmd := exec.Command("docker", args...) // #nosec G204: setup cleanup intentionally invokes the trusted docker CLI with structured args
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return u.Logger.NewError(err.Error(), "command", "docker", "args", args, "dir", dir)
	}
	return nil
}
