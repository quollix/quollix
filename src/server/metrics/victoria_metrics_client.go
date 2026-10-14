package metrics

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"server/apps_basic"
	"server/tools"

	"github.com/quollix/common/quollix/api"
	u "github.com/quollix/common/utils"
)

type VictoriaMetricsClient interface {
	Push([]api.MetricSeries) error
	Pull() ([]api.MetricSeries, error)
	Wipe() error
}

type VictoriaMetricsClientImpl struct {
	appRepository apps_basic.AppRepository
	httpClient    *http.Client
	baseUrl       string
}

func NewVictoriaMetricsClient(baseUrl string, repository apps_basic.AppRepository) *VictoriaMetricsClientImpl {
	return &VictoriaMetricsClientImpl{
		httpClient:    &http.Client{Timeout: 10 * time.Second},
		appRepository: repository,
		baseUrl:       baseUrl,
	}
}

func (c *VictoriaMetricsClientImpl) Push(records []api.MetricSeries) error {
	if len(records) == 0 {
		return nil
	}

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)

	for _, record := range records {
		if err := enc.Encode(record); err != nil {
			return u.Logger.NewError(err.Error())
		}
	}

	req, err := http.NewRequest(http.MethodPost, c.baseUrl+"/api/v1/import", &buf)
	if err != nil {
		return u.Logger.NewError(err.Error())
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.requestWithAuth(req)
	if err != nil {
		return err
	}
	defer u.Close(resp.Body)

	if resp.StatusCode != http.StatusNoContent {
		return tools.GetStatusCodeError(http.StatusNoContent, resp.StatusCode)
	}

	return nil
}

func (c *VictoriaMetricsClientImpl) Pull() ([]api.MetricSeries, error) {
	form := bytes.NewBufferString(`match[]={` + MetricNameLabel + `=~"` + containerMetricPrefix + `.*|` + VolumeUsedBytesMetric + `"}`)
	req, err := http.NewRequest(http.MethodPost, c.baseUrl+"/api/v1/export", form)
	if err != nil {
		return nil, u.Logger.NewError(err.Error())
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.requestWithAuth(req)
	if err != nil {
		return nil, err
	}
	defer u.Close(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, tools.GetStatusCodeError(http.StatusOK, resp.StatusCode)
	}

	var series []api.MetricSeries
	decoder := json.NewDecoder(resp.Body)
	for {
		var record api.MetricSeries
		if err := decoder.Decode(&record); err != nil {
			if errors.Is(err, io.EOF) {
				return series, nil
			}
			return nil, u.Logger.NewError(err.Error())
		}
		series = append(series, record)
	}
}

func (c *VictoriaMetricsClientImpl) Wipe() error {
	form := bytes.NewBufferString(`match[]={` + MetricNameLabel + `!=""}`)
	req, err := http.NewRequest(http.MethodPost, c.baseUrl+"/api/v1/admin/tsdb/delete_series", form)
	if err != nil {
		return u.Logger.NewError(err.Error())
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.requestWithAuth(req)
	if err != nil {
		return err
	}
	defer u.Close(resp.Body)

	if resp.StatusCode != http.StatusNoContent {
		return tools.GetStatusCodeError(http.StatusNoContent, resp.StatusCode)
	}

	return nil
}

func (c *VictoriaMetricsClientImpl) requestWithAuth(req *http.Request) (*http.Response, error) {
	app, err := c.appRepository.GetAppByName("grafana")
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth("admin", app.Secrets["SECRET_VICTORIAMETRICS_PASSWORD"])
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, u.Logger.NewError(err.Error())
	}
	return resp, nil
}
