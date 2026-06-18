package component

import (
	"crypto/tls"
	"server/apps_basic"
	"server/certificates"
	"server/tools"
	"testing"

	"github.com/quollix/common/assert"
	u "github.com/quollix/common/utils"
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

func InviteUserAndSetPassword(adminClient *QuollixClient, username, password, email string) {
	assert.Nil(adminClient.T, adminClient.Users.Invite(username, email))
	user := adminClient.Users.GetByUsername(username)
	assert.Equal(adminClient.T, 64, len(user.SetPasswordToken))

	anonymousClient := GetQuollixClient(adminClient.T)
	assert.Nil(adminClient.T, anonymousClient.Users.SetPasswordViaToken(password, user.SetPasswordToken))
}

func GetAppClient(t *testing.T, quollixClient *QuollixClient) *QuollixClient {
	appClient := GetQuollixClient(t)
	quollixCookie := *quollixClient.Parent.Cookie
	appClient.Parent.Cookie = &quollixCookie
	secret, err := quollixClient.Content.GetSecret()
	assert.Nil(t, err)
	assert.Nil(t, appClient.Content.ExchangeSecretForAppCookie(secret))
	return appClient
}

func RunAccessPoliciesTest(t *testing.T, adminClient *QuollixClient, testCases []AccessPolicyTestCase) {
	defer adminClient.Test.ResetTestState()

	anonymousClient := GetQuollixClient(t)
	InviteUserAndSetPassword(adminClient, SampleUsername, "userpassword", SampleUserEmail)

	userClient := GetQuollixClient(t)
	assert.Nil(t, userClient.Auth.Login(SampleUsername, "userpassword"))

	app, err := adminClient.Apps.InstallSample("2.0")
	assert.Nil(t, err)
	assert.Nil(t, adminClient.Apps.Start(app.AppId))

	type actor struct {
		name        string
		client      *QuollixClient
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
			app = adminClient.Apps.GetInstalledSample()
			assert.Equal(t, testCase.AccessPolicy, app.AccessPolicy)

			expectedAccessByActorName := map[string]bool{
				"admin":     testCase.ShouldAdminHaveAccess,
				"user":      testCase.ShouldUserHaveAccess,
				"anonymous": testCase.ShouldAnonymousHaveAccess,
			}

			for _, currentActor := range actors {
				assert.Equal(t, expectedAccessByActorName[currentActor.name], hasVisibleSampleApp(currentActor.client.Apps.ListInstalled()))

				contentClient := currentActor.client
				if !currentActor.isAnonymous {
					contentClient = GetAppClient(t, currentActor.client)
				}
				err := contentClient.Content.AssertSampleContent(currentActor.isAnonymous)
				if expectedAccessByActorName[currentActor.name] {
					assert.Nil(t, err)
				} else {
					u.AssertDeepStackErrorFromRequest(t, err, apps_basic.AccessDeniedError)
				}
			}
		})
	}
}

func assertBundleParsesAsTlsKeyPair(t *testing.T, pemBundleBytes []byte) {
	certPem, keyPem, err := certificates.SplitPemBundle(pemBundleBytes)
	assert.Nil(t, err)

	_, err = tls.X509KeyPair(certPem, keyPem)
	assert.Nil(t, err)
}
