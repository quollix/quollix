//go:build component

package component

import (
	"net/http"
	"testing"
	"time"

	"server/apps_basic"
	"server/ingress"
	"server/tests/api_client"
	"server/tools"
	"server/users"

	"github.com/quollix/common/assert"
	u "github.com/quollix/common/utils"
)

func TestSecretGeneration(t *testing.T) {
	cloud := GetClientAndLogin(t)
	defer cloud.Test.ResetTestState()
	secret, err := cloud.AppAccess.GetSecret()
	assert.Nil(t, err)
	assert.Equal(t, 64, len(secret))

	secret2, _ := cloud.AppAccess.GetSecret()
	assert.NotEqual(t, secret, secret2)
}

func TestOriginPolicyActive(t *testing.T) {
	client := api_client.NewQuollixClient()
	client.Parent.Origin = "http://other-domain.com"
	_, err := client.Parent.DoRequest("/api/hello", nil)
	assert.NotNil(t, err)
	u.AssertDeepStackErrorFromRequest(t, err, ingress.CrossRequestsToBrandAppNotAllowedErrorMessage)

	client = api_client.NewQuollixClient()
	client.Parent.Origin = "other-domain.com"
	_, err = client.Parent.DoRequest("/api/hello", nil)
	assert.NotNil(t, err)
	u.AssertDeepStackErrorFromRequest(t, err, ingress.InvalidOriginHeader)
}

func TestSignInHandlerSecurity(t *testing.T) {
	cloud := GetClientAndLogin(t)
	defer cloud.Test.ResetTestState()

	err := cloud.Auth.SignIn(tools.DefaultAdminName, "wrongpassword")
	assert.NotNil(t, err)
	u.AssertDeepStackErrorFromRequest(t, err, users.IncorrectLoginCredentialsError)

	err = cloud.Auth.SignIn("wrongadmin", tools.DefaultAdminPassword)
	assert.NotNil(t, err)
	u.AssertDeepStackErrorFromRequest(t, err, users.IncorrectLoginCredentialsError)
}

func TestSignInHandlerInputValidation(t *testing.T) {
	cloud := GetClientAndLogin(t)
	defer cloud.Test.ResetTestState()

	err := cloud.Auth.SignIn(tools.DefaultAdminName, "short")
	assert.NotNil(t, err)
	u.AssertInvalidInputError(t, err)

	err = cloud.Auth.SignIn("user!@#$", tools.DefaultAdminPassword)
	assert.NotNil(t, err)
	u.AssertInvalidInputError(t, err)
}

func TestCookieValidation(t *testing.T) {
	cloud := GetClientAndLogin(t)
	defer cloud.Test.ResetTestState()
	cloud.Parent.Cookie.Value += "a"

	err := cloud.Users.Invite(tools.DefaultAdminName, tools.DefaultAdminEmail)
	assert.NotNil(t, err)
	u.AssertInvalidInputError(t, err)
}

func TestSecretIsDeletedAfterExchangeAgainstCookie(t *testing.T) {
	cloud := GetClientAndLogin(t)
	defer cloud.Test.ResetTestState()
	app, err := InstallSample(t, cloud, "2.0")
	assert.Nil(t, err)
	assert.Nil(t, cloud.Apps.Start(app.AppId))
	secret, err := cloud.AppAccess.GetSecret()
	assert.Nil(t, err)
	err = AssertSampleAppContentUsingSecret(cloud, secret)
	assert.Nil(t, err)
	assert.NotEqual(t, secret, cloud.Parent.Cookie.Value)

	err = AssertSampleAppContentUsingSecret(cloud, secret)
	assert.NotNil(t, err)
	u.AssertDeepStackErrorFromRequest(t, err, "secret does not exist")
}

func TestSecretValidation(t *testing.T) {
	cloud := GetClientAndLogin(t)
	defer cloud.Test.ResetTestState()

	app, err := InstallSample(t, cloud, "2.0")
	assert.Nil(t, err)
	assert.Nil(t, cloud.Apps.Start(app.AppId))

	randomSecret := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	err = AssertSampleAppContentUsingSecret(cloud, randomSecret)
	assert.NotNil(t, err)
	u.AssertDeepStackErrorFromRequest(t, err, "secret does not exist")
}

func TestSecretsAreRandom(t *testing.T) {
	cloud := GetClientAndLogin(t)
	defer cloud.Test.ResetTestState()

	firstSecret, err := cloud.AppAccess.GetSecret()
	assert.Nil(t, err)
	secondSecret, err := cloud.AppAccess.GetSecret()
	assert.Nil(t, err)
	assert.Equal(t, len(firstSecret), len(secondSecret))
	assert.NotEqual(t, firstSecret, secondSecret)
}

