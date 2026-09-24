package app_migrations

import (
	"path/filepath"

	"server/tools"

	u "github.com/quollix/common/utils"
	"gopkg.in/yaml.v3"
)

const (
	PostgresDataVolumeChangedError = "postgres data volume changed"
	PostgresUserChangedError       = "postgres user changed"
)

var ExpectedAppDatabaseMigrationErrors = []string{
	InvalidPostgresEnvironmentError,
	InvalidPostgresDataVolumeError,
	MissingPostgresDataVolumeError,
	MultiplePostgresDataVolumesError,
	MultiplePostgresServicesError,
	MultipleRabbitMQServicesError,
	PostgresDataVolumeChangedError,
	PostgresUserChangedError,
}

const (
	postgresDumpFileName        = "dump.sql"
	postgresDumpMountPath       = "/quollix-migration"
	postgresDumpPathInContainer = postgresDumpMountPath + "/" + postgresDumpFileName
)

type AppDatabaseMigrator interface {
	ValidateForAppUpdate(oldComposeContent, newComposeContent []byte) error
	MigrateForAppUpdate(maintainer, appName string, oldComposeContent, newComposeContent []byte) error
}

type AppDatabaseMigratorImpl struct {
	ComposeDatabaseExtractor ComposeDatabaseExtractor
	CommandExecutor          AppMigrationCommandExecutor
	DockerService            tools.DockerService
	OsWrapper                u.OsWrapper
}

func (m *AppDatabaseMigratorImpl) ValidateForAppUpdate(oldComposeContent, newComposeContent []byte) error {
	_, _, _, err := m.getPostgresMajorUpdate(oldComposeContent, newComposeContent)
	if err != nil {
		return err
	}
	_, _, _, err = m.getRabbitMQVersionUpdate(oldComposeContent, newComposeContent)
	return err
}

func (m *AppDatabaseMigratorImpl) getPostgresMajorUpdate(oldComposeContent, newComposeContent []byte) (*PostgresService, *PostgresService, bool, error) {
	oldPostgres, oldPostgresExists, err := m.ComposeDatabaseExtractor.ExtractPostgresService(oldComposeContent)
	if err != nil {
		return nil, nil, false, err
	}
	if !oldPostgresExists {
		return nil, nil, false, nil
	}

	newPostgres, newPostgresExists, err := m.ComposeDatabaseExtractor.ExtractPostgresService(newComposeContent)
	if err != nil {
		return nil, nil, false, err
	}
	if !newPostgresExists {
		return nil, nil, false, nil
	}

	if oldPostgres.ServiceName != newPostgres.ServiceName || oldPostgres.ImageMajor == newPostgres.ImageMajor {
		return nil, nil, false, nil
	}
	if oldPostgres.User != newPostgres.User {
		return nil, nil, false, u.Logger.NewError(PostgresUserChangedError, "old_user", oldPostgres.User, "new_user", newPostgres.User)
	}
	if oldPostgres.DataVolume != newPostgres.DataVolume {
		return nil, nil, false, u.Logger.NewError(PostgresDataVolumeChangedError, "old_volume", oldPostgres.DataVolume, "new_volume", newPostgres.DataVolume)
	}

	return oldPostgres, newPostgres, true, nil
}

func (m *AppDatabaseMigratorImpl) getRabbitMQVersionUpdate(oldComposeContent, newComposeContent []byte) (*RabbitMQService, *RabbitMQService, bool, error) {
	oldRabbitMQ, oldRabbitMQExists, err := m.ComposeDatabaseExtractor.ExtractRabbitMQService(oldComposeContent)
	if err != nil {
		return nil, nil, false, err
	}
	if !oldRabbitMQExists {
		return nil, nil, false, nil
	}

	newRabbitMQ, newRabbitMQExists, err := m.ComposeDatabaseExtractor.ExtractRabbitMQService(newComposeContent)
	if err != nil {
		return nil, nil, false, err
	}
	if !newRabbitMQExists {
		return nil, nil, false, nil
	}

	if oldRabbitMQ.ServiceName != newRabbitMQ.ServiceName || oldRabbitMQ.ImageVersion == newRabbitMQ.ImageVersion {
		return nil, nil, false, nil
	}

	return oldRabbitMQ, newRabbitMQ, true, nil
}

func (m *AppDatabaseMigratorImpl) MigrateForAppUpdate(maintainer, appName string, oldComposeContent, newComposeContent []byte) error {
	oldPostgres, newPostgres, isPostgresMajorUpdate, err := m.getPostgresMajorUpdate(oldComposeContent, newComposeContent)
	if err != nil {
		return err
	}
	if isPostgresMajorUpdate {
		u.Logger.Info("migrating postgres app database", "service_name", oldPostgres.ServiceName, "old_major", oldPostgres.ImageMajor, "new_major", newPostgres.ImageMajor)
		if err = m.migratePostgres(maintainer, appName, oldComposeContent, newComposeContent, oldPostgres, newPostgres); err != nil {
			return err
		}
	}

	oldRabbitMQ, newRabbitMQ, isRabbitMQVersionUpdate, err := m.getRabbitMQVersionUpdate(oldComposeContent, newComposeContent)
	if err != nil {
		return err
	}
	if !isRabbitMQVersionUpdate {
		return nil
	}

	u.Logger.Info("migrating rabbitmq app service", "service_name", oldRabbitMQ.ServiceName, "old_version", oldRabbitMQ.ImageVersion, "new_version", newRabbitMQ.ImageVersion)
	return m.migrateRabbitMQ(maintainer, appName, oldComposeContent, oldRabbitMQ)
}

