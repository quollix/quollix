package src

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractGeneratedInitialAdminPassword_UsesDocumentedSearchText(t *testing.T) {
	logs := `{"level":"info","message":"INITIAL_ADMIN_PASSWORD environment variable is not set, generated random initial admin password","username":"administrator","password":"secret-password"}`

	password, err := extractGeneratedInitialAdminPassword(logs)

	require.NoError(t, err)
	assert.Equal(t, "secret-password", password)
}

func TestExtractGeneratedInitialAdminPassword_IgnoresOtherPasswordFields(t *testing.T) {
	logs := `{"level":"info","message":"some other log line","password":"wrong-password"}`

	_, err := extractGeneratedInitialAdminPassword(logs)

	require.Error(t, err)
}
