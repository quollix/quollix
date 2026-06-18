//go:build prod_profile

package component

import (
	"net/http"
	"server/tools"
	"testing"

	"github.com/quollix/common/assert"
	"github.com/quollix/deepstack"
)

func TestResetTestStateEndpointIsDisabledInProdProfile(t *testing.T) {
	cloud := GetQuollixClient(t)
	_, err := cloud.Parent.DoRequest(tools.Paths.BackendResetTestState, nil)
	assert.NotNil(t, err)
	deepStackError := err.(*deepstack.DeepStackError)
	assert.Equal(t, "request failed", deepStackError.Message)
	statusCode, ok := deepStackError.Context["status_code"]
	assert.True(t, ok)
	assert.Equal(t, http.StatusNotFound, statusCode)
}
