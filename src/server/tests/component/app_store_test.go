//go:build component

package component

import (
	"server/app_store"
	"server/tools"
	"testing"
	"time"

	"github.com/quollix/common/assert"
	"github.com/quollix/common/store"
	u "github.com/quollix/common/utils"
)

func TestAppStoreContainsBothSampleVersions(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()
	storeApps, err := client.Apps.SearchStore("", tools.SampleApp, true)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(storeApps))
	storeApp := storeApps[0]
	assert.Equal(t, tools.SampleApp, storeApp.AppName)
	assert.Equal(t, tools.SampleMaintainer, storeApp.Maintainer)
	assert.Equal(t, "2.0", storeApp.LatestVersionName)
	assert.Equal(t, tools.SampleAppVersion2CreationTimestamp, storeApp.LatestVersionCreationTimestamp)

	latestAppVersion := client.Apps.FindVersion(tools.SampleMaintainer, tools.SampleApp, "2.0")
	assertLeanVersion(t, latestAppVersion, "2.0", tools.SampleAppVersion2CreationTimestamp, tools.SampleAppVersion2ComposeYAML)

	oldAppVersion := client.Apps.FindVersion(tools.SampleMaintainer, tools.SampleApp, "1.0")
	assertLeanVersion(t, oldAppVersion, "1.0", tools.SampleAppVersion1CreationTimestamp, tools.SampleAppVersion1ComposeYAML)
}

func assertLeanVersion(t *testing.T, actual *store.LeanVersionDto, expectedName string, expectedCreationTimestamp time.Time, expectedContent string) {
	assert.Equal(t, expectedName, actual.Name)
	assert.Equal(t, expectedCreationTimestamp, actual.CreationTimestamp)
	assert.Equal(t, int64(len([]byte(expectedContent))), actual.SizeInBytes)
}

func TestInstallingExistingAppShouldFail(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()
	_, err := client.Apps.InstallSample("2.0")
	assert.Nil(t, err)
	_, err = client.Apps.InstallSample("2.0")
	assert.NotNil(t, err)
	u.AssertDeepStackErrorFromRequest(t, err, app_store.AppAlreadyInstalledError)
}

func TestAppStoreSearch(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()
	apps, err := client.Apps.SearchStore("", tools.SampleApp, true)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(apps))

	apps, err = client.Apps.SearchStore("", tools.SampleApp, false)
	assert.NotNil(t, err)
	u.AssertDeepStackErrorFromRequest(t, err, app_store.NoAppsFoundError)
	assert.Nil(t, apps)
}

func TestPhysicalAppDownloadToUsersPc(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()
	appDownload, err := client.Apps.DownloadVersion(tools.SampleMaintainer, tools.SampleApp, tools.SampleAppVersion2Name)
	assert.Nil(t, err)
	assert.Equal(t, "samplemaintainer_sampleapp_2.0_2021-01-01-01-00-00.yml", appDownload.FileName)
	assert.Equal(t, []byte(tools.SampleAppVersion2ComposeYAML), appDownload.Content)
}

func TestInstallShouldRejectVersionWithInvalidPackageSigning(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	_, err := client.Apps.InstallSample("1.5")
	assert.NotNil(t, err)
	u.AssertDeepStackErrorFromRequest(t, err, app_store.InvalidPackageSigningError)
}

func TestDownloadShouldRejectVersionWithInvalidPackageSigning(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	_, err := client.Apps.DownloadVersion(tools.SampleMaintainer, tools.SampleApp, "1.5")
	assert.NotNil(t, err)
	u.AssertDeepStackErrorFromRequest(t, err, app_store.InvalidPackageSigningError)
}
