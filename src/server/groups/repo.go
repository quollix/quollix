package groups

import (
	"server/tools"
	"testing"

	"github.com/lib/pq"
	u "github.com/quollix/common/utils"
)

type Group struct {
	Id   int
	Name string `validate:"default"`
}

type Member struct {
	Id   int
	Name string `validate:"default"`
}

type UsersByGroupMembership struct {
	In    []Member
	NotIn []Member
}

type AppsAccessByGroup struct {
	Granted    []string
	NotGranted []string
}

const GroupAlreadyExistsError = "group already exists"

type GroupRepository interface {
	DoesGroupExist(name string) (bool, error)
	CreateGroup(name string) (int, error)
	DeleteGroup(groupId int) error

	GetGroupById(groupId int) (Group, error)
	ListAllGroups() ([]Group, error)
	ListGroupsForUser(userId int) ([]Group, error)

	AddUsersToGroup(groupId int, userIds []int) error
	RemoveUsersFromGroup(groupId int, userIds []int) error
	ListUsersByGroupMembership(groupId int) (*UsersByGroupMembership, error)

	GrantAppAccess(groupId int, appNames []string) error
	RevokeAppAccess(groupId int, appNames []string) error
	ListAppsAccessByGroup(groupId int) (*AppsAccessByGroup, error)

	HasAccess(userId int, appName string) (bool, error)
}

type GroupRepositoryImpl struct {
	DbConnector tools.DatabaseConnector
}

func (r *GroupRepositoryImpl) DoesGroupExist(name string) (bool, error) {
	var exists bool
	err := r.DbConnector.GetDB().QueryRow(
		`SELECT EXISTS (SELECT 1 FROM groups WHERE group_name = $1)`,
		name,
	).Scan(&exists)
	if err != nil {
		return false, u.Logger.NewError(err.Error())
	}
	return exists, nil
}

func (r *GroupRepositoryImpl) ListGroupsForUser(userId int) ([]Group, error) {
	rows, err := r.DbConnector.GetDB().Query(
		`SELECT g.group_id, g.group_name
         FROM memberships m
         JOIN groups g ON g.group_id = m.group_id
         WHERE m.user_id = $1
         ORDER BY g.group_name`,
		userId,
	)
	if err != nil {
		return nil, u.Logger.NewError(err.Error())
	}
	defer u.Close(rows)

	var groups []Group
	for rows.Next() {
		var g Group
		if err := rows.Scan(&g.Id, &g.Name); err != nil {
			return nil, u.Logger.NewError(err.Error())
		}
		groups = append(groups, g)
	}
	if rows.Err() != nil {
		return nil, u.Logger.NewError(rows.Err().Error())
	}
	return groups, nil
}

func (r *GroupRepositoryImpl) CreateGroup(name string) (int, error) {
	var id int
	err := r.DbConnector.GetDB().QueryRow(
		`INSERT INTO groups (group_name) VALUES ($1) RETURNING group_id`,
		name,
	).Scan(&id)
	if err != nil {
		return -1, u.Logger.NewError(err.Error())
	}
	return id, nil
}

func (r *GroupRepositoryImpl) DeleteGroup(groupId int) error {
	_, err := r.DbConnector.GetDB().Exec(
		`DELETE FROM groups WHERE group_id = $1`,
		groupId,
	)
	if err != nil {
		return u.Logger.NewError(err.Error())
	}
	return nil
}

func (r *GroupRepositoryImpl) ListAllGroups() ([]Group, error) {
	rows, err := r.DbConnector.GetDB().Query(
		`SELECT group_id, group_name FROM groups ORDER BY group_name`,
	)
	if err != nil {
		return nil, u.Logger.NewError(err.Error())
	}
	defer u.Close(rows)

	var groups []Group
	for rows.Next() {
		var g Group
		if err := rows.Scan(&g.Id, &g.Name); err != nil {
			return nil, u.Logger.NewError(err.Error())
		}
		groups = append(groups, g)
	}
	if rows.Err() != nil {
		return nil, u.Logger.NewError(rows.Err().Error())
	}
	return groups, nil
}

func (r *GroupRepositoryImpl) AddUsersToGroup(groupId int, userIds []int) error {
	if len(userIds) == 0 {
		return nil
	}

	_, err := r.DbConnector.GetDB().Exec(
		`INSERT INTO memberships (group_id, user_id)
         SELECT $1, user_id
         FROM UNNEST($2::int[]) AS user_id
         ON CONFLICT DO NOTHING`,
		groupId,
		pq.Array(userIds),
	)
	if err != nil {
		return u.Logger.NewError(err.Error())
	}
	return nil
}

func (r *GroupRepositoryImpl) RemoveUsersFromGroup(groupId int, userIds []int) error {
	if len(userIds) == 0 {
		return nil
	}

	_, err := r.DbConnector.GetDB().Exec(
		`DELETE FROM memberships
         WHERE group_id = $1
           AND user_id = ANY($2::int[])`,
		groupId,
		pq.Array(userIds),
	)
	if err != nil {
		return u.Logger.NewError(err.Error())
	}
	return nil
}

