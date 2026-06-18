package tools

import (
	"database/sql"
	"path/filepath"

	u "github.com/quollix/common/utils"
)

const (
	DefaultPostgresPort         = "5432"
	PostgresApplicationDatabase = "application"
)

type DatabaseConnector interface {
	StartDatabaseAndConnect() error
	GetDB() *sql.DB
	Connect() error
}

type DatabaseConnectorImpl struct {
	Config            *GlobalConfig
	db                *sql.DB
	DockerService     DockerService
	DirectoryProvider DirectoryProvider
	DatabaseUtils     u.DatabaseUtils
}

func NewDatabaseConnector(
	config *GlobalConfig,
	dockerService DockerService,
	directoryProvider DirectoryProvider,
	databaseUtils u.DatabaseUtils,
) *DatabaseConnectorImpl {
	return &DatabaseConnectorImpl{
		Config:            config,
		DockerService:     dockerService,
		DirectoryProvider: directoryProvider,
		DatabaseUtils:     databaseUtils,
	}
}

func (d *DatabaseConnectorImpl) StartDatabaseAndConnect() error {
	d.DockerService.CreateDockerNetwork(u.OfficialMaintainer, u.OfficialDatabaseAppName)
	// since database it not reachable thorugh the Quollix proxy, we can set any host as value here
	d.DockerService.AttachBrandAppToNetwork(u.OfficialMaintainer, u.OfficialDatabaseAppName, "localhost")
	composeFilePath := filepath.Join(d.DirectoryProvider.GetOfficialDatabaseDockerDir(), "docker-compose.yml")
	if err := d.DockerService.StartAppContainer(u.OfficialMaintainer, u.OfficialDatabaseAppName, composeFilePath, nil); err != nil {
		return err
	}

	if err := d.DatabaseUtils.EnsureDatabaseExists(d.Config.DatabaseHostName, DefaultPostgresPort, PostgresApplicationDatabase); err != nil {
		return err
	}
	return d.Connect()
}

func (d *DatabaseConnectorImpl) GetDB() *sql.DB {
	if d.db == nil {
		u.Logger.Error("attempted to access database before connecting to it")
		return nil
	}
	return d.db
}

func (d *DatabaseConnectorImpl) Connect() error {
	err := d.DatabaseUtils.EnsureDatabaseExists(d.Config.DatabaseHostName, DefaultPostgresPort, PostgresApplicationDatabase)
	if err != nil {
		return err
	}

	d.db, err = d.DatabaseUtils.WaitForPostgresDb(d.Config.DatabaseHostName, DefaultPostgresPort, PostgresApplicationDatabase)
	if err != nil {
		return err
	}

	err = d.DatabaseUtils.RunMigrations(d.DirectoryProvider.GetMigrationsDir(), d.Config.DatabaseHostName, DefaultPostgresPort, PostgresApplicationDatabase)
	if err != nil {
		return err
	}

	u.Logger.Info("database initialized successfully")
	return nil
}
