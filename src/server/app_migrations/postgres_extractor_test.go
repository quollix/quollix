package app_migrations

import (
	"fmt"
	"testing"

	"github.com/quollix/common/assert"
	u "github.com/quollix/common/utils"
)

var composeDatabaseExtractor = &ComposeDatabaseExtractorImpl{}

func TestExtractPostgresService_NoPostgresReturnsNil(t *testing.T) {
	service, exists, err := composeDatabaseExtractor.ExtractPostgresService([]byte(`
services:
  app:
    image: nginx:latest
    container_name: sample_app
`))

	assert.Nil(t, err)
	assert.False(t, exists)
	assert.Nil(t, service)
}

func TestExtractPostgresService_OfficialPostgresImage(t *testing.T) {
	service, exists, err := composeDatabaseExtractor.ExtractPostgresService([]byte(`
services:
  postgres:
    image: postgres:17.5-alpine
    container_name: quollix_sample_postgres
    environment:
      POSTGRES_USER: sample
    volumes:
      - quollix_sample_postgres:/var/lib/postgresql/data
`))

	assert.Nil(t, err)
	assert.True(t, exists)
	assert.Equal(t, &PostgresService{
		ServiceName:   "postgres",
		ContainerName: "quollix_sample_postgres",
		ImageMajor:    17,
		User:          "sample",
		DataVolume:    "quollix_sample_postgres",
	}, service)
}

func TestExtractPostgresService_ZulipPostgresImage(t *testing.T) {
	service, exists, err := composeDatabaseExtractor.ExtractPostgresService([]byte(`
services:
  postgres:
    image: zulip/zulip-postgresql:14
    container_name: quollix_zulip_postgres
    environment:
      POSTGRES_USER: zulip
    volumes:
      - quollix_zulip_postgresql:/var/lib/postgresql/data:rw
`))

	assert.Nil(t, err)
	assert.True(t, exists)
	assert.Equal(t, &PostgresService{
		ServiceName:   "postgres",
		ContainerName: "quollix_zulip_postgres",
		ImageMajor:    14,
		User:          "zulip",
		DataVolume:    "quollix_zulip_postgresql",
	}, service)
}

func TestExtractPostgresService_PostgresImageVersionFormats(t *testing.T) {
	tests := []struct {
		image         string
		expectedMajor int
	}{
		{image: "postgres:17", expectedMajor: 17},
		{image: "postgres:17.0", expectedMajor: 17},
		{image: "postgres:17.0.0", expectedMajor: 17},
		{image: "postgres:17-alpine", expectedMajor: 17},
		{image: "postgres:17.0-alpine", expectedMajor: 17},
		{image: "postgres:18.0", expectedMajor: 18},
		{image: "postgres:18.0.0-alpine", expectedMajor: 18},
	}

	for _, test := range tests {
		t.Run(test.image, func(t *testing.T) {
			service, exists, err := composeDatabaseExtractor.ExtractPostgresService([]byte(fmt.Sprintf(`
services:
  postgres:
    image: %s
    container_name: quollix_sample_postgres
    environment:
      POSTGRES_USER: sample
    volumes:
      - quollix_sample_postgres:/var/lib/postgresql/data
`, test.image)))

			assert.Nil(t, err)
			assert.True(t, exists)
			assert.Equal(t, test.expectedMajor, service.ImageMajor)
		})
	}
}

func TestExtractPostgresService_ListEnvironment(t *testing.T) {
	service, exists, err := composeDatabaseExtractor.ExtractPostgresService([]byte(`
services:
  postgres:
    image: postgres:18.0-alpine
    container_name: quollix_sample_postgres
    environment:
      - POSTGRES_USER=sample
      - POSTGRES_DB=sample
    volumes:
      - quollix_sample_postgres:/var/lib/postgresql
`))

	assert.Nil(t, err)
	assert.True(t, exists)
	assert.Equal(t, "sample", service.User)
	assert.Equal(t, "quollix_sample_postgres", service.DataVolume)
}

func TestExtractPostgresService_DefaultsUserToPostgres(t *testing.T) {
	service, exists, err := composeDatabaseExtractor.ExtractPostgresService([]byte(`
services:
  postgres:
    image: postgres:18.0-alpine
    container_name: quollix_sample_postgres
    volumes:
      - quollix_sample_postgres:/var/lib/postgresql/data
`))

	assert.Nil(t, err)
	assert.True(t, exists)
	assert.Equal(t, "postgres", service.User)
}

