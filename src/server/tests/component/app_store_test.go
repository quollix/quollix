//go:build component

package component

import (
	"testing"
	"time"

	"server/app_store"
	"server/tools"

	"github.com/quollix/common/assert"
	api "github.com/quollix/common/quollix/api"
	"github.com/quollix/common/quollix/api_client"
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
	assert.True(t, storeApp.LatestVersionId > 0)
	assert.Equal(t, "2.0", storeApp.LatestVersionName)
	assert.Equal(t, tools.SampleAppVersion2CreationTimestamp, storeApp.LatestVersionCreationTimestamp)

	latestAppVersion, err := FindVersion(t, client, tools.SampleMaintainer, tools.SampleApp, "2.0")
	assert.Nil(t, err)
	assertLeanVersion(t, latestAppVersion, "2.0", tools.SampleAppVersion2CreationTimestamp, tools.SampleAppVersion2ComposeYAML)
	assert.True(t, latestAppVersion.IsMigrationCheckpoint)

	oldAppVersion, err := FindVersion(t, client, tools.SampleMaintainer, tools.SampleApp, "1.0")
	assert.Nil(t, err)
	assertLeanVersion(t, oldAppVersion, "1.0", tools.SampleAppVersion1CreationTimestamp, tools.SampleAppVersion1ComposeYAML)
	assert.False(t, oldAppVersion.IsMigrationCheckpoint)
}

func assertLeanVersion(t *testing.T, actual *store.LeanVersionDto, expectedName string, expectedCreationTimestamp time.Time, expectedContent string) {
	assert.True(t, actual.VersionId > 0)
	assert.Equal(t, expectedName, actual.Name)
	assert.Equal(t, expectedCreationTimestamp, actual.CreationTimestamp)
	assert.Equal(t, int64(len([]byte(expectedContent))), actual.SizeInBytes)
}

func TestInstallingAppFromStoreStartsApp(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()
	_, err := InstallSample(t, client, "2.0")
	assert.Nil(t, err)

	sampleApp := GetInstalledSample(t, client)
	assert.Equal(t, tools.SampleMaintainer, sampleApp.Maintainer)
	assert.True(t, sampleApp.IsRunning)
}

func TestInstallingAppFromStoreDownloadsMaintainerKey(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()
	publicKey, exists := getMaintainerPublicKey(t, client, tools.SampleMaintainer)
	assert.False(t, exists)
	assert.Nil(t, publicKey)

	_, err := InstallSample(t, client, "2.0")
	assert.Nil(t, err)

	publicKey, exists = getMaintainerPublicKey(t, client, tools.SampleMaintainer)
	assert.True(t, exists)
	assert.Equal(t, tools.SampleMaintainer, publicKey.Name)
	assert.Equal(t, u.OtherLocalTestingPublicKeyOpenSSH, publicKey.PublicKey)
	assert.Equal(t, u.OtherLocalTestingPublicKeyFingerprintSHA256, publicKey.Fingerprint)
	assert.False(t, publicKey.LastUpdatedAt.IsZero())
}

func TestInstallingAppFromStoreRefreshesMaintainerKeyAfterMismatch(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()
	assert.Nil(t, client.AppMaintainers.Add(tools.SampleMaintainer, u.LocalTestingPublicKeyOpenSSH))

	publicKey, exists := getMaintainerPublicKey(t, client, tools.SampleMaintainer)
	assert.True(t, exists)
	assert.Equal(t, u.LocalTestingPublicKeyOpenSSH, publicKey.PublicKey)
	assert.Equal(t, u.LocalTestingPublicKeyFingerprintSHA256, publicKey.Fingerprint)

	_, err := InstallSample(t, client, "2.0")
	assert.Nil(t, err)

	publicKey, exists = getMaintainerPublicKey(t, client, tools.SampleMaintainer)
	assert.True(t, exists)
	assert.Equal(t, u.OtherLocalTestingPublicKeyOpenSSH, publicKey.PublicKey)
	assert.Equal(t, u.OtherLocalTestingPublicKeyFingerprintSHA256, publicKey.Fingerprint)
}