func (m *AppDatabaseMigratorImpl) migratePostgres(
	maintainer, appName string,
	oldComposeContent, newComposeContent []byte,
	oldPostgres, newPostgres *PostgresService,
) error {
	tempDir, err := m.OsWrapper.GetTempDir()
	if err != nil {
		return err
	}
	defer m.removeMigrationTempDir(tempDir)

	oldComposePath, err := m.writePostgresMigrationComposeFile(tempDir, "old-docker-compose.yml", oldComposeContent, oldPostgres.ServiceName)
	if err != nil {
		return err
	}
	newComposePath, err := m.writePostgresMigrationComposeFile(tempDir, "new-docker-compose.yml", newComposeContent, newPostgres.ServiceName)
	if err != nil {
		return err
	}

	m.DockerService.CreateDockerNetwork(maintainer, appName)
	defer m.DockerService.RemoveNetwork(maintainer, appName)

	dumpFilePath := filepath.Join(tempDir, postgresDumpFileName)
	if err = m.runWithStartedMigrationService(maintainer, appName, oldComposePath, oldPostgres.ServiceName, func() error {
		if err = m.CommandExecutor.WaitUntilPostgresReady(oldPostgres.ContainerName, oldPostgres.User); err != nil {
			return err
		}
		return m.CommandExecutor.DumpPostgres(oldPostgres.ContainerName, oldPostgres.User, postgresDumpPathInContainer)
	}); err != nil {
		return err
	}

	if err = m.CommandExecutor.RecreateVolume(oldPostgres.DataVolume); err != nil {
		return err
	}

	if err = m.runWithStartedMigrationService(maintainer, appName, newComposePath, newPostgres.ServiceName, func() error {
		if err = m.CommandExecutor.WaitUntilPostgresReady(newPostgres.ContainerName, newPostgres.User); err != nil {
			return err
		}
		if err = m.CommandExecutor.ImportDump(newPostgres.ContainerName, newPostgres.User, postgresDumpPathInContainer); err != nil {
			return err
		}
		if err = m.OsWrapper.Remove(dumpFilePath); err != nil {
			return u.Logger.NewError(err.Error())
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func (m *AppDatabaseMigratorImpl) migrateRabbitMQ(maintainer, appName string, oldComposeContent []byte, rabbitMQ *RabbitMQService) error {
	tempDir, err := m.OsWrapper.GetTempDir()
	if err != nil {
		return err
	}
	defer m.removeMigrationTempDir(tempDir)

	oldComposePath, err := m.writeMigrationComposeFile(tempDir, "old-docker-compose.yml", oldComposeContent)
	if err != nil {
		return err
	}

	m.DockerService.CreateDockerNetwork(maintainer, appName)
	defer m.DockerService.RemoveNetwork(maintainer, appName)

	return m.runWithStartedMigrationService(maintainer, appName, oldComposePath, rabbitMQ.ServiceName, func() error {
		if err = m.CommandExecutor.WaitUntilRabbitMQReady(rabbitMQ.ContainerName); err != nil {
			return err
		}
		return m.CommandExecutor.EnableRabbitMQFeatureFlags(rabbitMQ.ContainerName)
	})
}

func (m *AppDatabaseMigratorImpl) runWithStartedMigrationService(maintainer, appName, composePath, serviceName string, action func() error) error {
	if err := m.CommandExecutor.StartService(maintainer, appName, composePath, serviceName); err != nil {
		return err
	}
	serviceRunning := true
	defer func() {
		if serviceRunning {
			m.CommandExecutor.StopCompose(maintainer, appName, composePath)
		}
	}()

	if err := action(); err != nil {
		return err
	}

	m.CommandExecutor.StopCompose(maintainer, appName, composePath)
	serviceRunning = false
	return nil
}

func (m *AppDatabaseMigratorImpl) removeMigrationTempDir(tempDir string) {
	if err := m.OsWrapper.RemoveAll(tempDir); err != nil {
		u.Logger.Error(err, "temp_dir", tempDir)
	}
}

func (m *AppDatabaseMigratorImpl) writePostgresMigrationComposeFile(tempDir, fileName string, content []byte, postgresServiceName string) (string, error) {
	contentWithDumpMount, err := addServiceVolumeMount(content, postgresServiceName, tempDir, postgresDumpMountPath)
	if err != nil {
		return "", err
	}
	return m.writeMigrationComposeFile(tempDir, fileName, contentWithDumpMount)
}

func addServiceVolumeMount(content []byte, serviceName, hostPath, containerPath string) ([]byte, error) {
	var compose map[string]any
	if err := yaml.Unmarshal(content, &compose); err != nil {
		return nil, u.Logger.NewError(err.Error())
	}

	services, ok := compose["services"].(map[string]any)
	if !ok {
		return nil, u.Logger.NewError("invalid compose services")
	}
	service, ok := services[serviceName].(map[string]any)
	if !ok {
		return nil, u.Logger.NewError("missing compose service", "service_name", serviceName)
	}

	volumeMount := hostPath + ":" + containerPath
	switch volumes := service["volumes"].(type) {
	case nil:
		service["volumes"] = []any{volumeMount}
	case []any:
		service["volumes"] = append(volumes, volumeMount)
	default:
		return nil, u.Logger.NewError("invalid compose service volumes", "service_name", serviceName)
	}

	marshaledCompose, err := yaml.Marshal(compose)
	if err != nil {
		return nil, u.Logger.NewError(err.Error())
	}
	return marshaledCompose, nil
}

func (m *AppDatabaseMigratorImpl) writeMigrationComposeFile(tempDir, fileName string, content []byte) (string, error) {
	composePath := filepath.Join(tempDir, fileName)
	if err := m.OsWrapper.WriteFile(composePath, content, 0o600); err != nil {
		return "", u.Logger.NewError(err.Error())
	}
	return composePath, nil
}
