package component

import (
	"context"
	"os/exec"
	"testing"
	"time"

	"github.com/quollix/common/assert"
	u "github.com/quollix/common/utils"
)

func WaitForGrafanaContainers(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := u.EventuallyWithTimeout(30*time.Second, 200*time.Millisecond, func() error {
		output, err := exec.CommandContext(ctx, "docker", "inspect", "--type", "container", "--format", "{{.Name}}",
			"quollix_grafana_postgres",
			"quollix_grafana_grafana",
			"quollix_grafana_victorialogs",
			"quollix_grafana_victoriametrics",
			"quollix_grafana_vector",
		).CombinedOutput()
		if err != nil {
			return u.Logger.NewError("failed to inspect Grafana containers", "error", err.Error(), "output", string(output))
		}
		return nil
	})
	assert.Nil(t, err)
}
