//go:build component

package component

import (
	"slices"
	"testing"
	"time"

	"server/tools"

	"github.com/quollix/common/assert"
	"github.com/quollix/common/quollix/api"
	"github.com/quollix/common/quollix/api_client"
	u "github.com/quollix/common/utils"
)

func TestMonitoring_CollectsSampleAppStdoutAndStderr(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	grafana := InstallGrafana(t, client)
	app, err := InstallSample(t, client, tools.SampleAppVersion1Name)
	assert.Nil(t, err)

	records := waitForSampleStartupLogs(t, client)
	firstBootID := assertSampleStartupPair(t, records)

	assert.Nil(t, client.Apps.Stop(app.AppId))
	assert.Nil(t, client.Logs.Wipe())
	assert.Nil(t, client.Apps.Start(app.AppId))

	recordsAfterRestart := waitForSampleStartupLogs(t, client, firstBootID)
	secondBootId := assertSampleStartupPair(t, recordsAfterRestart)
	assert.NotEqual(t, firstBootID, secondBootId)

	assert.Nil(t, client.Logs.Wipe())
	assert.Nil(t, client.Apps.Stop(grafana.AppId))
	assert.Nil(t, client.Apps.Stop(app.AppId))
	assert.Nil(t, client.Apps.Start(app.AppId))
	assert.Nil(t, client.Apps.Start(grafana.AppId))

	recordsAfterMonitoringRestart := waitForSampleStartupLogs(t, client, firstBootID, secondBootId)
	thirdBootId := assertSampleStartupPair(t, recordsAfterMonitoringRestart)
	assert.NotEqual(t, secondBootId, thirdBootId)

}

func assertSampleStartupPair(t *testing.T, records []api.LogRecord) string {
	assert.Equal(t, 2, len(records))
	bootID, ok := records[0].Attributes["boot_id"].(string)
	assert.True(t, ok)
	assert.NotEqual(t, "", bootID)
	for _, record := range records {
		assert.Equal(t, tools.SampleMaintainer, record.Maintainer)
		assert.Equal(t, tools.SampleApp, record.App)
		assert.Equal(t, tools.SampleApp, record.Service)
		assert.Equal(t, "sampleapp started", record.Message)
		assert.Equal(t, "info", record.Level)
		assert.True(t, record.Parsed)
		assert.False(t, record.ReadTimestamp.IsZero())
		assert.NotNil(t, record.LogTimestamp)
		assert.False(t, record.LogTimestamp.IsZero())
		assert.Equal(t, bootID, record.Attributes["boot_id"])
		assert.Equal(t, record.Attributes["stream"], record.Stream)
	}
	streams := []string{
		records[0].Attributes["stream"].(string),
		records[1].Attributes["stream"].(string),
	}
	slices.Sort(streams)
	assert.Equal(t, []string{"stderr", "stdout"}, streams)
	return bootID
}

func waitForSampleStartupLogs(t *testing.T, client *api_client.QuollixClient, previousBootIDs ...string) []api.LogRecord {
	var startupRecords []api.LogRecord
	err := u.EventuallyWithTimeout(20*time.Second, 200*time.Millisecond, func() error {
		if err := client.Logs.Flush(); err != nil {
			return err
		}
		records, err := client.Logs.Read()
		if err != nil {
			return err
		}
		startupRecords = nil
		for _, record := range records {
			bootID, _ := record.Attributes["boot_id"].(string)
			if slices.Contains(previousBootIDs, bootID) {
				continue
			}
			if record.Maintainer == tools.SampleMaintainer && record.App == tools.SampleApp && record.Message == "sampleapp started" {
				startupRecords = append(startupRecords, record)
			}
		}
		if len(startupRecords) < 2 {
			return u.Logger.NewError("waiting for stdout and stderr startup logs", "records", len(startupRecords))
		}
		return nil
	})
	assert.Nil(t, err)
	return startupRecords
}
