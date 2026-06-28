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

func TestConfigsServiceImpl_GetBaseDomainLoadsValue(t *testing.T) {
	testObjects := newConfigsServiceTestObjects(t)

	testObjects.ConfigsRepo.EXPECT().GetConfig(ConfigKeys.BaseDomain).Return("example.com", nil)

	host, err := testObjects.Service.GetBaseDomain()
	assert.Nil(t, err)
	assert.Equal(t, "example.com", host)

	testObjects.ConfigsRepo.AssertExpectations(t)
}

func TestConfigsServiceImpl_GetBaseDomainReturnsCachedValue(t *testing.T) {
	testObjects := newConfigsServiceTestObjects(t)
	testObjects.Service.baseDomain = "example.com"

	host, err := testObjects.Service.GetBaseDomain()
	assert.Nil(t, err)
	assert.Equal(t, "example.com", host)

	testObjects.ConfigsRepo.AssertExpectations(t)
}

func TestConfigsServiceImpl_SetBaseDomainUpdatesCache(t *testing.T) {
	testObjects := newConfigsServiceTestObjects(t)

	testObjects.ConfigsRepo.EXPECT().SetConfig(ConfigKeys.BaseDomain, "example.com").Return(nil)

	err := testObjects.Service.SetBaseDomain("example.com")
	assert.Nil(t, err)

	host, err := testObjects.Service.GetBaseDomain()
	assert.Nil(t, err)
	assert.Equal(t, "example.com", host)

	testObjects.ConfigsRepo.AssertExpectations(t)
}
