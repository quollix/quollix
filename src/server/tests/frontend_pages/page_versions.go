package frontend_pages

import (
	"server/tools"
	"sort"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/quollix/common/assert"
	u "github.com/quollix/common/utils"
)

type VersionsPage struct {
	Frame *FrameType
}

func (v *VersionsPage) AssertVersionsPageHeader(maintainer, appName string) *VersionsPage {
	v.Frame.Assert.PagePath(tools.Paths.FrontendVersions)
	assert.Equal(v.Frame.T, "Maintainer: "+maintainer, strings.TrimSpace(v.Frame.Page.MustElement("#versions-maintainer").MustText()))
	assert.Equal(v.Frame.T, "App Name: "+appName, strings.TrimSpace(v.Frame.Page.MustElement("#versions-app-name").MustText()))
	return v
}

func (v *VersionsPage) AssertVersionRowCount(expectedCount int) *VersionsPage {
	rows := v.Frame.Page.MustElements("#versions-results-body tr.version-row")
	assert.Equal(v.Frame.T, expectedCount, len(rows))
	return v
}

func (v *VersionsPage) AssertVersionNames(expected []string) *VersionsPage {
	rows := v.Frame.Page.MustElements("#versions-results-body tr.version-row")
	actual := make([]string, 0, len(rows))
	for _, row := range rows {
		versionAttr := row.MustAttribute("data-version-name")
		assert.NotNil(v.Frame.T, versionAttr)
		actual = append(actual, *versionAttr)
	}
	sort.Strings(actual)
	sortedExpected := append([]string(nil), expected...)
	sort.Strings(sortedExpected)
	assert.Equal(v.Frame.T, sortedExpected, actual)
	return v
}

func (v *VersionsPage) AssertVersionsAndCreationDates(expected map[string]time.Time) *VersionsPage {
	rows := v.Frame.Page.MustElements("#versions-results-body tr.version-row")
	assert.Equal(v.Frame.T, len(expected), len(rows))

	for _, row := range rows {
		versionAttr := row.MustAttribute("data-version-name")
		createdAttr := row.MustAttribute("data-created-at-pretty")
		assert.NotNil(v.Frame.T, versionAttr)
		assert.NotNil(v.Frame.T, createdAttr)

		expectedTime, exists := expected[*versionAttr]
		assert.True(v.Frame.T, exists)

		actualTime, parseErr := time.Parse(tools.PrettyFrontendTimeLayout, strings.TrimSpace(*createdAttr))
		assert.Nil(v.Frame.T, parseErr)
		assert.Equal(v.Frame.T, actualTime.UTC(), expectedTime.UTC())
	}

	return v
}

func (v *VersionsPage) InstallVersion(version string) *VersionsPage {
	row := v.findVersionRow(version)
	installButton, err := row.Element("button.version-install-button")
	assert.Nil(v.Frame.T, err)
	installButton.MustClick()
	return v
}

func (v *VersionsPage) SetVersionFilter(version string) *VersionsPage {
	v.Frame.Page.MustElement("#version-filter").MustInput(version)
	return v
}

func (v *VersionsPage) AssertVisibleVersionNames(expected []string) *VersionsPage {
	rows := v.Frame.Page.MustElements("#versions-results-body tr.version-row")
	visible := make([]string, 0, len(rows))
	for _, row := range rows {
		styleAttr := row.MustAttribute("style")
		if styleAttr != nil && strings.Contains(*styleAttr, "none") {
			continue
		}
		versionAttr := row.MustAttribute("data-version-name")
		assert.NotNil(v.Frame.T, versionAttr)
		visible = append(visible, *versionAttr)
	}
	sort.Strings(visible)
	sortedExpected := append([]string(nil), expected...)
	sort.Strings(sortedExpected)
	assert.Equal(v.Frame.T, sortedExpected, visible)
	return v
}

func (v *VersionsPage) InstallFilteredVersion() *VersionsPage {
	installButton, err := v.Frame.Page.Element("#version-filter-install-button")
	assert.Nil(v.Frame.T, err)
	installButton.MustClick()
	return v
}

func (v *VersionsPage) WaitUntilAppVersionInstalled(appName, expectedVersion string) *VersionsPage {
	err := tools.EventuallyWithTimeout(backupOperationTimeout, 50*time.Millisecond, func() error {
		apps, err := v.Frame.Client.Apps.ListInstalled()
		assert.Nil(v.Frame.T, err)
		for _, app := range apps {
			if app.AppName == appName && app.VersionName == expectedVersion {
				return nil
			}
		}
		return u.Logger.NewError(
			"expected installed app version not present yet",
			"app_name", appName,
			"expected_version", expectedVersion,
		)
	})
	assert.Nil(v.Frame.T, err)
	return v
}

func (v *VersionsPage) findVersionRow(version string) *rod.Element {
	rows := v.Frame.Page.MustElements("#versions-results-body tr.version-row")
	for _, row := range rows {
		versionAttr := row.MustAttribute("data-version-name")
		if versionAttr != nil && *versionAttr == version {
			return row
		}
	}
	assert.True(v.Frame.T, false)
	return nil
}
