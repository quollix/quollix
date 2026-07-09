package apps_basic

import (
	"server/groups"
	"server/tools"
	"server/users"

	u "github.com/quollix/common/utils"
)

const (
	AccessDeniedError        = "access denied"
	UnknownAccessPolicyError = "unknown access policy"
)

type Authorizer interface {
	Authorize(appPolicy string, level tools.UserAccessLevel, userId int, appName string) error
}

type AuthorizerImpl struct {
	GroupRepository groups.GroupRepository
}

func (a *AuthorizerImpl) Authorize(appPolicy string, userAccessLevel tools.UserAccessLevel, userId int, appName string) error {
	err := a.authorize(appPolicy, userAccessLevel, userId, appName)
	if err != nil {
		return u.Logger.AddContext(err, "app_policy", appPolicy, "user_access_level", userAccessLevel.String(), "user_id", userId, "app_name", appName)
	}
	return nil
}

func (a *AuthorizerImpl) authorize(appPolicy string, userAccessLevel tools.UserAccessLevel, userId int, appName string) error {
	switch appPolicy {
	case tools.Policies.PublicAccessPolicy:
		return nil
	case tools.Policies.AuthenticatedAccessPolicy:
		if userAccessLevel == tools.AnonymousLevel {
			return u.Logger.NewError(AccessDeniedError)
		}
		return nil
	case tools.Policies.AdminOnlyAccessPolicy:
		if userAccessLevel == tools.AdminLevel {
			return nil
		}
		return u.Logger.NewError(AccessDeniedError)
	case tools.Policies.GroupRestrictedAccessPolicy:
		if userAccessLevel == tools.AdminLevel {
			return nil
		}
		if userId == users.AnonymousUserId {
			return u.Logger.NewError(AccessDeniedError)
		}
		hasAccess, err := a.GroupRepository.HasAccess(userId, appName)
		if err != nil {
			return err
		}
		if hasAccess {
			return nil
		}
		return u.Logger.NewError(AccessDeniedError)
	default:
		return u.Logger.NewError(UnknownAccessPolicyError)
	}
}