func TestCookiesAreRandom(t *testing.T) {
	cloud := api_client.NewQuollixClient()
	defer cloud.Test.ResetTestState()
	assert.Nil(t, cloud.Auth.SignIn(tools.DefaultAdminName, tools.DefaultAdminPassword))

	cookie1 := cloud.Parent.Cookie.Value
	assert.Nil(t, cloud.Auth.SignIn(tools.DefaultAdminName, tools.DefaultAdminPassword))
	cookie2 := cloud.Parent.Cookie.Value
	assert.Equal(t, len(cookie1), len(cookie2))
	assert.NotEqual(t, cookie1, cookie2)
}

func TestAppSearchValidation(t *testing.T) {
	cloud := GetClientAndLogin(t)
	defer cloud.Test.ResetTestState()
	_, err := cloud.Apps.SearchStore("", "ab!", true)
	assert.NotNil(t, err)
	u.AssertInvalidInputError(t, err)
}

func TestRoleVerificationForEndpoints(t *testing.T) {
	adminClient := GetClientAndLogin(t)
	defer adminClient.Test.ResetTestState()
	InviteUserAndSetPassword(t, adminClient, SampleUsername, "password", SampleUserEmail)

	userClient := api_client.NewQuollixClient()
	assert.Nil(t, userClient.Auth.SignIn(SampleUsername, "password"))
	err := userClient.Apps.Start("123")
	assert.NotNil(t, err)
	u.AssertDeepStackErrorFromRequest(t, err, "Unauthorized")

	anonymousClient := api_client.NewQuollixClient() // no cookie is set
	err = anonymousClient.Apps.Start("123")
	assert.NotNil(t, err)
	u.AssertDeepStackErrorFromRequest(t, err, users.CookieNotFoundError)
}

func TestNullOriginHeaderIsAllowed(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()
	client.Parent.Origin = "null"
	InviteUserAndSetPassword(t, client, SampleUsername, "password", SampleUserEmail)
}

func TestLogin(t *testing.T) {
	client := api_client.NewQuollixClient()
	defer client.Test.ResetTestState()
	assert.Nil(t, client.Parent.Cookie)
	assert.Nil(t, client.Auth.SignIn(tools.DefaultAdminName, tools.DefaultAdminPassword))
	cookie := client.Parent.Cookie
	assert.NotNil(t, cookie)
	assert.Equal(t, 64, len(cookie.Value))
	assert.True(t, cookie.Expires.After(time.Now().AddDate(0, 0, tools.CookieExpirationTimeInDays-1)))
	assert.True(t, cookie.Expires.Before(time.Now().AddDate(0, 0, tools.CookieExpirationTimeInDays+1)))
	assert.True(t, cookie.HttpOnly)
}

func TestCookieExpirationDateRenewalWhenCheckingAuth(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	currentUser, err := client.Auth.GetCurrentUser()
	assert.Nil(t, err)
	cookieExpiration1 := currentUser.CookieExpirationDate
	assert.True(t, cookieExpiration1.After(time.Now().AddDate(0, 0, tools.CookieExpirationTimeInDays-1)))
	assert.True(t, cookieExpiration1.Before(time.Now().AddDate(0, 0, tools.CookieExpirationTimeInDays+1)))

	tools.WaitOneSecond()
	currentUser, err = client.Auth.GetCurrentUser()
	assert.Nil(t, err)
	cookieExpiration2 := currentUser.CookieExpirationDate
	assert.NotEqual(t, cookieExpiration1, cookieExpiration2)
	assert.True(t, cookieExpiration1.Before(cookieExpiration2))
}

func TestSecureCookieFlagsPresence(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	assertSecureCookieFlags(t, client.Parent.Cookie)

	InstallAndStartSample(t, client, "2.0")
	secret, err := client.AppAccess.GetSecret()
	assert.Nil(t, err)
	client.Parent.Cookie = nil
	proxyCookie, err := client.AppAccess.ExchangeSecretForAppAccessCookie(secret, sampleAppHttpsUrl+sampleEndpoint)
	assert.Nil(t, err)
	assertSecureCookieFlags(t, proxyCookie)
}

func assertSecureCookieFlags(t *testing.T, cookie *http.Cookie) {
	assert.NotNil(t, cookie)
	assert.True(t, cookie.HttpOnly)
	assert.True(t, cookie.Secure)
	assert.Equal(t, "/", cookie.Path)
	assert.Equal(t, http.SameSiteLaxMode, cookie.SameSite)
}

