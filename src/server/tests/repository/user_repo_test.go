//go:build integration

package repository

import (
	"server/tools"
	"server/users"
	"testing"
	"time"

	"github.com/quollix/common/assert"
)

func TestUserReadings(t *testing.T) {
	InitDeps()
	defer UserRepo.Wipe()

	expectedUser := GetSampleAdminUser()
	var err error
	expectedUser.Id, err = UserRepo.CreateUser(expectedUser)
	assert.Nil(t, err)

	actualUser, err := UserRepo.GetUserByUsername(expectedUser.Username)
	assert.Nil(t, err)
	assertUserEquality(t, expectedUser, actualUser)

	users, err := UserRepo.ListUsers()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(users))
	actualUser = &users[0]
	assertUserEquality(t, expectedUser, actualUser)

	userId := actualUser.Id
	actualUser, err = UserRepo.GetUserById(userId)
	assert.Nil(t, err)
	assertUserEquality(t, expectedUser, actualUser)
}

func TestUserDeletion(t *testing.T) {
	InitDeps()
	defer UserRepo.Wipe()

	expectedUser := GetSampleAdminUser()
	var err error
	expectedUser.Id, err = UserRepo.CreateUser(expectedUser)
	assert.Nil(t, err)

	users, err := UserRepo.ListUsers()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(users))
	userId := users[0].Id

	assert.Nil(t, UserRepo.DeleteUser(userId))

	users, err = UserRepo.ListUsers()
	assert.Nil(t, err)
	assert.Equal(t, 0, len(users))
}

func TestGetUserByToken(t *testing.T) {
	InitDeps()
	defer UserRepo.Wipe()

	expectedUser := GetSampleAdminUser()
	var err error
	expectedUser.Id, err = UserRepo.CreateUser(expectedUser)
	assert.Nil(t, err)

	actualUser, err := UserRepo.GetUserByToken(expectedUser.SetPasswordToken)
	assert.Nil(t, err)
	assertUserEquality(t, expectedUser, actualUser)

	actualUser, err = UserRepo.GetUserByToken("wrong-token")
	assert.NotNil(t, err)
	assert.Nil(t, actualUser)
}

func TestDoesUserExist(t *testing.T) {
	InitDeps()
	defer UserRepo.Wipe()

	doesUserExist, err := UserRepo.DoesUserExist("normal-user")
	assert.Nil(t, err)
	assert.False(t, doesUserExist)

	normalUser := GetSampleAdminUser()
	normalUser.IsAdmin = false
	normalUser.Username = "normal-user"
	normalUser.Email = "normal-user@example.com"
	_, err = UserRepo.CreateUser(normalUser)
	assert.Nil(t, err)
	doesUserExist, err = UserRepo.DoesUserExist("normal-user")
	assert.Nil(t, err)
	assert.True(t, doesUserExist)

	doesAdminExist, err := UserRepo.DoesAnyAdminUserExist()
	assert.Nil(t, err)
	assert.False(t, doesAdminExist)

	adminUser := GetSampleAdminUser()
	adminUser.Email = "admin2@example.com"
	_, err = UserRepo.CreateUser(adminUser)
	assert.Nil(t, err)

	doesAdminExist, err = UserRepo.DoesAnyAdminUserExist()
	assert.Nil(t, err)
	assert.True(t, doesAdminExist)
}

func TestUserRepositoryImpl_GetHighestGeneratedUsernameSuffix_ReturnsHighestMatchingSuffix(t *testing.T) {
	InitDeps()
	defer UserRepo.Wipe()

	insertTestUser(t, "tom")
	insertTestUser(t, "tom1")
	insertTestUser(t, "tom5")
	insertTestUser(t, "tom40")
	insertTestUser(t, "tommy50")

	suffix, exists, err := UserRepo.GetHighestGeneratedUsernameSuffix("tom", users.UsernameMaxLength)

	assert.Nil(t, err)
	assert.True(t, exists)
	assert.Equal(t, 40, suffix)
}

