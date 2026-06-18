package repository

import (
	"server/apps_basic"
	"server/backup_server"
	"server/configs"
	"server/di"
	"server/maintenance/retention"
	"server/tools"
	"server/users"
	"testing"
	"time"

	"github.com/quollix/common/assert"
	u "github.com/quollix/common/utils"
)

var (
	isInitialized = false

	DatabaseUtils     u.DatabaseUtils
	DatabaseConnector *tools.DatabaseConnectorImpl
	UserRepo          *users.UserRepositoryImpl
	SessionRepo       *users.SessionRepositoryImpl
	AppRepo           *apps_basic.AppRepositoryImpl
	ConfigRepo        *configs.ConfigsRepositoryImpl
	SshRepo           *backup_server.SshRepositoryImpl
	EmailRepo         *configs.EmailRepoImpl
	RetentionRepo     *retention.RetentionPolicyRepositoryImpl
	MaintenanceRepo   *configs.MaintenanceRepositoryImpl
)

func InitDeps() {
	if isInitialized {
		return
	}

	globalConfig := di.NewGlobalConfig()
	dirProvider := tools.NewDirectoryProvider(globalConfig)

	err := dirProvider.InitializeDirectories()
	if err != nil {
		panic(err)
	}

	DatabaseUtils = &u.DatabaseUtilsImpl{}
	dockerService := &tools.DockerServiceImpl{}
	DatabaseConnector = &tools.DatabaseConnectorImpl{
		Config:            globalConfig,
		DirectoryProvider: dirProvider,
		DockerService:     dockerService,
		DatabaseUtils:     DatabaseUtils,
	}

	DatabaseConnector.Config.DatabaseHostName = "localhost"
	err = DatabaseConnector.StartDatabaseAndConnect()
	if err != nil {
		panic(err)
	}

	UserRepo = &users.UserRepositoryImpl{DbProvider: DatabaseConnector}
	SessionRepo = &users.SessionRepositoryImpl{DbProvider: DatabaseConnector}
	AppRepo = &apps_basic.AppRepositoryImpl{DbProvider: DatabaseConnector}
	ConfigRepo = &configs.ConfigsRepositoryImpl{DbProvider: DatabaseConnector}
	SshRepo = &backup_server.SshRepositoryImpl{DbProvider: DatabaseConnector}
	EmailRepo = &configs.EmailRepoImpl{DatabaseConnector: DatabaseConnector}
	RetentionRepo = &retention.RetentionPolicyRepositoryImpl{DatabaseConnector: DatabaseConnector}
	MaintenanceRepo = &configs.MaintenanceRepositoryImpl{DatabaseConnector: DatabaseConnector}
	isInitialized = true
}

func GetSampleAdminUser() *tools.User {
	return users.NewUser(
		"admin",
		"admin@example.com",
		"hashed-password",
		"cookieValue",
		time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC),
		true,
		"set-password-token",
		time.Date(2021, time.February, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2022, time.March, 1, 0, 0, 0, 0, time.UTC),
	)
}

func AssertAppEquality(t *testing.T, expected, actual *apps_basic.RepoApp) {
	assert.Equal(t, expected.AppId, actual.AppId)
	assert.Equal(t, expected.Maintainer, actual.Maintainer)
	assert.Equal(t, expected.AppName, actual.AppName)
	assert.Equal(t, expected.VersionName, actual.VersionName)
	assert.Equal(t, expected.VersionCreationTimestamp, actual.VersionCreationTimestamp)
	assert.Equal(t, expected.VersionContent, actual.VersionContent)
	assert.Equal(t, expected.ShouldBeRunning, actual.ShouldBeRunning)
	assert.Equal(t, expected.AccessPolicy, actual.AccessPolicy)
	assert.Equal(t, expected.Port, actual.Port)
	assert.Equal(t, expected.ClientId, actual.ClientId)
	assert.Equal(t, expected.ClientSecret, actual.ClientSecret)
	assert.Equal(t, len(expected.Metadata), len(actual.Metadata))
	assert.Equal(t, expected.Metadata, actual.Metadata)
}
