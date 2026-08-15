//go:build component

package component

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"server/apps_basic"
	"server/ingress"
	"server/tools"
	"server/users"

	"github.com/quollix/common/quollix/api"
	"github.com/quollix/common/quollix/api_client"

	"github.com/quollix/common/assert"
	u "github.com/quollix/common/utils"
)

func TestSecretGeneration(t *testing.T) {
	cloud := GetClientAndLogin(t)
	defer cloud.Test.ResetTestState()

	_, err := InstallSample(t, cloud, "2.0")
	assert.Nil(t, err)

	secret, err := cloud.AppAccess.GetSecret(tools.SampleApp)
	assert.Nil(t, err)
	assert.Equal(t, 64, len(secret))

	secret2, err := cloud.AppAccess.GetSecret(tools.SampleApp)
	assert.Nil(t, err)
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
	_, err := InstallAndStartSample(t, cloud, "2.0")
	assert.Nil(t, err)
	secret, err := cloud.AppAccess.GetSecret(tools.SampleApp)
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

	_, err := InstallSample(t, cloud, "2.0")
	assert.Nil(t, err)

	randomSecret := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	err = AssertSampleAppContentUsingSecret(cloud, randomSecret)
	assert.NotNil(t, err)
	u.AssertDeepStackErrorFromRequest(t, err, "secret does not exist")
}

func TestAppOpenSecretDoesNotGrantAccessToDifferentAccessibleApp(t *testing.T) {
	adminClient := GetClientAndLogin(t)
	defer adminClient.Test.ResetTestState()
	sampleApp, sampleApp2 := prepareTwoRunningSampleApps(t, adminClient)
	assert.Nil(t, adminClient.Apps.SetAccessPolicy(sampleApp.AppId, api.Policies.AuthenticatedAccessPolicy))
	assert.Nil(t, adminClient.Apps.SetAccessPolicy(sampleApp2.AppId, api.Policies.AuthenticatedAccessPolicy))

	InviteUserAndSetPassword(t, adminClient, SampleUsername, SampleUserPassword, SampleUserEmail)
	userClient := api_client.NewQuollixClient()
	assert.Nil(t, userClient.Auth.SignIn(SampleUsername, SampleUserPassword))

	secret := getAppOpenSecret(t, userClient, tools.SampleApp)

	assertAppOpenSecretCannotBeExchangedWithUrl(t, userClient, secret, sampleApp2HttpsUrl+sampleEndpoint)
}

func TestAppOpenSecretDoesNotGrantAccessToDifferentRestrictedApp(t *testing.T) {
	adminClient := GetClientAndLogin(t)
	defer adminClient.Test.ResetTestState()
	sampleApp, sampleApp2 := prepareTwoRunningSampleApps(t, adminClient)
	assert.Nil(t, adminClient.Apps.SetAccessPolicy(sampleApp.AppId, api.Policies.AuthenticatedAccessPolicy))
	assert.Nil(t, adminClient.Apps.SetAccessPolicy(sampleApp2.AppId, api.Policies.AdminOnlyAccessPolicy))

	InviteUserAndSetPassword(t, adminClient, SampleUsername, SampleUserPassword, SampleUserEmail)
	userClient := api_client.NewQuollixClient()
	assert.Nil(t, userClient.Auth.SignIn(SampleUsername, SampleUserPassword))

	secret := getAppOpenSecret(t, userClient, tools.SampleApp)

	assertAppOpenSecretCannotBeExchangedWithUrl(t, userClient, secret, sampleApp2HttpsUrl+sampleEndpoint)
}

