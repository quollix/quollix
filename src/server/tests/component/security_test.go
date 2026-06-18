//go:build component

package component

import (
	"net/http"
	"server/apps_basic"
	"server/ingress"
	"server/tools"
	"server/users"
	"testing"
	"time"

	"crypto/tls"

	"github.com/quollix/common/assert"
	u "github.com/quollix/common/utils"
)

func TestSecretGeneration(t *testing.T) {
	cloud := GetClientAndLogin(t)
	defer cloud.Test.ResetTestState()
	secret, err := cloud.Content.GetSecret()
	assert.Nil(t, err)
	assert.Equal(t, 64, len(secret))

	secret2, _ := cloud.Content.GetSecret()
	assert.NotEqual(t, secret, secret2)
}

func TestOriginPolicyActive(t *testing.T) {
	client := GetQuollixClient(t)
	client.Parent.Origin = "http://other-domain.com"
	_, err := client.Parent.DoRequest("/api/hello", nil)
	assert.NotNil(t, err)
	u.AssertDeepStackErrorFromRequest(t, err, ingress.CrossRequestsToBrandAppNotAllowedErrorMessage)

	client = GetQuollixClient(t)
	client.Parent.Origin = "other-domain.com"
	_, err = client.Parent.DoRequest("/api/hello", nil)
	assert.NotNil(t, err)
	u.AssertDeepStackErrorFromRequest(t, err, ingress.InvalidOriginHeader)
}

func TestLoginHandlerSecurity(t *testing.T) {
	cloud := GetClientAndLogin(t)
	defer cloud.Test.ResetTestState()

	err := cloud.Auth.Login(tools.DefaultAdminName, "wrongpassword")
	assert.NotNil(t, err)
	u.AssertDeepStackErrorFromRequest(t, err, users.IncorrectLoginCredentialsError)

	err = cloud.Auth.Login("wrongadmin", tools.DefaultAdminPassword)
	assert.NotNil(t, err)
	u.AssertDeepStackErrorFromRequest(t, err, users.IncorrectLoginCredentialsError)
}

func TestLoginHandlerInputValidation(t *testing.T) {
	cloud := GetClientAndLogin(t)
	defer cloud.Test.ResetTestState()

	err := cloud.Auth.Login(tools.DefaultAdminName, "short")
	assert.NotNil(t, err)
	u.AssertInvalidInputError(t, err)

	err = cloud.Auth.Login("user!@#$", tools.DefaultAdminPassword)
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
	app, err := cloud.Apps.InstallSample("2.0")
	assert.Nil(t, err)
	assert.Nil(t, cloud.Apps.Start(app.AppId))
	secret, err := cloud.Content.GetSecret()
	assert.Nil(t, err)
	err = cloud.Content.AssertContentUsingSecret(secret)
	assert.Nil(t, err)
	assert.NotEqual(t, secret, cloud.Parent.Cookie.Value)

	err = cloud.Content.AssertContentUsingSecret(secret)
	assert.NotNil(t, err)
	u.AssertDeepStackErrorFromRequest(t, err, "secret does not exist")
}

func TestSecretValidation(t *testing.T) {
	cloud := GetClientAndLogin(t)
	defer cloud.Test.ResetTestState()

	app, err := cloud.Apps.InstallSample("2.0")
	assert.Nil(t, err)
	assert.Nil(t, cloud.Apps.Start(app.AppId))

	randomSecret := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	err = cloud.Content.AssertContentUsingSecret(randomSecret)
	assert.NotNil(t, err)
	u.AssertDeepStackErrorFromRequest(t, err, "secret does not exist")
}

func TestSecretsAreRandom(t *testing.T) {
	cloud := GetClientAndLogin(t)
	defer cloud.Test.ResetTestState()

	firstSecret, err := cloud.Content.GetSecret()
	assert.Nil(t, err)
	secondSecret, err := cloud.Content.GetSecret()
	assert.Nil(t, err)
	assert.Equal(t, len(firstSecret), len(secondSecret))
	assert.NotEqual(t, firstSecret, secondSecret)
}

func TestCookiesAreRandom(t *testing.T) {
	cloud := GetQuollixClient(t)
	defer cloud.Test.ResetTestState()
	assert.Nil(t, cloud.Auth.Login(tools.DefaultAdminName, tools.DefaultAdminPassword))

	cookie1 := cloud.Parent.Cookie.Value
	assert.Nil(t, cloud.Auth.Login(tools.DefaultAdminName, tools.DefaultAdminPassword))
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
	InviteUserAndSetPassword(adminClient, SampleUsername, "password", SampleUserEmail)

	userClient := GetQuollixClient(t)
	assert.Nil(t, userClient.Auth.Login(SampleUsername, "password"))
	err := userClient.Apps.Start("123")
	assert.NotNil(t, err)
	u.AssertDeepStackErrorFromRequest(t, err, "Unauthorized")

	anonymousClient := GetQuollixClient(t) // no cookie is set
	err = anonymousClient.Apps.Start("123")
	assert.NotNil(t, err)
	u.AssertDeepStackErrorFromRequest(t, err, users.CookieNotFoundError)
}

func TestNullOriginHeaderIsAllowed(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()
	client.Parent.Origin = "null"
	InviteUserAndSetPassword(client, SampleUsername, "password", SampleUserEmail)
}

