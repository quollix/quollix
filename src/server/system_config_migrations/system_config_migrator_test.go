package system_config_migrations

import (
	"testing"

	"server/configs"

	"github.com/quollix/common/assert"
)

type systemConfigMigratorTestObjects struct {
	Migrator           *SystemConfigMigratorImpl
	ConfigsRepo        *configs.ConfigsRepositoryMock
	MigrationsProvider *SystemConfigMigrationsProviderMock
}

func newSystemConfigMigratorTestObjects(t *testing.T) systemConfigMigratorTestObjects {
	configRepo := configs.NewConfigsRepositoryMock(t)
	migrationsProvider := NewSystemConfigMigrationsProviderMock(t)
	migrator := &SystemConfigMigratorImpl{
		ConfigsRepo:        configRepo,
		MigrationsProvider: migrationsProvider,
	}

	return systemConfigMigratorTestObjects{
		Migrator:           migrator,
		ConfigsRepo:        configRepo,
		MigrationsProvider: migrationsProvider,
	}
}

func TestSystemConfigMigratorRunsPendingMigrationsAndStoresVersions(t *testing.T) {
	testObjects := newSystemConfigMigratorTestObjects(t)
	executedSteps := []string{}

	testObjects.expectSystemConfigVersionMissing()
	testObjects.expectSystemConfigVersionWrite("1")
	testObjects.expectSystemConfigVersionWrite("2")
	testObjects.expectMigrations(&executedSteps, "1", "2")

	err := testObjects.Migrator.Run()

	assert.Nil(t, err)
	assert.Equal(t, []string{"1", "2"}, executedSteps)
}

func TestSystemConfigMigratorSkipsAlreadyAppliedMigrations(t *testing.T) {
	testObjects := newSystemConfigMigratorTestObjects(t)
	executedSteps := []string{}

	testObjects.expectSystemConfigVersionRead("1")
	testObjects.expectSystemConfigVersionWrite("2")
	testObjects.expectMigrations(&executedSteps, "1", "2")

	err := testObjects.Migrator.Run()

	assert.Nil(t, err)
	assert.Equal(t, []string{"2"}, executedSteps)
}

func (o systemConfigMigratorTestObjects) expectSystemConfigVersionMissing() {
	o.ConfigsRepo.EXPECT().IsConfigSet(configs.ConfigKeys.SystemConfigVersion).Return(false, nil)
}

func (o systemConfigMigratorTestObjects) expectSystemConfigVersionRead(version string) {
	o.ConfigsRepo.EXPECT().IsConfigSet(configs.ConfigKeys.SystemConfigVersion).Return(true, nil)
	o.ConfigsRepo.EXPECT().GetConfig(configs.ConfigKeys.SystemConfigVersion).Return(version, nil)
}

func (o systemConfigMigratorTestObjects) expectSystemConfigVersionWrite(version string) {
	o.ConfigsRepo.EXPECT().
		SetConfig(configs.ConfigKeys.SystemConfigVersion, version).
		Return(nil).
		Once()
}

func (o systemConfigMigratorTestObjects) expectMigrations(executedSteps *[]string, names ...string) {
	migrations := []func() error{}
	for _, name := range names {
		migrationName := name
		migrations = append(migrations, func() error {
			*executedSteps = append(*executedSteps, migrationName)
			return nil
		})
	}
	o.MigrationsProvider.EXPECT().List().Return(migrations)
}
