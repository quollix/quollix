package src

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractGeneratedInitialAdminPassword_UsesDocumentedSearchText(t *testing.T) {
	logs := `{"level":"info","message":"INITIAL_ADMIN_PASSWORD environment variable is not set, generated random initial admin password","username":"administrator","password":"secret-password"}`

	credentials, err := extractGeneratedInitialAdminCredentials(logs)

	require.NoError(t, err)
	assert.Equal(t, "administrator", credentials.Username)
	assert.Equal(t, "secret-password", credentials.Password)
}

func TestExtractGeneratedInitialAdminPassword_IgnoresOtherPasswordFields(t *testing.T) {
	logs := `{"level":"info","message":"some other log line","password":"wrong-password"}`

	_, err := extractGeneratedInitialAdminCredentials(logs)

	require.Error(t, err)
}
