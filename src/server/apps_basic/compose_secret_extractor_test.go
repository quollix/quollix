package apps_basic

import (
	"testing"

	"github.com/quollix/common/assert"
	u "github.com/quollix/common/utils"
)

func TestComposeSecretExtractorImpl_Extract_MapEnvironment(t *testing.T) {
	composeYaml := `
services:
  postgres:
    environment:
      POSTGRES_PASSWORD: "${SECRET_POSTGRES_PASSWORD}"
`

	secrets, err := (&ComposeSecretExtractorImpl{}).Extract([]byte(composeYaml))

	assert.Nil(t, err)
	assert.Equal(t, []string{"SECRET_POSTGRES_PASSWORD"}, secrets)
}

func TestComposeSecretExtractorImpl_Extract_ListEnvironment(t *testing.T) {
	composeYaml := `
services:
  postgres:
    environment:
      - POSTGRES_PASSWORD=${SECRET_POSTGRES_PASSWORD}
`

	secrets, err := (&ComposeSecretExtractorImpl{}).Extract([]byte(composeYaml))

	assert.Nil(t, err)
	assert.Equal(t, []string{"SECRET_POSTGRES_PASSWORD"}, secrets)
}

func TestComposeSecretExtractorImpl_Extract_EmbeddedString(t *testing.T) {
	composeYaml := `
services:
  hedgedoc:
    environment:
      CMD_DB_URL: postgres://hedgedoc:${SECRET_HEDGEDOC_POSTGRES_PASSWORD}@postgres:5432/hedgedoc
`

	secrets, err := (&ComposeSecretExtractorImpl{}).Extract([]byte(composeYaml))

	assert.Nil(t, err)
	assert.Equal(t, []string{"SECRET_HEDGEDOC_POSTGRES_PASSWORD"}, secrets)
}

func TestComposeSecretExtractorImpl_Extract_DeduplicatesAndSortsSecrets(t *testing.T) {
	composeYaml := `
services:
  postgres:
    environment:
      POSTGRES_PASSWORD: "${SECRET_POSTGRES_PASSWORD}"
  app:
    environment:
      APP_DB_PASSWORD: "${SECRET_POSTGRES_PASSWORD}"
      SESSION_SECRET: "${SECRET_SESSION_SECRET}"
`

	secrets, err := (&ComposeSecretExtractorImpl{}).Extract([]byte(composeYaml))

	assert.Nil(t, err)
	assert.Equal(t, []string{"SECRET_POSTGRES_PASSWORD", "SECRET_SESSION_SECRET"}, secrets)
}

func TestComposeSecretExtractorImpl_Extract_IgnoresNonSecretPlaceholders(t *testing.T) {
	composeYaml := `
services:
  app:
    environment:
      APP_HOME_URL: "https://app.${BASE_DOMAIN}"
      CLIENT_ID: "${CLIENT_ID}"
      CLIENT_SECRET: "${CLIENT_SECRET}"
      APP_SECRET: "${APP_SECRET}"
`

	secrets, err := (&ComposeSecretExtractorImpl{}).Extract([]byte(composeYaml))

	assert.Nil(t, err)
	assert.Equal(t, []string{}, secrets)
}

func TestComposeSecretExtractorImpl_Extract_InvalidYamlReturnsError(t *testing.T) {
	composeYaml := `
services:
  app:
    environment:
      key: value: invalid
`

	secrets, err := (&ComposeSecretExtractorImpl{}).Extract([]byte(composeYaml))

	assert.NotNil(t, err)
	assert.Nil(t, secrets)
}

func TestComposeSecretExtractorImpl_Extract_InvalidSecretPlaceholderReturnsError(t *testing.T) {
	composeYaml := `
services:
  app:
    environment:
      POSTGRES_PASSWORD: "${SECRET_postgres_PASSWORD}"
`

	secrets, err := (&ComposeSecretExtractorImpl{}).Extract([]byte(composeYaml))

	assert.Equal(t, InvalidSecretPlaceholderError, u.ExtractError(err))
	assert.Nil(t, secrets)
}
