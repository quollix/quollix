//go:build component

package component

import (
	"server/apps_basic"
	"server/groups"
	"server/tools"
	"testing"

	"github.com/quollix/common/assert"
	u "github.com/quollix/common/utils"
)

const sampleGroupName = "samplegroup"

func TestGroupsCRUDViaHTTP(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	group, err := client.Groups.CreateGroup(sampleGroupName)
	assert.Nil(t, err)

	allGroups, err := client.Groups.ListAllGroups()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(allGroups))
	assert.Equal(t, sampleGroupName, allGroups[0].Name)

	assert.Equal(t, group.Id, allGroups[0].Id)
	assert.Equal(t, group.Name, allGroups[0].Name)

	assert.Nil(t, client.Groups.DeleteGroup(allGroups[0].Id))

	allGroups, err = client.Groups.ListAllGroups()
	assert.Nil(t, err)
	assert.Equal(t, 0, len(allGroups))
}

func TestCreateGroupDuplicateViaHTTP_ReturnsGroupAlreadyExistsError(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	_, err := client.Groups.CreateGroup(sampleGroupName)
	assert.Nil(t, err)

	_, err = client.Groups.CreateGroup(sampleGroupName)
	u.AssertDeepStackErrorFromRequest(t, err, groups.GroupAlreadyExistsError)
}

func TestGroupMembershipCRUDViaHTTP(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	InviteUserAndSetPassword(client, SampleUsername, SampleUserPassword, SampleUserEmail)
	user := client.Users.GetByUsername(SampleUsername)

	group, err := client.Groups.CreateGroup(sampleGroupName)
	assert.Nil(t, err)

	usersByGroup, err := client.Groups.ListUsersByGroupMembership(group.Id)
	assert.Nil(t, err)
	assert.True(t, containsMember(usersByGroup.NotIn, SampleUsername))
	assert.False(t, containsMember(usersByGroup.In, SampleUsername))

	assert.Nil(t, client.Groups.AddUserToGroup(user.Id, group.Id))

	usersByGroup, err = client.Groups.ListUsersByGroupMembership(group.Id)
	assert.Nil(t, err)
	assert.True(t, containsMember(usersByGroup.In, SampleUsername))
	assert.False(t, containsMember(usersByGroup.NotIn, SampleUsername))

	assert.Nil(t, client.Groups.RemoveUserFromGroup(user.Id, group.Id))

	usersByGroup, err = client.Groups.ListUsersByGroupMembership(group.Id)
	assert.Nil(t, err)
	assert.True(t, containsMember(usersByGroup.NotIn, SampleUsername))
	assert.False(t, containsMember(usersByGroup.In, SampleUsername))
}

func containsMember(ms []groups.Member, name string) bool {
	for _, m := range ms {
		if m.Name == name {
			return true
		}
	}
	return false
}

func TestAppAccessCRUDViaHTTP(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	_, err := client.Apps.InstallSample("2.0")
	assert.Nil(t, err)

	group, err := client.Groups.CreateGroup(sampleGroupName)
	assert.Nil(t, err)

	appsAccessByGroup, err := client.Groups.ListAppsAccessByGroup(group.Id)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(appsAccessByGroup.Granted))
	assert.Equal(t, 1, len(appsAccessByGroup.NotGranted))
	assert.Equal(t, tools.SampleApp, appsAccessByGroup.NotGranted[0])

	assert.Nil(t, client.Groups.GrantAppAccess(group.Id, tools.SampleApp))

	appsAccessByGroup, err = client.Groups.ListAppsAccessByGroup(group.Id)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(appsAccessByGroup.Granted))
	assert.Equal(t, 0, len(appsAccessByGroup.NotGranted))
	assert.Equal(t, tools.SampleApp, appsAccessByGroup.Granted[0])

	assert.Nil(t, client.Groups.RevokeAppAccess(group.Id, tools.SampleApp))

	appsAccessByGroup, err = client.Groups.ListAppsAccessByGroup(group.Id)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(appsAccessByGroup.Granted))
	assert.Equal(t, 1, len(appsAccessByGroup.NotGranted))
	assert.Equal(t, tools.SampleApp, appsAccessByGroup.NotGranted[0])
}

