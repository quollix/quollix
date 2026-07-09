//go:build integration

package repository

import (
	"server/apps_basic"
	"server/groups"
	"slices"
	"testing"

	"github.com/quollix/common/assert"
)

const (
	tableMemberships = "memberships"
	tableAppAccess   = "app_access"

	sampleGroup = "sample-group"
	sampleApp   = "sampleapp"
)

var (
	isGroupRepoInitialized = false
	groupsRepo             *groups.GroupRepositoryImpl
)

func InitGroupDeps() {
	if isGroupRepoInitialized {
		return
	}
	InitDeps()
	groupsRepo = &groups.GroupRepositoryImpl{DbConnector: DatabaseConnector}
	isGroupRepoInitialized = true
}

func wipeAll() {
	AppRepo.Wipe()
	UserRepo.Wipe()
	groupsRepo.Wipe()
}

func createSampleUser(t *testing.T) (int, string) {
	adminUser := GetSampleAdminUser()
	var err error
	adminUser.Id, err = UserRepo.CreateUser(adminUser)
	assert.Nil(t, err)
	assert.NotEqual(t, 0, adminUser.Id)
	return adminUser.Id, adminUser.Username
}

func setupUserGroupWithAccess(t *testing.T, groupName, appName string) (int, int) {
	uid, _ := createSampleUser(t)
	gid, err := groupsRepo.CreateGroup(groupName)
	assert.Nil(t, err)

	assert.Nil(t, groupsRepo.AddUsersToGroup(gid, []int{uid}))
	assert.Nil(t, groupsRepo.GrantAppAccess(gid, []string{appName}))

	return uid, gid
}

func TestGroupCRUD(t *testing.T) {
	InitGroupDeps()

	groups, err := groupsRepo.ListAllGroups()
	assert.Nil(t, err)
	assert.Equal(t, 0, len(groups))

	id, err := groupsRepo.CreateGroup(sampleGroup)
	assert.Nil(t, err)

	groups, err = groupsRepo.ListAllGroups()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(groups))
	assert.Equal(t, id, groups[0].Id)
	assert.Equal(t, sampleGroup, groups[0].Name)

	err = groupsRepo.DeleteGroup(id)
	assert.Nil(t, err)

	groups, err = groupsRepo.ListAllGroups()
	assert.Nil(t, err)
	assert.Equal(t, 0, len(groups))
}

func TestCreateGroupDuplicate(t *testing.T) {
	InitGroupDeps()
	defer wipeAll()

	_, err := groupsRepo.CreateGroup(sampleGroup)
	assert.Nil(t, err)

	_, err = groupsRepo.CreateGroup(sampleGroup)
	assert.NotNil(t, err)
}

func TestDoesGroupExist(t *testing.T) {
	InitGroupDeps()
	defer wipeAll()

	exists, err := groupsRepo.DoesGroupExist(sampleGroup)
	assert.Nil(t, err)
	assert.False(t, exists)

	_, err = groupsRepo.CreateGroup(sampleGroup)
	assert.Nil(t, err)

	exists, err = groupsRepo.DoesGroupExist(sampleGroup)
	assert.Nil(t, err)
	assert.True(t, exists)
}

func TestMemberCRUD(t *testing.T) {
	InitGroupDeps()
	defer wipeAll()

	uid, uname := createSampleUser(t)
	gid, err := groupsRepo.CreateGroup(sampleGroup)
	assert.Nil(t, err)

	assert.Nil(t, groupsRepo.AddUsersToGroup(gid, []int{uid}))

	usersByGroup, err := groupsRepo.ListUsersByGroupMembership(gid)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(usersByGroup.In))
	assert.Equal(t, uid, usersByGroup.In[0].Id)
	assert.Equal(t, uname, usersByGroup.In[0].Name)

	assert.Nil(t, groupsRepo.RemoveUsersFromGroup(gid, []int{uid}))

	usersByGroup, err = groupsRepo.ListUsersByGroupMembership(gid)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(usersByGroup.In))
}

