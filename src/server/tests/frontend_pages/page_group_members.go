package frontend_pages

import (
	"fmt"
	"server/tools"

	"github.com/quollix/common/assert"
)

type GroupMembersPage struct {
	Frame *FrameType
}

func (g *GroupMembersPage) ClickBack() *GroupsPage {
	backButton, err := g.Frame.Page.Element("#group-members-back-button")
	assert.Nil(g.Frame.T, err)
	g.Frame.Browser.DoAndWaitDOMContentLoaded(func() {
		backButton.MustClick()
	})
	g.Frame.Assert.PagePath(tools.Paths.FrontendGroups)
	return g.Frame.Pages.GroupsPage
}

func (g *GroupMembersPage) SetNonMembersFilter(value string) *GroupMembersPage {
	g.selectionTable("#nonMembersTableBody").SetFilter("#nonMembersFilterInput", value)
	return g
}

func (g *GroupMembersPage) SetMembersFilter(value string) *GroupMembersPage {
	g.selectionTable("#membersTableBody").SetFilter("#membersFilterInput", value)
	return g
}

func (g *GroupMembersPage) AssertUserInNonMembers(username string) *GroupMembersPage {
	g.selectionTable("#nonMembersTableBody").AssertRowPresent(username, true)
	return g
}

func (g *GroupMembersPage) AssertUserNotInNonMembers(username string) *GroupMembersPage {
	g.selectionTable("#nonMembersTableBody").AssertRowPresent(username, false)
	return g
}

func (g *GroupMembersPage) AssertUserInMembers(username string) *GroupMembersPage {
	g.selectionTable("#membersTableBody").AssertRowPresent(username, true)
	return g
}

func (g *GroupMembersPage) AssertUserNotInMembers(username string) *GroupMembersPage {
	g.selectionTable("#membersTableBody").AssertRowPresent(username, false)
	return g
}

func (g *GroupMembersPage) AssertNonMemberRowVisible(username string, expectedVisible bool) *GroupMembersPage {
	g.selectionTable("#nonMembersTableBody").AssertRowVisible(username, expectedVisible)
	return g
}

func (g *GroupMembersPage) AssertMemberRowVisible(username string, expectedVisible bool) *GroupMembersPage {
	g.selectionTable("#membersTableBody").AssertRowVisible(username, expectedVisible)
	return g
}

func (g *GroupMembersPage) AssertNonMembersSelectAllChecked(expected bool) *GroupMembersPage {
	g.selectionTable("#nonMembersTableBody").AssertCheckboxChecked("#non-members-select-all-checkbox", expected)
	return g
}

func (g *GroupMembersPage) AssertMembersSelectAllChecked(expected bool) *GroupMembersPage {
	g.selectionTable("#membersTableBody").AssertCheckboxChecked("#members-select-all-checkbox", expected)
	return g
}

func (g *GroupMembersPage) SetNonMembersSelectAll(checked bool) *GroupMembersPage {
	g.selectionTable("#nonMembersTableBody").SetCheckbox("#non-members-select-all-checkbox", checked)
	return g
}

func (g *GroupMembersPage) SetMembersSelectAll(checked bool) *GroupMembersPage {
	g.selectionTable("#membersTableBody").SetCheckbox("#members-select-all-checkbox", checked)
	return g
}

func (g *GroupMembersPage) AssertNonMemberChecked(username string, expected bool) *GroupMembersPage {
	g.selectionTable("#nonMembersTableBody").AssertRowCheckboxChecked(username, expected)
	return g
}

func (g *GroupMembersPage) AssertMemberChecked(username string, expected bool) *GroupMembersPage {
	g.selectionTable("#membersTableBody").AssertRowCheckboxChecked(username, expected)
	return g
}

func (g *GroupMembersPage) SetNonMemberChecked(username string, checked bool) *GroupMembersPage {
	g.selectionTable("#nonMembersTableBody").SetRowCheckbox(username, checked)
	return g
}

func (g *GroupMembersPage) SetMemberChecked(username string, checked bool) *GroupMembersPage {
	g.selectionTable("#membersTableBody").SetRowCheckbox(username, checked)
	return g
}

func (g *GroupMembersPage) ClickAddSelected() *GroupMembersPage {
	addButton, err := g.Frame.Page.Element("#group-members-add-button")
	assert.Nil(g.Frame.T, err)
	g.Frame.Browser.DoAndWaitDOMContentLoaded(func() {
		addButton.MustClick()
	})
	g.Frame.Assert.PagePath(tools.Paths.FrontendGroupMembers)
	return g
}

func (g *GroupMembersPage) ClickRemoveSelected() *GroupMembersPage {
	removeButton, err := g.Frame.Page.Element("#group-members-remove-button")
	assert.Nil(g.Frame.T, err)
	g.Frame.Browser.DoAndWaitDOMContentLoaded(func() {
		removeButton.MustClick()
	})
	g.Frame.Assert.PagePath(tools.Paths.FrontendGroupMembers)
	return g
}

func (g *GroupMembersPage) selectionTable(tableBodySelector string) SelectionTableHelper {
	return SelectionTableHelper{
		Frame:          g.Frame,
		TableBody:      tableBodySelector,
		RowSelector:    userRowSelector,
		RowCheckboxCSS: "input.member-checkbox",
	}
}

func userRowSelector(username string) string {
	return fmt.Sprintf(`tr[data-name="%s"]`, username)
}
