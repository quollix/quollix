package metrics

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"time"

	"server/apps_basic"
	"server/tools"

	"github.com/quollix/common/quollix/api"
	u "github.com/quollix/common/utils"
)

type VictoriaMetricsClient interface {
	Push([]api.MetricSeries) error
	Pull() ([]api.MetricSeries, error)
	Query(string) (VictoriaMetricsQueryResult, error)
	Wipe() error
}

type VictoriaMetricsQueryResult struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Metric map[string]string   `json:"metric"`
			Value  []json.RawMessage   `json:"value"`
			Values [][]json.RawMessage `json:"values"`
		} `json:"result"`
	} `json:"data"`
}

type VictoriaMetricsClientImpl struct {
	appRepository       apps_basic.AppRepository
	httpClient          *http.Client
	baseUrl             string
	forceFlushAfterPush bool
}

func NewVictoriaMetricsClient(baseUrl string, repository apps_basic.AppRepository, forceFlushAfterPush bool) *VictoriaMetricsClientImpl {
	return &VictoriaMetricsClientImpl{
		httpClient:          &http.Client{Timeout: 10 * time.Second},
		appRepository:       repository,
		baseUrl:             baseUrl,
		forceFlushAfterPush: forceFlushAfterPush,
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

	if resp.StatusCode != http.StatusNoContent {
		u.Close(resp.Body)
		return tools.GetStatusCodeError(http.StatusNoContent, resp.StatusCode)
	}
	u.Close(resp.Body)
	if c.forceFlushAfterPush {
		return c.forceFlush()
	}

	return nil
}

func (c *VictoriaMetricsClientImpl) forceFlush() error {
	req, err := http.NewRequest(http.MethodPost, c.baseUrl+"/internal/force_flush", nil)
	if err != nil {
		return u.Logger.NewError(err.Error())
	}
	resp, err := c.requestWithAuth(req)
	if err != nil {
		return err
	}
	defer u.Close(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return tools.GetStatusCodeError(http.StatusOK, resp.StatusCode)
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

func (c *VictoriaMetricsClientImpl) Query(expression string) (VictoriaMetricsQueryResult, error) {
	var result VictoriaMetricsQueryResult
	req, err := http.NewRequest(http.MethodGet, c.baseUrl+"/api/v1/query?query="+url.QueryEscape(expression), nil)
	if err != nil {
		return result, u.Logger.NewError(err.Error())
	}
	resp, err := c.requestWithAuth(req)
	if err != nil {
		return result, err
	}
	defer u.Close(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return result, tools.GetStatusCodeError(http.StatusOK, resp.StatusCode)
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return result, u.Logger.NewError(err.Error())
	}
	return result, nil
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
