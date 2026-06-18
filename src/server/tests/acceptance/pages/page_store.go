//go:build acceptance

package pages

import (
	"server/tools"

	"github.com/go-rod/rod"
	"github.com/quollix/common/assert"
)

type StorePage struct {
	Frame *FrameType
}

func (l *StorePage) InstallSampleApp() *StorePage {
	l.Frame.AssertPagePath(tools.Paths.FrontendStore)
	l.EnableUnofficialSearchAndConfirm().
		SetMaintainerFilter("samplemaintainer").
		SetSearchAppName("sampleapp").
		Search().
		AssertSearchRowCount(1).
		AssertSearchContainsResult("samplemaintainer", "sampleapp", "2.0").
		InstallFromResult("samplemaintainer", "sampleapp")

	l.Frame.AssertSnackbarVisibleWithTextEventually("Installation successful")

	return l
}

func (l *StorePage) SetSearchAppName(appName string) *StorePage {
	l.Frame.page.MustElement("#app-input").MustInput(appName)
	return l
}

func (l *StorePage) Search() *StorePage {
	l.Frame.DoAndWaitDOMContentLoaded(func() {
		l.Frame.page.MustElement("#search-button").MustClick()
	})
	return l
}

func (l *StorePage) EnableUnofficialSearchAndConfirm() *StorePage {
	l.Frame.page.MustElement("#unofficial").MustClick()
	l.Frame.ConfirmDialog()
	checked, _, hasErr := l.Frame.page.Has("#unofficial:checked")
	assert.Nil(l.Frame.t, hasErr)
	assert.True(l.Frame.t, checked)
	return l
}

func (l *StorePage) SetMaintainerFilter(maintainer string) *StorePage {
	l.Frame.page.MustElement("#maintainer-input").MustInput(maintainer)
	return l
}

func (l *StorePage) AssertNoSearchRows() *StorePage {
	rows := l.Frame.page.MustElements(`#store-results-body tr.store-result-row`)
	assert.Equal(l.Frame.t, 0, len(rows))
	return l
}

func (l *StorePage) AssertSearchRowCount(expectedCount int) *StorePage {
	rows := l.Frame.page.MustElements(`#store-results-body tr.store-result-row`)
	assert.Equal(l.Frame.t, expectedCount, len(rows))
	return l
}

func (l *StorePage) AssertSearchContainsResult(maintainer, appName, latestVersion string) *StorePage {
	apps := l.searchApps()
	found := false
	for _, app := range apps {
		if app.Maintainer == maintainer && app.AppName == appName && app.LatestVersionName == latestVersion {
			found = true
			break
		}
	}
	assert.True(l.Frame.t, found)
	return l
}

func (l *StorePage) AssertSearchResultCreatedAt(maintainer, appName, expectedCreatedAt string) *StorePage {
	row := l.findSearchResultRow(maintainer, appName)
	assert.Equal(l.Frame.t, expectedCreatedAt, row.MustElement(".store-result-created-at").MustText())
	return l
}

func (l *StorePage) InstallFromResult(maintainer, appName string) *StorePage {
	row := l.findSearchResultRow(maintainer, appName)
	installButton, err := row.Element("button.store-install-button")
	assert.Nil(l.Frame.t, err)
	installButton.MustClick()
	return l
}

func (l *StorePage) OpenVersionsFromResult(maintainer, appName string) *VersionsPage {
	row := l.findSearchResultRow(maintainer, appName)
	versionButton, err := row.Element("button.store-version-button")
	assert.Nil(l.Frame.t, err)
	l.Frame.DoAndWaitDOMContentLoaded(func() {
		versionButton.MustClick()
	})
	return l.Frame.VersionsPage
}

type storeSearchResult struct {
	Maintainer        string
	AppName           string
	LatestVersionName string
}

func (l *StorePage) searchApps() []storeSearchResult {
	rows := l.Frame.page.MustElements(`#store-results-body tr.store-result-row`)
	out := make([]storeSearchResult, 0, len(rows))

	for _, r := range rows {
		m := r.MustAttribute("data-maintainer")
		a := r.MustAttribute("data-app")
		assert.NotNil(l.Frame.t, m)
		assert.NotNil(l.Frame.t, a)

		v := r.MustElement(".version-button").MustText()

		out = append(out, storeSearchResult{
			Maintainer:        *m,
			AppName:           *a,
			LatestVersionName: v,
		})
	}
	return out
}

func (l *StorePage) findSearchResultRow(maintainer, appName string) *rod.Element {
	rows := l.Frame.page.MustElements(`#store-results-body tr.store-result-row`)
	for _, row := range rows {
		rowMaintainer := row.MustAttribute("data-maintainer")
		rowApp := row.MustAttribute("data-app")
		assert.NotNil(l.Frame.t, rowMaintainer)
		assert.NotNil(l.Frame.t, rowApp)
		if *rowMaintainer == maintainer && *rowApp == appName {
			return row
		}
	}
	assert.True(l.Frame.t, false)
	return nil
}
