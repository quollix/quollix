//go:build component

package component

import (
	"database/sql"
	"fmt"
	"net/url"
	"os/exec"
	"strings"
	"testing"
	"time"

	"server/app_migrations"
	"server/apps_basic"
	"server/tools"

	"github.com/quollix/common/assert"
	"github.com/quollix/common/quollix/api"
	"github.com/quollix/common/quollix/api_client"
	u "github.com/quollix/common/utils"

	_ "github.com/lib/pq"
)

const (
	samplePostgresUser  = "postgresapp"
	samplePostgresDb    = "postgresapp"
	samplePostgresValue = "persisted before postgres migration"
)

func TestPostgresMajorUpdateMigratesData(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	appBeforeUpdate, err := installSamplePostgresApp(t, client, tools.SamplePostgresAppVersion17Name)
	assert.Nil(t, err)
	assert.Nil(t, client.Apps.Start(appBeforeUpdate.AppId))

	db := openSamplePostgresDb(t, client)
	writeSamplePostgresValue(t, db, samplePostgresValue)
	assert.Equal(t, samplePostgresValue, readSamplePostgresValue(t, db))
	assert.Nil(t, db.Close())

	assert.Nil(t, client.Apps.Update(appBeforeUpdate.AppId))

	appAfterUpdate := getInstalledSamplePostgresApp(t, client)
	assert.Equal(t, tools.SamplePostgresAppVersion18Name, appAfterUpdate.VersionName)
	assert.True(t, appAfterUpdate.IsRunning)
	assertSamplePostgresContainerUsesImage(t, "postgres:18.0-alpine")

	db = openSamplePostgresDb(t, client)
	defer db.Close()
	assert.Equal(t, samplePostgresValue, readSamplePostgresValue(t, db))
}

func TestUploadedPostgresMajorUpdateMigratesData(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	appBeforeUpdate, err := uploadSamplePostgresApp(t, client, tools.SamplePostgresAppVersion17Name, tools.SamplePostgresAppVersion17CreationTimestamp, tools.SamplePostgresAppVersion17ComposeYAML)
	assert.Nil(t, err)
	assert.Nil(t, client.Apps.Start(appBeforeUpdate.AppId))

	db := openSamplePostgresDb(t, client)
	writeSamplePostgresValue(t, db, samplePostgresValue)
	assert.Equal(t, samplePostgresValue, readSamplePostgresValue(t, db))
	assert.Nil(t, db.Close())

	_, err = uploadSamplePostgresApp(t, client, tools.SamplePostgresAppVersion18Name, tools.SamplePostgresAppVersion18CreationTimestamp, tools.SamplePostgresAppVersion18ComposeYAML)
	assert.Nil(t, err)

	appAfterUpdate := getInstalledSamplePostgresApp(t, client)
	assert.Equal(t, tools.SamplePostgresAppVersion18Name, appAfterUpdate.VersionName)
	assert.True(t, appAfterUpdate.IsRunning)
	assertSamplePostgresContainerUsesImage(t, "postgres:18.0-alpine")

	db = openSamplePostgresDb(t, client)
	defer db.Close()
	assert.Equal(t, samplePostgresValue, readSamplePostgresValue(t, db))
}

func TestUploadedPostgresMajorUpdateWithInvalidPreflightIsRejected(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	_, err := uploadSamplePostgresApp(t, client, tools.SamplePostgresAppVersion17Name, tools.SamplePostgresAppVersion17CreationTimestamp, tools.SamplePostgresAppVersion17ComposeYAML)
	assert.Nil(t, err)
	configureBackupRepo(t, client)

	tests := []struct {
		name          string
		composeYAML   string
		expectedError string
	}{
		{
			name:          "changed user",
			composeYAML:   strings.Replace(tools.SamplePostgresAppVersion18ComposeYAML, "POSTGRES_USER: postgresapp", "POSTGRES_USER: postgresapp_changed", 1),
			expectedError: app_migrations.PostgresUserChangedError,
		},
		{
			name: "changed data volume",
			composeYAML: strings.ReplaceAll(
				tools.SamplePostgresAppVersion18ComposeYAML,
				"samplemaintainer_postgresapp_data",
				"samplemaintainer_postgresapp_datachanged",
			),
			expectedError: app_migrations.PostgresDataVolumeChangedError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err = uploadSamplePostgresApp(t, client, tools.SamplePostgresAppVersion18Name, tools.SamplePostgresAppVersion18CreationTimestamp, test.composeYAML)
			u.AssertDeepStackErrorFromRequest(t, err, test.expectedError)

			appAfterRejectedUpdate := getInstalledSamplePostgresApp(t, client)
			assert.Equal(t, tools.SamplePostgresAppVersion17Name, appAfterRejectedUpdate.VersionName)
		})
	}

	backups, err := client.Backups.ListByApp(tools.SampleMaintainer, tools.SamplePostgresApp)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(backups))
}

