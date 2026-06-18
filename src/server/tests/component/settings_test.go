//go:build component

package component

import (
	"reflect"
	"server/backup_server"
	"testing"

	"github.com/quollix/common/assert"
)

func TestAssertContentUsingHttps(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()
	_, err := client.Apps.InstallSample("2.0")
	assert.Nil(t, err)
	app := client.Apps.GetInstalledSample()

	assert.Nil(t, client.Apps.Start(app.AppId))
	appClient := GetAppClient(t, client)
	assert.Nil(t, appClient.Content.AssertContent("this is version 2.0"))

	client.Parent.RootUrl = "https://localhost"
	appClient = GetAppClient(t, client)
	assert.Nil(t, appClient.Content.AssertContent("this is version 2.0"))
}

func TestHostSettings(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	host, err := client.Settings.GetHostValue()
	assert.Nil(t, err)
	assert.Equal(t, "localhost", host)
	client.Settings.SetHostValue("localhost2")
	host, err = client.Settings.GetHostValue()
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
	settings.SshKnownHosts = client.Settings.GetKnownHosts(settings)
	assert.Nil(t, client.Settings.SaveSshConfigs(settings))
	configs, err := client.Settings.ReadSshConfigs()
	assert.Nil(t, err)
	reflect.DeepEqual(configs, backup_server.GetSampleRemoteRepo())
}
