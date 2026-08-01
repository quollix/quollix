package frontend

import (
	"testing"

	"github.com/quollix/common/assert"
)

func TestParseAppOpenPath_DefaultsToRoot(t *testing.T) {
	appPath, err := parseAppOpenPath("")

	assert.Nil(t, err)
	assert.Equal(t, "/", appPath.Path)
	assert.Equal(t, "", appPath.RawQuery)
}

func TestParseAppOpenPath_PreservesPathAndQuery(t *testing.T) {
	appPath, err := parseAppOpenPath("/documents/view?id=123&tab=files")

	assert.Nil(t, err)
	assert.Equal(t, "/documents/view", appPath.Path)
	assert.Equal(t, "id=123&tab=files", appPath.RawQuery)
}

func TestParseAppOpenPath_RejectsPathWithoutLeadingSlash(t *testing.T) {
	_, err := parseAppOpenPath("documents/view")

	assert.NotNil(t, err)
}

func TestHasValidNextURL_AcceptsEmptyAndRelativeRequestURI(t *testing.T) {
	assert.True(t, hasValidNextURL(""))
	assert.True(t, hasValidNextURL("/installed-apps"))
	assert.True(t, hasValidNextURL("/authorize?client_id=app&state=abc"))
}

func TestHasValidNextURL_RejectsExternalAndInvalidRequestURI(t *testing.T) {
	assert.False(t, hasValidNextURL("https://example.invalid"))
	assert.False(t, hasValidNextURL("//example.invalid"))
	assert.False(t, hasValidNextURL("installed-apps"))
	assert.False(t, hasValidNextURL("/invalid path"))
}
