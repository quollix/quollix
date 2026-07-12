//go:build component

package component

import (
	"encoding/json"
	"testing"

	"server/tests/api_client"
	"server/tools"

	"github.com/quollix/common/assert"
)

func TestHealthEndpoint(t *testing.T) {
	client := api_client.NewQuollixClient()

	responseBody, err := client.Parent.DoRequest(tools.Paths.BackendHealth, nil)
	assert.Nil(t, err)

	var healthResponse map[string]string
	assert.Nil(t, json.Unmarshal(responseBody, &healthResponse))
	assert.Equal(t, "ok", healthResponse["status"])
}
