package component

import (
	"encoding/json"
	"server/apps_basic"
	"server/tools"

	"github.com/quollix/common/assert"
	"github.com/quollix/common/store"
	u "github.com/quollix/common/utils"
)

type QuollixAppsClient struct {
	quollix *QuollixClient
}

func (c *QuollixAppsClient) InstallSample(version string) (*apps_basic.AppDto, error) {
	if err := c.InstallFromStore(tools.SampleMaintainer, tools.SampleApp, version); err != nil {
		return nil, err
	}

	installedApps := c.ListInstalled()
	for _, installedApp := range installedApps {
		if installedApp.AppName == tools.SampleApp {
			return &installedApp, nil
		}
	}
	return nil, u.Logger.NewError("Sample app not found")
}

func (c *QuollixAppsClient) InstallFromStore(maintainer, appName, version string) error {
	_, err := c.quollix.Parent.DoRequest(tools.Paths.BackendStoreVersionsInstall, store.VersionTree{
		Maintainer:  maintainer,
		AppName:     appName,
		VersionName: version,
	})
	return err
}

func (c *QuollixAppsClient) SearchStore(maintainerName, appName string, searchForUnofficialApps bool) ([]store.AppWithLatestVersion, error) {
	appSearchRequest := store.SearchRequest{
		MaintainerSearchTerm: maintainerName,
		AppSearchTerm:        appName,
		ShowUnofficialApps:   searchForUnofficialApps,
	}
	responseBody, err := c.quollix.Parent.DoRequest(tools.Paths.BackendStoreSearch, appSearchRequest)
	if err != nil {
		return nil, err
	}
	var apps []store.AppWithLatestVersion
	err = json.Unmarshal(responseBody, &apps)
	if err != nil {
		return nil, err
	}
	return apps, nil
}

func (c *QuollixAppsClient) FindVersion(userName, appName, versionName string) *store.LeanVersionDto {
	responseBody, err := c.quollix.Parent.DoRequest(tools.Paths.BackendStoreVersionsList, store.AppTree{
		Maintainer: userName,
		AppName:    appName,
	})
	assert.Nil(c.quollix.T, err)
	var versions []store.LeanVersionDto
	err = json.Unmarshal(responseBody, &versions)
	assert.Nil(c.quollix.T, err)
	for _, v := range versions {
		if v.Name == versionName {
			return &v
		}
	}
	c.quollix.T.Fatal("No version found")
	return nil
}

func (c *QuollixAppsClient) ListInstalled() []apps_basic.AppDto {
	responseBody, err := c.quollix.Parent.DoRequest(tools.Paths.BackendAppsList, nil)
	assert.Nil(c.quollix.T, err)
	var apps []apps_basic.AppDto
	err = json.Unmarshal(responseBody, &apps)
	assert.Nil(c.quollix.T, err)
	return apps
}

func (c *QuollixAppsClient) GetInstalledSample() *apps_basic.AppDto {
	var sampleApp *apps_basic.AppDto
	apps := c.ListInstalled()
	for _, app := range apps {
		if app.AppName == tools.SampleApp {
			sampleApp = &app
			break
		}
	}
	assert.NotNil(c.quollix.T, sampleApp)
	return sampleApp
}

func (c *QuollixAppsClient) Delete(appId string) error {
	_, err := c.quollix.Parent.DoRequest(tools.Paths.BackendAppsDelete, tools.NumberString{Value: appId})
	if err != nil {
		return err
	}
	return nil
}

func (c *QuollixAppsClient) Update(appId string) error {
	_, err := c.quollix.Parent.DoRequest(tools.Paths.BackendAppsUpdate, tools.NumberString{Value: appId})
	if err != nil {
		return err
	}
	return nil
}

func (c *QuollixAppsClient) Start(appId string) error {
	_, err := c.quollix.Parent.DoRequest(tools.Paths.BackendAppsStart, tools.NumberString{Value: appId})
	if err != nil {
		return err
	}
	return nil
}

func (c *QuollixAppsClient) Stop(appId string) error {
	_, err := c.quollix.Parent.DoRequest(tools.Paths.BackendAppsStop, tools.NumberString{Value: appId})
	if err != nil {
		return err
	}
	return nil
}

func (c *QuollixAppsClient) SetAccessPolicy(appId, accessPolicy string) error {
	_, err := c.quollix.Parent.DoRequest(tools.Paths.BackendAppsChangeAccessPolicy, apps_basic.ChangeAccessPolicyRequest{
		AppId:        appId,
		AccessPolicy: accessPolicy,
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *QuollixAppsClient) UpdateMaintenanceSettings(appId string, autoUpdatesEnabled, autoBackupsEnabled bool) {
	_, err := c.quollix.Parent.DoRequest(tools.Paths.BackendAppAutomaticMaintenanceSettings, apps_basic.AutoMaintenanceSettingsResponse{
		AppId:                   appId,
		AutomaticUpdatesEnabled: autoUpdatesEnabled,
		AutomaticBackupsEnabled: autoBackupsEnabled,
	})
	assert.Nil(c.quollix.T, err)
}

func (c *QuollixAppsClient) UploadVersionFile(file tools.BinaryFile) error {
	_, err := c.quollix.Parent.DoRequest(tools.Paths.BackendAppUploadToApplication, file)
	return err
}

func (c *QuollixAppsClient) DownloadVersionFile(appId string) *tools.BinaryFile {
	resp, err := c.quollix.Parent.DoRequestWithFullResponse(tools.Paths.BackendAppDownloadFromApplication, tools.NumberString{Value: appId})
	assert.Nil(c.quollix.T, err)
	defer u.Close(resp.Body)
	var versionFile tools.BinaryFile
	err = json.NewDecoder(resp.Body).Decode(&versionFile)
	assert.Nil(c.quollix.T, err)
	return &versionFile
}

func (c *QuollixAppsClient) GetCurrentOperations() ([]string, bool) {
	body, err := c.quollix.Parent.DoRequest(tools.Paths.BackendAppOperationInfo, nil)
	assert.Nil(c.quollix.T, err)
	var out apps_basic.AppOperationInfoResponse
	assert.Nil(c.quollix.T, json.Unmarshal(body, &out))
	return out.Operations, out.IsOngoing
}

func (c *QuollixAppsClient) DownloadVersion(maintainer, appName, versionName string) (*tools.BinaryFile, error) {
	responseBody, err := c.quollix.Parent.DoRequest(tools.Paths.BackendStoreVersionsDownload, store.VersionTree{
		Maintainer:  maintainer,
		AppName:     appName,
		VersionName: versionName,
	})
	if err != nil {
		return nil, err
	}
	var download tools.BinaryFile
	err = json.Unmarshal(responseBody, &download)
	if err != nil {
		return nil, err
	}
	return &download, nil
}
