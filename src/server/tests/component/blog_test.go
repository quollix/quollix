//go:build component

package component

import (
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"server/tools"

	"github.com/quollix/common/assert"
)

func TestMarkBlogCheckedRedirectsAndUpdatesAdmin(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	users := ListUsers(t, client)
	assert.Equal(t, 1, len(users))
	assert.Equal(t, time.Time{}, users[0].LastBlogCheck)

	statusCode, redirectURL, err := client.Blog.MarkChecked()
	assert.Nil(t, err)
	assert.Equal(t, http.StatusSeeOther, statusCode)

	destination, err := url.Parse(redirectURL)
	assert.Nil(t, err)
	assert.Equal(t, "quollix.localhost", destination.Host)

	page, err := client.Frontend.GetPage(destination.Path)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, page.StatusCode)
	assert.Equal(t, "Mock Quollix blog page for component tests", strings.TrimSpace(page.Body))

	updatedAdmin := GetRequiredUserByUsername(t, client, tools.DefaultAdminName)
	assert.True(t, updatedAdmin.LastBlogCheck.After(time.Now().Add(-1*time.Minute)))
	assert.True(t, updatedAdmin.LastBlogCheck.Before(time.Now().Add(1*time.Minute)))
}