func TestQuollixClient_FullAccessViaGroupFlow(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	app, err := client.Apps.InstallSample("2.0")
	assert.Nil(t, err)
	assert.Nil(t, client.Apps.Start(app.AppId))

	assert.Nil(t, client.Apps.SetAccessPolicy(app.AppId, tools.Policies.GroupRestrictedAccessPolicy))

	InviteUserAndSetPassword(client, SampleUsername, SampleUserPassword, SampleUserEmail)
	user := client.Users.GetByUsername(SampleUsername)

	group, err := client.Groups.CreateGroup(sampleGroupName)
	assert.Nil(t, err)
	assert.Nil(t, client.Groups.AddUserToGroup(user.Id, group.Id))

	userClient := GetQuollixClient(t)
	assert.Nil(t, userClient.Auth.Login(SampleUsername, SampleUserPassword))

	userAppClient := GetAppClient(t, userClient)
	err = userAppClient.Content.AssertContent("this is version 2.0")
	u.AssertDeepStackErrorFromRequest(t, err, apps_basic.AccessDeniedError)

	assert.Nil(t, client.Groups.GrantAppAccess(group.Id, tools.SampleApp))

	userAppClient = GetAppClient(t, userClient)
	err = userAppClient.Content.AssertContent("this is version 2.0")
	assert.Nil(t, err)
}

func TestOIDC_IDToken_GroupsClaimIsSetWhenUserAddedToGroup(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	app, err := client.Apps.InstallSample("2.0")
	assert.Nil(t, err)

	addAdminUserToGroup(t, client)

	ctx := NewTestContext(t)
	ctx.LoginAsAdmin()
	authRes, verifier := ctx.AuthorizeWithPKCE("openid", app.ClientId)
	tokens := ctx.ExchangeCodeForTokens(authRes.Code, verifier, app.ClientId, app.ClientSecret, ClientAuthMethodBasic)

	pubKey := FetchPublicKeyFromJWKS(t, ctx)
	claims := VerifyIDToken(t, tokens.IDToken, pubKey, ctx.Config.Issuer, app.ClientId)

	assert.Equal(t, []string{"samplegroup"}, claims.Groups)
}

func TestOIDC_Userinfo_GroupsAttributeIsSetWhenUserAddedToGroup(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	app, err := client.Apps.InstallSample("2.0")
	assert.Nil(t, err)

	addAdminUserToGroup(t, client)

	ctx := NewTestContext(t)
	ctx.LoginAsAdmin()
	authRes, verifier := ctx.AuthorizeWithPKCE("openid", app.ClientId)
	tokens := ctx.ExchangeCodeForTokens(authRes.Code, verifier, app.ClientId, app.ClientSecret, ClientAuthMethodBasic)

	uinfo := ctx.FetchUserinfo(tokens.AccessToken)
	assert.Equal(t, []string{"samplegroup"}, uinfo.Groups)
}

func TestAccessPolicy_GroupRestricted(t *testing.T) {
	adminClient := GetClientAndLogin(t)
	RunAccessPoliciesTest(t, adminClient, []AccessPolicyTestCase{
		{
			AccessPolicy:              tools.Policies.GroupRestrictedAccessPolicy,
			ShouldAdminHaveAccess:     true,
			ShouldUserHaveAccess:      false,
			ShouldAnonymousHaveAccess: false,
		},
	})
}

func addAdminUserToGroup(t *testing.T, client *QuollixClient) {
	users := client.Users.List()
	assert.Equal(t, 1, len(users))

	group, err := client.Groups.CreateGroup("samplegroup")
	assert.Nil(t, err)
	assert.Nil(t, client.Groups.AddUserToGroup(users[0].Id, group.Id))
}
