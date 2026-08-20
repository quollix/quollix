package app_migrations

import (
	"testing"

	"server/tools"

	"github.com/quollix/common/assert"
	u "github.com/quollix/common/utils"
	"github.com/stretchr/testify/mock"
)

const (
	sampleMaintainer = "sample-maintainer"
	sampleApp        = "sample-app"
	sampleProject    = sampleMaintainer + "_" + sampleApp
)

var oldComposeContent = []byte("old-compose")
var newComposeContent = []byte("new-compose")

type appDatabaseMigratorTestObjects struct {
	Migrator                 *AppDatabaseMigratorImpl
	ComposeDatabaseExtractor *ComposeDatabaseExtractorMock
	CommandExecutor          *AppMigrationCommandExecutorMock
	DockerService            *tools.DockerServiceMock
	OsWrapper                *tools.CommonOsWrapperMock
}

func newAppDatabaseMigratorTestObjects(t *testing.T) appDatabaseMigratorTestObjects {
	composeDatabaseExtractor := NewComposeDatabaseExtractorMock(t)
	commandExecutor := NewAppMigrationCommandExecutorMock(t)
	dockerService := tools.NewDockerServiceMock(t)
	osWrapper := tools.NewCommonOsWrapperMock(t)

	return appDatabaseMigratorTestObjects{
		Migrator: &AppDatabaseMigratorImpl{
			ComposeDatabaseExtractor: composeDatabaseExtractor,
			CommandExecutor:          commandExecutor,
			DockerService:            dockerService,
			OsWrapper:                osWrapper,
		},
		ComposeDatabaseExtractor: composeDatabaseExtractor,
		CommandExecutor:          commandExecutor,
		DockerService:            dockerService,
		OsWrapper:                osWrapper,
	}
}

func TestMigrateForAppUpdate_NoPostgresMajorUpgradeDoesNothing(t *testing.T) {
	testObjects := newAppDatabaseMigratorTestObjects(t)

	testObjects.ComposeDatabaseExtractor.EXPECT().ExtractPostgresService(oldComposeContent).Return(postgresService(17), true, nil)
	testObjects.ComposeDatabaseExtractor.EXPECT().ExtractPostgresService(newComposeContent).Return(postgresService(17), true, nil)
	testObjects.ComposeDatabaseExtractor.EXPECT().ExtractRabbitMQService(oldComposeContent).Return(nil, false, nil)

	err := testObjects.Migrator.MigrateForAppUpdate(sampleMaintainer, sampleApp, oldComposeContent, newComposeContent)

	assert.Nil(t, err)
}

func TestValidateForAppUpdate_ChangedPostgresUserReturnsError(t *testing.T) {
	testObjects := newAppDatabaseMigratorTestObjects(t)

	newPostgresService := postgresService(18)
	newPostgresService.User = "other"
	testObjects.ComposeDatabaseExtractor.EXPECT().ExtractPostgresService(oldComposeContent).Return(postgresService(17), true, nil)
	testObjects.ComposeDatabaseExtractor.EXPECT().ExtractPostgresService(newComposeContent).Return(newPostgresService, true, nil)

	err := testObjects.Migrator.ValidateForAppUpdate(oldComposeContent, newComposeContent)

	assert.Equal(t, PostgresUserChangedError, u.ExtractError(err))
}

func TestValidateForAppUpdate_ChangedPostgresVolumeReturnsError(t *testing.T) {
	testObjects := newAppDatabaseMigratorTestObjects(t)

	newPostgresService := postgresService(18)
	newPostgresService.DataVolume = "quollix_sample_postgres_new"
	testObjects.ComposeDatabaseExtractor.EXPECT().ExtractPostgresService(oldComposeContent).Return(postgresService(17), true, nil)
	testObjects.ComposeDatabaseExtractor.EXPECT().ExtractPostgresService(newComposeContent).Return(newPostgresService, true, nil)

	err := testObjects.Migrator.ValidateForAppUpdate(oldComposeContent, newComposeContent)

	assert.Equal(t, PostgresDataVolumeChangedError, u.ExtractError(err))
}

