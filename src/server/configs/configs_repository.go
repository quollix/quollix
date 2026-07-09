package configs

import (
	"server/tools"

	u "github.com/quollix/common/utils"
)

type ConfigsRepository interface {
	IsConfigSet(key string) (bool, error)
	GetConfig(key string) (string, error)
	SetConfig(configFieldName string, value string) error
	DeleteConfig(key string) error
}

type ConfigsRepositoryImpl struct {
	DbProvider tools.DatabaseConnector
}

func (c *ConfigsRepositoryImpl) IsConfigSet(key string) (bool, error) {
	var exists bool
	err := c.DbProvider.GetDB().QueryRow("SELECT EXISTS(SELECT 1 FROM configs WHERE key = $1)", key).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (c *ConfigsRepositoryImpl) GetConfig(key string) (string, error) {
	var value string
	err := c.DbProvider.GetDB().QueryRow("SELECT value FROM configs WHERE key = $1", key).Scan(&value)
	if err != nil {
		return "", err
	}
	return value, nil
}

func (c *ConfigsRepositoryImpl) SetConfig(configFieldName string, value string) error {
	_, err := c.DbProvider.GetDB().Exec("INSERT INTO configs (key, value) VALUES ($1, $2) ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value", configFieldName, value)
	if err != nil {
		return u.Logger.NewError(err.Error(), tools.ConfigKeyField, configFieldName)
	}
	return nil
}

func (c *ConfigsRepositoryImpl) DeleteConfig(key string) error {
	_, err := c.DbProvider.GetDB().Exec("DELETE FROM configs WHERE key = $1", key)
	if err != nil {
		return u.Logger.NewError(err.Error(), tools.ConfigKeyField, key)
	}
	return nil
}

// only for testing purposes
func (c *ConfigsRepositoryImpl) Wipe() {
	_, err := c.DbProvider.GetDB().Exec("DELETE FROM configs")
	if err != nil {
		u.Logger.Error(err)
	}
}