func (r *GroupRepositoryImpl) GrantAppAccess(groupId int, appNames []string) error {
	if len(appNames) == 0 {
		return nil
	}

	_, err := r.DbConnector.GetDB().Exec(
		`INSERT INTO app_access (group_id, app_name)
         SELECT $1, app_name
         FROM UNNEST($2::text[]) AS app_name
         ON CONFLICT DO NOTHING`,
		groupId,
		pq.Array(appNames),
	)
	if err != nil {
		return u.Logger.NewError(err.Error())
	}
	return nil
}

func (r *GroupRepositoryImpl) RevokeAppAccess(groupId int, appNames []string) error {
	if len(appNames) == 0 {
		return nil
	}

	_, err := r.DbConnector.GetDB().Exec(
		`DELETE FROM app_access
         WHERE group_id = $1
           AND app_name = ANY($2::text[])`,
		groupId,
		pq.Array(appNames),
	)
	if err != nil {
		return u.Logger.NewError(err.Error())
	}
	return nil
}

func (r *GroupRepositoryImpl) HasAccess(userId int, appName string) (bool, error) {
	var exists bool
	err := r.DbConnector.GetDB().QueryRow(
		`SELECT EXISTS (
            SELECT 1
            FROM memberships m
            JOIN app_access a ON a.group_id = m.group_id
            WHERE m.user_id = $1
              AND a.app_name = $2
        )`,
		userId, appName,
	).Scan(&exists)
	if err != nil {
		return false, u.Logger.NewError(err.Error())
	}
	return exists, nil
}

// only used for testing
func (r *GroupRepositoryImpl) Wipe() {
	_, err := r.DbConnector.GetDB().Exec(`
        DELETE FROM app_access;
        DELETE FROM memberships;
        DELETE FROM groups;
    `)
	if err != nil {
		u.Logger.Error(err)
	}
}

func (r *GroupRepositoryImpl) AssertTableEmpty(t *testing.T, table string) {
	var count int
	err := r.DbConnector.GetDB().QueryRow(`SELECT COUNT(*) FROM ` + table).Scan(&count)
	if err != nil {
		t.Fatalf("counting rows in %s failed: %v", table, err)
	}
	if count != 0 {
		t.Fatalf("expected table %s to be empty, but found %d rows", table, count)
	}
}

func (r *GroupRepositoryImpl) ListUsersByGroupMembership(groupId int) (*UsersByGroupMembership, error) {
	rows, err := r.DbConnector.GetDB().Query(
		`SELECT u.user_id, u.username,
		        EXISTS (
		          SELECT 1
		          FROM memberships m
		          WHERE m.group_id = $1 AND m.user_id = u.user_id
		        ) AS in_group
		 FROM users u
		 ORDER BY u.username`,
		groupId,
	)
	if err != nil {
		return nil, u.Logger.NewError(err.Error())
	}
	defer u.Close(rows)

	res := &UsersByGroupMembership{}
	for rows.Next() {
		var m Member
		var inGroup bool
		if err := rows.Scan(&m.Id, &m.Name, &inGroup); err != nil {
			return nil, u.Logger.NewError(err.Error())
		}
		if inGroup {
			res.In = append(res.In, m)
		} else {
			res.NotIn = append(res.NotIn, m)
		}
	}
	if rows.Err() != nil {
		return nil, u.Logger.NewError(rows.Err().Error())
	}
	return res, nil
}

func (r *GroupRepositoryImpl) ListAppsAccessByGroup(groupId int) (*AppsAccessByGroup, error) {
	rows, err := r.DbConnector.GetDB().Query(
		`SELECT a.app_name,
		        EXISTS (
		          SELECT 1
		          FROM app_access aa
		          WHERE aa.group_id = $1
		            AND aa.app_name = a.app_name
		        ) AS granted
		 FROM apps a
		 ORDER BY a.app_name`,
		groupId,
	)
	if err != nil {
		return nil, u.Logger.NewError(err.Error())
	}
	defer u.Close(rows)

	res := &AppsAccessByGroup{}
	for rows.Next() {
		var name string
		var granted bool
		if err := rows.Scan(&name, &granted); err != nil {
			return nil, u.Logger.NewError(err.Error())
		}
		if granted {
			res.Granted = append(res.Granted, name)
		} else {
			res.NotGranted = append(res.NotGranted, name)
		}
	}
	if rows.Err() != nil {
		return nil, u.Logger.NewError(rows.Err().Error())
	}
	return res, nil
}

func (r *GroupRepositoryImpl) GetGroupById(groupId int) (Group, error) {
	var group Group
	err := r.DbConnector.GetDB().QueryRow(
		`SELECT group_id, group_name FROM groups WHERE group_id = $1`,
		groupId,
	).Scan(&group.Id, &group.Name)
	if err != nil {
		return Group{}, u.Logger.NewError(err.Error())
	}
	return group, nil
}
