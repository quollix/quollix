//go:build acceptance

package acceptance

import (
	"net/http"
	"os"
	"server/tests/acceptance/pages"
	"server/tests/component"
	"server/tools"
	"testing"
	"time"

	"github.com/quollix/common/assert"
)

const (
	username        = "sampleuser"
	sampleUserEmail = "sampleuser@example.invalid"
)

func TestMain(m *testing.M) {
	defer pages.CloseBrowser()
	code := m.Run()
	os.Exit(code)
}

func checkAuthWithCookie(t *testing.T, cookie *http.Cookie) error {
	client := component.GetQuollixClient(t)
	client.Parent.Cookie = cookie
	_, err := client.Parent.DoRequest(tools.Paths.BackendCheckAuth, nil)
	return err
}

func assertFrontendTimestampSet(t *testing.T, value string) {
	assert.NotEqual(t, "", value)
	timestamp, err := time.Parse(tools.PrettyFrontendTimeLayout, value)
	assert.True(t, timestamp.After(time.Unix(0, 0)))
	assert.Nil(t, err)
}

func assertFrontendRelativeTimeSet(t *testing.T, value string) {
	assert.NotEqual(t, "", value)
	assert.True(t, value == "0s ago" || len(value) > len(" ago"))
}
