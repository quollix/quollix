package frontend

import (
	"net/url"
	"testing"

	frontendpages "server/frontend/pages"

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

func TestParseStorePageQuery_UsesMaintainerOnlyForUnofficialSearch(t *testing.T) {
	const (
		sampleMaintainer = "samplemaintainer"
		sampleApp        = "sampleapp"
	)

	testCases := []struct {
		name               string
		query              url.Values
		expectedMaintainer string
		expectedUnofficial bool
	}{
		{
			name:               "unofficial search keeps maintainer",
			query:              newStorePageQueryForTest(sampleMaintainer, sampleApp, true),
			expectedMaintainer: sampleMaintainer,
			expectedUnofficial: true,
		},
		{
			name:               "official search clears stale maintainer",
			query:              newStorePageQueryForTest(sampleMaintainer, sampleApp, false),
			expectedMaintainer: "",
			expectedUnofficial: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			maintainerName, appName, showUnofficial, isSearch := parseStorePageQuery(testCase.query)

			assert.Equal(t, testCase.expectedMaintainer, maintainerName)
			assert.Equal(t, sampleApp, appName)
			assert.Equal(t, testCase.expectedUnofficial, showUnofficial)
			assert.True(t, isSearch)
		})
	}
}

func newStorePageQueryForTest(maintainerName, appName string, showUnofficial bool) url.Values {
	query := url.Values{}
	query.Set(frontendpages.QueryParams.Store.MaintainerName, maintainerName)
	query.Set(frontendpages.QueryParams.Store.AppName, appName)
	query.Set(frontendpages.QueryParams.Store.IsSearch, "true")
	if showUnofficial {
		query.Set(frontendpages.QueryParams.Store.ShowUnofficial, "true")
	}
	return query
}
