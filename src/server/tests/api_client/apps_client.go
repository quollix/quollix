package api_client

import (
	"encoding/json"
	"server/apps_basic"
	"server/tools"

	"github.com/quollix/common/store"
	u "github.com/quollix/common/utils"
)

type QuollixAppsClient struct {
	quollix *QuollixClient
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

func (c *QuollixAppsClient) ListVersions(userName, appName string) ([]store.LeanVersionDto, error) {
	responseBody, err := c.quollix.Parent.DoRequest(tools.Paths.BackendStoreVersionsList, store.AppTree{
		Maintainer: userName,
		AppName:    appName,
	})
	if err != nil {
		return nil, err
	}
	var versions []store.LeanVersionDto
	if err = json.Unmarshal(responseBody, &versions); err != nil {
		return nil, err
	}
	return versions, nil
}

func (c *QuollixAppsClient) ListInstalled() ([]apps_basic.AppDto, error) {
	responseBody, err := c.quollix.Parent.DoRequest(tools.Paths.BackendAppsList, nil)
	if err != nil {
		return nil, err
	}
	var apps []apps_basic.AppDto
	err = json.Unmarshal(responseBody, &apps)
	if err != nil {
		return nil, err
	}
	return apps, nil
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

func (c *QuollixAppsClient) UpdateMaintenanceSettings(appId string, autoUpdatesEnabled, autoBackupsEnabled bool) error {
	_, err := c.quollix.Parent.DoRequest(tools.Paths.BackendAppAutomaticMaintenanceSettings, apps_basic.AutoMaintenanceSettingsResponse{
		AppId:                   appId,
		AutomaticUpdatesEnabled: autoUpdatesEnabled,
		AutomaticBackupsEnabled: autoBackupsEnabled,
	})
	return err
}

func (c *QuollixAppsClient) UploadVersionFile(file tools.BinaryFile) error {
	_, err := c.quollix.Parent.DoRequest(tools.Paths.BackendAppUploadToApplication, file)
	return err
}

func (c *QuollixAppsClient) DownloadVersionFile(appId string) (*tools.BinaryFile, error) {
	resp, err := c.quollix.Parent.DoRequestWithFullResponse(tools.Paths.BackendAppDownloadFromApplication, tools.NumberString{Value: appId})
	if err != nil {
		return nil, err
	}
	defer u.Close(resp.Body)
	var versionFile tools.BinaryFile
	err = json.NewDecoder(resp.Body).Decode(&versionFile)
	if err != nil {
		return nil, err
	}
	return &versionFile, nil
}

func (c *QuollixAppsClient) GetCurrentOperations() ([]string, bool, error) {
	body, err := c.quollix.Parent.DoRequest(tools.Paths.BackendAppOperationInfo, nil)
	if err != nil {
		return nil, false, err
	}
	var out apps_basic.AppOperationInfoResponse
	if err = json.Unmarshal(body, &out); err != nil {
		return nil, false, err
	}
	return out.Operations, out.IsOngoing, nil
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
