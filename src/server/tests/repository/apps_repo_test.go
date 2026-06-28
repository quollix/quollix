//go:build integration

package repository

import (
	"testing"
	"time"

	"server/apps_basic"
	"server/tools"

	"github.com/quollix/common/assert"
	u "github.com/quollix/common/utils"
)

func TestAppCreateAndRead(t *testing.T) {
	InitDeps()
	defer AppRepo.Wipe()

	initialApps, err := AppRepo.ListApps()
	assert.Nil(t, err)

	expectedApp := apps_basic.GetSampleApp()
	expectedApp.AppId, err = AppRepo.CreateApp(expectedApp)
	assert.Nil(t, err)

	appById, err := AppRepo.GetAppById(expectedApp.AppId)
	assert.Nil(t, err)
	AssertAppEquality(t, expectedApp, appById)

	actualByName, err := AppRepo.GetAppByName(expectedApp.AppName)
	assert.Nil(t, err)
	AssertAppEquality(t, expectedApp, actualByName)
	assert.Equal(t, 16, len(appById.ClientId))
	assert.Equal(t, 64, len(actualByName.ClientSecret))

	appsList, err := AppRepo.ListApps()
	assert.Nil(t, err)
	assert.Equal(t, len(initialApps)+1, len(appsList))
	var appFromList *apps_basic.RepoApp
	for _, app := range appsList {
		if app.AppId == expectedApp.AppId {
			appCopy := app
			appFromList = &appCopy
			break
		}
	}
	assert.NotNil(t, appFromList)
	AssertAppEquality(t, expectedApp, appFromList)
}

func TestAppDeletion(t *testing.T) {
	InitDeps()
	defer AppRepo.Wipe()

	initialApps, err := AppRepo.ListApps()
	assert.Nil(t, err)

	app := apps_basic.GetSampleApp()
	app.AppId, err = AppRepo.CreateApp(app)
	assert.Nil(t, err)

	apps, err := AppRepo.ListApps()
	assert.Nil(t, err)
	assert.Equal(t, len(initialApps)+1, len(apps))

	assert.Nil(t, AppRepo.DeleteApp(app.AppId))

	apps, err = AppRepo.ListApps()
	assert.Nil(t, err)
	assert.Equal(t, len(initialApps), len(apps))
}

func TestDoesAppExist(t *testing.T) {
	InitDeps()
	defer AppRepo.Wipe()

	app := apps_basic.GetSampleApp()
	exists, err := AppRepo.DoesAppExist(app.AppName)
	assert.Nil(t, err)
	assert.False(t, exists)

	_, err = AppRepo.CreateApp(app)
	assert.Nil(t, err)
	exists, err = AppRepo.DoesAppExist(app.AppName)
	assert.Nil(t, err)
	assert.True(t, exists)
}

func TestUpdateApp(t *testing.T) {
	InitDeps()
	defer AppRepo.Wipe()

	app := apps_basic.GetSampleApp()
	var err error
	app.AppId, err = AppRepo.CreateApp(app)
	assert.Nil(t, err)

	clientCredentialsGenerator := &apps_basic.ClientCredentialsGeneratorImpl{
		AuthHelper: &u.AuthHelperImpl{},
	}
	clientId, clientSecret, err := clientCredentialsGenerator.Generate()
	assert.Nil(t, err)
	updatedApp := apps_basic.NewRepoApp(
		"updated-maintainer",
		"updated-app-name",
		"v2.0.0",
		tools.Policies.GroupRestrictedAccessPolicy,
		"8080",
		clientId,
		clientSecret,
		time.Date(2025, time.January, 1, 12, 0, 0, 0, time.UTC),
		[]byte("updated-content"),
		false,
		false,
		false,
	)
	updatedApp.AppId = app.AppId

	assert.Nil(t, AppRepo.UpdateApp(updatedApp))

	actual, err := AppRepo.GetAppById(app.AppId)
	assert.Nil(t, err)
	AssertAppEquality(t, updatedApp, actual)
}

func TestGetAppByClientId(t *testing.T) {
	InitDeps()
	defer AppRepo.Wipe()

	expected := apps_basic.GetSampleApp()
	var err error
	expected.AppId, err = AppRepo.CreateApp(expected)
	assert.Nil(t, err)

	actual, exists, err := AppRepo.GetAppByClientId(expected.ClientId)
	assert.Nil(t, err)
	assert.True(t, exists)
	AssertAppEquality(t, expected, actual)
}

func TestGetAppByClientIdNotFound(t *testing.T) {
	InitDeps()
	defer AppRepo.Wipe()

	actual, exists, err := AppRepo.GetAppByClientId("missing-client-id")
	assert.Nil(t, err)
	assert.False(t, exists)
	assert.Nil(t, actual)
}

func TestGetAppRequestData(t *testing.T) {
	InitDeps()
	defer AppRepo.Wipe()

	expected := apps_basic.GetSampleApp()
	_, err := AppRepo.CreateApp(expected)
	assert.Nil(t, err)

	actual, err := AppRepo.GetAppRequestData(expected.AppName)
	assert.Nil(t, err)

	assert.Equal(t, expected.Maintainer, actual.Maintainer)
	assert.Equal(t, expected.AppName, actual.AppName)
	assert.Equal(t, expected.AccessPolicy, actual.AccessPolicy)
	assert.Equal(t, expected.Port, actual.Port)
}

func TestDoesAppWithMaintainerExist(t *testing.T) {
	InitDeps()
	defer AppRepo.Wipe()

	app := apps_basic.GetSampleApp()

	exists, err := AppRepo.DoesAppWithMaintainerExist(app.Maintainer, app.AppName)
	assert.Nil(t, err)
	assert.False(t, exists)

	_, err = AppRepo.CreateApp(app)
	assert.Nil(t, err)

	exists, err = AppRepo.DoesAppWithMaintainerExist(app.Maintainer, app.AppName)
	assert.Nil(t, err)
	assert.True(t, exists)

	exists, err = AppRepo.DoesAppWithMaintainerExist("different-maintainer", app.AppName)
	assert.Nil(t, err)
	assert.False(t, exists)

	exists, err = AppRepo.DoesAppWithMaintainerExist(app.Maintainer, "different-app-name")
	assert.Nil(t, err)
	assert.False(t, exists)

	exists, err = AppRepo.DoesAppWithMaintainerExist("different-maintainer", "different-app-name")
	assert.Nil(t, err)
	assert.False(t, exists)
}
