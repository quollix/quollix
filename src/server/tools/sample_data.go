package tools

import (
	"fmt"
	"time"
)

var (
	SampleAppDockerNetwork = fmt.Sprintf("%s_%s", SampleMaintainer, SampleApp)
	SampleAppDockerVolume  = fmt.Sprintf("%s_%s_data", SampleMaintainer, SampleApp)
	SampleAppContainerName = fmt.Sprintf("%s_%s_%s", SampleMaintainer, SampleApp, SampleApp)

	SamplePostgresAppDockerNetwork = fmt.Sprintf("%s_%s", SampleMaintainer, SamplePostgresApp)
	SamplePostgresAppDockerVolume  = fmt.Sprintf("%s_%s_data", SampleMaintainer, SamplePostgresApp)
	SamplePostgresAppContainerName = fmt.Sprintf("%s_%s_%s", SampleMaintainer, SamplePostgresApp, SamplePostgresApp)

	SampleRabbitMQAppDockerNetwork = fmt.Sprintf("%s_%s", SampleMaintainer, SampleRabbitMQApp)
	SampleRabbitMQAppDockerVolume  = fmt.Sprintf("%s_%s_data", SampleMaintainer, SampleRabbitMQApp)
	SampleRabbitMQAppContainerName = fmt.Sprintf("%s_%s_%s", SampleMaintainer, SampleRabbitMQApp, SampleRabbitMQApp)

	SampleAppCreationTimestamp         = time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	SampleAppVersion0CreationTimestamp = SampleAppCreationTimestamp.Add(-24 * time.Hour)
	SampleAppVersion1CreationTimestamp = SampleAppCreationTimestamp.Add(-time.Hour)
	SampleAppVersion2CreationTimestamp = SampleAppCreationTimestamp.Add(+time.Hour)

	SamplePostgresAppCreationTimestamp          = time.Date(2021, 1, 2, 0, 0, 0, 0, time.UTC)
	SamplePostgresAppVersion17CreationTimestamp = SamplePostgresAppCreationTimestamp.Add(-time.Hour)
	SamplePostgresAppVersion18CreationTimestamp = SamplePostgresAppCreationTimestamp.Add(+time.Hour)

	SampleRabbitMQAppCreationTimestamp           = time.Date(2021, 1, 3, 0, 0, 0, 0, time.UTC)
	SampleRabbitMQAppVersion311CreationTimestamp = SampleRabbitMQAppCreationTimestamp.Add(-time.Hour)
	SampleRabbitMQAppVersion312CreationTimestamp = SampleRabbitMQAppCreationTimestamp.Add(+time.Hour)
)

const (
	SampleMaintainer  = "samplemaintainer"
	SampleApp         = "sampleapp"
	SamplePostgresApp = "postgresapp"
	SampleRabbitMQApp = "rabbitmqapp"

	SampleAppVersion0Name           = "0.0"
	SampleAppVersion1Name           = "1.0"
	SampleAppVersion2Name           = "2.0"
	SamplePostgresAppVersion17Name  = "17.0"
	SamplePostgresAppVersion18Name  = "18.0"
	SampleRabbitMQAppVersion311Name = "3.11"
	SampleRabbitMQAppVersion312Name = "3.12"

	TestSshServerHost             = "dummy_backup_server"
	TestSshServerPort             = "2222"
	TestSshServerBackupsDirectory = "/config/backups" // the actual backups are store also in the config folder, which is a little odd. maybe improve naming? where is the /config path coming from?
)

