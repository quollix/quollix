package apps_basic

import (
	"testing"

	"github.com/quollix/common/assert"
	u "github.com/quollix/common/utils"
)

var extractor = &ComposeExtractorImpl{}

func TestComposeExtractorImpl_ExtractSecrets_MapEnvironment(t *testing.T) {
	composeYaml := `
services:
  postgres:
    environment:
      POSTGRES_PASSWORD: "${SECRET_POSTGRES_PASSWORD}"
`

	secrets, err := extractor.ExtractSecrets([]byte(composeYaml))

	assert.Nil(t, err)
	assert.Equal(t, []string{"SECRET_POSTGRES_PASSWORD"}, secrets)
}

func TestComposeExtractorImpl_ExtractSecrets_ListEnvironment(t *testing.T) {
	composeYaml := `
services:
  postgres:
    environment:
      - POSTGRES_PASSWORD=${SECRET_POSTGRES_PASSWORD}
`

	secrets, err := extractor.ExtractSecrets([]byte(composeYaml))

	assert.Nil(t, err)
	assert.Equal(t, []string{"SECRET_POSTGRES_PASSWORD"}, secrets)
}

func TestComposeExtractorImpl_ExtractSecrets_EmbeddedString(t *testing.T) {
	composeYaml := `
services:
  hedgedoc:
    environment:
      CMD_DB_URL: postgres://hedgedoc:${SECRET_HEDGEDOC_POSTGRES_PASSWORD}@postgres:5432/hedgedoc
`

	secrets, err := extractor.ExtractSecrets([]byte(composeYaml))

	assert.Nil(t, err)
	assert.Equal(t, []string{"SECRET_HEDGEDOC_POSTGRES_PASSWORD"}, secrets)
}

func TestComposeExtractorImpl_ExtractSecrets_DeduplicatesAndSortsSecrets(t *testing.T) {
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

	secrets, err := extractor.ExtractSecrets([]byte(composeYaml))

	assert.Nil(t, err)
	assert.Equal(t, []string{"SECRET_POSTGRES_PASSWORD", "SECRET_SESSION_SECRET"}, secrets)
}

func TestComposeExtractorImpl_ExtractSecretSet_ReturnsTwoSecrets(t *testing.T) {
	composeYaml := `
services:
  app:
    environment:
      FIRST_SECRET: "${SECRET_FIRST}"
      SECOND_SECRET: "${SECRET_SECOND}"
`

	secretSet, err := extractor.ExtractSecretSet([]byte(composeYaml))

	assert.Nil(t, err)
	assert.Equal(t, map[string]bool{"SECRET_FIRST": true, "SECRET_SECOND": true}, secretSet)
}

func TestComposeExtractorImpl_ExtractSecrets_IgnoresNonSecretPlaceholders(t *testing.T) {
	composeYaml := `
services:
  app:
    environment:
      APP_HOME_URL: "https://app.${BASE_DOMAIN}"
      CLIENT_ID: "${CLIENT_ID}"
      CLIENT_SECRET: "${CLIENT_SECRET}"
      APP_SECRET: "${APP_SECRET}"
`

	secrets, err := extractor.ExtractSecrets([]byte(composeYaml))

	assert.Nil(t, err)
	assert.Equal(t, []string{}, secrets)
}

func TestComposeExtractorImpl_ExtractSecrets_InvalidYamlReturnsError(t *testing.T) {
	composeYaml := `
services:
  app:
    environment:
      key: value: invalid
`

	secrets, err := extractor.ExtractSecrets([]byte(composeYaml))

	assert.NotNil(t, err)
	assert.Nil(t, secrets)
}

func TestComposeExtractorImpl_ExtractSecrets_InvalidSecretPlaceholderReturnsError(t *testing.T) {
	composeYaml := `
services:
  app:
    environment:
      POSTGRES_PASSWORD: "${SECRET_postgres_PASSWORD}"
`

	secrets, err := extractor.ExtractSecrets([]byte(composeYaml))

	assert.Equal(t, InvalidSecretPlaceholderError, u.ExtractError(err))
	assert.Nil(t, secrets)
}
