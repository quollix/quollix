package alerts

import (
	"testing"

	"github.com/quollix/common/assert"
	"github.com/quollix/common/quollix/api"
	u "github.com/quollix/common/utils"
)

func TestAlertSettingsService_SaveAllowsZeroPendingTime(t *testing.T) {
	repository := NewAlertSettingsRepositoryMock(t)
	service := &AlertSettingsServiceImpl{Repository: repository}
	settings := DefaultAlertSettings()
	settings.HostUsagePendingMinutes = 0
	repository.EXPECT().Write(settings).Return(nil)

	assert.Nil(t, service.Save(settings))
}

func TestAlertSettingsService_SaveRejectsInvalidSettings(t *testing.T) {
	testCases := []struct {
		name   string
		update func(*api.AlertSettings)
	}{
		{name: "trigger below minimum", update: func(settings *api.AlertSettings) { settings.HostUsageTriggerPercent = 9 }},
		{name: "trigger above maximum", update: func(settings *api.AlertSettings) { settings.HostUsageTriggerPercent = 101 }},
		{name: "recovery below minimum", update: func(settings *api.AlertSettings) { settings.HostUsageRecoveryPercent = 9 }},
		{name: "recovery above maximum", update: func(settings *api.AlertSettings) { settings.HostUsageRecoveryPercent = 101 }},
		{name: "recovery equals trigger", update: func(settings *api.AlertSettings) {
			settings.HostUsageRecoveryPercent = settings.HostUsageTriggerPercent
		}},
		{name: "recovery above trigger", update: func(settings *api.AlertSettings) {
			settings.HostUsageRecoveryPercent = settings.HostUsageTriggerPercent + 1
		}},
		{name: "negative pending time", update: func(settings *api.AlertSettings) { settings.HostUsagePendingMinutes = -1 }},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			repository := NewAlertSettingsRepositoryMock(t)
			service := &AlertSettingsServiceImpl{Repository: repository}
			settings := DefaultAlertSettings()
			testCase.update(settings)

			err := service.Save(settings)
			assert.Equal(t, InvalidAlertSettingsError, u.ExtractError(err))
		})
	}
}
