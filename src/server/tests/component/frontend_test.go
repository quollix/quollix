//go:build component

package component

import (
	"net/http"
	"os"
	"path/filepath"
	"server/tests/api_client"
	"server/tools"
	"strings"
	"testing"

	"github.com/quollix/common/assert"
	u "github.com/quollix/common/utils"
)

func TestFrontendReload(t *testing.T) {
	dir, err := u.FindDir("server")
	assert.Nil(t, err)
	frameFilePath := filepath.Join(dir, tools.FrontendResourcesPath, tools.FrontendFramePathInResources)

	client := GetClientAndLogin(t)

	originalContent, err := os.ReadFile(frameFilePath)
	assert.Nil(t, err)

	defer func() {
		writeErr := os.WriteFile(frameFilePath, originalContent, 0600)
		assert.Nil(t, writeErr)
	}()

	updatedContent := strings.Replace(string(originalContent), "sample-test-string</span>", "sample-test-string2</span>", 1)
	err = os.WriteFile(frameFilePath, []byte(updatedContent), 0600)
	assert.Nil(t, err)
	assert.Nil(t, client.Frontend.Reload())

	taskFrontendContain(t, client, "sample-test-string2</span>")
}

func taskFrontendContain(t *testing.T, client *api_client.QuollixClient, expectedSubstring string) {
	responseBody, err := client.Parent.DoRequest("/", nil)
	assert.Nil(t, err)

	normalizedBody := strings.Join(strings.Fields(string(responseBody)), "")
	assert.True(t, strings.Contains(normalizedBody, expectedSubstring))
}

func TestFrontendAdminPageAndMissingPage(t *testing.T) {
	adminClient := GetClientAndLogin(t)
	defer adminClient.Test.ResetTestState()

	InviteUserAndSetPassword(t, adminClient, SampleUsername, "password", SampleUserEmail)

	userClient := api_client.NewQuollixClient()
	assert.Nil(t, userClient.Auth.SignIn(SampleUsername, "password"))

	anonymousClient := api_client.NewQuollixClient()

	adminPage := tools.Paths.FrontendUsers
	missingPage := "/page-that-does-not-exist"

	adminResponse, err := adminClient.Frontend.GetPage(adminPage)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, adminResponse.StatusCode)
	assert.True(t, strings.Contains(adminResponse.Body, "Users</h2>"))

	userResponse, err := userClient.Frontend.GetPage(adminPage)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, userResponse.StatusCode)
	assert.Equal(t, tools.Paths.FrontendSignIn, userResponse.Header.Get("Location"))

	anonymousResponse, err := anonymousClient.Frontend.GetPage(adminPage)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, anonymousResponse.StatusCode)
	assert.Equal(t, tools.Paths.FrontendSignIn, anonymousResponse.Header.Get("Location"))

	for _, client := range []*api_client.QuollixClient{adminClient, userClient, anonymousClient} {
		response, responseErr := client.Frontend.GetPage(missingPage)
		assert.Nil(t, responseErr)
		assert.Equal(t, http.StatusBadRequest, response.StatusCode)
		assert.True(t, strings.Contains(response.Body, tools.PageCouldNotBeLoadedTitle))
	}
}