func TestMigrateForAppUpdate_PostgresMajorUpgradeUsesComposeServiceDumpAndRestore(t *testing.T) {
	testObjects := newAppDatabaseMigratorTestObjects(t)

	testObjects.ComposeDatabaseExtractor.EXPECT().ExtractPostgresService(oldComposeContent).Return(postgresService(17), true, nil)
	testObjects.ComposeDatabaseExtractor.EXPECT().ExtractPostgresService(newComposeContent).Return(postgresService(18), true, nil)
	testObjects.ComposeDatabaseExtractor.EXPECT().ExtractRabbitMQService(oldComposeContent).Return(nil, false, nil)
	testObjects.DockerService.EXPECT().CreateDockerNetwork(sampleMaintainer, sampleApp).Return()
	testObjects.DockerService.EXPECT().RemoveNetwork(sampleMaintainer, sampleApp).Return()
	testObjects.CommandExecutor.EXPECT().StartService(sampleMaintainer, sampleApp, mock.Anything, "postgres").Return(nil)
	testObjects.CommandExecutor.EXPECT().StopCompose(sampleMaintainer, sampleApp, mock.Anything).Return()
	testObjects.CommandExecutor.EXPECT().WaitUntilPostgresReady("quollix_sample_postgres", "sample").Return(nil)
	testObjects.CommandExecutor.EXPECT().DumpPostgres("quollix_sample_postgres", "sample", postgresDumpPathInContainer).Return(nil)
	testObjects.CommandExecutor.EXPECT().CopyFromContainer("quollix_sample_postgres", postgresDumpPathInContainer, mock.Anything).Return(nil)
	testObjects.CommandExecutor.EXPECT().RemoveFileInContainer("quollix_sample_postgres", postgresDumpPathInContainer).Return(nil)
	testObjects.CommandExecutor.EXPECT().RecreateVolume("quollix_sample_postgres").Return(nil)
	testObjects.CommandExecutor.EXPECT().CopyToContainer(mock.Anything, "quollix_sample_postgres", postgresDumpPathInContainer).Return(nil)
	testObjects.CommandExecutor.EXPECT().ImportDump("quollix_sample_postgres", "sample", postgresDumpPathInContainer).Return(nil)
	testObjects.OsWrapper.EXPECT().Remove(mock.Anything).Return(nil)

	err := testObjects.Migrator.MigrateForAppUpdate(sampleMaintainer, sampleApp, oldComposeContent, newComposeContent)

	assert.Nil(t, err)
}

func TestMigrateForAppUpdate_RabbitMQVersionUpdateEnablesFeatureFlags(t *testing.T) {
	testObjects := newAppDatabaseMigratorTestObjects(t)

	testObjects.ComposeDatabaseExtractor.EXPECT().ExtractPostgresService(oldComposeContent).Return(nil, false, nil)
	testObjects.ComposeDatabaseExtractor.EXPECT().ExtractRabbitMQService(oldComposeContent).Return(rabbitMQService("3.11.18-management-alpine"), true, nil)
	testObjects.ComposeDatabaseExtractor.EXPECT().ExtractRabbitMQService(newComposeContent).Return(rabbitMQService("3.12.14-management-alpine"), true, nil)
	testObjects.DockerService.EXPECT().CreateDockerNetwork(sampleMaintainer, sampleApp).Return()
	testObjects.DockerService.EXPECT().RemoveNetwork(sampleMaintainer, sampleApp).Return()
	testObjects.CommandExecutor.EXPECT().StartService(sampleMaintainer, sampleApp, mock.Anything, "rabbitmq").Return(nil)
	testObjects.CommandExecutor.EXPECT().WaitUntilRabbitMQReady("quollix_sample_rabbitmq").Return(nil)
	testObjects.CommandExecutor.EXPECT().EnableRabbitMQFeatureFlags("quollix_sample_rabbitmq").Return(nil)
	testObjects.CommandExecutor.EXPECT().StopCompose(sampleMaintainer, sampleApp, mock.Anything).Return()

	err := testObjects.Migrator.MigrateForAppUpdate(sampleMaintainer, sampleApp, oldComposeContent, newComposeContent)

	assert.Nil(t, err)
}

func postgresService(imageMajor int) *PostgresService {
	return &PostgresService{
		ServiceName:   "postgres",
		ContainerName: "quollix_sample_postgres",
		ImageMajor:    imageMajor,
		User:          "sample",
		DataVolume:    "quollix_sample_postgres",
	}
}

func rabbitMQService(imageVersion string) *RabbitMQService {
	return &RabbitMQService{
		ServiceName:   "rabbitmq",
		ContainerName: "quollix_sample_rabbitmq",
		ImageVersion:  imageVersion,
	}
}