func TestExtractPostgresService_MultiplePostgresServicesReturnsError(t *testing.T) {
	service, exists, err := composeDatabaseExtractor.ExtractPostgresService([]byte(`
services:
  postgres:
    image: postgres:17.5-alpine
    container_name: quollix_sample_postgres
    volumes:
      - quollix_sample_postgres:/var/lib/postgresql/data
  second-postgres:
    image: postgres:17.5-alpine
    container_name: quollix_sample_postgres_two
    volumes:
      - quollix_sample_postgres_two:/var/lib/postgresql/data
`))

	assert.Equal(t, MultiplePostgresServicesError, u.ExtractError(err))
	assert.False(t, exists)
	assert.Nil(t, service)
}

func TestExtractPostgresService_MissingDataVolumeReturnsError(t *testing.T) {
	service, exists, err := composeDatabaseExtractor.ExtractPostgresService([]byte(`
services:
  postgres:
    image: postgres:17.5-alpine
    container_name: quollix_sample_postgres
`))

	assert.Equal(t, MissingPostgresDataVolumeError, u.ExtractError(err))
	assert.False(t, exists)
	assert.Nil(t, service)
}

func TestExtractPostgresService_MultipleDataVolumesReturnsError(t *testing.T) {
	service, exists, err := composeDatabaseExtractor.ExtractPostgresService([]byte(`
services:
  postgres:
    image: postgres:17.5-alpine
    container_name: quollix_sample_postgres
    volumes:
      - quollix_sample_postgres:/var/lib/postgresql/data
      - quollix_sample_uploads:/uploads
`))

	assert.Equal(t, MultiplePostgresDataVolumesError, u.ExtractError(err))
	assert.False(t, exists)
	assert.Nil(t, service)
}

func TestExtractPostgresService_InvalidDataVolumeReturnsError(t *testing.T) {
	service, exists, err := composeDatabaseExtractor.ExtractPostgresService([]byte(`
services:
  postgres:
    image: postgres:17.5-alpine
    container_name: quollix_sample_postgres
    volumes:
      - quollix_sample_postgres
`))

	assert.Equal(t, InvalidPostgresDataVolumeError, u.ExtractError(err))
	assert.False(t, exists)
	assert.Nil(t, service)
}

func TestExtractRabbitMQService_NoRabbitMQReturnsNil(t *testing.T) {
	service, exists, err := composeDatabaseExtractor.ExtractRabbitMQService([]byte(`
services:
  app:
    image: nginx:latest
    container_name: sample_app
`))

	assert.Nil(t, err)
	assert.False(t, exists)
	assert.Nil(t, service)
}

func TestExtractRabbitMQService_OfficialRabbitMQImage(t *testing.T) {
	service, exists, err := composeDatabaseExtractor.ExtractRabbitMQService([]byte(`
services:
  rabbitmq:
    image: rabbitmq:3.12.14-management-alpine
    container_name: quollix_sample_rabbitmq
`))

	assert.Nil(t, err)
	assert.True(t, exists)
	assert.Equal(t, &RabbitMQService{
		ServiceName:   "rabbitmq",
		ContainerName: "quollix_sample_rabbitmq",
		ImageVersion:  "3.12.14-management-alpine",
	}, service)
}

func TestExtractRabbitMQService_MultipleRabbitMQServicesReturnsError(t *testing.T) {
	service, exists, err := composeDatabaseExtractor.ExtractRabbitMQService([]byte(`
services:
  rabbitmq:
    image: rabbitmq:3.12.14-management-alpine
    container_name: quollix_sample_rabbitmq
  rabbitmq-two:
    image: rabbitmq:3.12.14-management-alpine
    container_name: quollix_sample_rabbitmq_two
`))

	assert.Equal(t, MultipleRabbitMQServicesError, u.ExtractError(err))
	assert.False(t, exists)
	assert.Nil(t, service)
}

func TestExtractRabbitMQService_IgnoresNonOfficialRabbitMQImage(t *testing.T) {
	service, exists, err := composeDatabaseExtractor.ExtractRabbitMQService([]byte(`
services:
  rabbitmq:
    image: sample/rabbitmq:3.12.14-management-alpine
    container_name: quollix_sample_rabbitmq
`))

	assert.Nil(t, err)
	assert.False(t, exists)
	assert.Nil(t, service)
}
