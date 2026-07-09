//go:build integration

package repository

import (
	"server/oidc_client"
	test_tools "server/tests/test_tools"
	"server/tools"
	"testing"
	"time"

	"github.com/quollix/common/assert"
)

var (
	sampleLastOidcAuthenticatedAt  = time.Date(2026, time.June, 20, 10, 30, 0, 0, time.UTC)
	updatedLastOidcAuthenticatedAt = time.Date(2026, time.June, 21, 11, 45, 0, 0, time.UTC)
)

func TestUserAuthMethodRepository_CRUD(t *testing.T) {
	user, provider := initUserAuthMethodRepoTest(t)
	defer cleanupUserAuthMethodRepoTest()

	method := getSampleUserAuthMethod(user.Id, provider.Id)
	var err error
	method.Id, err = UserAuthMethodRepo.CreateUserAuthMethod(method)
	assert.Nil(t, err)

	actualMethod, found, err := UserAuthMethodRepo.GetUserAuthMethodByProviderAndSubject(provider.Id, method.ExternalSubject)
	assert.Nil(t, err)
	assert.True(t, found)
	assertUserAuthMethodEquality(t, method, actualMethod)

	methods, err := UserAuthMethodRepo.ListUserAuthMethodsByUserId(user.Id)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(methods))
	assertUserAuthMethodEquality(t, method, &methods[0])

	assert.Nil(t, UserAuthMethodRepo.UpdateLastOidcAuthenticatedAt(method.Id, updatedLastOidcAuthenticatedAt))
	actualMethod, found, err = UserAuthMethodRepo.GetUserAuthMethodByProviderAndSubject(provider.Id, method.ExternalSubject)
	assert.Nil(t, err)
	assert.True(t, found)
	assert.Equal(t, updatedLastOidcAuthenticatedAt, actualMethod.LastOidcAuthenticatedAt)

	assert.Nil(t, UserAuthMethodRepo.DeleteUserAuthMethod(method.Id))
	methods, err = UserAuthMethodRepo.ListUserAuthMethodsByUserId(user.Id)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(methods))

	actualMethod, found, err = UserAuthMethodRepo.GetUserAuthMethodByProviderAndSubject(provider.Id, method.ExternalSubject)
	assert.Nil(t, err)
	assert.False(t, found)
	assert.Nil(t, actualMethod)
}

func TestUserAuthMethodRepository_DuplicateProviderSubjectReturnsError(t *testing.T) {
	user, provider := initUserAuthMethodRepoTest(t)
	defer cleanupUserAuthMethodRepoTest()

	method := getSampleUserAuthMethod(user.Id, provider.Id)
	_, err := UserAuthMethodRepo.CreateUserAuthMethod(method)
	assert.Nil(t, err)

	otherUser := GetSampleAdminUser()
	otherUser.Username = "other-user"
	otherUser.Email = "other-user@example.com"
	otherUser.Id, err = UserRepo.CreateUser(otherUser)
	assert.Nil(t, err)

	method.UserId = otherUser.Id
	_, err = UserAuthMethodRepo.CreateUserAuthMethod(method)
	assert.NotNil(t, err)
}

func TestUserAuthMethodRepository_DuplicateUserProviderReturnsError(t *testing.T) {
	user, provider := initUserAuthMethodRepoTest(t)
	defer cleanupUserAuthMethodRepoTest()

	method := getSampleUserAuthMethod(user.Id, provider.Id)
	_, err := UserAuthMethodRepo.CreateUserAuthMethod(method)
	assert.Nil(t, err)

	method.ExternalSubject = "other-subject"
	_, err = UserAuthMethodRepo.CreateUserAuthMethod(method)
	assert.NotNil(t, err)
}

func TestUserAuthMethodRepository_UserDeletionCascades(t *testing.T) {
	user, provider := initUserAuthMethodRepoTest(t)
	defer cleanupUserAuthMethodRepoTest()

	method := getSampleUserAuthMethod(user.Id, provider.Id)
	_, err := UserAuthMethodRepo.CreateUserAuthMethod(method)
	assert.Nil(t, err)

	assert.Nil(t, UserRepo.DeleteUser(user.Id))
	methods, err := UserAuthMethodRepo.ListUserAuthMethodsByUserId(user.Id)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(methods))
}

func TestUserAuthMethodRepository_ProviderDeletionCascades(t *testing.T) {
	user, provider := initUserAuthMethodRepoTest(t)
	defer cleanupUserAuthMethodRepoTest()

	method := getSampleUserAuthMethod(user.Id, provider.Id)
	_, err := UserAuthMethodRepo.CreateUserAuthMethod(method)
	assert.Nil(t, err)

	assert.Nil(t, OidcAuthProviderRepo.DeleteProvider(provider.Id))
	methods, err := UserAuthMethodRepo.ListUserAuthMethodsByUserId(user.Id)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(methods))
}

func initUserAuthMethodRepoTest(t *testing.T) (*tools.User, *oidc_client.OidcAuthProviderDto) {
	InitDeps()

	user := GetSampleAdminUser()
	var err error
	user.Id, err = UserRepo.CreateUser(user)
	assert.Nil(t, err)

	provider := test_tools.GetSampleOidcAuthProvider()
	provider.Id, err = OidcAuthProviderRepo.CreateProvider(provider)
	assert.Nil(t, err)

	return user, provider
}

func cleanupUserAuthMethodRepoTest() {
	UserAuthMethodRepo.Wipe()
	OidcAuthProviderRepo.Wipe()
	UserRepo.Wipe()
}

func getSampleUserAuthMethod(userId, providerId int) *oidc_client.UserAuthMethod {
	return &oidc_client.UserAuthMethod{
		UserId:                  userId,
		OidcAuthProviderId:      providerId,
		ExternalSubject:         "external-subject",
		LastOidcAuthenticatedAt: sampleLastOidcAuthenticatedAt,
	}
}

func assertUserAuthMethodEquality(t *testing.T, expected, actual *oidc_client.UserAuthMethod) {
	assert.Equal(t, expected.Id, actual.Id)
	assert.Equal(t, expected.UserId, actual.UserId)
	assert.Equal(t, expected.OidcAuthProviderId, actual.OidcAuthProviderId)
	assert.Equal(t, expected.ExternalSubject, actual.ExternalSubject)
	assert.Equal(t, expected.LastOidcAuthenticatedAt, actual.LastOidcAuthenticatedAt)
}
