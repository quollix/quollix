package component

import (
	"context"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/quollix/common/assert"
	"github.com/quollix/common/quollix/api_client"
	u "github.com/quollix/common/utils"
)

const grafanaTestCompose = `# This file is licensed under the 0BSD License: https://opensource.org/license/0bsd

services:
  grafana:
    # Caddy is only a dummy main service for app conventions. No web server is needed.
    image: caddy:2.10.2-alpine@sha256:4c6e91c6ed0e2fa03efd5b44747b625fec79bc9cd06ac5235a779726618e530d
    container_name: quollix_grafana_grafana
    command: [sleep, infinity]
    labels:
      quollix.port: 8080

  victorialogs:
    image: victoriametrics/victoria-logs:v1.53.0@sha256:251121fa882af99b95ba0c230a4a2f412ea602d2698c64a96c58dc9842bb755d
    container_name: quollix_grafana_victorialogs
    command:
      - -httpAuth.username=admin
      - -httpAuth.password=${SECRET_VICTORIALOGS_PASSWORD}
      - -loggerFormat=json
      - -delete.enable
      - -storageDataPath=/victoria-logs-data
      - -retentionPeriod=${LOGS_RETENTION_DAYS}d
      - -retention.maxDiskSpaceUsageBytes=${LOGS_MAX_DISK_SPACE_GB}GB
    volumes:
      - quollix_grafana_victorialogs:/victoria-logs-data

  victoriametrics:
    image: victoriametrics/victoria-metrics:v1.154.1@sha256:c7884f32920dd45a9d86542a717227e63af03e060e53bbf8031e6e6fd0e1e05d
    container_name: quollix_grafana_victoriametrics
    command:
      - -httpAuth.username=admin
      - -httpAuth.password=${SECRET_VICTORIAMETRICS_PASSWORD}
      - -loggerFormat=json
      - -storageDataPath=/victoria-metrics-data
      - -retentionPeriod=${METRICS_RETENTION_DAYS}d
    volumes:
      - quollix_grafana_victoriametrics:/victoria-metrics-data

  vector:
    image: timberio/vector:0.60.0-debian@sha256:3ccb4bbd36e38734394e158db1411e91a69d076ed25013f8de385046a382d36a
    container_name: quollix_grafana_vector
    hostname: quollix-host
    depends_on:
      - victorialogs
      - victoriametrics
    environment:
      PROCFS_ROOT: /host/proc
      SYSFS_ROOT: /host/sys
    command: [--config, /etc/vector/vector.yaml, --log-format, json]
    volumes:
      - /:/host:ro,rslave
      - /var/log/journal:/var/log/journal:ro
      - /run/log/journal:/run/log/journal:ro
      - quollix_grafana_vector:/var/lib/vector
    configs:
      - source: vector
        target: /etc/vector/vector.yaml

volumes:
  quollix_grafana_victorialogs: null
  quollix_grafana_victoriametrics: null
  quollix_grafana_vector: null

configs:
  vector:
    content: |
      data_dir: /var/lib/vector
      sources:
        host_journal:
          type: journald
          current_boot_only: true
          extra_args: [--merge]
        host_metrics:
          type: host_metrics
          collectors: [cpu, memory, filesystem]
          scrape_interval_secs: ${HOST_METRICS_INTERVAL_SECONDS}
          filesystem:
            mountpoints:
              includes: [/host, /host/**]
              excludes: [/host/dev/**, /host/proc/**, /host/sys/**, /host/var/lib/docker/**, /host/var/lib/containerd/**]
            filesystems:
              excludes: [autofs, binfmt_misc, cgroup, cgroup2, debugfs, devpts, devtmpfs, fusectl, mqueue, overlay, proc, pstore, securityfs, sysfs, tracefs]
      sinks:
        victoriametrics:
          type: prometheus_remote_write
          inputs: [host_metrics]
          endpoint: http://victoriametrics:8428/api/v1/write
          healthcheck:
            enabled: false
          auth:
            strategy: basic
            user: admin
            password: "${SECRET_VICTORIAMETRICS_PASSWORD}"
        victorialogs:
          type: http
          inputs: [host_journal]
          uri: "http://victorialogs:9428/insert/jsonline?_stream_fields=source_type,host,_SYSTEMD_UNIT&_msg_field=message&_time_field=timestamp"
          auth:
            strategy: basic
            user: admin
            password: "${SECRET_VICTORIALOGS_PASSWORD}"
          compression: gzip
          encoding:
            codec: json
          framing:
            method: newline_delimited
          acknowledgements:
            enabled: true
          healthcheck:
            enabled: false
`

func WaitForMonitoringReady(t *testing.T, client *api_client.QuollixClient) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := u.EventuallyWithTimeout(30*time.Second, 200*time.Millisecond, func() error {
		output, err := exec.CommandContext(ctx, "docker", "inspect", "--type", "container", "--format", "{{if not .State.Running}}{{.Name}}: {{.State.Status}}{{end}}",
			"quollix_grafana_grafana",
			"quollix_grafana_victorialogs",
			"quollix_grafana_victoriametrics",
			"quollix_grafana_vector",
		).CombinedOutput()
		if err != nil {
			return u.Logger.NewError("failed to inspect Grafana containers", "error", err.Error(), "output", string(output))
		}
		if strings.TrimSpace(string(output)) != "" {
			return u.Logger.NewError("waiting for Grafana containers to run", "output", string(output))
		}
		if _, err := client.Metrics.Read(); err != nil {
			return u.Logger.AddContext(err, "readiness", "VictoriaMetrics")
		}
		if _, err := client.Logs.Read(); err != nil {
			return u.Logger.AddContext(err, "readiness", "VictoriaLogs")
		}
		return nil
	})
	assert.Nil(t, err)
}
