package oidc_provider

import (
	"testing"

	"server/configs"
	"server/groups"
	"server/tools"
	"server/users"

	"github.com/quollix/common/assert"
)

type oidcClientServiceTestObjects struct {
	Service         *OidcClientServiceImpl
	ConfigsRepo     *configs.ConfigsRepositoryMock
	ConfigsService  *configs.ConfigsServiceMock
	UserRepo        *users.UserRepositoryMock
	GroupRepository *groups.GroupRepositoryMock
	Clock           *ClockMock
}

func newOidcClientServiceTestObjects(t *testing.T) oidcClientServiceTestObjects {
	configsRepo := configs.NewConfigsRepositoryMock(t)
	configsService := configs.NewConfigsServiceMock(t)
	userRepo := users.NewUserRepositoryMock(t)
	groupRepository := groups.NewGroupRepositoryMock(t)
	clock := NewClockMock(t)
	oidcEmailService := &configs.OidcEmailExposureServiceImpl{ConfigsRepo: configsRepo}

	service := &OidcClientServiceImpl{
		ConfigsService:   configsService,
		OidcEmailService: oidcEmailService,
		UserRepository:   userRepo,
		GroupRepository:  groupRepository,
		Clock:            clock,
	}

	return oidcClientServiceTestObjects{
		Service:         service,
		ConfigsRepo:     configsRepo,
		ConfigsService:  configsService,
		UserRepo:        userRepo,
		GroupRepository: groupRepository,
		Clock:           clock,
	}
}

func expectUserAndGroups(
	testObjects oidcClientServiceTestObjects,
	userId int,
	isAdmin bool,
	returnedGroupNames []string,
	username string,
	email string,
) {
	testObjects.UserRepo.EXPECT().
		GetUserById(userId).
		Return(&tools.User{
			Id:       userId,
			IsAdmin:  isAdmin,
			Username: username,
			Email:    email,
		}, nil)

	returnedGroups := make([]groups.Group, 0, len(returnedGroupNames))
	for index, groupName := range returnedGroupNames {
		returnedGroups = append(returnedGroups, groups.Group{Id: index + 1, Name: groupName})
	}
	testObjects.GroupRepository.EXPECT().
		ListGroupsForUser(userId).
		Return(returnedGroups, nil)
}

func expectOidcRealEmailExposure(testObjects oidcClientServiceTestObjects, value bool) {
	configValue := "false"
	if value {
		configValue = "true"
	}
	testObjects.ConfigsRepo.EXPECT().
		GetConfig(configs.ConfigKeys.ExposeRealEmailInOidcToken).
		Return(configValue, nil)
}

func TestOidcClientServiceImpl_GetOidcUserInfo_WhenAdmin_ReturnsAdminGroupsAndProfileFields(t *testing.T) {
	testObjects := newOidcClientServiceTestObjects(t)

	expectUserAndGroups(
		testObjects,
		7,
		true,
		[]string{"admins", "devs"},
		"user-7",
		"user7@example.com",
	)
	expectOidcRealEmailExposure(testObjects, false)

	userInfo, err := testObjects.Service.GetOidcUserInfo("7")
	assert.Nil(t, err)

	assert.Equal(t, "7", userInfo.Sub)
	assert.Equal(t, "admin", userInfo.Role)
	assert.Equal(t, []string{"admins", "devs"}, userInfo.Groups)
	assert.Equal(t, "user-7", userInfo.Name)
	assert.Equal(t, "user-7", userInfo.PreferredUsername)
	assert.Equal(t, "user-7@example.invalid", userInfo.Email)
}

func TestOidcClientServiceImpl_GetOidcUserInfo_WhenNotAdmin_ReturnsUserGroupsAndProfileFields(t *testing.T) {
	testObjects := newOidcClientServiceTestObjects(t)

	expectUserAndGroups(
		testObjects,
		7,
		false,
		[]string{"samplegroup"},
		"user-7",
		"user7@example.com",
	)
	expectOidcRealEmailExposure(testObjects, false)

	userInfo, err := testObjects.Service.GetOidcUserInfo("7")
	assert.Nil(t, err)

	assert.Equal(t, "7", userInfo.Sub)
	assert.Equal(t, "user", userInfo.Role)
	assert.Equal(t, []string{"samplegroup"}, userInfo.Groups)
	assert.Equal(t, "user-7", userInfo.Name)
	assert.Equal(t, "user-7", userInfo.PreferredUsername)
	assert.Equal(t, "user-7@example.invalid", userInfo.Email)
}

func TestOidcClientServiceImpl_GetOidcUserInfo_WhenRealEmailExposureEnabled_ReturnsStoredEmail(t *testing.T) {
	testObjects := newOidcClientServiceTestObjects(t)

	expectUserAndGroups(
		testObjects,
		7,
		false,
		[]string{"samplegroup"},
		"user-7",
		"user7@example.com",
	)
	expectOidcRealEmailExposure(testObjects, true)

	userInfo, err := testObjects.Service.GetOidcUserInfo("7")
	assert.Nil(t, err)

	assert.Equal(t, "user7@example.com", userInfo.Email)
}

func TestOidcClientServiceImpl_BuildIdTokenClaims_UsesClockIssuerAndCopiesUserClaims(t *testing.T) {
	testObjects := newOidcClientServiceTestObjects(t)

	testObjects.Clock.EXPECT().Now().Return(sampleTime)
	testObjects.ConfigsService.EXPECT().GetBaseDomain().Return("localhost", nil)
	expectUserAndGroups(
		testObjects,
		7,
		true,
		[]string{"admins"},
		"user-7",
		"user7@example.com",
	)
	expectOidcRealEmailExposure(testObjects, false)

	claims, err := testObjects.Service.BuildIdTokenClaims("7", "client-1", "nonce-1")
	assert.Nil(t, err)

	assert.Equal(t, "7", claims.Sub)
	assert.Equal(t, "https://quollix.localhost", claims.Iss)
	assert.Equal(t, "client-1", claims.Aud)
	assert.Equal(t, "nonce-1", claims.Nonce)
	assert.Equal(t, "admin", claims.Role)
	assert.Equal(t, []string{"admins"}, claims.Groups)

	assert.Equal(t, sampleTime.Add(accessTokenTTL).Unix(), claims.Exp)
	assert.Equal(t, sampleTime.Unix(), claims.Iat)

	assert.Equal(t, "user-7", claims.Name)
	assert.Equal(t, "user-7", claims.PreferredUsername)
	assert.Equal(t, "user-7@example.invalid", claims.Email)
}
