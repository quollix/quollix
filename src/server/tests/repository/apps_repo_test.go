//go:build integration

package repository

import (
	"testing"
	"time"

	"server/apps_basic"

	"github.com/quollix/common/assert"
	"github.com/quollix/common/quollix/api"
	u "github.com/quollix/common/utils"
)

func TestAppCreateAndRead(t *testing.T) {
	InitDeps()
	defer AppRepo.Wipe()

	initialApps, err := AppRepo.ListApps()
	assert.Nil(t, err)

	expectedApp := apps_basic.GetSampleApp()
	expectedApp.VersionCreationTimestamp = time.Date(2026, time.July, 10, 12, 34, 56, 789123000, time.UTC)
	expectedApp.Secrets = map[string]string{
		"SECRET_POSTGRES_PASSWORD": "postgres-secret",
		"SECRET_SESSION_SECRET":    "session-secret",
	}
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
	assert.Equal(t, 64, len(actualByName.AppSecret))

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
	app.Secrets = map[string]string{
		"SECRET_POSTGRES_PASSWORD": "postgres-secret",
	}
	app.AppId, err = AppRepo.CreateApp(app)
	assert.Nil(t, err)

	apps, err := AppRepo.ListApps()
	assert.Nil(t, err)
	assert.Equal(t, len(initialApps)+1, len(apps))

	assert.Nil(t, AppRepo.DeleteApp(app.AppId))

	apps, err = AppRepo.ListApps()
	assert.Nil(t, err)
	assert.Equal(t, len(initialApps), len(apps))

	secretCount := 0
	assert.Nil(t, DatabaseConnector.GetDB().QueryRow("SELECT COUNT(*) FROM app_secrets WHERE app_id = $1", app.AppId).Scan(&secretCount))
	assert.Equal(t, 0, secretCount)
}

func TestAppListLoadsSecretsForMultipleApps(t *testing.T) {
	InitDeps()
	defer AppRepo.Wipe()

	firstApp := apps_basic.GetSampleApp()
	firstApp.Secrets = map[string]string{
		"SECRET_FIRST_PASSWORD": "first-secret",
	}
	var err error
	firstApp.AppId, err = AppRepo.CreateApp(firstApp)
	assert.Nil(t, err)

	secondApp := apps_basic.GetSampleApp()
	secondApp.AppName = "second-app"
	secondApp.ClientId = "1234567890abcdef"
	secondApp.ClientSecret = "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
	secondApp.AppSecret = "fedcba0987654321fedcba0987654321fedcba0987654321fedcba0987654321"
	secondApp.VersionContent = []byte("second-content")
	secondApp.Secrets = map[string]string{
		"SECRET_SECOND_PASSWORD": "second-secret",
		"SECRET_SECOND_TOKEN":    "second-token",
	}
	secondApp.AppId, err = AppRepo.CreateApp(secondApp)
	assert.Nil(t, err)

	apps, err := AppRepo.ListApps()
	assert.Nil(t, err)

	actualById := map[int]*apps_basic.RepoApp{}
	for _, app := range apps {
		appCopy := app
		actualById[app.AppId] = &appCopy
	}
	AssertAppEquality(t, firstApp, actualById[firstApp.AppId])
	AssertAppEquality(t, secondApp, actualById[secondApp.AppId])
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

	authHelper := &u.AuthHelperImpl{}
	clientCredentialsGenerator := &apps_basic.ClientCredentialsGeneratorImpl{
		AuthHelper: authHelper,
	}
	clientId, clientSecret, err := clientCredentialsGenerator.Generate()
	assert.Nil(t, err)
	appSecret, err := authHelper.GenerateSecret()
	assert.Nil(t, err)
	updatedApp := apps_basic.NewRepoApp(
		"updated-maintainer",
		"updated-app-name",
		"v2.0.0",
		api.Policies.GroupRestrictedAccessPolicy,
		"8080",
		clientId,
		clientSecret,
		appSecret,
		time.Date(2025, time.January, 1, 12, 0, 0, 987654000, time.UTC),
		[]byte("updated-content"),
		false,
		false,
		false,
	)
	updatedApp.AppId = app.AppId
	updatedApp.Secrets = map[string]string{
		"SECRET_UPDATED_PASSWORD": "updated-secret",
	}

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
