package maintenance

import (
	"server/configs"
	"testing"
	"time"

	"server/tools"

	"github.com/quollix/common/assert"
	"github.com/stretchr/testify/mock"
)

type maintenanceServiceTestObjects struct {
	Service          *MaintenanceServiceImpl
	Repository       *configs.MaintenanceRepositoryMock
	OsWrapper        *tools.CommonOsWrapperMock
	TimezoneProvider *tools.TimezoneProviderMock
	AgentHelper      *AgentHelperMock
}

func newMaintenanceServiceTestObjects(t *testing.T) maintenanceServiceTestObjects {
	repository := configs.NewMaintenanceRepositoryMock(t)
	osWrapper := tools.NewCommonOsWrapperMock(t)
	timezoneProvider := tools.NewTimezoneProviderMock(t)
	agentHelper := NewAgentHelperMock(t)

	service := &MaintenanceServiceImpl{
		MaintenanceRepository: repository,
		OsWrapper:             osWrapper,
		TimezoneProvider:      timezoneProvider,
		AgentHelper:           agentHelper,
	}

	return maintenanceServiceTestObjects{
		Service:          service,
		Repository:       repository,
		OsWrapper:        osWrapper,
		TimezoneProvider: timezoneProvider,
		AgentHelper:      agentHelper,
	}
}

func assertMaintenanceServiceExpectations(t *testing.T, testObjects maintenanceServiceTestObjects) {
	testObjects.Repository.AssertExpectations(t)
	testObjects.OsWrapper.AssertExpectations(t)
	testObjects.TimezoneProvider.AssertExpectations(t)
	testObjects.AgentHelper.AssertExpectations(t)
}

func TestSaveMaintenanceConfig_HappyPath(t *testing.T) {
	testObjects := newMaintenanceServiceTestObjects(t)
	defer assertMaintenanceServiceExpectations(t, testObjects)

	timezone := "Europe/Berlin"
	hour := 2

	nowUtc := time.Date(2026, 2, 5, 12, 0, 0, 0, time.UTC)
	newNextMaintenanceAtUtc := time.Date(2026, 2, 6, 1, 0, 0, 0, time.UTC)

	databaseConfig := &configs.MaintenanceConfig{
		MaintenanceWindowStartHour: 5,
		NextMaintenanceAt:          time.Date(2026, 2, 5, 11, 0, 0, 0, time.UTC),
		IanaTimezone:               "UTC",
	}

	expectedConfig := &configs.MaintenanceConfig{
		IanaTimezone:               timezone,
		MaintenanceWindowStartHour: hour,
		NextMaintenanceAt:          newNextMaintenanceAtUtc,
	}

	testObjects.TimezoneProvider.EXPECT().IsIanaTimezoneValid(timezone).Return(true)
	testObjects.Repository.EXPECT().GetMaintenanceConfig().Return(databaseConfig, nil)
	testObjects.OsWrapper.EXPECT().Now().Return(nowUtc)
	testObjects.AgentHelper.EXPECT().CalculateNextMaintenanceAtUtc(nowUtc, timezone, hour).Return(&newNextMaintenanceAtUtc, nil)
	testObjects.Repository.EXPECT().
		SetMaintenanceConfig(mock.Anything).
		Run(func(configPassedToDependency *configs.MaintenanceConfig) {
			assert.Equal(t, expectedConfig, configPassedToDependency)
		}).
		Return(nil)

	err := testObjects.Service.SaveMaintenanceConfig(timezone, hour)
	assert.Nil(t, err)
}
