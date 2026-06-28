package api_client

import (
	"encoding/json"
	"server/backup_server"
	"server/tools"
)

type QuollixSettingsClient struct {
	quollix *QuollixClient
}

func (c *QuollixSettingsClient) SetBaseDomainValue(baseDomain string) error {
	_, err := c.quollix.Parent.DoRequest(tools.Paths.BackendSettingsBaseDomainSave, tools.NumberString{Value: baseDomain})
	return err
}

func (c *QuollixSettingsClient) GetBaseDomainValue() (string, error) {
	responseBody, err := c.quollix.Parent.DoRequest(tools.Paths.BackendSettingsBaseDomainRead, nil)
	if err != nil {
		return "", err
	}
	var baseDomain string
	err = json.Unmarshal(responseBody, &baseDomain)
	if err != nil {
		return "", err
	}
	return baseDomain, nil
}

func (c *QuollixSettingsClient) ReadSshConfigs() (*tools.BackupServerConfigs, error) {
	responseBody, err := c.quollix.Parent.DoRequest(tools.Paths.BackendSettingsSshRead, nil)
	if err != nil {
		return nil, err
	}
	var backupServerConfig *tools.BackupServerConfigs
	err = json.Unmarshal(responseBody, &backupServerConfig)
	if err != nil {
		return nil, err
	}
	return backupServerConfig, nil
}

func (c *QuollixSettingsClient) SaveSshConfigs(repo *tools.BackupServerConfigs) error {
	err := c.TestSshAccess(repo)
	if err != nil {
		return err
	}
	_, err = c.quollix.Parent.DoRequest(tools.Paths.BackendSettingsSshSave, repo)
	if err != nil {
		return err
	}
	return nil
}

func (c *QuollixSettingsClient) TestSshAccess(repo *tools.BackupServerConfigs) error {
	_, err := c.quollix.Parent.DoRequest(tools.Paths.BackendSettingsSshTestAccess, repo.ConvertToSshConnectionTestRequest())
	return err
}

func (c *QuollixSettingsClient) GetKnownHosts(repo *tools.BackupServerConfigs) (string, error) {
	knownHostsRequest := backup_server.KnownHostsRequest{
		Host: repo.Host,
		Port: repo.SshPort,
	}
	responseBody, err := c.quollix.Parent.DoRequest(tools.Paths.BackendSettingsGetSshKnownHosts, knownHostsRequest)
	if err != nil {
		return "", err
	}
	var knownHostsStruct tools.SingleString
	err = json.Unmarshal(responseBody, &knownHostsStruct)
	if err != nil {
		return "", err
	}
	return knownHostsStruct.Value, nil
}
