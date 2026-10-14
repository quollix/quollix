package logging

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"server/apps_basic"
	"server/tools"

	"github.com/quollix/common/quollix/api"
	u "github.com/quollix/common/utils"
)

type VictoriaLogsClient interface {
	Push(records []api.LogRecord) error
	Pull() ([]api.LogRecord, error)
	Wipe() error
}

type VictoriaLogsClientImpl struct {
	AppRepository apps_basic.AppRepository
	Client        *http.Client
	BaseURL       string
}

func (s *VictoriaLogsClientImpl) Push(records []api.LogRecord) error {
	if len(records) == 0 {
		return nil
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	for _, r := range records {
		if err := enc.Encode(r); err != nil {
			return u.Logger.NewError(err.Error())
		}
	}
	params := url.Values{
		// Group logs for efficient stream filtering
		"_stream_fields": {"maintainer,app,service,stream,level,parsed"},
	}
	resp, err := s.authenticatedRequest(http.MethodPost, "/insert/jsonline?"+params.Encode(), "application/json", &buf)
	if err != nil {
		return err
	}
	defer u.Close(resp.Body)
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return tools.GetStatusCodeError(http.StatusOK, resp.StatusCode)
	}
	return nil
}

func (s *VictoriaLogsClientImpl) Pull() ([]api.LogRecord, error) {
	resp, err := s.authenticatedRequest(http.MethodPost, "/select/logsql/query", "application/x-www-form-urlencoded", strings.NewReader(url.Values{"query": {"*"}}.Encode()))
	if err != nil {
		return nil, err
	}
	defer u.Close(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, tools.GetStatusCodeError(http.StatusOK, resp.StatusCode)
	}

	dec := json.NewDecoder(resp.Body)
	var out []api.LogRecord

	for {
		var row map[string]json.RawMessage
		if err := dec.Decode(&row); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, u.Logger.NewError(err.Error())
		}
		record, err := decodeLogRecord(row)
		if err != nil {
			return nil, err
		}
		out = append(out, record)
	}

	return out, nil
}

func decodeLogRecord(row map[string]json.RawMessage) (api.LogRecord, error) {
	if _, exists := row["attributes"]; !exists {
		if extra, ok := row["extra"]; ok {
			row["attributes"] = extra
		}
	}
	delete(row, "extra")
	for _, key := range []string{"attributes", "parsed"} {
		value, exists := row[key]
		if !exists {
			continue
		}
		var encoded string
		if err := json.Unmarshal(value, &encoded); err == nil {
			if encoded == "" {
				delete(row, key)
			} else {
				row[key] = json.RawMessage(encoded)
			}
		}
	}
	data, err := json.Marshal(row)
	if err != nil {
		return api.LogRecord{}, u.Logger.NewError(err.Error())
	}
	var record api.LogRecord
	if err := json.Unmarshal(data, &record); err != nil {
		return api.LogRecord{}, u.Logger.NewError(err.Error())
	}
	// VictoriaLogs flattening loses nesting and scalar types, so reconstructing them would risk changing the original data.
	for key, value := range row {
		if name, ok := strings.CutPrefix(key, "attributes."); ok {
			var attribute any
			if err := json.Unmarshal(value, &attribute); err != nil {
				return api.LogRecord{}, u.Logger.NewError(err.Error())
			}
			if record.Attributes == nil {
				record.Attributes = make(map[string]any)
			}
			record.Attributes[name] = attribute
		}
	}
	return record, nil
}

func (s *VictoriaLogsClientImpl) Wipe() error {
	resp, err := s.authenticatedRequest(http.MethodPost, "/delete/run_task?filter=*", "", nil)
	if err != nil {
		return err
	}
	defer u.Close(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return tools.GetStatusCodeError(http.StatusOK, resp.StatusCode)
	}
	var task struct {
		ID string `json:"task_id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&task); err != nil {
		return u.Logger.NewError(err.Error())
	}
	if task.ID == "" {
		return u.Logger.NewError("deletion task ID is missing")
	}
	return u.EventuallyWithTimeout(20*time.Second, 200*time.Millisecond, func() error {
		response, err := s.authenticatedRequest(http.MethodGet, "/delete/active_tasks", "", nil)
		if err != nil {
			return err
		}
		defer u.Close(response.Body)
		if response.StatusCode != http.StatusOK {
			return tools.GetStatusCodeError(http.StatusOK, response.StatusCode)
		}
		var tasks []struct {
			ID string `json:"task_id"`
		}
		if err := json.NewDecoder(response.Body).Decode(&tasks); err != nil {
			return u.Logger.NewError(err.Error())
		}
		for _, activeTask := range tasks {
			if activeTask.ID == task.ID {
				return u.Logger.NewError("waiting for log deletion")
			}
		}
		return nil
	})
}

func (s *VictoriaLogsClientImpl) authenticatedRequest(method, path, contentType string, body io.Reader) (*http.Response, error) {
	app, err := s.AppRepository.GetAppByName("grafana")
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(method, s.BaseURL+path, body)
	if err != nil {
		return nil, u.Logger.NewError(err.Error())
	}
	req.SetBasicAuth("admin", app.Secrets["SECRET_VICTORIALOGS_PASSWORD"])
	req.Header.Set("Content-Type", contentType)
	response, err := s.Client.Do(req)
	if err != nil {
		return nil, u.Logger.NewError(err.Error())
	}
	return response, nil
}
