package oidc_client

import (
	"server/tools"
	"time"

	u "github.com/quollix/common/utils"
)

const userAuthMethodSelect = `
	SELECT
		user_auth_method_id,
		user_id,
		oidc_auth_provider_id,
		external_subject,
		last_oidc_authenticated_at
	FROM user_auth_methods
`

type UserAuthMethod struct {
	Id                      int
	UserId                  int
	OidcAuthProviderId      int
	ExternalSubject         string
	LastOidcAuthenticatedAt time.Time
}

type UserAuthMethodRepository interface {
	CreateUserAuthMethod(method *UserAuthMethod) (int, error)
	GetUserAuthMethodByProviderAndSubject(providerId int, externalSubject string) (*UserAuthMethod, bool, error)
	ListUserAuthMethodsByUserId(userId int) ([]UserAuthMethod, error)
	UpdateLastOidcAuthenticatedAt(methodId int, authenticatedAt time.Time) error
	DeleteUserAuthMethod(methodId int) error
}

type UserAuthMethodRepositoryImpl struct {
	DbConnector tools.DatabaseConnector
}

func (r *UserAuthMethodRepositoryImpl) CreateUserAuthMethod(method *UserAuthMethod) (int, error) {
	var id int
	err := r.DbConnector.GetDB().QueryRow(
		`INSERT INTO user_auth_methods (
			user_id,
			oidc_auth_provider_id,
			external_subject,
			last_oidc_authenticated_at
		)
		VALUES ($1, $2, $3, $4)
		RETURNING user_auth_method_id`,
		method.UserId,
		method.OidcAuthProviderId,
		method.ExternalSubject,
		method.LastOidcAuthenticatedAt,
	).Scan(&id)
	if err != nil {
		return 0, u.Logger.NewError(err.Error())
	}
	return id, nil
}

func (r *UserAuthMethodRepositoryImpl) GetUserAuthMethodByProviderAndSubject(providerId int, externalSubject string) (*UserAuthMethod, bool, error) {
	methods, err := r.queryUserAuthMethods(
		`WHERE oidc_auth_provider_id = $1 AND external_subject = $2`,
		providerId,
		externalSubject,
	)
	if err != nil {
		return nil, false, err
	}
	if len(methods) == 0 {
		return nil, false, nil
	}
	return &methods[0], true, nil
}

func (r *UserAuthMethodRepositoryImpl) ListUserAuthMethodsByUserId(userId int) ([]UserAuthMethod, error) {
	return r.queryUserAuthMethods(`WHERE user_id = $1 ORDER BY oidc_auth_provider_id`, userId)
}

func (r *UserAuthMethodRepositoryImpl) queryUserAuthMethods(where string, args ...any) ([]UserAuthMethod, error) {
	rows, err := r.DbConnector.GetDB().Query(userAuthMethodSelect+" "+where, args...)
	if err != nil {
		return nil, u.Logger.NewError(err.Error())
	}
	defer u.Close(rows)

	var methods []UserAuthMethod
	for rows.Next() {
		var method UserAuthMethod
		if err := rows.Scan(
			&method.Id,
			&method.UserId,
			&method.OidcAuthProviderId,
			&method.ExternalSubject,
			&method.LastOidcAuthenticatedAt,
		); err != nil {
			return nil, u.Logger.NewError(err.Error())
		}
		method.LastOidcAuthenticatedAt = method.LastOidcAuthenticatedAt.UTC()
		methods = append(methods, method)
	}
	if rows.Err() != nil {
		return nil, u.Logger.NewError(rows.Err().Error())
	}
	return methods, nil
}

func (r *UserAuthMethodRepositoryImpl) UpdateLastOidcAuthenticatedAt(methodId int, authenticatedAt time.Time) error {
	_, err := r.DbConnector.GetDB().Exec(
		`UPDATE user_auth_methods
         SET last_oidc_authenticated_at = $2
         WHERE user_auth_method_id = $1`,
		methodId,
		authenticatedAt,
	)
	if err != nil {
		return u.Logger.NewError(err.Error())
	}
	return nil
}

func (r *UserAuthMethodRepositoryImpl) DeleteUserAuthMethod(methodId int) error {
	_, err := r.DbConnector.GetDB().Exec(`DELETE FROM user_auth_methods WHERE user_auth_method_id = $1`, methodId)
	if err != nil {
		return u.Logger.NewError(err.Error())
	}
	return nil
}

func (r *UserAuthMethodRepositoryImpl) Wipe() {
	_, err := r.DbConnector.GetDB().Exec(`DELETE FROM user_auth_methods`)
	if err != nil {
		u.Logger.Error(err)
	}
}