func installSamplePostgresApp(t *testing.T, client *api_client.QuollixClient, version string) (*api.AdminAppDto, error) {
	if err := client.Apps.InstallFromStore(tools.SampleMaintainer, tools.SamplePostgresApp, version); err != nil {
		return nil, err
	}
	return getInstalledSamplePostgresApp(t, client), nil
}

func uploadSamplePostgresApp(t *testing.T, client *api_client.QuollixClient, version string, creationTimestamp time.Time, composeYAML string) (*api.AdminAppDto, error) {
	versionFile := api.BinaryFile{
		FileName: getSamplePostgresFileNameForAppUpload(version, creationTimestamp),
		Content:  []byte(composeYAML),
	}
	if err := client.Apps.UploadVersionFile(versionFile); err != nil {
		return nil, err
	}
	return getInstalledSamplePostgresApp(t, client), nil
}

func getSamplePostgresFileNameForAppUpload(version string, creationTimestamp time.Time) string {
	return fmt.Sprintf("%s_%s_%s_%s.yml", tools.SampleMaintainer, tools.SamplePostgresApp, version, creationTimestamp.Format(apps_basic.VersionFileUploadTimestampLayout))
}

func getInstalledSamplePostgresApp(t *testing.T, client *api_client.QuollixClient) *api.AdminAppDto {
	for _, app := range ListInstalledApps(t, client) {
		if app.AppName == tools.SamplePostgresApp {
			return &app
		}
	}
	assert.Nil(t, u.Logger.NewError("sample postgres app not found"))
	return nil
}

func openSamplePostgresDb(t *testing.T, client *api_client.QuollixClient) *sql.DB {
	password, err := client.Apps.GetInstalledAppSecret(tools.SamplePostgresApp, "SECRET_POSTGRES_PASSWORD")
	assert.Nil(t, err)

	db, err := sql.Open("postgres", samplePostgresDsn(password))
	assert.Nil(t, err)

	deadline := time.Now().Add(5 * time.Second)
	for {
		err = db.Ping()
		if err == nil {
			return db
		}
		if time.Now().After(deadline) {
			assert.Nil(t, err)
			return db
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func samplePostgresDsn(password string) string {
	dsn := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(samplePostgresUser, password),
		Host:   "127.0.0.1:5433",
		Path:   samplePostgresDb,
	}
	query := dsn.Query()
	query.Set("sslmode", "disable")
	dsn.RawQuery = query.Encode()
	return dsn.String()
}

func writeSamplePostgresValue(t *testing.T, db *sql.DB, value string) {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS sample_values (
			id integer PRIMARY KEY,
			value text NOT NULL
		)
	`)
	assert.Nil(t, err)

	_, err = db.Exec(`
		INSERT INTO sample_values (id, value)
		VALUES (1, $1)
		ON CONFLICT (id) DO UPDATE SET value = EXCLUDED.value
	`, value)
	assert.Nil(t, err)
}

func readSamplePostgresValue(t *testing.T, db *sql.DB) string {
	var value string
	err := db.QueryRow("SELECT value FROM sample_values WHERE id = 1").Scan(&value)
	assert.Nil(t, err)
	return value
}

func assertSamplePostgresContainerUsesImage(t *testing.T, expectedImage string) {
	cmd := exec.Command("docker", "inspect", "--format", "{{.Config.Image}}", tools.SamplePostgresAppContainerName) // #nosec G204 (CWE-78): component test inspects a known fixture container
	output, err := cmd.Output()
	assert.Nil(t, err)
	assert.Equal(t, expectedImage, strings.TrimSpace(string(output)))
}
