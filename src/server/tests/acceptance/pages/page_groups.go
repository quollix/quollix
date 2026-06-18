//go:build acceptance

package pages

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
	g.Frame.AssertPagePath(tools.Paths.FrontendGroups)
	nameInput, err := g.Frame.Page().Element("#create-group-name-input")
	assert.Nil(g.Frame.TestingT(), err)
	nameInput.MustInput(groupName)

	createButton, err := g.Frame.Page().Element("#create-group-button")
	assert.Nil(g.Frame.TestingT(), err)
	g.Frame.DoAndWaitDOMContentLoaded(func() {
		createButton.MustClick()
	})
	g.Frame.AssertPagePath(tools.Paths.FrontendGroups)
	return g
}

func (g *GroupsPage) DeleteGroup(groupName string) *GroupsPage {
	row := g.getRequiredGroupRowElement(groupName)
	deleteButton, err := row.Element("button.group-delete-button")
	assert.Nil(g.Frame.TestingT(), err)

	g.Frame.DoAndWaitDOMContentLoaded(func() {
		deleteButton.MustClick()
		g.Frame.ConfirmDialog()
	})
	g.Frame.AssertPagePath(tools.Paths.FrontendGroups)
	return g
}

func (g *GroupsPage) OpenGroupMembersPage(groupName string) *GroupMembersPage {
	row := g.getRequiredGroupRowElement(groupName)
	manageMembersButton, err := row.Element("button.group-manage-members-button")
	assert.Nil(g.Frame.TestingT(), err)
	g.Frame.DoAndWaitDOMContentLoaded(func() {
		manageMembersButton.MustClick()
	})
	g.Frame.AssertPagePath(tools.Paths.FrontendGroupMembers)
	return g.Frame.GroupMembersPage
}

func (g *GroupsPage) OpenGroupAppsPage(groupName string) *GroupAppsPage {
	row := g.getRequiredGroupRowElement(groupName)
	manageAppsButton, err := row.Element("button.group-manage-apps-button")
	assert.Nil(g.Frame.TestingT(), err)
	g.Frame.DoAndWaitDOMContentLoaded(func() {
		manageAppsButton.MustClick()
	})
	g.Frame.AssertPagePath(tools.Paths.FrontendGroupApps)
	return g.Frame.GroupAppsPage
}

func (g *GroupsPage) ListGroups() []GroupRow {
	rows, err := g.Frame.Page().Elements("tr.group-row")
	assert.Nil(g.Frame.TestingT(), err)

	out := make([]GroupRow, 0, len(rows))
	for _, row := range rows {
		nameCell, err := row.Element(".group-name-cell")
		assert.Nil(g.Frame.TestingT(), err)
		name, err := nameCell.Text()
		assert.Nil(g.Frame.TestingT(), err)

		manageMembersButton, err := row.Element("button.group-manage-members-button")
		assert.Nil(g.Frame.TestingT(), err)
		assert.NotNil(g.Frame.TestingT(), manageMembersButton)

		manageAppsButton, err := row.Element("button.group-manage-apps-button")
		assert.Nil(g.Frame.TestingT(), err)
		assert.NotNil(g.Frame.TestingT(), manageAppsButton)

		deleteButton, err := row.Element("button.group-delete-button")
		assert.Nil(g.Frame.TestingT(), err)
		assert.NotNil(g.Frame.TestingT(), deleteButton)

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
	g.Frame.TestingT().Fatalf("group not found: %s", groupName)
	return nil
}

func (g *GroupsPage) AssertGroupAbsent(groupName string) *GroupsPage {
	groups := g.ListGroups()
	for _, group := range groups {
		assert.NotEqual(g.Frame.TestingT(), groupName, group.Name)
	}
	return g
}

func (g *GroupsPage) getRequiredGroupRowElement(groupName string) *rod.Element {
	rows, err := g.Frame.Page().Elements("tr.group-row")
	assert.Nil(g.Frame.TestingT(), err)

	for _, row := range rows {
		nameAttr, err := row.Attribute("data-group-name")
		assert.Nil(g.Frame.TestingT(), err)
		assert.NotNil(g.Frame.TestingT(), nameAttr)
		if strings.TrimSpace(*nameAttr) == groupName {
			return row
		}
	}
	g.Frame.TestingT().Fatalf("group row not found: %s", groupName)
	return nil
}
