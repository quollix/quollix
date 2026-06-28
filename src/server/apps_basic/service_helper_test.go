package apps_basic

import (
	"errors"
	"server/tools"
	"strconv"
	"testing"

	"github.com/quollix/common/assert"
	u "github.com/quollix/common/utils"
)

type appServiceHelperTestObjects struct {
	AppServiceHelper *AppServiceHelperImpl
	AuthorizerMock   *AuthorizerMock
	AppDetectorMock  *AppDetectorMock
}

func setupAppServiceHelperTest(t *testing.T) appServiceHelperTestObjects {
	authorizerMock := NewAuthorizerMock(t)
	appDetectorMock := NewAppDetectorMock(t)

	return appServiceHelperTestObjects{
		AppServiceHelper: &AppServiceHelperImpl{
			Authorizer:  authorizerMock,
			AppDetector: appDetectorMock,
		},
		AuthorizerMock:  authorizerMock,
		AppDetectorMock: appDetectorMock,
	}
}

func TestIsAppVisibleToUser_AdminAlwaysVisible(t *testing.T) {
	testObjects := setupAppServiceHelperTest(t)
	repoApp := GetSampleApp()

	isVisible := testObjects.AppServiceHelper.IsAppVisibleToUser(sampleUserId, tools.AdminLevel, *repoApp)
	assert.True(t, isVisible)
}

func TestIsAppVisibleToUser_NotRunningNotVisible(t *testing.T) {
	testObjects := setupAppServiceHelperTest(t)
	repoApp := GetSampleApp()
	repoApp.ShouldBeRunning = false

	isVisible := testObjects.AppServiceHelper.IsAppVisibleToUser(sampleUserId, tools.UserLevel, *repoApp)
	assert.False(t, isVisible)
}

func TestIsAppVisibleToUser_SystemAppNotVisible(t *testing.T) {
	testObjects := setupAppServiceHelperTest(t)
	repoApp := GetSampleApp()

	testObjects.AppDetectorMock.EXPECT().IsSystemApp(sampleAppName).Return(true)

	isVisible := testObjects.AppServiceHelper.IsAppVisibleToUser(sampleUserId, tools.UserLevel, *repoApp)
	assert.False(t, isVisible)
}

func TestIsAppVisibleToUser_AuthorizedVisible(t *testing.T) {
	testObjects := setupAppServiceHelperTest(t)
	repoApp := GetSampleApp()

	testObjects.AppDetectorMock.EXPECT().IsSystemApp(sampleAppName).Return(false)
	testObjects.AuthorizerMock.EXPECT().
		Authorize(tools.Policies.PublicAccessPolicy, tools.UserLevel, sampleUserId, sampleAppName).
		Return(nil)

	isVisible := testObjects.AppServiceHelper.IsAppVisibleToUser(sampleUserId, tools.UserLevel, *repoApp)
	assert.True(t, isVisible)
}

func TestIsAppVisibleToUser_UnauthorizedNotVisible(t *testing.T) {
	testObjects := setupAppServiceHelperTest(t)
	repoApp := GetSampleApp()

	testObjects.AppDetectorMock.EXPECT().IsSystemApp(sampleAppName).Return(false)
	testObjects.AuthorizerMock.EXPECT().
		Authorize(tools.Policies.PublicAccessPolicy, tools.UserLevel, sampleUserId, sampleAppName).
		Return(errors.New("denied"))

	isVisible := testObjects.AppServiceHelper.IsAppVisibleToUser(sampleUserId, tools.UserLevel, *repoApp)
	assert.False(t, isVisible)
}