func TestAppSessionCookieIsSeparatedFromQuollixSessionCookie(t *testing.T) {
	quollixClient := GetClientAndLogin(t)
	defer quollixClient.Test.ResetTestState()
	app, err := InstallSample(t, quollixClient, "2.0")
	assert.Nil(t, err)
	assert.Nil(t, quollixClient.Apps.Start(app.AppId))
	assert.Nil(t, quollixClient.Apps.SetAccessPolicy(app.AppId, tools.Policies.AuthenticatedAccessPolicy))

	quollixCookie := *quollixClient.Parent.Cookie
	err = AssertSampleAppContentWithCookie(&quollixCookie)
	assert.NotNil(t, err)
	u.AssertDeepStackErrorFromRequest(t, err, apps_basic.AccessDeniedError)

	authenticatedUser, err := quollixClient.Auth.GetCurrentUser()
	assert.Nil(t, err)
	assert.Equal(t, tools.DefaultAdminName, authenticatedUser.Username)

	appClient := GetAppClient(t, quollixClient)
	appCookie := *appClient.Parent.Cookie
	assert.NotEqual(t, quollixCookie.Value, appCookie.Value)
	assert.Nil(t, AssertSampleAppContentWithCookie(&appCookie))

	_, err = appClient.Auth.GetCurrentUser()
	assert.NotNil(t, err)
	u.AssertDeepStackErrorFromRequest(t, err, users.CookieNotFoundError)
}

func TestLogoutInvalidatesQuollixAndAppSessions(t *testing.T) {
	quollixClient := GetClientAndLogin(t)
	defer quollixClient.Test.ResetTestState()
	prepareAuthenticatedSampleApp(t, quollixClient)

	appClient := GetAppClient(t, quollixClient)
	appCookie := *appClient.Parent.Cookie

	assert.Nil(t, quollixClient.Auth.SignOut())
	_, err := quollixClient.Auth.GetCurrentUser()
	assert.NotNil(t, err)
	u.AssertDeepStackErrorFromRequest(t, err, users.CookieNotFoundError)

	err = AssertSampleAppContentWithCookie(&appCookie)
	assert.NotNil(t, err)
	u.AssertDeepStackErrorFromRequest(t, err, apps_basic.AccessDeniedError)
}

func TestUserDeletionInvalidatesQuollixAndAppSessions(t *testing.T) {
	adminClient := GetClientAndLogin(t)
	defer adminClient.Test.ResetTestState()
	prepareAuthenticatedSampleApp(t, adminClient)
	InviteUserAndSetPassword(t, adminClient, SampleUsername, SampleUserPassword, SampleUserEmail)

	userClient := api_client.NewQuollixClient()
	assert.Nil(t, userClient.Auth.SignIn(SampleUsername, SampleUserPassword))
	appClient := GetAppClient(t, userClient)
	appCookie := *appClient.Parent.Cookie
	user := GetRequiredUserByUsername(t, adminClient, SampleUsername)

	assert.Nil(t, adminClient.Users.Delete(user.Id))
	_, err := userClient.Auth.GetCurrentUser()
	assert.NotNil(t, err)
	u.AssertDeepStackErrorFromRequest(t, err, users.CookieNotFoundError)

	err = AssertSampleAppContentWithCookie(&appCookie)
	assert.NotNil(t, err)
	u.AssertDeepStackErrorFromRequest(t, err, apps_basic.AccessDeniedError)
}

func TestPasswordResetInvalidatesQuollixAndAppSessions(t *testing.T) {
	adminClient := GetClientAndLogin(t)
	defer adminClient.Test.ResetTestState()
	prepareAuthenticatedSampleApp(t, adminClient)
	InviteUserAndSetPassword(t, adminClient, SampleUsername, SampleUserPassword, SampleUserEmail)

	userClient := api_client.NewQuollixClient()
	assert.Nil(t, userClient.Auth.SignIn(SampleUsername, SampleUserPassword))
	appClient := GetAppClient(t, userClient)
	appCookie := *appClient.Parent.Cookie
	user := GetRequiredUserByUsername(t, adminClient, SampleUsername)

	assert.Nil(t, adminClient.Users.ResetPassword(user.Id))
	_, err := userClient.Auth.GetCurrentUser()
	assert.NotNil(t, err)
	u.AssertDeepStackErrorFromRequest(t, err, users.CookieNotFoundError)

	err = AssertSampleAppContentWithCookie(&appCookie)
	assert.NotNil(t, err)
	u.AssertDeepStackErrorFromRequest(t, err, apps_basic.AccessDeniedError)
}

func prepareAuthenticatedSampleApp(t *testing.T, client *api_client.QuollixClient) {
	app, err := InstallSample(t, client, "2.0")
	assert.Nil(t, err)
	assert.Nil(t, client.Apps.Start(app.AppId))
	assert.Nil(t, client.Apps.SetAccessPolicy(app.AppId, tools.Policies.AuthenticatedAccessPolicy))
}
