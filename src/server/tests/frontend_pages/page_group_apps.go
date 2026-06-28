package frontend_pages

import (
	"fmt"
	"server/tools"

	"github.com/quollix/common/assert"
)

type GroupAppsPage struct {
	Frame *FrameType
}

func (g *GroupAppsPage) ClickBack() *GroupsPage {
	backButton, err := g.Frame.Page.Element("#group-apps-back-button")
	assert.Nil(g.Frame.T, err)
	g.Frame.Browser.DoAndWaitDOMContentLoaded(func() {
		backButton.MustClick()
	})
	g.Frame.Assert.PagePath(tools.Paths.FrontendGroups)
	return g.Frame.Pages.GroupsPage
}

func (g *GroupAppsPage) SetNoAccessFilter(value string) *GroupAppsPage {
	g.selectionTable("#noAccessGrantedTableBody").SetFilter("#noAccessFilterInput", value)
	return g
}

func (g *GroupAppsPage) SetGrantedFilter(value string) *GroupAppsPage {
	g.selectionTable("#accessGrantedTableBody").SetFilter("#grantedFilterInput", value)
	return g
}

func (g *GroupAppsPage) AssertAppInNoAccess(appName string) *GroupAppsPage {
	g.selectionTable("#noAccessGrantedTableBody").AssertRowPresent(appName, true)
	return g
}

func (g *GroupAppsPage) AssertAppNotInNoAccess(appName string) *GroupAppsPage {
	g.selectionTable("#noAccessGrantedTableBody").AssertRowPresent(appName, false)
	return g
}

func (g *GroupAppsPage) AssertAppInGranted(appName string) *GroupAppsPage {
	g.selectionTable("#accessGrantedTableBody").AssertRowPresent(appName, true)
	return g
}

func (g *GroupAppsPage) AssertAppNotInGranted(appName string) *GroupAppsPage {
	g.selectionTable("#accessGrantedTableBody").AssertRowPresent(appName, false)
	return g
}

func (g *GroupAppsPage) AssertNoAccessRowVisible(appName string, expectedVisible bool) *GroupAppsPage {
	g.selectionTable("#noAccessGrantedTableBody").AssertRowVisible(appName, expectedVisible)
	return g
}

func (g *GroupAppsPage) AssertGrantedRowVisible(appName string, expectedVisible bool) *GroupAppsPage {
	g.selectionTable("#accessGrantedTableBody").AssertRowVisible(appName, expectedVisible)
	return g
}

func (g *GroupAppsPage) AssertNoAccessSelectAllChecked(expected bool) *GroupAppsPage {
	g.selectionTable("#noAccessGrantedTableBody").AssertCheckboxChecked("#group-apps-no-access-select-all-checkbox", expected)
	return g
}

func (g *GroupAppsPage) AssertGrantedSelectAllChecked(expected bool) *GroupAppsPage {
	g.selectionTable("#accessGrantedTableBody").AssertCheckboxChecked("#group-apps-access-granted-select-all-checkbox", expected)
	return g
}

func (g *GroupAppsPage) SetNoAccessSelectAll(checked bool) *GroupAppsPage {
	g.selectionTable("#noAccessGrantedTableBody").SetCheckbox("#group-apps-no-access-select-all-checkbox", checked)
	return g
}

func (g *GroupAppsPage) SetGrantedSelectAll(checked bool) *GroupAppsPage {
	g.selectionTable("#accessGrantedTableBody").SetCheckbox("#group-apps-access-granted-select-all-checkbox", checked)
	return g
}

func (g *GroupAppsPage) AssertNoAccessChecked(appName string, expected bool) *GroupAppsPage {
	g.selectionTable("#noAccessGrantedTableBody").AssertRowCheckboxChecked(appName, expected)
	return g
}

func (g *GroupAppsPage) AssertGrantedChecked(appName string, expected bool) *GroupAppsPage {
	g.selectionTable("#accessGrantedTableBody").AssertRowCheckboxChecked(appName, expected)
	return g
}

func (g *GroupAppsPage) SetNoAccessChecked(appName string, checked bool) *GroupAppsPage {
	g.selectionTable("#noAccessGrantedTableBody").SetRowCheckbox(appName, checked)
	return g
}

func (g *GroupAppsPage) SetGrantedChecked(appName string, checked bool) *GroupAppsPage {
	g.selectionTable("#accessGrantedTableBody").SetRowCheckbox(appName, checked)
	return g
}

func (g *GroupAppsPage) ClickAddSelected() *GroupAppsPage {
	addButton, err := g.Frame.Page.Element("#group-apps-add-button")
	assert.Nil(g.Frame.T, err)
	g.Frame.Browser.DoAndWaitDOMContentLoaded(func() {
		addButton.MustClick()
	})
	g.Frame.Assert.PagePath(tools.Paths.FrontendGroupApps)
	return g
}

func (g *GroupAppsPage) ClickRemoveSelected() *GroupAppsPage {
	removeButton, err := g.Frame.Page.Element("#group-apps-remove-button")
	assert.Nil(g.Frame.T, err)
	g.Frame.Browser.DoAndWaitDOMContentLoaded(func() {
		removeButton.MustClick()
	})
	g.Frame.Assert.PagePath(tools.Paths.FrontendGroupApps)
	return g
}

func (g *GroupAppsPage) selectionTable(tableBodySelector string) SelectionTableHelper {
	return SelectionTableHelper{
		Frame:          g.Frame,
		TableBody:      tableBodySelector,
		RowSelector:    appRowSelector,
		RowCheckboxCSS: "input.app-checkbox",
	}
}

func appRowSelector(appName string) string {
	return fmt.Sprintf(`tr[data-name="%s"]`, appName)
}
