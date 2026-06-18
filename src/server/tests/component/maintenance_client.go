package component

import (
	"encoding/json"
	"server/configs"
	"server/maintenance"
	"server/maintenance/retention"
	"server/tools"

	"github.com/quollix/common/assert"
)

type QuollixMaintenanceClient struct {
	quollix *QuollixClient
}

func (c *QuollixMaintenanceClient) SaveConfigs(request *maintenance.MaintenanceConfigDto) error {
	_, err := c.quollix.Parent.DoRequestWithFullResponse(tools.Paths.BackendMaintenanceConfigsSave, request)
	return err
}

func (c *QuollixMaintenanceClient) ReadConfigs() (*configs.MaintenanceConfig, error) {
	body, err := c.quollix.Parent.DoRequest(tools.Paths.BackendMaintenanceConfigsRead, nil)
	if err != nil {
		return nil, err
	}
	var config *configs.MaintenanceConfig
	err = json.Unmarshal(body, &config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func (c *QuollixMaintenanceClient) ReadRetentionPolicy() *retention.RetentionPolicy {
	body, err := c.quollix.Parent.DoRequest(tools.Paths.BackendMaintenanceRetentionPolicyRead, nil)
	assert.Nil(c.quollix.T, err)

	var policy retention.RetentionPolicy
	err = json.Unmarshal(body, &policy)
	assert.Nil(c.quollix.T, err)

	return &policy
}

func (c *QuollixMaintenanceClient) SaveRetentionPolicy(policy *retention.RetentionPolicy) error {
	_, err := c.quollix.Parent.DoRequestWithFullResponse(tools.Paths.BackendMaintenanceRetentionPolicySave, policy)
	return err
}

func (c *QuollixMaintenanceClient) ExecuteJob() {
	_, err := c.quollix.Parent.DoRequest(tools.Paths.BackendMaintenanceTriggerMaintenanceJob, nil)
	assert.Nil(c.quollix.T, err)
}
