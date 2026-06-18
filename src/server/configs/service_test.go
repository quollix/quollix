package configs

import (
	"testing"

	"github.com/quollix/common/assert"
)

type configsServiceTestObjects struct {
	Service     *ConfigsServiceImpl
	ConfigsRepo *ConfigsRepositoryMock
}

func newConfigsServiceTestObjects(t *testing.T) configsServiceTestObjects {
	configsRepo := NewConfigsRepositoryMock(t)
	service := &ConfigsServiceImpl{
		ConfigsRepo: configsRepo,
	}

	return configsServiceTestObjects{
		Service:     service,
		ConfigsRepo: configsRepo,
	}
}

func TestConfigsServiceImpl_GetServerHostLoadsValue(t *testing.T) {
	testObjects := newConfigsServiceTestObjects(t)

	testObjects.ConfigsRepo.EXPECT().GetConfig(ConfigKeys.ServerHost).Return("example.com", nil)

	host, err := testObjects.Service.GetServerHost()
	assert.Nil(t, err)
	assert.Equal(t, "example.com", host)

	testObjects.ConfigsRepo.AssertExpectations(t)
}

func TestConfigsServiceImpl_GetServerHostReturnsCachedValue(t *testing.T) {
	testObjects := newConfigsServiceTestObjects(t)
	testObjects.Service.serverHost = "example.com"

	host, err := testObjects.Service.GetServerHost()
	assert.Nil(t, err)
	assert.Equal(t, "example.com", host)

	testObjects.ConfigsRepo.AssertExpectations(t)
}

func TestConfigsServiceImpl_SetServerHostUpdatesCache(t *testing.T) {
	testObjects := newConfigsServiceTestObjects(t)

	testObjects.ConfigsRepo.EXPECT().SetConfig(ConfigKeys.ServerHost, "example.com").Return(nil)

	err := testObjects.Service.SetServerHost("example.com")
	assert.Nil(t, err)

	host, err := testObjects.Service.GetServerHost()
	assert.Nil(t, err)
	assert.Equal(t, "example.com", host)

	testObjects.ConfigsRepo.AssertExpectations(t)
}
