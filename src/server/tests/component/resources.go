package component

import (
	"crypto/tls"
	"testing"

	"server/apps_basic"
	"server/certificates"
	"server/oidc_client"
	"server/oidc_provider"
	"server/tests/api_client"
	"server/tools"

	"github.com/quollix/common/assert"
	"github.com/quollix/common/store"
	u "github.com/quollix/common/utils"
)

const (
	SampleUsername      = "user"
	SampleUserPassword  = "userpassword"
	SampleUserEmail     = "user@example.invalid"
	SampleRealUserEmail = "user@example.com"
)

type AccessPolicyTestCase struct {
	AccessPolicy              string
	ShouldAdminHaveAccess     bool
	ShouldUserHaveAccess      bool
	ShouldAnonymousHaveAccess bool
}

func hasVisibleSampleApp(apps []apps_basic.AppDto) bool {
	for _, app := range apps {
		if app.AppName == tools.SampleApp {
			return true
		}
	}
	return false
}

func GetClientAndLogin(t *testing.T) *api_client.QuollixClient {
	client := api_client.NewQuollixClient()
	assert.Nil(t, client.Auth.SignIn(tools.DefaultAdminName, tools.DefaultAdminPassword))
	return client
}

func InviteUserAndSetPassword(t *testing.T, adminClient *api_client.QuollixClient, username, password, email string) {
	assert.Nil(t, adminClient.Users.Invite(username, email))
	user := GetRequiredUserByUsername(t, adminClient, username)
	assert.Equal(t, 64, len(user.SetPasswordToken))

	anonymousClient := api_client.NewQuollixClient()
	assert.Nil(t, anonymousClient.Users.SetPasswordViaToken(password, user.SetPasswordToken))
}

func GetAppClient(t *testing.T, quollixClient *api_client.QuollixClient) *api_client.QuollixClient {
	appClient := api_client.NewQuollixClient()
	quollixCookie := *quollixClient.Parent.Cookie
	appClient.Parent.Cookie = &quollixCookie
	secret, err := quollixClient.AppAccess.GetSecret()
	assert.Nil(t, err)
	assert.Nil(t, ExchangeAppAccessSecretForCookie(appClient, secret))
	return appClient
}

func RunAccessPoliciesTest(t *testing.T, adminClient *api_client.QuollixClient, testCases []AccessPolicyTestCase) {
	defer func() {
		assert.Nil(t, adminClient.Test.ResetTestState())
	}()

	anonymousClient := api_client.NewQuollixClient()
	InviteUserAndSetPassword(t, adminClient, SampleUsername, "userpassword", SampleUserEmail)

	userClient := api_client.NewQuollixClient()
	assert.Nil(t, userClient.Auth.SignIn(SampleUsername, "userpassword"))

	app, err := InstallSample(t, adminClient, "2.0")
	assert.Nil(t, err)
	assert.Nil(t, adminClient.Apps.Start(app.AppId))

	type actor struct {
		name        string
		client      *api_client.QuollixClient
		isAnonymous bool
	}

	actors := []actor{
		{"admin", adminClient, false},
		{"user", userClient, false},
		{"anonymous", anonymousClient, true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.AccessPolicy, func(t *testing.T) {
			assert.Nil(t, adminClient.Apps.SetAccessPolicy(app.AppId, testCase.AccessPolicy))
			app = GetInstalledSample(t, adminClient)
			assert.Equal(t, testCase.AccessPolicy, app.AccessPolicy)

			expectedAccessByActorName := map[string]bool{
				"admin":     testCase.ShouldAdminHaveAccess,
				"user":      testCase.ShouldUserHaveAccess,
				"anonymous": testCase.ShouldAnonymousHaveAccess,
			}

			for _, currentActor := range actors {
				assert.Equal(t, expectedAccessByActorName[currentActor.name], hasVisibleSampleApp(ListInstalledApps(t, currentActor.client)))

				contentClient := currentActor.client
				if !currentActor.isAnonymous {
					contentClient = GetAppClient(t, currentActor.client)
				}
				err := AssertSampleAppDefaultContent(contentClient, currentActor.isAnonymous)
				if expectedAccessByActorName[currentActor.name] {
					assert.Nil(t, err)
				} else {
					u.AssertDeepStackErrorFromRequest(t, err, apps_basic.AccessDeniedError)
				}
			}
		})
	}
}

func ListUsers(t *testing.T, client *api_client.QuollixClient) []tools.User {
	users, err := client.Users.List()
	assert.Nil(t, err)
	return users
}

func GetRequiredUserByUsername(t *testing.T, client *api_client.QuollixClient, username string) tools.User {
	user, exists, err := client.Users.GetByUsername(username)
	assert.Nil(t, err)
	assert.True(t, exists)
	return user
}

func ListInstalledApps(t *testing.T, client *api_client.QuollixClient) []apps_basic.AppDto {
	apps, err := client.Apps.ListInstalled()
	assert.Nil(t, err)
	return apps
}

func InstallSample(t *testing.T, client *api_client.QuollixClient, version string) (*apps_basic.AppDto, error) {
	if err := client.Apps.InstallFromStore(tools.SampleMaintainer, tools.SampleApp, version); err != nil {
		return nil, err
	}
	return GetInstalledSample(t, client), nil
}

func InstallAndStartSample(t *testing.T, client *api_client.QuollixClient, version string) *apps_basic.AppDto {
	app, err := InstallSample(t, client, version)
	assert.Nil(t, err)
	assert.Nil(t, client.Apps.Start(app.AppId))
	return app
}

func GetInstalledSample(t *testing.T, client *api_client.QuollixClient) *apps_basic.AppDto {
	for _, app := range ListInstalledApps(t, client) {
		if app.AppName == tools.SampleApp {
			return &app
		}
	}
	assert.Nil(t, u.Logger.NewError("sample app not found"))
	return nil
}

func FindVersion(t *testing.T, client *api_client.QuollixClient, userName, appName, versionName string) (*store.LeanVersionDto, error) {
	versions, err := client.Apps.ListVersions(userName, appName)
	assert.Nil(t, err)
	for _, version := range versions {
		if version.Name == versionName {
			return &version, nil
		}
	}
	return nil, u.Logger.NewError("version not found")
}

func CreateSampleBackup(t *testing.T, client *api_client.QuollixClient) error {
	return client.Backups.Create(GetInstalledSample(t, client).AppId)
}

func GetOnlyOidcAuthProvider(t *testing.T, client *api_client.QuollixClient) oidc_client.OidcAuthProviderDto {
	providers, err := client.OidcProviders.List()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(providers))
	return providers[0]
}

func GetOnlyOidcRelyingParty(t *testing.T, client *api_client.QuollixClient) oidc_provider.OidcRelyingPartyDto {
	clients, err := client.OidcClients.List()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(clients))
	return clients[0]
}

func AssertBundleParsesAsTlsKeyPair(t *testing.T, pemBundleBytes []byte) {
	certPem, keyPem, err := certificates.SplitPemBundle(pemBundleBytes)
	assert.Nil(t, err)

	_, err = tls.X509KeyPair(certPem, keyPem)
	assert.Nil(t, err)
}
