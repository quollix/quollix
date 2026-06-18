//go:build acceptance

package pages

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
	v.Frame.AssertPagePath(tools.Paths.FrontendVersions)
	assert.Equal(v.Frame.t, "Maintainer: "+maintainer, strings.TrimSpace(v.Frame.page.MustElement("#versions-maintainer").MustText()))
	assert.Equal(v.Frame.t, "App Name: "+appName, strings.TrimSpace(v.Frame.page.MustElement("#versions-app-name").MustText()))
	return v
}

func (v *VersionsPage) AssertVersionRowCount(expectedCount int) *VersionsPage {
	rows := v.Frame.page.MustElements("#versions-results-body tr.version-row")
	assert.Equal(v.Frame.t, expectedCount, len(rows))
	return v
}

func (v *VersionsPage) AssertVersionNames(expected []string) *VersionsPage {
	rows := v.Frame.page.MustElements("#versions-results-body tr.version-row")
	actual := make([]string, 0, len(rows))
	for _, row := range rows {
		versionAttr := row.MustAttribute("data-version-name")
		assert.NotNil(v.Frame.t, versionAttr)
		actual = append(actual, *versionAttr)
	}
	sort.Strings(actual)
	sortedExpected := append([]string(nil), expected...)
	sort.Strings(sortedExpected)
	assert.Equal(v.Frame.t, sortedExpected, actual)
	return v
}

func (v *VersionsPage) AssertVersionsAndCreationDates(expected map[string]time.Time) *VersionsPage {
	rows := v.Frame.page.MustElements("#versions-results-body tr.version-row")
	assert.Equal(v.Frame.t, len(expected), len(rows))

	for _, row := range rows {
		versionAttr := row.MustAttribute("data-version-name")
		createdAttr := row.MustAttribute("data-created-at-pretty")
		assert.NotNil(v.Frame.t, versionAttr)
		assert.NotNil(v.Frame.t, createdAttr)

		expectedTime, exists := expected[*versionAttr]
		assert.True(v.Frame.t, exists)

		actualTime, parseErr := time.Parse(tools.PrettyFrontendTimeLayout, strings.TrimSpace(*createdAttr))
		assert.Nil(v.Frame.t, parseErr)
		assert.Equal(v.Frame.t, actualTime.UTC(), expectedTime.UTC())
	}

	return v
}

func (v *VersionsPage) InstallVersion(version string) *VersionsPage {
	row := v.findVersionRow(version)
	installButton, err := row.Element("button.version-install-button")
	assert.Nil(v.Frame.t, err)
	installButton.MustClick()
	return v
}

func (v *VersionsPage) SetVersionFilter(version string) *VersionsPage {
	v.Frame.page.MustElement("#version-filter").MustInput(version)
	return v
}

func (v *VersionsPage) AssertVisibleVersionNames(expected []string) *VersionsPage {
	rows := v.Frame.page.MustElements("#versions-results-body tr.version-row")
	visible := make([]string, 0, len(rows))
	for _, row := range rows {
		styleAttr := row.MustAttribute("style")
		if styleAttr != nil && strings.Contains(*styleAttr, "none") {
			continue
		}
		versionAttr := row.MustAttribute("data-version-name")
		assert.NotNil(v.Frame.t, versionAttr)
		visible = append(visible, *versionAttr)
	}
	sort.Strings(visible)
	sortedExpected := append([]string(nil), expected...)
	sort.Strings(sortedExpected)
	assert.Equal(v.Frame.t, sortedExpected, visible)
	return v
}

func (v *VersionsPage) InstallFilteredVersion() *VersionsPage {
	installButton, err := v.Frame.page.Element("#version-filter-install-button")
	assert.Nil(v.Frame.t, err)
	installButton.MustClick()
	return v
}

func (v *VersionsPage) WaitUntilAppVersionInstalled(appName, expectedVersion string) *VersionsPage {
	err := tools.EventuallyWithTimeout(backupOperationTimeout, 50*time.Millisecond, func() error {
		apps := v.Frame.Client.Apps.ListInstalled()
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
	assert.Nil(v.Frame.t, err)
	return v
}

func (v *VersionsPage) findVersionRow(version string) *rod.Element {
	rows := v.Frame.page.MustElements("#versions-results-body tr.version-row")
	for _, row := range rows {
		versionAttr := row.MustAttribute("data-version-name")
		if versionAttr != nil && *versionAttr == version {
			return row
		}
	}
	assert.True(v.Frame.t, false)
	return nil
}
