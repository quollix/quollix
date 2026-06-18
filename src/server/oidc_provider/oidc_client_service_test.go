package oidc_provider

import (
	"server/apps_basic"
	"server/configs"
	"server/groups"
	"server/tools"
	"server/users"
	"testing"

	"github.com/quollix/common/assert"
	u "github.com/quollix/common/utils"
)

const (
	oidcClientServiceClientId    = "abcdef"
	oidcClientServiceRedirectUri = "https://sampleapp.localhost/callback"
)

type oidcClientServiceTestObjects struct {
	Service         *OidcClientServiceImpl
	AppRepo         *apps_basic.AppRepositoryMock
	ConfigsRepo     *configs.ConfigsRepositoryMock
	UserRepo        *users.UserRepositoryMock
	GroupRepository *groups.GroupRepositoryMock
	Clock           *ClockMock
}

func newOidcClientServiceTestObjects(t *testing.T) oidcClientServiceTestObjects {
	appRepo := apps_basic.NewAppRepositoryMock(t)
	configsRepo := configs.NewConfigsRepositoryMock(t)
	userRepo := users.NewUserRepositoryMock(t)
	groupRepository := groups.NewGroupRepositoryMock(t)
	clock := NewClockMock(t)

	service := &OidcClientServiceImpl{
		AppRepository:   appRepo,
		ConfigsRepo:     configsRepo,
		UserRepository:  userRepo,
		GroupRepository: groupRepository,
		Clock:           clock,
	}

	return oidcClientServiceTestObjects{
		Service:         service,
		AppRepo:         appRepo,
		ConfigsRepo:     configsRepo,
		UserRepo:        userRepo,
		GroupRepository: groupRepository,
		Clock:           clock,
	}
}

func expectAppAndHostForRedirect(testObjects oidcClientServiceTestObjects) {
	testObjects.AppRepo.EXPECT().
		GetAppByClientId(oidcClientServiceClientId).
		Return(&apps_basic.RepoApp{AppName: "sampleapp"}, nil)

	testObjects.ConfigsRepo.EXPECT().
		GetConfig(configs.ConfigKeys.ServerHost).
		Return("localhost", nil)
}

func TestOidcClientServiceImpl_CheckClientIdAndRedirectUri_HappyPath(t *testing.T) {
	testObjects := newOidcClientServiceTestObjects(t)
	expectAppAndHostForRedirect(testObjects)

	err := testObjects.Service.CheckClientIdAndRedirectUri(oidcClientServiceClientId, oidcClientServiceRedirectUri)
	assert.Nil(t, err)
}

func TestOidcClientServiceImpl_CheckClientIdAndRedirectUri_WhenWrongScheme_ReturnsError(t *testing.T) {
	testObjects := newOidcClientServiceTestObjects(t)
	expectAppAndHostForRedirect(testObjects)

	err := testObjects.Service.CheckClientIdAndRedirectUri(oidcClientServiceClientId, "http://sampleapp.localhost/callback")
	assert.Equal(t, "wrong redirect_uri scheme", u.ExtractError(err))
}

func TestOidcClientServiceImpl_CheckClientIdAndRedirectUri_WhenWrongHost_ReturnsError(t *testing.T) {
	testObjects := newOidcClientServiceTestObjects(t)
	expectAppAndHostForRedirect(testObjects)

	err := testObjects.Service.CheckClientIdAndRedirectUri(oidcClientServiceClientId, "https://evil.localhost/callback")
	assert.Equal(t, "wrong redirect_uri host", u.ExtractError(err))
}

func TestOidcClientServiceImpl_CheckClientIdAndRedirectUri_WhenWrongPort_ReturnsError(t *testing.T) {
	testObjects := newOidcClientServiceTestObjects(t)
	expectAppAndHostForRedirect(testObjects)

	err := testObjects.Service.CheckClientIdAndRedirectUri(oidcClientServiceClientId, "https://sampleapp.localhost:8443/callback")
	assert.Equal(t, "wrong redirect_uri port", u.ExtractError(err))
}

func TestOidcClientServiceImpl_CheckClientIdAndRedirectUri_WhenInvalidUrl_ReturnsError(t *testing.T) {
	testObjects := newOidcClientServiceTestObjects(t)
	expectAppAndHostForRedirect(testObjects)

	err := testObjects.Service.CheckClientIdAndRedirectUri(oidcClientServiceClientId, "https://%zz")
	assert.Equal(t, "invalid redirect_uri format", u.ExtractError(err))
}

