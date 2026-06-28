//go:build component

package component

import (
	"reflect"
	"server/backup_server"
	"testing"

	"github.com/quollix/common/assert"
)

func TestAssertSampleAppContentUsingHttps(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()
	_, err := InstallSample(t, client, "2.0")
	assert.Nil(t, err)
	app := GetInstalledSample(t, client)

	assert.Nil(t, client.Apps.Start(app.AppId))
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
