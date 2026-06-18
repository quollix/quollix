package groups

import (
	"net/http"
	"server/tools"
	"strconv"

	u "github.com/quollix/common/utils"
	"github.com/quollix/common/validation"
)

type GroupHandler struct {
	Repo GroupRepository
}

var expectedGroupCreationErrors = u.MapOf(GroupAlreadyExistsError)

func (g *GroupHandler) CreateGroupHandler(w http.ResponseWriter, r *http.Request) {
	groupString, ok := validation.ReadBody[tools.DefaultString](w, r)
	if !ok {
		return
	}

	exists, err := g.Repo.DoesGroupExist(groupString.Value)
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}
	if exists {
		u.WriteResponseError(w, expectedGroupCreationErrors, u.Logger.NewError(GroupAlreadyExistsError))
		return
	}

	_, err = g.Repo.CreateGroup(groupString.Value)
	if err != nil {
		u.WriteResponseError(w, nil, err)
	}
}

func (g *GroupHandler) DeleteGroupHandler(w http.ResponseWriter, r *http.Request) {
	groupString, ok := validation.ReadBody[tools.NumberString](w, r)
	if !ok {
		return
	}
	groupId, err := strconv.Atoi(groupString.Value)
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}
	if err = g.Repo.DeleteGroup(groupId); err != nil {
		u.WriteResponseError(w, nil, err)
	}
}

type GroupIdAndUserIds struct {
	GroupId string   `json:"group_id" validate:"number"`
	UserIds []string `json:"user_ids" validate:"number"`
}

func (g *GroupHandler) AddUserToGroupHandler(w http.ResponseWriter, r *http.Request) {
	groupIdAndUserIds, ok := validation.ReadBody[GroupIdAndUserIds](w, r)
	if !ok {
		return
	}

	groupId, err := strconv.Atoi(groupIdAndUserIds.GroupId)
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}

	var userIds []int
	for _, userIdString := range groupIdAndUserIds.UserIds {
		userId, err := strconv.Atoi(userIdString)
		if err != nil {
			u.WriteResponseError(w, nil, err)
			return
		}
		userIds = append(userIds, userId)
	}

	if err := g.Repo.AddUsersToGroup(groupId, userIds); err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}
}

func (g *GroupHandler) RemoveUserFromGroupHandler(w http.ResponseWriter, r *http.Request) {
	groupIdAndUserIds, ok := validation.ReadBody[GroupIdAndUserIds](w, r)
	if !ok {
		return
	}

	groupId, err := strconv.Atoi(groupIdAndUserIds.GroupId)
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}

	userIds := make([]int, 0, len(groupIdAndUserIds.UserIds))
	for _, userIdString := range groupIdAndUserIds.UserIds {
		userId, err := strconv.Atoi(userIdString)
		if err != nil {
			u.WriteResponseError(w, nil, err)
			return
		}
		userIds = append(userIds, userId)
	}

	if err := g.Repo.RemoveUsersFromGroup(groupId, userIds); err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}
}

type GroupIdAndAppNames struct {
	GroupId  string   `json:"group_id" validate:"number"`
	AppNames []string `json:"app_names" validate:"default"`
}

func (g *GroupHandler) GrantAppAccessHandler(w http.ResponseWriter, r *http.Request) {
	appNameAndGroupId, ok := validation.ReadBody[GroupIdAndAppNames](w, r)
	if !ok {
		return
	}
	groupId, err := strconv.Atoi(appNameAndGroupId.GroupId)
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}
	if err = g.Repo.GrantAppAccess(groupId, appNameAndGroupId.AppNames); err != nil {
		u.WriteResponseError(w, nil, err)
	}
}

func (g *GroupHandler) RevokeAppAccessHandler(w http.ResponseWriter, r *http.Request) {
	GroupIdAndAppNames, ok := validation.ReadBody[GroupIdAndAppNames](w, r)
	if !ok {
		return
	}
	groupId, err := strconv.Atoi(GroupIdAndAppNames.GroupId)
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}
	if err = g.Repo.RevokeAppAccess(groupId, GroupIdAndAppNames.AppNames); err != nil {
		u.WriteResponseError(w, nil, err)
	}
}

func (g *GroupHandler) ListUsersByGroupMembership(w http.ResponseWriter, r *http.Request) {
	groupString, ok := validation.ReadBody[tools.NumberString](w, r)
	if !ok {
		return
	}
	groupId, err := strconv.Atoi(groupString.Value)
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}
	usersByGroup, err := g.Repo.ListUsersByGroupMembership(groupId)
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}
	u.SendJsonResponse(w, usersByGroup)
}

func (g *GroupHandler) ListAppsAccessByGroup(w http.ResponseWriter, r *http.Request) {
	groupString, ok := validation.ReadBody[tools.NumberString](w, r)
	if !ok {
		return
	}
	groupId, err := strconv.Atoi(groupString.Value)
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}
	appsAccess, err := g.Repo.ListAppsAccessByGroup(groupId)
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}

	appsAccess.Granted = filterApps(appsAccess.Granted)
	appsAccess.NotGranted = filterApps(appsAccess.NotGranted)

	u.SendJsonResponse(w, appsAccess)
}

func filterApps(in []string) []string {
	out := in[:0]
	for _, appName := range in {
		if appName != u.OfficialDatabaseAppName {
			out = append(out, appName)
		}
	}
	return out
}

func (g *GroupHandler) ListAllGroupsHandler(w http.ResponseWriter, r *http.Request) {
	groups, err := g.Repo.ListAllGroups()
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}
	u.SendJsonResponse(w, groups)
}