func TestUserRepositoryImpl_GetHighestGeneratedUsernameSuffix_ReturnsFalseWhenNoMatchExists(t *testing.T) {
	InitDeps()
	defer UserRepo.Wipe()

	suffix, exists, err := UserRepo.GetHighestGeneratedUsernameSuffix("tom", users.UsernameMaxLength)

	assert.Nil(t, err)
	assert.False(t, exists)
	assert.Equal(t, 0, suffix)
}

func TestUserRepositoryImpl_GetHighestGeneratedUsernameSuffix_IgnoresSuffixesWhenBaseUsernameIsFree(t *testing.T) {
	InitDeps()
	defer UserRepo.Wipe()

	insertTestUser(t, "tom1")
	insertTestUser(t, "tom40")

	suffix, exists, err := UserRepo.GetHighestGeneratedUsernameSuffix("tom", users.UsernameMaxLength)

	assert.Nil(t, err)
	assert.False(t, exists)
	assert.Equal(t, 0, suffix)
}

func TestUserRepositoryImpl_GetHighestGeneratedUsernameSuffix_HandlesTruncatedGeneratedUsernames(t *testing.T) {
	InitDeps()
	defer UserRepo.Wipe()

	insertTestUser(t, "abcdefghijklmnopqrst")
	insertTestUser(t, "abcdefghijklmnopqrs1")
	insertTestUser(t, "abcdefghijklmnopqr12")

	suffix, exists, err := UserRepo.GetHighestGeneratedUsernameSuffix("abcdefghijklmnopqrst", users.UsernameMaxLength)

	assert.Nil(t, err)
	assert.True(t, exists)
	assert.Equal(t, 12, suffix)
}

func TestUpdateUser(t *testing.T) {
	InitDeps()
	defer UserRepo.Wipe()

	sampleUser := GetSampleAdminUser()
	var err error
	sampleUser.Id, err = UserRepo.CreateUser(sampleUser)
	assert.Nil(t, err)

	updatedUser := &tools.User{
		Id:                             sampleUser.Id,
		Username:                       "updated-name",
		Email:                          "updated-name@example.com",
		HashedPassword:                 "other-hashed-password",
		IsAdmin:                        false,
		IsEnabled:                      false,
		SetPasswordToken:               "other-set-password-token",
		SetPasswordTokenExpirationDate: time.Date(1991, time.January, 1, 0, 0, 0, 0, time.UTC),
	}

	assert.Nil(t, UserRepo.UpdateUser(updatedUser))
	actualUser, err := UserRepo.GetUserByUsername("updated-name")
	assert.Nil(t, err)
	assertUserEquality(t, updatedUser, actualUser)
}

func TestDoesEmailExist(t *testing.T) {
	InitDeps()
	defer UserRepo.Wipe()

	exists, err := UserRepo.DoesEmailExist("missing@example.com")
	assert.Nil(t, err)
	assert.False(t, exists)

	user := GetSampleAdminUser()
	user.Email = "existing@example.com"
	_, err = UserRepo.CreateUser(user)
	assert.Nil(t, err)

	exists, err = UserRepo.DoesEmailExist("existing@example.com")
	assert.Nil(t, err)
	assert.True(t, exists)
}

func assertUserEquality(t *testing.T, expectedUser *tools.User, actualUser *tools.User) {
	assert.Equal(t, expectedUser.Id, actualUser.Id)
	assert.Equal(t, expectedUser.Username, actualUser.Username)
	assert.Equal(t, expectedUser.Email, actualUser.Email)
	assert.Equal(t, expectedUser.HashedPassword, actualUser.HashedPassword)
	assert.Equal(t, expectedUser.IsAdmin, actualUser.IsAdmin)
	assert.Equal(t, expectedUser.IsEnabled, actualUser.IsEnabled)
	assert.Equal(t, expectedUser.SetPasswordToken, actualUser.SetPasswordToken)
	assert.Equal(t, expectedUser.SetPasswordTokenExpirationDate, actualUser.SetPasswordTokenExpirationDate)
	assert.Equal(t, expectedUser.CreationDate, actualUser.CreationDate)
}

func insertTestUser(t *testing.T, username string) {
	user := GetSampleAdminUser()
	user.Username = username
	user.Email = username + "@example.invalid"
	_, err := UserRepo.CreateUser(user)
	assert.Nil(t, err)
}