func TestUserCannotGetAppOpenSecretForRestrictedApp(t *testing.T) {
	adminClient := GetClientAndLogin(t)
	defer adminClient.Test.ResetTestState()
	_, sampleApp2 := prepareTwoRunningSampleApps(t, adminClient)
	assert.Nil(t, adminClient.Apps.SetAccessPolicy(sampleApp2.AppId, api.Policies.AdminOnlyAccessPolicy))

	InviteUserAndSetPassword(t, adminClient, SampleUsername, SampleUserPassword, SampleUserEmail)
	userClient := api_client.NewQuollixClient()
	assert.Nil(t, userClient.Auth.SignIn(SampleUsername, SampleUserPassword))

	_, err := userClient.AppAccess.GetSecret(sampleApp2Name)
	assert.NotNil(t, err)
	u.AssertDeepStackErrorFromRequest(t, err, apps_basic.AccessDeniedError)
}

func TestUserCannotGetFrontendAppOpenSecretForRestrictedApp(t *testing.T) {
	adminClient := GetClientAndLogin(t)
	defer adminClient.Test.ResetTestState()
	sampleApp, err := InstallAndStartSample(t, adminClient, tools.SampleAppVersion2Name)
	assert.Nil(t, err)
	assert.Nil(t, adminClient.Apps.SetAccessPolicy(sampleApp.AppId, api.Policies.AdminOnlyAccessPolicy))

	InviteUserAndSetPassword(t, adminClient, SampleUsername, SampleUserPassword, SampleUserEmail)
	userClient := api_client.NewQuollixClient()
	assert.Nil(t, userClient.Auth.SignIn(SampleUsername, SampleUserPassword))

	response := getAppOpenResponse(t, userClient, tools.SampleApp)

	assertAppUnavailablePage(t, response.StatusCode, response.Body)
	assert.Equal(t, "", response.Header.Get("Location"))
}

func TestAppPolicyChangeRevokesExistingUserAppSessionCookie(t *testing.T) {
	adminClient := GetClientAndLogin(t)
	defer adminClient.Test.ResetTestState()
	sampleApp, err := InstallAndStartSample(t, adminClient, tools.SampleAppVersion2Name)
	assert.Nil(t, err)
	assert.Nil(t, adminClient.Apps.SetAccessPolicy(sampleApp.AppId, api.Policies.AuthenticatedAccessPolicy))

	InviteUserAndSetPassword(t, adminClient, SampleUsername, SampleUserPassword, SampleUserEmail)
	userClient := api_client.NewQuollixClient()
	assert.Nil(t, userClient.Auth.SignIn(SampleUsername, SampleUserPassword))
	appClient := GetAppClient(t, userClient)
	appCookie := *appClient.Parent.Cookie
	assert.Nil(t, AssertSampleAppContentWithCookie(&appCookie))

	assert.Nil(t, adminClient.Apps.SetAccessPolicy(sampleApp.AppId, api.Policies.AdminOnlyAccessPolicy))

	AssertSampleAppUnavailablePage(t, &appCookie)
}