func TestMaintainerPublicKeyCanBeAddedListedAndDeleted(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	assert.Nil(t, client.AppMaintainers.Add(tools.SampleMaintainer, u.LocalTestingPublicKeyOpenSSH))
	publicKey, exists := getMaintainerPublicKey(t, client, tools.SampleMaintainer)
	assert.True(t, exists)
	assert.Equal(t, tools.SampleMaintainer, publicKey.Name)
	assert.Equal(t, u.LocalTestingPublicKeyOpenSSH, publicKey.PublicKey)
	assert.Equal(t, u.LocalTestingPublicKeyFingerprintSHA256, publicKey.Fingerprint)
	assert.False(t, publicKey.LastUpdatedAt.IsZero())

	assert.Nil(t, client.AppMaintainers.Delete(tools.SampleMaintainer))
	publicKey, exists = getMaintainerPublicKey(t, client, tools.SampleMaintainer)
	assert.False(t, exists)
	assert.Nil(t, publicKey)
}

func TestInstallingSameOrOlderAppVersionFromStoreShouldFail(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()
	_, err := InstallSample(t, client, "1.0")
	assert.Nil(t, err)

	_, err = InstallSample(t, client, "1.0")
	assert.NotNil(t, err)
	u.AssertDeepStackErrorFromRequest(t, err, app_store.CanNotInstallOlderAppVersionOverNewerOrEqual)

	_, err = InstallSample(t, client, "0.0")
	assert.NotNil(t, err)
	u.AssertDeepStackErrorFromRequest(t, err, app_store.CanNotInstallOlderAppVersionOverNewerOrEqual)
}

func TestInstallingNewerAppVersionFromStoreUpdatesAppAndCreatesBackup(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()
	_, err := InstallSample(t, client, "1.0")
	assert.Nil(t, err)
	configureBackupRepo(t, client)

	appBackups, err := client.Backups.ListByApp(tools.SampleMaintainer, tools.SampleApp)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(appBackups))

	_, err = InstallSample(t, client, "2.0")
	assert.Nil(t, err)
	sampleApp := GetInstalledSample(t, client)
	assert.Equal(t, "2.0", sampleApp.VersionName)
	assert.True(t, sampleApp.IsRunning)

	appBackups, err = client.Backups.ListByApp(tools.SampleMaintainer, tools.SampleApp)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(appBackups))
	assert.Equal(t, "1.0", appBackups[0].VersionName)
	assert.Equal(t, tools.PreUpdateBackupDescription, appBackups[0].Description)
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
	storeVersion, err := FindVersion(t, client, tools.SampleMaintainer, tools.SampleApp, tools.SampleAppVersion2Name)
	assert.Nil(t, err)
	appDownload, err := client.Apps.DownloadStoreVersion(storeVersion.VersionId)
	assert.Nil(t, err)
	assert.Equal(t, "samplemaintainer_sampleapp_2.0_2021-01-01-01-00-00.yml", appDownload.FileName)
	assert.Equal(t, []byte(tools.SampleAppVersion2ComposeYAML), appDownload.Content)
}

func TestInstallShouldRejectVersionWithInvalidPackageSigning(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	_, err := InstallSample(t, client, "1.5")
	assert.NotNil(t, err)
	u.AssertDeepStackErrorFromRequest(t, err, app_store.InvalidPackageSigningError)
}

func TestDownloadShouldRejectVersionWithInvalidPackageSigning(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	storeVersion, err := FindVersion(t, client, tools.SampleMaintainer, tools.SampleApp, "1.5")
	assert.Nil(t, err)
	_, err = client.Apps.DownloadStoreVersion(storeVersion.VersionId)
	assert.NotNil(t, err)
	u.AssertDeepStackErrorFromRequest(t, err, app_store.InvalidPackageSigningError)
}

func getMaintainerPublicKey(t *testing.T, client *api_client.QuollixClient, maintainer string) (*api.MaintainerPublicKeyDto, bool) {
	publicKeys, err := client.AppMaintainers.List()
	assert.Nil(t, err)
	for index := range publicKeys {
		publicKey := &publicKeys[index]
		if publicKey.Name == maintainer {
			return publicKey, true
		}
	}
	return nil, false
}
