//go:build victoria

package victoria

import (
	"net/http"
	"testing"
	"time"

	"server/alerts"
	"server/apps_basic"
	"server/logging"
	"server/metrics"

	"github.com/quollix/common/assert"
)

const (
	victoriaMetricsURL = "http://127.0.0.1:18428"
	victoriaLogsURL    = "http://127.0.0.1:19428"
)

type victoriaTestClients struct {
	Metrics      *metrics.VictoriaMetricsClientImpl
	Logs         *logging.VictoriaLogsClientImpl
	AlertQueries alerts.VictoriaAlertQueries
}

func newVictoriaTestClients(t *testing.T) *victoriaTestClients {
	repository := apps_basic.NewAppRepositoryMock(t)
	repository.EXPECT().GetAppByName("grafana").Return(&apps_basic.RepoApp{}, nil).Maybe()
	metricsClient := metrics.NewVictoriaMetricsClient(victoriaMetricsURL, repository, true)
	logsClient := &logging.VictoriaLogsClientImpl{
		AppRepository: repository,
		Client:        &http.Client{Timeout: 10 * time.Second},
		BaseURL:       victoriaLogsURL,
	}
	clients := &victoriaTestClients{
		Metrics:      metricsClient,
		Logs:         logsClient,
		AlertQueries: alerts.NewVictoriaAlertQueries(metricsClient, logsClient),
	}
	clients.Wipe(t)
	t.Cleanup(func() { clients.Wipe(t) })
	return clients
}

func (c *victoriaTestClients) Wipe(t *testing.T) {
	assert.Nil(t, c.Logs.Wipe())
	assert.Nil(t, c.Metrics.Wipe())
}