func TestAddUserToGroupDuplicateShouldDoNothing(t *testing.T) {
	InitGroupDeps()
	defer wipeAll()

	uid, _ := createSampleUser(t)
	gid, err := groupsRepo.CreateGroup(sampleGroup)
	assert.Nil(t, err)

	assert.Nil(t, groupsRepo.AddUsersToGroup(gid, []int{uid}))
	assert.Nil(t, groupsRepo.AddUsersToGroup(gid, []int{uid}))

	usersByGroup, err := groupsRepo.ListUsersByGroupMembership(gid)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(usersByGroup.In))
	assert.Equal(t, uid, usersByGroup.In[0].Id)
}

func TestAppAccessCRUD(t *testing.T) {
	InitGroupDeps()
	defer wipeAll()

	_, err := AppRepo.CreateApp(apps_basic.GetSampleApp())
	assert.Nil(t, err)

	gid, err := groupsRepo.CreateGroup(sampleGroup)
	assert.Nil(t, err)

	appsAccessByGroup, err := groupsRepo.ListAppsAccessByGroup(gid)
	assert.Nil(t, err)
	assert.False(t, slices.Contains(appsAccessByGroup.Granted, sampleApp))
	assert.True(t, slices.Contains(appsAccessByGroup.NotGranted, sampleApp))

	assert.Nil(t, groupsRepo.GrantAppAccess(gid, []string{sampleApp}))

	appsAccessByGroup, err = groupsRepo.ListAppsAccessByGroup(gid)
	assert.Nil(t, err)
	assert.True(t, slices.Contains(appsAccessByGroup.Granted, sampleApp))
	assert.False(t, slices.Contains(appsAccessByGroup.NotGranted, sampleApp))
}

func TestGrantAppAccessDuplicateShouldDoNothing(t *testing.T) {
	InitGroupDeps()
	defer wipeAll()

	_, err := AppRepo.CreateApp(apps_basic.GetSampleApp())
	assert.Nil(t, err)

	gid, err := groupsRepo.CreateGroup(sampleGroup)
	assert.Nil(t, err)

	assert.Nil(t, groupsRepo.GrantAppAccess(gid, []string{sampleApp}))
	assert.Nil(t, groupsRepo.GrantAppAccess(gid, []string{sampleApp}))

	appsAccessByGroup, err := groupsRepo.ListAppsAccessByGroup(gid)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(appsAccessByGroup.Granted))
	assert.Equal(t, sampleApp, appsAccessByGroup.Granted[0])
}

func TestRemoveUserFromGroupRevokesAppAccess(t *testing.T) {
	InitGroupDeps()
	defer wipeAll()

	uid, gid := setupUserGroupWithAccess(t, sampleGroup, sampleApp)

	has, err := groupsRepo.HasAccess(uid, sampleApp)
	assert.Nil(t, err)
	assert.True(t, has)

	assert.Nil(t, groupsRepo.RemoveUsersFromGroup(gid, []int{uid}))

	has, err = groupsRepo.HasAccess(uid, sampleApp)
	assert.Nil(t, err)
	assert.False(t, has)
}

func TestRemovingAppFromGroupRevokesAppAccess(t *testing.T) {
	InitGroupDeps()
	defer wipeAll()

	uid, gid := setupUserGroupWithAccess(t, sampleGroup, sampleApp)

	has, err := groupsRepo.HasAccess(uid, sampleApp)
	assert.Nil(t, err)
	assert.True(t, has)

	assert.Nil(t, groupsRepo.RevokeAppAccess(gid, []string{sampleApp}))

	has, err = groupsRepo.HasAccess(uid, sampleApp)
	assert.Nil(t, err)
	assert.False(t, has)
}

func TestAppNotExistingIsTolerated(t *testing.T) {
	InitGroupDeps()
	defer wipeAll()
	uid, _ := createSampleUser(t)
	has, err := groupsRepo.HasAccess(uid, sampleApp)
	assert.Nil(t, err)
	assert.False(t, has)
}

func TestCascadeDeleteGroup(t *testing.T) {
	InitGroupDeps()
	defer wipeAll()

	uid, _ := createSampleUser(t)
	gid, err := groupsRepo.CreateGroup(sampleGroup)
	assert.Nil(t, err)

	assert.Nil(t, groupsRepo.AddUsersToGroup(gid, []int{uid}))
	assert.Nil(t, groupsRepo.GrantAppAccess(gid, []string{sampleApp}))

	assert.Nil(t, groupsRepo.DeleteGroup(gid))

	groupsRepo.AssertTableEmpty(t, tableMemberships)
	groupsRepo.AssertTableEmpty(t, tableAppAccess)
}

