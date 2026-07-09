package frontend_pages

import (
	"server/tools"
	"strings"

	"github.com/go-rod/rod"
	"github.com/quollix/common/assert"
)

type GroupsPage struct {
	Frame *FrameType
}

type GroupRow struct {
	Name string
}

func (g *GroupsPage) CreateGroup(groupName string) *GroupsPage {
	g.Frame.Assert.PagePath(tools.Paths.FrontendGroups)
	nameInput, err := g.Frame.Page.Element("#create-group-name-input")
	assert.Nil(g.Frame.T, err)
	nameInput.MustInput(groupName)

	createButton, err := g.Frame.Page.Element("#create-group-button")
	assert.Nil(g.Frame.T, err)
	g.Frame.Browser.DoAndWaitDOMContentLoaded(func() {
		createButton.MustClick()
	})
	g.Frame.Assert.PagePath(tools.Paths.FrontendGroups)
	return g
}

func (g *GroupsPage) DeleteGroup(groupName string) *GroupsPage {
	row := g.getRequiredGroupRowElement(groupName)
	deleteButton, err := row.Element("button.group-delete-button")
	assert.Nil(g.Frame.T, err)

	g.Frame.Browser.DoAndWaitDOMContentLoaded(func() {
		deleteButton.MustClick()
		g.Frame.Browser.ConfirmDialog()
	})
	g.Frame.Assert.PagePath(tools.Paths.FrontendGroups)
	return g
}

func (g *GroupsPage) OpenGroupMembersPage(groupName string) *GroupMembersPage {
	row := g.getRequiredGroupRowElement(groupName)
	manageMembersButton, err := row.Element("button.group-manage-members-button")
	assert.Nil(g.Frame.T, err)
	g.Frame.Browser.DoAndWaitDOMContentLoaded(func() {
		manageMembersButton.MustClick()
	})
	g.Frame.Assert.PagePath(tools.Paths.FrontendGroupMembers)
	return g.Frame.Pages.GroupMembersPage
}

func (g *GroupsPage) OpenGroupAppsPage(groupName string) *GroupAppsPage {
	row := g.getRequiredGroupRowElement(groupName)
	manageAppsButton, err := row.Element("button.group-manage-apps-button")
	assert.Nil(g.Frame.T, err)
	g.Frame.Browser.DoAndWaitDOMContentLoaded(func() {
		manageAppsButton.MustClick()
	})
	g.Frame.Assert.PagePath(tools.Paths.FrontendGroupApps)
	return g.Frame.Pages.GroupAppsPage
}

func (g *GroupsPage) ListGroups() []GroupRow {
	rows, err := g.Frame.Page.Elements("tr.group-row")
	assert.Nil(g.Frame.T, err)

	out := make([]GroupRow, 0, len(rows))
	for _, row := range rows {
		nameCell, err := row.Element(".group-name-cell")
		assert.Nil(g.Frame.T, err)
		name, err := nameCell.Text()
		assert.Nil(g.Frame.T, err)

		manageMembersButton, err := row.Element("button.group-manage-members-button")
		assert.Nil(g.Frame.T, err)
		assert.NotNil(g.Frame.T, manageMembersButton)

		manageAppsButton, err := row.Element("button.group-manage-apps-button")
		assert.Nil(g.Frame.T, err)
		assert.NotNil(g.Frame.T, manageAppsButton)

		deleteButton, err := row.Element("button.group-delete-button")
		assert.Nil(g.Frame.T, err)
		assert.NotNil(g.Frame.T, deleteButton)

		out = append(out, GroupRow{Name: strings.TrimSpace(name)})
	}
	return out
}

func (g *GroupsPage) GetRequiredGroup(groupName string) *GroupRow {
	groups := g.ListGroups()
	for _, group := range groups {
		if group.Name == groupName {
			groupCopy := group
			return &groupCopy
		}
	}
	g.Frame.T.Fatalf("group not found: %s", groupName)
	return nil
}

func (g *GroupsPage) AssertGroupAbsent(groupName string) *GroupsPage {
	groups := g.ListGroups()
	for _, group := range groups {
		assert.NotEqual(g.Frame.T, groupName, group.Name)
	}
	return g
}

func (g *GroupsPage) getRequiredGroupRowElement(groupName string) *rod.Element {
	rows, err := g.Frame.Page.Elements("tr.group-row")
	assert.Nil(g.Frame.T, err)

	for _, row := range rows {
		nameAttr, err := row.Attribute("data-group-name")
		assert.Nil(g.Frame.T, err)
		assert.NotNil(g.Frame.T, nameAttr)
		if strings.TrimSpace(*nameAttr) == groupName {
			return row
		}
	}
	g.Frame.T.Fatalf("group row not found: %s", groupName)
	return nil
}