func TestAppPolicyChangeRevokesAnonymousPublicAccess(t *testing.T) {
	adminClient := GetClientAndLogin(t)
	defer adminClient.Test.ResetTestState()
	sampleApp, err := InstallAndStartSample(t, adminClient, tools.SampleAppVersion2Name)
	assert.Nil(t, err)
	assert.Nil(t, adminClient.Apps.SetAccessPolicy(sampleApp.AppId, api.Policies.PublicAccessPolicy))
	assert.Nil(t, AssertSampleAppDefaultContent(api_client.NewQuollixClient(), true))

	assert.Nil(t, adminClient.Apps.SetAccessPolicy(sampleApp.AppId, api.Policies.AdminOnlyAccessPolicy))

	AssertSampleAppOpenRedirect(t, nil)
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

	u.WaitMillis(1000)
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

	_, err := InstallSample(t, client, "2.0")
	assert.Nil(t, err)
	secret, err := client.AppAccess.GetSecret(tools.SampleApp)
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

const (
	sampleApp2Name     = "sampleapp2"
	sampleApp2HttpsUrl = "https://sampleapp2.localhost"
)

func prepareTwoRunningSampleApps(t *testing.T, adminClient *api_client.QuollixClient) (*api.AdminAppDto, *api.AdminAppDto) {
	sampleApp, err := InstallAndStartSample(t, adminClient, tools.SampleAppVersion2Name)
	assert.Nil(t, err)
	sampleApp2 := uploadRenamedSampleApp(t, adminClient, sampleApp2Name)
	assert.Nil(t, adminClient.Apps.Start(sampleApp2.AppId))
	return sampleApp, &sampleApp2
}

func uploadRenamedSampleApp(t *testing.T, client *api_client.QuollixClient, appName string) api.AdminAppDto {
	versionFile := api.BinaryFile{
		FileName: renamedSampleAppFileName(appName),
		Content:  renamedSampleAppContent(appName),
	}
	assert.Nil(t, client.Apps.UploadVersionFile(versionFile))
	return getRequiredInstalledAppByName(t, client, appName)
}

func renamedSampleAppFileName(appName string) string {
	return fmt.Sprintf("%s_%s_%s_%s.yml",
		tools.SampleMaintainer,
		appName,
		tools.SampleAppVersion2Name,
		tools.SampleAppVersion2CreationTimestamp.Format(apps_basic.VersionFileUploadTimestampLayout),
	)
}

func renamedSampleAppContent(appName string) []byte {
	content := strings.ReplaceAll(tools.SampleAppVersion2ComposeYAML, tools.SampleApp, appName)
	content = strings.ReplaceAll(content, "image: "+appName+":local", "image: "+tools.SampleApp+":local")
	return []byte(content)
}

func getRequiredInstalledAppByName(t *testing.T, client *api_client.QuollixClient, appName string) api.AdminAppDto {
	app, exists := findAppByName(ListInstalledApps(t, client), appName)
	assert.True(t, exists)
	return app
}

func getAppOpenSecret(t *testing.T, client *api_client.QuollixClient, appName string) string {
	response := getAppOpenResponse(t, client, appName)
	assert.Equal(t, http.StatusFound, response.StatusCode)

	redirectURL, err := url.Parse(response.Header.Get("Location"))
	assert.Nil(t, err)
	assert.Equal(t, appName+".localhost", redirectURL.Host)
	secret := redirectURL.Query().Get(api.BrandAppQuerySecretName)
	assert.Equal(t, 64, len(secret))
	return secret
}

func getAppOpenResponse(t *testing.T, client *api_client.QuollixClient, appName string) *api_client.FrontendResponse {
	appOpenPath := api.Paths.FrontendAppOpen + "?app=" + url.QueryEscape(appName) + "&path=/"
	response, err := client.Frontend.GetPage(appOpenPath)
	assert.Nil(t, err)
	return response
}

func assertAppOpenSecretCannotBeExchangedWithUrl(t *testing.T, client *api_client.QuollixClient, secret string, appURL string) {
	_, err := client.AppAccess.ExchangeSecretForAppAccessCookie(secret, appURL)
	assert.NotNil(t, err)
	u.AssertDeepStackErrorFromRequest(t, err, users.SecretDoesNotExistError)
}

func TestAppSessionCookieIsSeparatedFromQuollixSessionCookie(t *testing.T) {
	quollixClient := GetClientAndLogin(t)
	defer quollixClient.Test.ResetTestState()
	app, err := InstallSample(t, quollixClient, "2.0")
	assert.Nil(t, err)
	assert.Nil(t, quollixClient.Apps.Start(app.AppId))
	assert.Nil(t, quollixClient.Apps.SetAccessPolicy(app.AppId, api.Policies.AuthenticatedAccessPolicy))

	quollixCookie := *quollixClient.Parent.Cookie
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

	AssertSampleAppOpenRedirect(t, &appCookie)
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

	AssertSampleAppOpenRedirect(t, &appCookie)
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

	AssertSampleAppOpenRedirect(t, &appCookie)
}

func prepareAuthenticatedSampleApp(t *testing.T, client *api_client.QuollixClient) {
	app, err := InstallAndStartSample(t, client, "2.0")
	assert.Nil(t, err)
	assert.Nil(t, client.Apps.SetAccessPolicy(app.AppId, api.Policies.AuthenticatedAccessPolicy))
}
