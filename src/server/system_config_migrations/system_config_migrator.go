package system_config_migrations

import (
	"strconv"

	"server/configs"

	u "github.com/quollix/common/utils"
)

type SystemConfigMigrationsProvider interface {
	List() []func() error
}

type SystemConfigMigrator interface {
	Run() error
}

type SystemConfigMigratorImpl struct {
	ConfigsRepo        configs.ConfigsRepository
	MigrationsProvider SystemConfigMigrationsProvider
}

func (s *SystemConfigMigratorImpl) Run() error {
	currentVersion, err := s.readVersion()
	if err != nil {
		return err
	}

	for index, migration := range s.MigrationsProvider.List() {
		targetVersion := index + 1
		if targetVersion <= currentVersion {
			continue
		}

		u.Logger.Info("running system config migration", "version", targetVersion)
		if err := migration(); err != nil {
			return err
		}
		if err := s.ConfigsRepo.SetConfig(configs.ConfigKeys.SystemConfigVersion, strconv.Itoa(targetVersion)); err != nil {
			return err
		}
	}
	return nil
}

func (s *SystemConfigMigratorImpl) readVersion() (int, error) {
	isSet, err := s.ConfigsRepo.IsConfigSet(configs.ConfigKeys.SystemConfigVersion)
	if err != nil {
		return 0, err
	}
	if !isSet {
		return 0, nil
	}

	versionValue, err := s.ConfigsRepo.GetConfig(configs.ConfigKeys.SystemConfigVersion)
	if err != nil {
		return 0, err
	}
	version, err := strconv.Atoi(versionValue)
	if err != nil {
		return 0, u.Logger.NewError("invalid system config version", "version", versionValue)
	}
	if version < 0 {
		return 0, u.Logger.NewError("invalid system config version", "version", versionValue)
	}
	return version, nil
}
