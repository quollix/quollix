package api_client

import (
	"encoding/json"
	"server/tools"
)

type QuollixBackupsClient struct {
	quollix *QuollixClient
}

func (c *QuollixBackupsClient) Create(appId string) error {
	_, err := c.quollix.Parent.DoRequest(tools.Paths.BackendBackupsCreate, tools.NumberString{Value: appId})
	if err != nil {
		return err
	}
	return nil
}

func (c *QuollixBackupsClient) ListByApp(maintainer, appName string) ([]tools.BackupInfo, error) {
	backupListRequest := tools.MaintainerAndApp{
		Maintainer: maintainer,
		AppName:    appName,
	}
	responseBody, err := c.quollix.Parent.DoRequest(tools.Paths.BackendBackupsList, backupListRequest)
	if err != nil {
		return nil, err
	}
	var backups []tools.BackupInfo
	if responseBody == nil {
		return backups, nil
	}
	err = json.Unmarshal(responseBody, &backups)
	if err != nil {
		return nil, err
	}
	return backups, nil
}

func (c *QuollixBackupsClient) Delete(backupIds []string) error {
	deleteBackupRequest := tools.BackupsOperationRequest{BackupIds: backupIds}
	_, err := c.quollix.Parent.DoRequest(tools.Paths.BackendBackupsDelete, deleteBackupRequest)
	if err != nil {
		return err
	}
	return nil
}

func (c *QuollixBackupsClient) Restore(backupId string) error {
	_, err := c.quollix.Parent.DoRequest(tools.Paths.BackendBackupsRestore, tools.BackupOperationRequest{BackupId: backupId})
	if err != nil {
		return err
	}
	return nil
}

func (c *QuollixBackupsClient) ListAppsInRepository() ([]tools.MaintainerAndApp, error) {
	responseBody, err := c.quollix.Parent.DoRequest(tools.Paths.BackendBackupsListApps, nil)
	if err != nil {
		return nil, err
	}
	var apps []tools.MaintainerAndApp
	err = json.Unmarshal(responseBody, &apps)
	if err != nil {
		return nil, err
	}
	return apps, nil
}

func (c *QuollixBackupsClient) PurgeServer(repo *tools.BackupServerConfigs) error {
	purgeRequest := &tools.SshConnectionRequest{
		Host:          repo.Host,
		SshPort:       repo.SshPort,
		SshKnownHosts: repo.SshKnownHosts,
		SshUser:       repo.SshUser,
		SshPassword:   repo.SshPassword,
	}
	_, err := c.quollix.Parent.DoRequest(tools.Paths.BackendBackupsPurgeBackupServer, purgeRequest)
	return err
}