func TestCascadeDeleteUserRemovesMemberships(t *testing.T) {
	InitGroupDeps()
	defer wipeAll()

	uid, _ := createSampleUser(t)
	gid, err := groupsRepo.CreateGroup(sampleGroup)
	assert.Nil(t, err)

	assert.Nil(t, groupsRepo.AddUsersToGroup(gid, []int{uid}))
	assert.Nil(t, groupsRepo.GrantAppAccess(gid, []string{sampleApp}))
	assert.Nil(t, UserRepo.DeleteUser(uid))

	groupsRepo.AssertTableEmpty(t, tableMemberships)
}

func TestListGroupsOfUser(t *testing.T) {
	InitGroupDeps()
	defer wipeAll()

	uid, _ := createSampleUser(t)
	gid, err := groupsRepo.CreateGroup(sampleGroup)
	assert.Nil(t, err)

	groups, err := groupsRepo.ListGroupsForUser(uid)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(groups))

	assert.Nil(t, groupsRepo.AddUsersToGroup(gid, []int{uid}))
	groups, err = groupsRepo.ListGroupsForUser(uid)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(groups))
	assert.Equal(t, gid, groups[0].Id)
	assert.Equal(t, sampleGroup, groups[0].Name)

	assert.Nil(t, groupsRepo.DeleteGroup(gid))
	groups, err = groupsRepo.ListGroupsForUser(uid)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(groups))
}

func TestListUsersByGroupMembership(t *testing.T) {
	InitGroupDeps()
	defer wipeAll()

	uid, uname := createSampleUser(t)
	gid, err := groupsRepo.CreateGroup(sampleGroup)
	assert.Nil(t, err)

	usersByGroup, err := groupsRepo.ListUsersByGroupMembership(gid)
	assert.Nil(t, err)
	assert.True(t, containsMember(usersByGroup.NotIn, uname))
	assert.False(t, containsMember(usersByGroup.In, uname))

	assert.Nil(t, groupsRepo.AddUsersToGroup(gid, []int{uid}))

	usersByGroup, err = groupsRepo.ListUsersByGroupMembership(gid)
	assert.Nil(t, err)
	assert.True(t, containsMember(usersByGroup.In, uname))
	assert.False(t, containsMember(usersByGroup.NotIn, uname))
}

func containsMember(members []groups.Member, name string) bool {
	for _, member := range members {
		if member.Name == name {
			return true
		}
	}
	return false
}

func TestListAppsAccessByGroup(t *testing.T) {
	InitGroupDeps()
	defer wipeAll()

	_, err := AppRepo.CreateApp(apps_basic.GetSampleApp())
	assert.Nil(t, err)

	gid, err := groupsRepo.CreateGroup(sampleGroup)
	assert.Nil(t, err)

	appsAccessByGroup, err := groupsRepo.ListAppsAccessByGroup(gid)
	assert.Nil(t, err)
	assert.True(t, slices.Contains(appsAccessByGroup.NotGranted, sampleApp))
	assert.False(t, slices.Contains(appsAccessByGroup.Granted, sampleApp))

	assert.Nil(t, groupsRepo.GrantAppAccess(gid, []string{sampleApp}))

	appsAccessByGroup, err = groupsRepo.ListAppsAccessByGroup(gid)
	assert.Nil(t, err)
	assert.True(t, slices.Contains(appsAccessByGroup.Granted, sampleApp))
	assert.False(t, slices.Contains(appsAccessByGroup.NotGranted, sampleApp))
}

func TestGetGroupById(t *testing.T) {
	InitGroupDeps()
	defer wipeAll()

	gid, err := groupsRepo.CreateGroup(sampleGroup)
	assert.Nil(t, err)

	group, err := groupsRepo.GetGroupById(gid)
	assert.Nil(t, err)
	assert.Equal(t, gid, group.Id)
	assert.Equal(t, sampleGroup, group.Name)
}
