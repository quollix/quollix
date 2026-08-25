//go:build component

package component

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
	"testing"
	"time"

	"server/tools"

	"github.com/quollix/common/assert"
	"github.com/quollix/common/quollix/api"
	"github.com/quollix/common/quollix/api_client"
	u "github.com/quollix/common/utils"
)

func TestRabbitMQVersionUpdateEnablesFeatureFlags(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	appBeforeUpdate, err := installSampleRabbitMQApp(t, client, tools.SampleRabbitMQAppVersion311Name)
	assert.Nil(t, err)
	assert.Nil(t, client.Apps.Start(appBeforeUpdate.AppId))
	assertSampleRabbitMQReady(t)
	assertSampleRabbitMQManagementReady(t)
	createSampleRabbitMQMessage(t)

	assert.Nil(t, client.Apps.Update(appBeforeUpdate.AppId))

	appAfterUpdate := getInstalledSampleRabbitMQApp(t, client)
	assert.Equal(t, tools.SampleRabbitMQAppVersion312Name, appAfterUpdate.VersionName)
	assert.True(t, appAfterUpdate.IsRunning)
	assertSampleRabbitMQContainerUsesImage(t, "rabbitmq:3.12.14-management-alpine")
	assertSampleRabbitMQReady(t)
	assert.Equal(t, sampleRabbitMQMessagePayload, readSampleRabbitMQMessage(t))
}

const (
	sampleRabbitMQManagementURL  = "http://127.0.0.1:15673"
	sampleRabbitMQQueueName      = "quollix-migration-test"
	sampleRabbitMQMessagePayload = "message-before-upgrade"
)

type rabbitMQPublishResponse struct {
	Routed bool `json:"routed"`
}

type rabbitMQGetMessageResponse struct {
	Payload string `json:"payload"`
}

func installSampleRabbitMQApp(t *testing.T, client *api_client.QuollixClient, version string) (*api.AdminAppDto, error) {
	storeVersion, err := FindVersion(t, client, tools.SampleMaintainer, tools.SampleRabbitMQApp, version)
	if err != nil {
		return nil, err
	}
	if err := client.Apps.InstallFromStoreVersion(storeVersion.VersionId); err != nil {
		return nil, err
	}
	return getInstalledSampleRabbitMQApp(t, client), nil
}

func getInstalledSampleRabbitMQApp(t *testing.T, client *api_client.QuollixClient) *api.AdminAppDto {
	for _, app := range ListInstalledApps(t, client) {
		if app.AppName == tools.SampleRabbitMQApp {
			return &app
		}
	}
	assert.Nil(t, u.Logger.NewError("sample rabbitmq app not found"))
	return nil
}

func assertSampleRabbitMQReady(t *testing.T) {
	deadline := time.Now().Add(10 * time.Second)
	for {
		cmd := exec.Command("docker", "exec", tools.SampleRabbitMQAppContainerName, "rabbitmq-diagnostics", "-q", "check_running") // #nosec G204 (CWE-78): component test checks a known fixture container
		err := cmd.Run()
		if err == nil {
			return
		}
		if time.Now().After(deadline) {
			assert.Nil(t, err)
			return
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func assertSampleRabbitMQManagementReady(t *testing.T) {
	deadline := time.Now().Add(10 * time.Second)
	for {
		_, statusCode, err := callSampleRabbitMQManagementAPI("GET", "/api/overview", nil)
		if err == nil && statusCode == http.StatusOK {
			return
		}
		if time.Now().After(deadline) {
			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, statusCode)
			return
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func createSampleRabbitMQMessage(t *testing.T) {
	requestSampleRabbitMQManagementAPI(t, "PUT", "/api/queues/%2F/"+sampleRabbitMQQueueName, map[string]any{
		"durable":   true,
		"arguments": map[string]any{},
	}, http.StatusCreated, http.StatusNoContent)

	responseBody := requestSampleRabbitMQManagementAPI(t, "POST", "/api/exchanges/%2F/amq.default/publish", map[string]any{
		"properties": map[string]any{
			"delivery_mode": 2,
		},
		"routing_key":      sampleRabbitMQQueueName,
		"payload":          sampleRabbitMQMessagePayload,
		"payload_encoding": "string",
	}, http.StatusOK)

	var response rabbitMQPublishResponse
	assert.Nil(t, json.Unmarshal(responseBody, &response))
	assert.True(t, response.Routed)
}

func readSampleRabbitMQMessage(t *testing.T) string {
	responseBody := requestSampleRabbitMQManagementAPI(t, "POST", "/api/queues/%2F/"+sampleRabbitMQQueueName+"/get", map[string]any{
		"count":    1,
		"ackmode":  "ack_requeue_false",
		"encoding": "auto",
		"truncate": 50000,
	}, http.StatusOK)

	var messages []rabbitMQGetMessageResponse
	assert.Nil(t, json.Unmarshal(responseBody, &messages))
	assert.Equal(t, 1, len(messages))
	return messages[0].Payload
}

func requestSampleRabbitMQManagementAPI(t *testing.T, method string, path string, requestBody any, expectedStatusCodes ...int) []byte {
	responseBody, statusCode, err := callSampleRabbitMQManagementAPI(method, path, requestBody)
	assert.Nil(t, err)
	if !containsStatusCode(expectedStatusCodes, statusCode) {
		assert.Nil(t, u.Logger.NewError(fmt.Sprintf("unexpected rabbitmq management api status code %d: %s", statusCode, string(responseBody))))
	}
	return responseBody
}

func callSampleRabbitMQManagementAPI(method string, path string, requestBody any) ([]byte, int, error) {
	body, err := json.Marshal(requestBody)
	if err != nil {
		return nil, 0, err
	}

	request, err := http.NewRequest(method, sampleRabbitMQManagementURL+path, bytes.NewReader(body))
	if err != nil {
		return nil, 0, err
	}
	request.SetBasicAuth("guest", "guest")
	request.Header.Set("Content-Type", "application/json")

	client := http.Client{Timeout: 5 * time.Second}
	response, err := client.Do(request)
	if err != nil {
		return nil, 0, err
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, response.StatusCode, err
	}
	return responseBody, response.StatusCode, nil
}

func containsStatusCode(statusCodes []int, statusCode int) bool {
	for _, expectedStatusCode := range statusCodes {
		if expectedStatusCode == statusCode {
			return true
		}
	}
	return false
}

func assertSampleRabbitMQContainerUsesImage(t *testing.T, expectedImage string) {
	cmd := exec.Command("docker", "inspect", "--format", "{{.Config.Image}}", tools.SampleRabbitMQAppContainerName) // #nosec G204 (CWE-78): component test inspects a known fixture container
	output, err := cmd.Output()
	assert.Nil(t, err)
	assert.Equal(t, expectedImage, strings.TrimSpace(string(output)))
}
