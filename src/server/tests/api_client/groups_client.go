package api_client

import (
	"encoding/json"
	"server/groups"
	"server/tools"
	"strconv"

	u "github.com/quollix/common/utils"
)

type GroupsClient struct {
	quollix *QuollixClient
}

func (c *GroupsClient) ListAppsAccessByGroup(groupId int) (*groups.AppsAccessByGroup, error) {
	body, err := c.quollix.Parent.DoRequest(tools.Paths.BackendGroupsListAppsAccessByGroup, tools.NumberString{Value: strconv.Itoa(groupId)})
	if err != nil {
		return nil, err
	}
	var out groups.AppsAccessByGroup
	return &out, json.Unmarshal(body, &out)
}

func (c *GroupsClient) CreateGroup(name string) (*groups.Group, error) {
	_, err := c.quollix.Parent.DoRequest(tools.Paths.BackendGroupsCreate, tools.DefaultString{Value: name})
	if err != nil {
		return nil, err
	}

	allGroups, err := c.ListAllGroups()
	if err != nil {
		return nil, err
	}
	for _, group := range allGroups {
		if group.Name == name {
			return &group, nil
		}
	}
	return nil, u.Logger.NewError("group not found")
}

func (c *GroupsClient) DeleteGroup(groupId int) error {
	_, err := c.quollix.Parent.DoRequest(tools.Paths.BackendGroupsDelete, tools.NumberString{Value: strconv.Itoa(groupId)})
	return err
}

func (c *GroupsClient) ListAllGroups() ([]groups.Group, error) {
	body, err := c.quollix.Parent.DoRequest(tools.Paths.BackendListAllGroups, nil)
	if err != nil {
		return nil, err
	}
	var out []groups.Group
	return out, json.Unmarshal(body, &out)
}

func (c *GroupsClient) AddUserToGroup(userId, groupId int) error {
	_, err := c.quollix.Parent.DoRequest(tools.Paths.BackendGroupsAddUsers, groups.GroupIdAndUserIds{
		GroupId: strconv.Itoa(groupId),
		UserIds: []string{strconv.Itoa(userId)},
	})
	return err
}

func (c *GroupsClient) RemoveUserFromGroup(userId, groupId int) error {
	_, err := c.quollix.Parent.DoRequest(tools.Paths.BackendGroupsRemoveUsers, groups.GroupIdAndUserIds{
		GroupId: strconv.Itoa(groupId),
		UserIds: []string{strconv.Itoa(userId)},
	})
	return err
}

func (c *GroupsClient) ListUsersByGroupMembership(groupId int) (*groups.UsersByGroupMembership, error) {
	body, err := c.quollix.Parent.DoRequest(tools.Paths.BackendGroupsListUsersByMembership, tools.NumberString{Value: strconv.Itoa(groupId)})
	if err != nil {
		return nil, err
	}
	var out groups.UsersByGroupMembership
	return &out, json.Unmarshal(body, &out)
}

func (c *GroupsClient) GrantAppAccess(groupId int, appName string) error {
	_, err := c.quollix.Parent.DoRequest(tools.Paths.BackendGroupsGrantGroupAccessToApps, groups.GroupIdAndAppNames{
		GroupId:  strconv.Itoa(groupId),
		AppNames: []string{appName},
	})
	return err
}

func (c *GroupsClient) RevokeAppAccess(groupId int, appName string) error {
	_, err := c.quollix.Parent.DoRequest(tools.Paths.BackendGroupsRevokeGroupAccessToApps, groups.GroupIdAndAppNames{
		GroupId:  strconv.Itoa(groupId),
		AppNames: []string{appName},
	})
	return err
}