func TestLogin(t *testing.T) {
	client := GetQuollixClient(t)
	defer client.Test.ResetTestState()
	assert.Nil(t, client.Parent.Cookie)
	assert.Nil(t, client.Auth.Login(tools.DefaultAdminName, tools.DefaultAdminPassword))
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

	sampleApp, err := client.Apps.InstallSample("2.0")
	assert.Nil(t, err)
	assert.Nil(t, client.Apps.Start(sampleApp.AppId))
	secret, err := client.Content.GetSecret()
	assert.Nil(t, err)
	client.Parent.Cookie = nil
	rawClient := http.Client{
		CheckRedirect: func(request *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	resp, err := rawClient.Get(getAppRequestUrlWithSecret(sampleEndpoint, secret))
	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, 1, len(resp.Cookies()))
	proxyCookie := resp.Cookies()[0]
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
	app, err := quollixClient.Apps.InstallSample("2.0")
	assert.Nil(t, err)
	assert.Nil(t, quollixClient.Apps.Start(app.AppId))
	assert.Nil(t, quollixClient.Apps.SetAccessPolicy(app.AppId, tools.Policies.AuthenticatedAccessPolicy))

	quollixCookie := *quollixClient.Parent.Cookie
	err = quollixClient.Content.AssertSampleContentWithCookie(&quollixCookie)
	assert.NotNil(t, err)
	u.AssertDeepStackErrorFromRequest(t, err, apps_basic.AccessDeniedError)

	authenticatedUser, err := quollixClient.Auth.GetCurrentUser()
	assert.Nil(t, err)
	assert.Equal(t, tools.DefaultAdminName, authenticatedUser.Username)

	appClient := GetAppClient(t, quollixClient)
	appCookie := *appClient.Parent.Cookie
	assert.NotEqual(t, quollixCookie.Value, appCookie.Value)
	assert.Nil(t, appClient.Content.AssertSampleContentWithCookie(&appCookie))

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

	assert.Nil(t, quollixClient.Auth.Logout())
	_, err := quollixClient.Auth.GetCurrentUser()
	assert.NotNil(t, err)
	u.AssertDeepStackErrorFromRequest(t, err, users.CookieNotFoundError)

	err = appClient.Content.AssertSampleContentWithCookie(&appCookie)
	assert.NotNil(t, err)
	u.AssertDeepStackErrorFromRequest(t, err, apps_basic.AccessDeniedError)
}

func TestUserDeletionInvalidatesQuollixAndAppSessions(t *testing.T) {
	adminClient := GetClientAndLogin(t)
	defer adminClient.Test.ResetTestState()
	prepareAuthenticatedSampleApp(t, adminClient)
	InviteUserAndSetPassword(adminClient, SampleUsername, SampleUserPassword, SampleUserEmail)

	userClient := GetQuollixClient(t)
	assert.Nil(t, userClient.Auth.Login(SampleUsername, SampleUserPassword))
	appClient := GetAppClient(t, userClient)
	appCookie := *appClient.Parent.Cookie
	user := adminClient.Users.GetByUsername(SampleUsername)

	assert.Nil(t, adminClient.Users.Delete(user.Id))
	_, err := userClient.Auth.GetCurrentUser()
	assert.NotNil(t, err)
	u.AssertDeepStackErrorFromRequest(t, err, users.CookieNotFoundError)

	err = appClient.Content.AssertSampleContentWithCookie(&appCookie)
	assert.NotNil(t, err)
	u.AssertDeepStackErrorFromRequest(t, err, apps_basic.AccessDeniedError)
}

func TestPasswordResetInvalidatesQuollixAndAppSessions(t *testing.T) {
	adminClient := GetClientAndLogin(t)
	defer adminClient.Test.ResetTestState()
	prepareAuthenticatedSampleApp(t, adminClient)
	InviteUserAndSetPassword(adminClient, SampleUsername, SampleUserPassword, SampleUserEmail)

	userClient := GetQuollixClient(t)
	assert.Nil(t, userClient.Auth.Login(SampleUsername, SampleUserPassword))
	appClient := GetAppClient(t, userClient)
	appCookie := *appClient.Parent.Cookie
	user := adminClient.Users.GetByUsername(SampleUsername)

	assert.Nil(t, adminClient.Users.ResetPassword(user.Id))
	_, err := userClient.Auth.GetCurrentUser()
	assert.NotNil(t, err)
	u.AssertDeepStackErrorFromRequest(t, err, users.CookieNotFoundError)

	err = appClient.Content.AssertSampleContentWithCookie(&appCookie)
	assert.NotNil(t, err)
	u.AssertDeepStackErrorFromRequest(t, err, apps_basic.AccessDeniedError)
}

func prepareAuthenticatedSampleApp(t *testing.T, client *QuollixClient) {
	app, err := client.Apps.InstallSample("2.0")
	assert.Nil(t, err)
	assert.Nil(t, client.Apps.Start(app.AppId))
	assert.Nil(t, client.Apps.SetAccessPolicy(app.AppId, tools.Policies.AuthenticatedAccessPolicy))
}

func TestRedirectsHttpToHttps(t *testing.T) {
	httpClient := &http.Client{
		CheckRedirect: func(request *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	httpRequest, err := http.NewRequest(http.MethodGet, "http://localhost/some/path?value=1", nil)
	assert.Nil(t, err)

	httpResponse, err := httpClient.Do(httpRequest)
	assert.Nil(t, err)
	defer u.Close(httpResponse.Body)

	assert.Equal(t, http.StatusPermanentRedirect, httpResponse.StatusCode)

	expectedLocation := "https://localhost/some/path?value=1"
	actualLocation := httpResponse.Header.Get("Location")
	assert.Equal(t, expectedLocation, actualLocation)
}
