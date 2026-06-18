//go:build acceptance

package acceptance

import (
	"server/tests/acceptance/pages"
	"server/tools"
	"testing"

	"github.com/quollix/common/assert"
)

const sampleGroupName = "samplegroup"

func TestGroupsPageCreateAndDeleteGroup(t *testing.T) {
	frame := pages.Setup(t)
	defer frame.Client.Test.ResetTestState()

	groupsPage := frame.OpenGroupsPage()
	assert.Equal(t, 0, len(groupsPage.ListGroups()))

	groupsPage.CreateGroup(sampleGroupName)
	groups := groupsPage.ListGroups()
	assert.Equal(t, 1, len(groups))
	group := groupsPage.GetRequiredGroup(sampleGroupName)
	assert.Equal(t, sampleGroupName, group.Name)

	groupsPage.DeleteGroup(sampleGroupName)
	groupsPage.AssertGroupAbsent(sampleGroupName)
	assert.Equal(t, 0, len(groupsPage.ListGroups()))
}

func TestGroupMembersPageMembershipFlow(t *testing.T) {
	frame := pages.Setup(t)
	defer frame.Client.Test.ResetTestState()

	groupsPage := frame.OpenGroupsPage()
	groupsPage.CreateGroup(sampleGroupName)

	groupMembersPage := groupsPage.OpenGroupMembersPage(sampleGroupName)
	groupMembersPage.ClickBack()

	groupMembersPage = frame.OpenGroupsPage().OpenGroupMembersPage(sampleGroupName)

	groupMembersPage.
		AssertUserInNonMembers(tools.DefaultAdminName).
		AssertUserNotInMembers(tools.DefaultAdminName).
		AssertNonMembersSelectAllChecked(false).
		AssertNonMemberChecked(tools.DefaultAdminName, false)

	groupMembersPage.
		SetNonMembersFilter("does-not-exist").
		AssertNonMemberRowVisible(tools.DefaultAdminName, false).
		SetNonMembersFilter(tools.DefaultAdminName).
		AssertNonMemberRowVisible(tools.DefaultAdminName, true).
		SetNonMembersFilter("")

	groupMembersPage.
		SetNonMembersSelectAll(true).
		AssertNonMemberChecked(tools.DefaultAdminName, true).
		SetNonMembersSelectAll(false).
		AssertNonMemberChecked(tools.DefaultAdminName, false)

	groupMembersPage.
		SetNonMemberChecked(tools.DefaultAdminName, true).
		ClickAddSelected().
		AssertUserNotInNonMembers(tools.DefaultAdminName).
		AssertUserInMembers(tools.DefaultAdminName).
		AssertMembersSelectAllChecked(false).
		AssertMemberChecked(tools.DefaultAdminName, false)

	groupMembersPage.
		SetMembersSelectAll(true).
		AssertMemberChecked(tools.DefaultAdminName, true).
		SetMembersSelectAll(false).
		AssertMemberChecked(tools.DefaultAdminName, false)

	groupMembersPage.
		SetMemberChecked(tools.DefaultAdminName, true).
		ClickRemoveSelected().
		AssertUserInNonMembers(tools.DefaultAdminName).
		AssertUserNotInMembers(tools.DefaultAdminName)
}

func TestGroupAppsPageAccessFlow(t *testing.T) {
	frame := pages.Setup(t)
	defer frame.Client.Test.ResetTestState()

	_, err := frame.Client.Apps.InstallSample("2.0")
	assert.Nil(t, err)

	groupsPage := frame.OpenGroupsPage()
	groupsPage.CreateGroup(sampleGroupName)

	groupAppsPage := groupsPage.OpenGroupAppsPage(sampleGroupName)
	groupAppsPage.ClickBack()

	groupAppsPage = frame.OpenGroupsPage().OpenGroupAppsPage(sampleGroupName)

	groupAppsPage.
		AssertAppInNoAccess(tools.SampleApp).
		AssertAppNotInGranted(tools.SampleApp).
		AssertNoAccessSelectAllChecked(false).
		AssertNoAccessChecked(tools.SampleApp, false)

	groupAppsPage.
		SetNoAccessFilter("does-not-exist").
		AssertNoAccessRowVisible(tools.SampleApp, false).
		SetNoAccessFilter(tools.SampleApp).
		AssertNoAccessRowVisible(tools.SampleApp, true).
		SetNoAccessFilter("")

	groupAppsPage.
		SetNoAccessSelectAll(true).
		AssertNoAccessChecked(tools.SampleApp, true).
		SetNoAccessSelectAll(false).
		AssertNoAccessChecked(tools.SampleApp, false)

	groupAppsPage.
		SetNoAccessChecked(tools.SampleApp, true).
		ClickAddSelected().
		AssertAppNotInNoAccess(tools.SampleApp).
		AssertAppInGranted(tools.SampleApp).
		AssertGrantedSelectAllChecked(false).
		AssertGrantedChecked(tools.SampleApp, false)

	groupAppsPage.
		SetGrantedSelectAll(true).
		AssertGrantedChecked(tools.SampleApp, true).
		SetGrantedSelectAll(false).
		AssertGrantedChecked(tools.SampleApp, false)

	groupAppsPage.
		SetGrantedChecked(tools.SampleApp, true).
		ClickRemoveSelected().
		AssertAppInNoAccess(tools.SampleApp).
		AssertAppNotInGranted(tools.SampleApp)
}