func TestConvertToAppDtos(t *testing.T) {
	testObjects := setupAppServiceHelperTest(t)

	officialDatabaseApp := GetSampleApp()
	officialDatabaseApp.AppId = 123
	officialDatabaseApp.AppName = "postgres"
	officialDatabaseApp.ShouldBeRunning = true

	customApp := GetSampleApp()
	customApp.AppId = 456
	customApp.Maintainer = "custom-maintainer"
	customApp.AppName = "my-app"
	customApp.ShouldBeRunning = false

	testObjects.AppDetectorMock.EXPECT().IsOfficialDatabaseApp(officialDatabaseApp.AppName).Return(true)
	testObjects.AppDetectorMock.EXPECT().IsOfficialApp(officialDatabaseApp.Maintainer).Return(true)
	testObjects.AppDetectorMock.EXPECT().IsOfficialDatabaseApp(customApp.AppName).Return(false)
	testObjects.AppDetectorMock.EXPECT().IsOfficialApp(customApp.Maintainer).Return(false)

	appDtos := testObjects.AppServiceHelper.ConvertToAppDtos([]RepoApp{*officialDatabaseApp, *customApp})

	assert.Equal(t, 2, len(appDtos))
	expectedOfficialDatabaseApp := AppDto{
		AppId:                    strconv.Itoa(officialDatabaseApp.AppId),
		Maintainer:               officialDatabaseApp.Maintainer,
		AppName:                  officialDatabaseApp.AppName,
		VersionName:              officialDatabaseApp.VersionName,
		AccessPolicy:             officialDatabaseApp.AccessPolicy,
		Port:                     officialDatabaseApp.Port,
		VersionCreationTimestamp: officialDatabaseApp.VersionCreationTimestamp,
		VersionContent:           officialDatabaseApp.VersionContent,
		ClientId:                 officialDatabaseApp.ClientId,
		ClientSecret:             officialDatabaseApp.ClientSecret,
		AutomaticBackupsEnabled:  officialDatabaseApp.AutomaticBackupsEnabled,
		AutomaticUpdatesEnabled:  officialDatabaseApp.AutomaticUpdatesEnabled,
		IsRunning:                true,
		IsOfficialDatabaseApp:    true,
		IsOfficial:               true,
		DocsUrl:                  tools.InstalledAppDocsUrl(officialDatabaseApp.AppName),
	}
	expectedCustomApp := AppDto{
		AppId:                    strconv.Itoa(customApp.AppId),
		Maintainer:               customApp.Maintainer,
		AppName:                  customApp.AppName,
		VersionName:              customApp.VersionName,
		AccessPolicy:             customApp.AccessPolicy,
		Port:                     customApp.Port,
		VersionCreationTimestamp: customApp.VersionCreationTimestamp,
		VersionContent:           customApp.VersionContent,
		ClientId:                 customApp.ClientId,
		ClientSecret:             customApp.ClientSecret,
		AutomaticBackupsEnabled:  customApp.AutomaticBackupsEnabled,
		AutomaticUpdatesEnabled:  customApp.AutomaticUpdatesEnabled,
		IsRunning:                false,
		IsOfficialDatabaseApp:    false,
		IsOfficial:               false,
		DocsUrl:                  "",
	}
	assert.Equal(t, expectedOfficialDatabaseApp, appDtos[0])
	assert.Equal(t, expectedCustomApp, appDtos[1])
}

func TestGetPortFromComposeYaml_ValidComposeReturnsPort(t *testing.T) {
	testObjects := setupAppServiceHelperTest(t)

	composeContent := []byte(`
services:
  sampleapp:
    labels:
      quollix.port: 8080
`)

	port, err := testObjects.AppServiceHelper.GetPortFromComposeYaml(composeContent, "sampleapp")
	assert.Nil(t, err)
	assert.Equal(t, "8080", port)
}

func TestGetPortFromComposeYaml_ServiceNotFoundReturnsError(t *testing.T) {
	testObjects := setupAppServiceHelperTest(t)

	composeContent := []byte(`
services:
  another-service:
    labels:
      quollix.port: 8080
`)

	port, err := testObjects.AppServiceHelper.GetPortFromComposeYaml(composeContent, "sampleapp")
	assert.Equal(t, "", port)
	assert.NotNil(t, err)
	assert.Equal(t, "service not found in docker-compose.yml", u.ExtractError(err))
}

func TestGetPortFromComposeYaml_MissingPortLabelReturnsError(t *testing.T) {
	testObjects := setupAppServiceHelperTest(t)

	composeContent := []byte(`
services:
  sampleapp:
    labels:
      another.label: 1
`)

	port, err := testObjects.AppServiceHelper.GetPortFromComposeYaml(composeContent, "sampleapp")
	assert.Equal(t, "", port)
	assert.NotNil(t, err)
	assert.Equal(t, "could not find quollix.port label in docker-compose.yml", u.ExtractError(err))
}