var (
	SampleAppVersion0ComposeYAML = buildSampleAppComposeYAML(sampleAppComposeOptions{
		Version:           "1.0",
		Port:              "3000",
		VersionSecretName: "SECRET_SAMPLE_VERSION_ZERO",
	})
	SampleAppVersion1ComposeYAML = buildSampleAppComposeYAML(sampleAppComposeOptions{
		Version:           "1.0",
		Port:              "3000",
		VersionSecretName: "SECRET_SAMPLE_VERSION_ONE",
		ExtraEnvironment:  sampleAppLegacyPasswordEnvironment("password"),
		VolumeMapping:     sampleAppVolumeMapping,
		VolumeDeclaration: sampleAppVolumeDeclaration,
	})
	SampleAppVersion2ComposeYAML = buildSampleAppComposeYAML(sampleAppComposeOptions{
		Version:           "2.0",
		Port:              "3001",
		VersionSecretName: "SECRET_SAMPLE_VERSION_TWO",
		ExtraEnvironment:  sampleAppLegacyPasswordEnvironment("${SECRET_SAMPLE_MIGRATED_PASSWORD}"),
		VolumeMapping:     sampleAppVolumeMapping,
		VolumeDeclaration: sampleAppVolumeDeclaration,
	})

	SamplePostgresAppVersion17ComposeYAML  = buildSamplePostgresAppComposeYAML("17.0-alpine", "/var/lib/postgresql/data")
	SamplePostgresAppVersion18ComposeYAML  = buildSamplePostgresAppComposeYAML("18.0-alpine", "/var/lib/postgresql")
	SampleRabbitMQAppVersion311ComposeYAML = buildSampleRabbitMQAppComposeYAML("3.11.18-management-alpine", sampleRabbitMQLegacyFeatureFlagsEnvironment)
	SampleRabbitMQAppVersion312ComposeYAML = buildSampleRabbitMQAppComposeYAML("3.12.14-management-alpine", "")
)

type sampleAppComposeOptions struct {
	Version           string
	Port              string
	VersionSecretName string
	ExtraEnvironment  string
	VolumeMapping     string
	VolumeDeclaration string
}

const (
	sampleAppVolumeMapping     = "    volumes:\n      - samplemaintainer_sampleapp_data:/data\n"
	sampleAppVolumeDeclaration = "\nvolumes:\n  samplemaintainer_sampleapp_data:\n"

	// This keeps the RabbitMQ migration component test from always passing when migration logic is disabled: a fresh unclustered node would enable all stable flags by default, so we simulate a legacy 3.11 node where flags required by 3.12 still need to be enabled by the migrator.
	sampleRabbitMQLegacyFeatureFlagsEnvironment = "      RABBITMQ_FEATURE_FLAGS: implicit_default_bindings,quorum_queue,virtual_host_metadata,maintenance_mode_status,user_limits\n"
)

func sampleAppLegacyPasswordEnvironment(value string) string {
	return fmt.Sprintf("      - SAMPLE_LEGACY_PASSWORD=%s\n", value)
}

func buildSampleAppComposeYAML(options sampleAppComposeOptions) string {
	return fmt.Sprintf(`services:
  sampleapp:
    image: sampleapp:local
    container_name: samplemaintainer_sampleapp_sampleapp
    environment:
      - VERSION=%s
      - SERVER_URL=https://sampleapp.${BASE_DOMAIN}
      - OIDC_CLIENT_ID=${CLIENT_ID}
      - OIDC_CLIENT_SECRET=${CLIENT_SECRET}
      - APP_SECRET=${APP_SECRET}
      - SECRET_SAMPLE_SHARED=${SECRET_SAMPLE_SHARED}
      - %s=${%s}
%s      - IANA_TIMEZONE=${IANA_TIMEZONE}
      - PORT=%s
%s    labels:
      quollix.port: %s
%s`, options.Version, options.VersionSecretName, options.VersionSecretName, options.ExtraEnvironment, options.Port, options.VolumeMapping, options.Port, options.VolumeDeclaration)
}

func buildSamplePostgresAppComposeYAML(postgresTag, postgresDataPath string) string {
	return fmt.Sprintf(`services:
  postgresapp:
    image: postgres:%s
    container_name: samplemaintainer_postgresapp_postgresapp
    environment:
      POSTGRES_DB: postgresapp
      POSTGRES_USER: postgresapp
      POSTGRES_PASSWORD: "${SECRET_POSTGRES_PASSWORD}"
    ports:
      - 127.0.0.1:5433:5432
    volumes:
      - samplemaintainer_postgresapp_data:%s
    labels:
      quollix.port: 5432

volumes:
  samplemaintainer_postgresapp_data:
`, postgresTag, postgresDataPath)
}

func buildSampleRabbitMQAppComposeYAML(rabbitMQTag string, extraEnvironment string) string {
	return fmt.Sprintf(`services:
  rabbitmqapp:
    image: rabbitmq:%s
    container_name: samplemaintainer_rabbitmqapp_rabbitmqapp
    environment:
      RABBITMQ_NODENAME: rabbit@rabbitmqapp
%s
    ports:
      - 127.0.0.1:15673:15672
    volumes:
      - samplemaintainer_rabbitmqapp_data:/var/lib/rabbitmq
    labels:
      quollix.port: 15672

volumes:
  samplemaintainer_rabbitmqapp_data:
`, rabbitMQTag, extraEnvironment)
}