func TestOidcClientServiceImpl_CheckClientIdAndRedirectUri_WhenPortIs443_ReturnsNil(t *testing.T) {
	testObjects := newOidcClientServiceTestObjects(t)
	expectAppAndHostForRedirect(testObjects)

	err := testObjects.Service.CheckClientIdAndRedirectUri(oidcClientServiceClientId, "https://sampleapp.localhost:443/callback")
	assert.Nil(t, err)
}

func TestOidcClientServiceImpl_AuthenticateClient_WhenSecretMatches_ReturnsNil(t *testing.T) {
	testObjects := newOidcClientServiceTestObjects(t)

	clientSecret := "secret-1"
	testObjects.AppRepo.EXPECT().
		GetAppByClientId(oidcClientServiceClientId).
		Return(&apps_basic.RepoApp{ClientSecret: clientSecret}, nil)

	err := testObjects.Service.AuthenticateClient(oidcClientServiceClientId, clientSecret)
	assert.Nil(t, err)
}

func TestOidcClientServiceImpl_AuthenticateClient_WhenSecretMismatch_ReturnsError(t *testing.T) {
	testObjects := newOidcClientServiceTestObjects(t)

	testObjects.AppRepo.EXPECT().
		GetAppByClientId(oidcClientServiceClientId).
		Return(&apps_basic.RepoApp{ClientSecret: "expected-secret"}, nil)

	err := testObjects.Service.AuthenticateClient(oidcClientServiceClientId, "wrong-secret")
	assert.Equal(t, "wrong client secret", u.ExtractError(err))
}

func TestOidcClientServiceImpl_AuthenticateClient_WhenStoredSecretMissing_ReturnsError(t *testing.T) {
	testObjects := newOidcClientServiceTestObjects(t)

	testObjects.AppRepo.EXPECT().
		GetAppByClientId(oidcClientServiceClientId).
		Return(&apps_basic.RepoApp{ClientSecret: ""}, nil)

	err := testObjects.Service.AuthenticateClient(oidcClientServiceClientId, "secret-1")
	assert.Equal(t, "confidential client missing stored client secret", u.ExtractError(err))
}

func TestOidcClientServiceImpl_AuthenticateClient_WhenRequestSecretMissing_ReturnsError(t *testing.T) {
	testObjects := newOidcClientServiceTestObjects(t)

	testObjects.AppRepo.EXPECT().
		GetAppByClientId(oidcClientServiceClientId).
		Return(&apps_basic.RepoApp{ClientSecret: "expected-secret"}, nil)

	err := testObjects.Service.AuthenticateClient(oidcClientServiceClientId, "")
	assert.Equal(t, "confidential client missing client secret", u.ExtractError(err))
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

	userInfo, err := testObjects.Service.GetOidcUserInfo("7")
	assert.Nil(t, err)

	assert.Equal(t, "7", userInfo.Sub)
	assert.Equal(t, "admin", userInfo.Role)
	assert.Equal(t, []string{"admins", "devs"}, userInfo.Groups)
	assert.Equal(t, "user-7", userInfo.Name)
	assert.Equal(t, "user-7", userInfo.PreferredUsername)
	assert.Equal(t, "user7@example.com", userInfo.Email)
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

	userInfo, err := testObjects.Service.GetOidcUserInfo("7")
	assert.Nil(t, err)

	assert.Equal(t, "7", userInfo.Sub)
	assert.Equal(t, "user", userInfo.Role)
	assert.Equal(t, []string{"samplegroup"}, userInfo.Groups)
	assert.Equal(t, "user-7", userInfo.Name)
	assert.Equal(t, "user-7", userInfo.PreferredUsername)
	assert.Equal(t, "user7@example.com", userInfo.Email)
}

func TestOidcClientServiceImpl_BuildIdTokenClaims_UsesClockIssuerAndCopiesUserClaims(t *testing.T) {
	testObjects := newOidcClientServiceTestObjects(t)

	testObjects.Clock.EXPECT().Now().Return(sampleTime)
	testObjects.ConfigsRepo.EXPECT().GetConfig(configs.ConfigKeys.ServerHost).Return("localhost", nil)
	expectUserAndGroups(
		testObjects,
		7,
		true,
		[]string{"admins"},
		"user-7",
		"user7@example.com",
	)

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
	assert.Equal(t, "user7@example.com", claims.Email)
}
