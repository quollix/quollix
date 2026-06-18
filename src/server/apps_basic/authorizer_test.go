package apps_basic

import (
	"errors"
	"server/groups"
	"server/tools"
	"server/users"
	"testing"

	"github.com/quollix/common/assert"
	u "github.com/quollix/common/utils"
)

const (
	sampleAppName = "sampleapp"
	sampleUserId  = 123
)

func getMockedAuthorizer(t *testing.T) (*AuthorizerImpl, *groups.GroupRepositoryMock) {
	groupRepositoryMock := groups.NewGroupRepositoryMock(t)
	authorizer := &AuthorizerImpl{
		GroupRepository: groupRepositoryMock,
	}
	return authorizer, groupRepositoryMock
}

func TestPublicPolicyAllowed(t *testing.T) {
	authorizer, _ := getMockedAuthorizer(t)

	err := authorizer.Authorize(tools.Policies.PublicAccessPolicy, tools.AnonymousLevel, users.AnonymousUserId, sampleAppName)
	assert.Nil(t, err)
}

func TestAuthenticatedPolicyAnonymousDenied(t *testing.T) {
	authorizer, _ := getMockedAuthorizer(t)

	err := authorizer.Authorize(tools.Policies.AuthenticatedAccessPolicy, tools.AnonymousLevel, users.AnonymousUserId, sampleAppName)
	assert.Equal(t, AccessDeniedError, u.ExtractError(err))
}

func TestAuthenticatedPolicyUserAllowed(t *testing.T) {
	authorizer, _ := getMockedAuthorizer(t)

	err := authorizer.Authorize(tools.Policies.AuthenticatedAccessPolicy, tools.UserLevel, sampleUserId, sampleAppName)
	assert.Nil(t, err)
}

func TestAdminOnlyPolicyAdminAllowed(t *testing.T) {
	authorizer, _ := getMockedAuthorizer(t)

	err := authorizer.Authorize(tools.Policies.AdminOnlyAccessPolicy, tools.AdminLevel, sampleUserId, sampleAppName)
	assert.Nil(t, err)
}

func TestAdminOnlyPolicyUserDenied(t *testing.T) {
	authorizer, _ := getMockedAuthorizer(t)

	err := authorizer.Authorize(tools.Policies.AdminOnlyAccessPolicy, tools.AnonymousLevel, sampleUserId, sampleAppName)
	assert.Equal(t, AccessDeniedError, u.ExtractError(err))

	err = authorizer.Authorize(tools.Policies.AdminOnlyAccessPolicy, tools.UserLevel, sampleUserId, sampleAppName)
	assert.Equal(t, AccessDeniedError, u.ExtractError(err))
}

func TestGroupRestrictedPolicyAdminAllowed(t *testing.T) {
	authorizer, _ := getMockedAuthorizer(t)

	err := authorizer.Authorize(tools.Policies.GroupRestrictedAccessPolicy, tools.AdminLevel, sampleUserId, sampleAppName)
	assert.Nil(t, err)
}

func TestGroupRestrictedPolicyAnonymousUserIdDenied(t *testing.T) {
	authorizer, _ := getMockedAuthorizer(t)

	err := authorizer.Authorize(tools.Policies.GroupRestrictedAccessPolicy, tools.UserLevel, users.AnonymousUserId, sampleAppName)
	assert.Equal(t, AccessDeniedError, u.ExtractError(err))
}

func TestGroupRestrictedPolicyHasAccessAllowed(t *testing.T) {
	authorizer, groupRepositoryMock := getMockedAuthorizer(t)
	groupRepositoryMock.EXPECT().HasAccess(sampleUserId, sampleAppName).Return(true, nil)

	err := authorizer.Authorize(tools.Policies.GroupRestrictedAccessPolicy, tools.UserLevel, sampleUserId, sampleAppName)
	assert.Nil(t, err)
}

func TestGroupRestrictedPolicyNoAccessDenied(t *testing.T) {
	authorizer, groupRepositoryMock := getMockedAuthorizer(t)
	groupRepositoryMock.EXPECT().HasAccess(sampleUserId, sampleAppName).Return(false, nil)

	err := authorizer.Authorize(tools.Policies.GroupRestrictedAccessPolicy, tools.UserLevel, sampleUserId, sampleAppName)
	assert.Equal(t, AccessDeniedError, u.ExtractError(err))
}

func TestGroupRestrictedPolicyRepoErrorReturned(t *testing.T) {
	authorizer, groupRepositoryMock := getMockedAuthorizer(t)

	repositoryError := errors.New("repository error")
	groupRepositoryMock.EXPECT().HasAccess(sampleUserId, sampleAppName).Return(false, repositoryError)

	err := authorizer.Authorize(tools.Policies.GroupRestrictedAccessPolicy, tools.UserLevel, sampleUserId, sampleAppName)
	assert.NotNil(t, err)
	assert.Equal(t, repositoryError.Error(), u.ExtractError(err))
}

func TestUnknownPolicyDenied(t *testing.T) {
	authorizer, _ := getMockedAuthorizer(t)

	err := authorizer.Authorize("some-unknown-policy", tools.UserLevel, sampleUserId, sampleAppName)
	assert.Equal(t, UnknownAccessPolicyError, u.ExtractError(err))
}
