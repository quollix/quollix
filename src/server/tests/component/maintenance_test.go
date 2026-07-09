//go:build component

package component

import (
	"server/maintenance"
	"server/maintenance/retention"
	"server/tests/api_client"
	"server/tools"
	"testing"
	"time"

	"github.com/quollix/common/assert"
	u "github.com/quollix/common/utils"
)

func TestDefaultAutoMaintenanceSettingsForApps(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	apps := ListInstalledApps(t, client)
	assert.Equal(t, 1, len(apps))
	postgresApp := apps[0]
	assert.Equal(t, u.OfficialDatabaseAppName, postgresApp.AppName)
	assert.True(t, postgresApp.AutomaticBackupsEnabled)
	assert.False(t, postgresApp.AutomaticUpdatesEnabled)

	sampleApp, err := InstallSample(t, client, "2.0")
	assert.Nil(t, err)
	assert.True(t, sampleApp.AutomaticUpdatesEnabled)
	assert.True(t, sampleApp.AutomaticBackupsEnabled)
}

func TestAppAutoMaintenanceSettingsChanges(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	sampleApp, err := InstallSample(t, client, "2.0")
	assert.Nil(t, err)
	assert.True(t, sampleApp.AutomaticUpdatesEnabled)
	assert.True(t, sampleApp.AutomaticBackupsEnabled)

	assert.Nil(t, client.Apps.UpdateMaintenanceSettings(sampleApp.AppId, true, false))

	sampleApp = GetInstalledSample(t, client)
	assert.True(t, sampleApp.AutomaticUpdatesEnabled)
	assert.False(t, sampleApp.AutomaticBackupsEnabled)

	assert.Nil(t, client.Apps.UpdateMaintenanceSettings(sampleApp.AppId, false, false))

	sampleApp = GetInstalledSample(t, client)
	assert.False(t, sampleApp.AutomaticUpdatesEnabled)
	assert.False(t, sampleApp.AutomaticBackupsEnabled)
}

func TestMaintenanceSettingsHappyPath(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	maintenanceConfig, err := client.Maintenance.ReadConfigs()
	assert.Nil(t, err)
	assert.Equal(t, 2, maintenanceConfig.MaintenanceWindowStartHour)
	assert.Equal(t, "Europe/London", maintenanceConfig.IanaTimezone)

	assertTimeIsWithinLocalHourWindow(t, maintenanceConfig.NextMaintenanceAt, "Europe/London", 2)

	expected := &maintenance.MaintenanceConfigDto{
		IanaTimezone:               "Europe/Berlin",
		MaintenanceWindowStartHour: 6,
	}
	assert.Nil(t, client.Maintenance.SaveConfigs(expected))

	maintenanceConfig, err = client.Maintenance.ReadConfigs()
	assert.Nil(t, err)
	assert.Equal(t, 6, maintenanceConfig.MaintenanceWindowStartHour)
	assert.Equal(t, "Europe/Berlin", maintenanceConfig.IanaTimezone)

	assertTimeIsWithinLocalHourWindow(t, maintenanceConfig.NextMaintenanceAt, "Europe/Berlin", 6)
}

func assertTimeIsWithinLocalHourWindow(t *testing.T, timestamp time.Time, ianaTimezone string, startHourInclusive int) {
	location, err := time.LoadLocation(ianaTimezone)
	assert.Nil(t, err)
	localTimestamp := timestamp.In(location)
	assert.Equal(t, localTimestamp.Hour(), startHourInclusive)
}

func TestMaintenanceSettingsInvalidInput(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	testCases := []struct {
		name     string
		timezone string
		hour     int
		err      string
	}{
		{"invalid timezone", "Timezone/NotExisting", 6, maintenance.InvalidIanaTimezoneErrorMessage},
		{"hour too small", "Europe/Berlin", -1, maintenance.InvalidMaintenanceWindowStartHourErrorMessage},
		{"hour too large", "Europe/Berlin", 24, maintenance.InvalidMaintenanceWindowStartHourErrorMessage},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			maintenanceConfig := &maintenance.MaintenanceConfigDto{
				IanaTimezone:               testCase.timezone,
				MaintenanceWindowStartHour: testCase.hour,
			}
			err := client.Maintenance.SaveConfigs(maintenanceConfig)
			u.AssertDeepStackErrorFromRequest(t, err, testCase.err)
		})
	}
}

func TestReadingAndSavingRetentionPolicy(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	defaultPolicy := retention.GetDefaultRetentionPolicy()
	newPolicy := &retention.RetentionPolicy{
		KeepDaily:     defaultPolicy.KeepDaily + 1,
		KeepWeekly:    defaultPolicy.KeepWeekly + 1,
		KeepMonthly:   defaultPolicy.KeepMonthly + 1,
		KeepYearly:    defaultPolicy.KeepYearly + 1,
		KeepPreUpdate: defaultPolicy.KeepPreUpdate + 1,
	}

	policyFromBackend, err := client.Maintenance.ReadRetentionPolicy()
	assert.Nil(t, err)
	assert.Equal(t, defaultPolicy, policyFromBackend)

	assert.Nil(t, client.Maintenance.SaveRetentionPolicy(newPolicy))

	policyFromBackend, err = client.Maintenance.ReadRetentionPolicy()
	assert.Nil(t, err)
	assert.Equal(t, newPolicy, policyFromBackend)
}

func TestMaintenanceJobExecutionConductsUpdatesAndCreatesBackups(t *testing.T) {
	client := prepareSshRemoteServerSetup(t)
	defer client.Test.ResetTestState()

	_, err := InstallSample(t, client, "1.0")
	assert.Nil(t, err)

	assertBackupCount(t, client, u.OfficialMaintainer, u.OfficialDatabaseAppName, 0)
	assertBackupCount(t, client, tools.SampleMaintainer, tools.SampleApp, 0)

	assert.Nil(t, client.Maintenance.ExecuteJob())

	sampleApp := GetInstalledSample(t, client)
	assert.Equal(t, "2.0", sampleApp.VersionName)

	assertBackupCount(t, client, u.OfficialMaintainer, u.OfficialDatabaseAppName, 1)
	assertBackupCount(t, client, tools.SampleMaintainer, tools.SampleApp, 2)
}

func assertBackupCount(t *testing.T, client *api_client.QuollixClient, maintainer string, appName string, expectedCount int) []tools.BackupInfo {
	backups, err := client.Backups.ListByApp(maintainer, appName)
	assert.Nil(t, err)
	assert.Equal(t, expectedCount, len(backups))
	return backups
}

func TestSavingRetentionPolicyInvalidInput(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	testCases := []struct {
		name   string
		policy retention.RetentionPolicy
	}{
		{"negative KeepPreUpdate", retention.RetentionPolicy{KeepPreUpdate: -1}},
		{"negative KeepDaily", retention.RetentionPolicy{KeepDaily: -1}},
		{"negative KeepWeekly", retention.RetentionPolicy{KeepWeekly: -1}},
		{"negative KeepMonthly", retention.RetentionPolicy{KeepMonthly: -1}},
		{"negative KeepYearly", retention.RetentionPolicy{KeepYearly: -1}},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := client.Maintenance.SaveRetentionPolicy(&testCase.policy)
			u.AssertDeepStackErrorFromRequest(t, err, maintenance.NegativeRetentionValuesErrors)
		})
	}
}
