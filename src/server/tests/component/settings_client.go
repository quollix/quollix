package component

import (
	"encoding/json"
	"server/backup_server"
	"server/tools"

	"github.com/quollix/common/assert"
)

type QuollixSettingsClient struct {
	quollix *QuollixClient
}

func (c *QuollixSettingsClient) SetHostValue(host string) {
	_, err := c.quollix.Parent.DoRequest(tools.Paths.BackendSettingsHostSave, tools.NumberString{Value: host})
	assert.Nil(c.quollix.T, err)
}

func (c *QuollixSettingsClient) GetHostValue() (string, error) {
	responseBody, err := c.quollix.Parent.DoRequest(tools.Paths.BackendSettingsHostRead, nil)
	if err != nil {
		return "", err
	}
	var host string
	err = json.Unmarshal(responseBody, &host)
	if err != nil {
		return "", err
	}
	return host, nil
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

func (c *QuollixSettingsClient) GetKnownHosts(repo *tools.BackupServerConfigs) string {
	knownHostsRequest := backup_server.KnownHostsRequest{
		Host: repo.Host,
		Port: repo.SshPort,
	}
	responseBody, err := c.quollix.Parent.DoRequest(tools.Paths.BackendSettingsGetSshKnownHosts, knownHostsRequest)
	assert.Nil(c.quollix.T, err)
	var knownHostsStruct tools.SingleString
	err = json.Unmarshal(responseBody, &knownHostsStruct)
	assert.Nil(c.quollix.T, err)
	return knownHostsStruct.Value
}
