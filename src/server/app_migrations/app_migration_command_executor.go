package app_migrations

import (
	"time"

	"server/tools"

	u "github.com/quollix/common/utils"
)

const (
	databaseReadyAttempts     = 30
	postgresMaintenanceDbName = "postgres"
)

type AppMigrationCommandExecutor interface {
	StartService(maintainer, appName, composePath, serviceName string) error
	StopCompose(maintainer, appName, composePath string)
	WaitUntilPostgresReady(containerName, postgresUser string) error
	DumpPostgres(containerName, postgresUser, dumpPathInContainer string) error
	ImportDump(containerName, postgresUser, dumpPathInContainer string) error
	RecreateVolume(volume string) error
	WaitUntilRabbitMQReady(containerName string) error
	EnableRabbitMQFeatureFlags(containerName string) error
}

type AppMigrationCommandExecutorImpl struct {
	CommandRunner tools.CommandRunner
}

func (e *AppMigrationCommandExecutorImpl) StartService(maintainer, appName, composePath, serviceName string) error {
	_, err := e.CommandRunner.RunCommand("docker", "compose", "-p", maintainer+"_"+appName, "-f", composePath, "up", "-d", "--no-deps", serviceName)
	return err
}

func (e *AppMigrationCommandExecutorImpl) StopCompose(maintainer, appName, composePath string) {
	if _, err := e.CommandRunner.RunCommand("docker", "compose", "-p", maintainer+"_"+appName, "-f", composePath, "down"); err != nil {
		u.Logger.Error(err, "maintainer", maintainer, "app_name", appName, "compose_path", composePath)
	}
}

func (e *AppMigrationCommandExecutorImpl) WaitUntilPostgresReady(containerName, postgresUser string) error {
	var lastErr error
	for range databaseReadyAttempts {
		// Verify the same maintenance database connection that pg_dumpall restore uses.
		_, err := e.CommandRunner.RunCommand("docker", "exec", containerName, "psql", "-U", postgresUser, "-d", postgresMaintenanceDbName, "-c", "SELECT 1")
		if err == nil {
			return nil
		}
		lastErr = err
		time.Sleep(time.Second)
	}
	return lastErr
}

func (e *AppMigrationCommandExecutorImpl) DumpPostgres(containerName, postgresUser, dumpPathInContainer string) error {
	_, err := e.CommandRunner.RunCommand("docker", "exec", containerName, "pg_dumpall", "-U", postgresUser, "-f", dumpPathInContainer)
	return err
}

func (e *AppMigrationCommandExecutorImpl) ImportDump(containerName, postgresUser, dumpPathInContainer string) error {
	// pg_dumpall restores should connect to a maintenance DB, not the app DB being recreated.
	_, err := e.CommandRunner.RunCommand("docker", "exec", containerName, "psql", "-U", postgresUser, "-d", postgresMaintenanceDbName, "-f", dumpPathInContainer)
	return err
}

func (e *AppMigrationCommandExecutorImpl) RecreateVolume(volume string) error {
	if _, err := e.CommandRunner.RunCommand("docker", "volume", "rm", "-f", volume); err != nil {
		return err
	}
	_, err := e.CommandRunner.RunCommand("docker", "volume", "create", volume)
	return err
}

func (e *AppMigrationCommandExecutorImpl) WaitUntilRabbitMQReady(containerName string) error {
	var lastErr error
	for range databaseReadyAttempts {
		_, err := e.CommandRunner.RunCommand("docker", "exec", containerName, "rabbitmq-diagnostics", "-q", "check_running")
		if err == nil {
			return nil
		}
		lastErr = err
		time.Sleep(time.Second)
	}
	return lastErr
}

func (e *AppMigrationCommandExecutorImpl) EnableRabbitMQFeatureFlags(containerName string) error {
	_, err := e.CommandRunner.RunCommand("docker", "exec", containerName, "rabbitmqctl", "enable_feature_flag", "all")
	return err
}
