package users

import (
	"fmt"
	"server/tools"

	u "github.com/quollix/common/utils"
)

var (
	userColumnsNoId = `
       username,
       email,
       hashed_password,
       is_admin,
       is_enabled,
       set_password_token,
       set_password_token_expiration_date,
       creation_date
	`

	userSelect = fmt.Sprintf(`SELECT users.user_id, %s FROM users`, userColumnsNoId)
	userInsert = fmt.Sprintf(`INSERT INTO users (%s) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING user_id`, userColumnsNoId)
	userUpdate = fmt.Sprintf(`UPDATE users SET (%s) = ($2, $3, $4, $5, $6, $7, $8, $9) WHERE user_id = $1`, userColumnsNoId)
)

type UserRepository interface {
	CreateUser(info *tools.User) (int, error)

	DeleteUser(userId int) error
	DoesAnyAdminUserExist() (bool, error)
	ListUsers() ([]tools.User, error)
	DoesUserExist(user string) (bool, error)
	GetHighestGeneratedUsernameSuffix(username string, maxUsernameLength int) (int, bool, error)

	GetUserById(userId int) (*tools.User, error)
	GetUserByUsername(username string) (*tools.User, error)
	UpdateUser(info *tools.User) error
	GetUserByToken(token string) (*tools.User, error)
	DoesEmailExist(email string) (bool, error)
}

type UserRepositoryImpl struct {
	DbProvider tools.DatabaseConnector
}

func (r *UserRepositoryImpl) DoesEmailExist(email string) (bool, error) {
	var exists bool
	err := r.DbProvider.GetDB().QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)", email).Scan(&exists)
	if err != nil {
		return false, u.Logger.NewError(err.Error())
	}
	return exists, nil
}

func (r *UserRepositoryImpl) searchUserBy(field string, value any) (*tools.User, error) {
	users, err := r.queryUsers(fmt.Sprintf("WHERE %s = $1", field), value)
	if err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return nil, u.Logger.NewError(UserNotFoundError)
	}
	return &users[0], nil
}

func (r *UserRepositoryImpl) queryUsers(where string, args ...any) ([]tools.User, error) {
	query := userSelect
	if where != "" {
		query += " " + where
	}

	return r.scanUsers(query, args...)
}

func (r *UserRepositoryImpl) scanUsers(query string, args ...any) ([]tools.User, error) {
	rows, err := r.DbProvider.GetDB().Query(query, args...)
	if err != nil {
		return nil, u.Logger.NewError(err.Error())
	}
	defer u.Close(rows)

	var users []tools.User
	for rows.Next() {
		var user tools.User
		if err = rows.Scan(
			&user.Id,
			&user.Username,
			&user.Email,
			&user.HashedPassword,
			&user.IsAdmin,
			&user.IsEnabled,
			&user.SetPasswordToken,
			&user.SetPasswordTokenExpirationDate,
			&user.CreationDate,
		); err != nil {
			return nil, u.Logger.NewError(err.Error())
		}
		normalizeUserTimes(&user)
		users = append(users, user)
	}
	return users, nil
}

func (r *UserRepositoryImpl) GetUserById(userId int) (*tools.User, error) {
	return r.searchUserBy("user_id", userId)
}

func (r *UserRepositoryImpl) GetUserByUsername(username string) (*tools.User, error) {
	return r.searchUserBy("username", username)
}

func (r *UserRepositoryImpl) UpdateUser(info *tools.User) error {
	_, err := r.DbProvider.GetDB().Exec(
		userUpdate,
		info.Id,
		info.Username,
		info.Email,
		info.HashedPassword,
		info.IsAdmin,
		info.IsEnabled,
		info.SetPasswordToken,
		info.SetPasswordTokenExpirationDate,
		info.CreationDate,
	)
	if err != nil {
		return u.Logger.NewError(err.Error())
	}
	return nil
}

func (r *UserRepositoryImpl) DoesAnyAdminUserExist() (bool, error) {
	var exists bool
	err := r.DbProvider.GetDB().QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE is_admin = $1)", true).Scan(&exists)
	if err != nil {
		return false, u.Logger.NewError(err.Error())
	}
	return exists, nil
}

func (r *UserRepositoryImpl) CreateUser(user *tools.User) (int, error) {
	var id int
	err := r.DbProvider.GetDB().QueryRow(userInsert, userInsertArgs(user)...).Scan(&id)
	if err != nil {
		return 0, u.Logger.NewError(err.Error())
	}
	return id, nil
}

func userInsertArgs(user *tools.User) []any {
	return []any{
		user.Username,
		user.Email,
		user.HashedPassword,
		user.IsAdmin,
		user.IsEnabled,
		user.SetPasswordToken,
		user.SetPasswordTokenExpirationDate,
		user.CreationDate,
	}
}

func (r *UserRepositoryImpl) DeleteUser(userId int) error {
	_, err := r.DbProvider.GetDB().Exec("DELETE FROM users WHERE user_id = $1", userId)
	if err != nil {
		return u.Logger.NewError(err.Error())
	}
	return nil
}

func (r *UserRepositoryImpl) DoesUserExist(user string) (bool, error) {
	var exists bool
	err := r.DbProvider.GetDB().QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)", user).Scan(&exists)
	if err != nil {
		return false, u.Logger.NewError(err.Error())
	}
	return exists, nil
}

func (r *UserRepositoryImpl) GetHighestGeneratedUsernameSuffix(username string, maxUsernameLength int) (int, bool, error) {
	query := `
WITH base_username AS (
	SELECT EXISTS(SELECT 1 FROM users WHERE username = $1) AS exists
),
matching_usernames AS (
	SELECT 0 AS suffix
	FROM users
	WHERE username = $1

	UNION ALL

	SELECT trailing_suffix::integer AS suffix
	FROM (
		SELECT username, substring(username FROM '([0-9]+)$') AS trailing_suffix
		FROM users
		WHERE username ~ '[0-9]$'
	) usernames_with_suffix
	WHERE (SELECT exists FROM base_username)
		AND length(trailing_suffix) < $2
		AND username = substring($1 FROM 1 FOR $2 - length(trailing_suffix)) || trailing_suffix
)
SELECT COALESCE(MAX(suffix), 0), (SELECT exists FROM base_username)
FROM matching_usernames`

	var suffix int
	var exists bool
	err := r.DbProvider.GetDB().QueryRow(query, username, maxUsernameLength).Scan(&suffix, &exists)
	if err != nil {
		return 0, false, u.Logger.NewError(err.Error())
	}
	return suffix, exists, nil
}

func (r *UserRepositoryImpl) ListUsers() ([]tools.User, error) {
	return r.queryUsers("")
}

func (r *UserRepositoryImpl) Wipe() {
	_, err := r.DbProvider.GetDB().Exec("DELETE FROM users")
	if err != nil {
		u.Logger.Error(err)
	}
}

func normalizeUserTimes(user *tools.User) {
	user.SetPasswordTokenExpirationDate = user.SetPasswordTokenExpirationDate.UTC()
	user.CreationDate = user.CreationDate.UTC()
}

func (r *UserRepositoryImpl) GetUserByToken(token string) (*tools.User, error) {
	return r.searchUserBy("set_password_token", token)
}
