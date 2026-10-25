package setup

import (
	"net/http"
	"os"
	"os/exec"
	"sync"

	"server/app_store"
	"server/apps_basic"
	"server/configs"
	"server/tools"

	u "github.com/quollix/common/utils"
)

var (
	testContainersToRemove = []string{
		tools.SampleAppContainerName,
		tools.SamplePostgresAppContainerName,
		tools.SampleRabbitMQAppContainerName,
		"quollix_grafana_grafana",
		"quollix_grafana_victorialogs",
		"quollix_grafana_victoriametrics",
	}
	testVolumesToRemove = []string{
		tools.SampleAppDockerVolume,
		tools.SamplePostgresAppDockerVolume,
		tools.SampleRabbitMQAppDockerVolume,
		"quollix_grafana_victorialogs",
		"quollix_grafana_victoriametrics",
	}
)

type TestStateResetHandler struct {
	ContainerMetricsMonitor    apps_basic.ContainerMetricsMonitor
	VolumeMetricsMonitor       apps_basic.VolumeMetricsMonitor
	ContainerLogMonitor        apps_basic.ContainerLogMonitor
	DirectoryProvider          tools.DirectoryProvider
	DatabaseSnapshotRepository u.DatabaseSnapshotRepository
	ConfigsService             configs.ConfigsService
	MaintainerRepository       *app_store.MaintainerRepositoryImpl
}

func (d *TestStateResetHandler) ResetTestStateHandler(w http.ResponseWriter, r *http.Request) {
	d.resetTestState()
}

func (d *TestStateResetHandler) resetTestState() {
	d.ContainerMetricsMonitor.StopMonitoring()
	d.VolumeMetricsMonitor.StopMonitoring()
	d.ContainerLogMonitor.StopMonitoring()

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
		err := runDockerCmd(".", append([]string{"rm", "-f"}, testContainersToRemove...)...)
		if err != nil {
			u.Logger.Error(err.Error())
		}
		err = runDockerCmd(".", append([]string{"volume", "rm", "-f"}, testVolumesToRemove...)...)
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
