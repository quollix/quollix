//go:build component

package component

import (
	"reflect"
	"testing"

	"server/backup_server"
	"server/configs"

	"github.com/quollix/common/assert"
)

func TestAssertSampleAppContentUsingHttps(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()
	_, err := InstallAndStartSample(t, client, "2.0")
	assert.Nil(t, err)

	appClient := GetAppClient(t, client)
	assert.Nil(t, AssertSampleAppContent(appClient, "this is version 2.0"))

	client.Parent.RootUrl = "https://localhost"
	appClient = GetAppClient(t, client)
	assert.Nil(t, AssertSampleAppContent(appClient, "this is version 2.0"))
}

func TestBaseDomainSettings(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	host, err := client.Settings.GetBaseDomainValue()
	assert.Nil(t, err)
	assert.Equal(t, "localhost", host)
	assert.Nil(t, client.Settings.SetBaseDomainValue("localhost2"))
	host, err = client.Settings.GetBaseDomainValue()
	assert.Nil(t, err)
	assert.Equal(t, "localhost2", host)
}

func TestSshConfigsSetting(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	initialConfigs, err := client.Settings.ReadSshConfigs()
	assert.Nil(t, err)
	reflect.DeepEqual(initialConfigs, backup_server.GetSampleRemoteRepo())

	settings := backup_server.GetSampleRemoteRepo()
	settings.SshKnownHosts, err = client.Settings.GetKnownHosts(settings)
	assert.Nil(t, err)
	assert.Nil(t, client.Settings.SaveSshConfigs(settings))
	configs, err := client.Settings.ReadSshConfigs()
	assert.Nil(t, err)
	reflect.DeepEqual(configs, backup_server.GetSampleRemoteRepo())
}

func TestObservabilitySettings(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()
	settings, err := client.Settings.ReadObservabilitySettings()
	assert.Nil(t, err)
	assert.Equal(t, configs.NewObservabilitySettings(30, 5, 30, 300, 300, 3600), settings)
	expected := configs.NewObservabilitySettings(14, 10, 60, 420, 660, 5400)
	assert.Nil(t, client.Settings.SaveObservabilitySettings(expected))
	settings, err = client.Settings.ReadObservabilitySettings()
	assert.Nil(t, err)
	assert.Equal(t, expected, settings)
}
