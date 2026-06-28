//go:build prod_profile

package component

import (
	"net/http"
	"server/tests/api_client"
	"server/tools"
	"testing"

	"github.com/quollix/common/assert"
	"github.com/quollix/deepstack"
)

func TestDevelopmentRoutesAreAbsentInProdProfile(t *testing.T) {
	cloud := api_client.NewQuollixClient()
	assert.Nil(t, cloud.Auth.SignIn(tools.DefaultAdminName, tools.DefaultAdminPassword))

	developmentRoutePaths := []string{
		tools.Paths.BackendResetTestState,
		tools.Paths.BackendReloadFrontendTemplatesFromFileSystem,
		tools.Paths.BackendStoreReloadPublishedApps,
	}

	for _, path := range developmentRoutePaths {
		t.Run(path, func(t *testing.T) {
			_, err := cloud.Parent.DoRequest(path, nil)
			assert.NotNil(t, err)
			deepStackError := err.(*deepstack.DeepStackError)
			assert.Equal(t, "request failed", deepStackError.Message)
			statusCode, ok := deepStackError.Context["status_code"]
			assert.True(t, ok)
			assert.Equal(t, http.StatusNotFound, statusCode)
		})
	}
}
