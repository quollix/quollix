//go:build component

package component

import (
	"net/http"
	"os"
	"path/filepath"
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
	assert.Nil(client.T, err)

	defer func() {
		writeErr := os.WriteFile(frameFilePath, originalContent, 0600)
		assert.Nil(client.T, writeErr)
	}()

	updatedContent := strings.Replace(string(originalContent), "sample-test-string</span>", "sample-test-string2</span>", 1)
	err = os.WriteFile(frameFilePath, []byte(updatedContent), 0600)
	assert.Nil(client.T, err)
	client.Frontend.Reload()

	taskFrontendContain(client, "sample-test-string2</span>")
}

func taskFrontendContain(client *QuollixClient, expectedSubstring string) {
	responseBody, err := client.Parent.DoRequest("/", nil)
	assert.Nil(client.T, err)

	normalizedBody := strings.Join(strings.Fields(string(responseBody)), "")
	assert.True(client.T, strings.Contains(normalizedBody, expectedSubstring))
}

func TestFrontendAdminPageAndMissingPage(t *testing.T) {
	adminClient := GetClientAndLogin(t)
	defer adminClient.Test.ResetTestState()

	InviteUserAndSetPassword(adminClient, SampleUsername, "password", SampleUserEmail)

	userClient := GetQuollixClient(t)
	assert.Nil(t, userClient.Auth.Login(SampleUsername, "password"))

	anonymousClient := GetQuollixClient(t)

	adminPage := tools.Paths.FrontendUsers
	missingPage := "/page-that-does-not-exist"

	adminResponse, err := adminClient.Frontend.GetPage(adminPage)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, adminResponse.StatusCode)
	assert.True(t, strings.Contains(adminResponse.Body, "Users</h2>"))

	userResponse, err := userClient.Frontend.GetPage(adminPage)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, userResponse.StatusCode)
	assert.Equal(t, tools.Paths.FrontendLogin, userResponse.Header.Get("Location"))

	anonymousResponse, err := anonymousClient.Frontend.GetPage(adminPage)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, anonymousResponse.StatusCode)
	assert.Equal(t, tools.Paths.FrontendLogin, anonymousResponse.Header.Get("Location"))

	for _, client := range []*QuollixClient{adminClient, userClient, anonymousClient} {
		response, responseErr := client.Frontend.GetPage(missingPage)
		assert.Nil(t, responseErr)
		assert.Equal(t, http.StatusBadRequest, response.StatusCode)
		assert.True(t, strings.Contains(response.Body, tools.PageCouldNotBeLoadedTitle))
	}
}
