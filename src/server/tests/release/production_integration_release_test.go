//go:build release

package release

import (
	"testing"

	"server/tests/component"

	"github.com/quollix/common/assert"
	commonStore "github.com/quollix/common/store"
	u "github.com/quollix/common/utils"
)

const (
	storeURL         = "https://store.quollix.org"
	expectedStoreApp = "nextcloud"
)

func TestReleaseProductionStoreIntegration(t *testing.T) {
	u.Logger.Info("checking production store integration")

	client := commonStore.AppStoreClientImpl{
		Parent: u.ComponentClient{
			RootUrl:           storeURL,
			VerifyCertificate: true,
		},
	}
	apps, err := client.SearchForApps("", expectedStoreApp, false)
	assert.Nil(t, err)
	_, err = findStoreApp(apps, expectedStoreApp)
	assert.Nil(t, err)
}

func TestReleaseLocalQuollixStoreIntegration(t *testing.T) {
	u.Logger.Info("checking local Quollix integration with production store")

	client := component.GetClientAndLogin(t)
	apps, err := client.Apps.SearchStore("", expectedStoreApp, false)
	assert.Nil(t, err)
	app, err := findStoreApp(apps, expectedStoreApp)
	assert.Nil(t, err)
	assert.Nil(t, client.Apps.InstallFromStore(app.Maintainer, app.AppName, app.LatestVersionName))
}

func findStoreApp(apps []commonStore.AppWithLatestVersion, expectedApp string) (*commonStore.AppWithLatestVersion, error) {
	for i := range apps {
		app := &apps[i]
		if app.AppName == expectedApp {
			return app, nil
		}
	}
	return nil, u.Logger.NewError("expected store app was not found in apps", expectedApp, storeAppNames(apps))
}

func storeAppNames(apps []commonStore.AppWithLatestVersion) []string {
	appNames := make([]string, 0, len(apps))
	for _, app := range apps {
		appNames = append(appNames, app.AppName)
	}
	return appNames
}
