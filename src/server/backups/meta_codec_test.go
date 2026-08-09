package backups

import (
	"os"
	"path/filepath"
	"server/apps_basic"
	"server/tools"
	"testing"
	"time"

	"github.com/quollix/common/assert"
)

var (
	versionCreationTimestamp = time.Unix(1735689600, 0).UTC()
)

func TestMetaCodecImpl_SaveLoadDeleteFile(t *testing.T) {
	tempDir := t.TempDir()
	metaPath := filepath.Join(tempDir, "meta.yaml")

	metaToSave := NewMetaData(
		"clientIdValue",
		"clientSecretValue",
		"appSecretValue",
		"accessPolicyValue",
		"8080",
		versionCreationTimestamp,
		true,
		true,
		map[string]string{
			"SECRET_POSTGRES_PASSWORD": "postgresPassword",
			"SECRET_SESSION_SECRET":    "sessionSecret",
		},
	)

	clientCredentialsGeneratorMock := apps_basic.NewClientCredentialsGeneratorMock(t)
	metaCodec := &MetaCodecImpl{
		ClientCredentialsCreator: clientCredentialsGeneratorMock,
		AuthHelper:               tools.NewAuthHelperMock(t),
	}

	assert.Nil(t, metaCodec.Save(metaPath, metaToSave))

	loadedMeta, err := metaCodec.Load(metaPath)
	assert.Nil(t, err)

	assert.Equal(t, metaToSave.AccessPolicy, loadedMeta.AccessPolicy)
	assert.Equal(t, metaToSave.Port, loadedMeta.Port)
	assert.True(t, metaToSave.VersionCreationTimestamp.Equal(loadedMeta.VersionCreationTimestamp))
	assert.Equal(t, metaToSave.ClientId, loadedMeta.ClientId)
	assert.Equal(t, metaToSave.ClientSecret, loadedMeta.ClientSecret)
	assert.Equal(t, metaToSave.AppSecret, loadedMeta.AppSecret)
	assert.Equal(t, metaToSave.Secrets, loadedMeta.Secrets)

	assert.Nil(t, os.Remove(metaPath))

	_, err = os.Stat(metaPath)
	assert.True(t, os.IsNotExist(err))
}

func TestMetaCodecImpl_LoadGeneratesCredentialsIfMissing(t *testing.T) {
	tempDir := t.TempDir()
	metaPath := filepath.Join(tempDir, "meta.yaml")

	versionCreationTimestamp := time.Unix(1735689600, 0).UTC()

	metaWithoutCredentials := NewMetaData(
		"",
		"",
		"",
		"accessPolicyValue",
		"8080",
		versionCreationTimestamp,
		true,
		true,
		nil,
	)

	clientCredentialsGeneratorMock := apps_basic.NewClientCredentialsGeneratorMock(t)
	clientCredentialsGeneratorMock.EXPECT().Generate().Return("generatedClientId", "generatedClientSecret", nil)
	authHelperMock := tools.NewAuthHelperMock(t)
	authHelperMock.EXPECT().GenerateSecret().Return("generatedAppSecret", nil)

	metaCodec := &MetaCodecImpl{
		ClientCredentialsCreator: clientCredentialsGeneratorMock,
		AuthHelper:               authHelperMock,
	}

	assert.Nil(t, metaCodec.Save(metaPath, metaWithoutCredentials))

	loadedMeta, err := metaCodec.Load(metaPath)
	assert.Nil(t, err)

	assert.Equal(t, "generatedClientId", loadedMeta.ClientId)
	assert.Equal(t, "generatedClientSecret", loadedMeta.ClientSecret)
	assert.Equal(t, "generatedAppSecret", loadedMeta.AppSecret)

	clientCredentialsGeneratorMock.AssertExpectations(t)
}

// Deprecated: remove this legacy metadata coverage when APP_SECRET backup compatibility is removed.
func TestMetaCodecImpl_LoadLegacyMetadataWithoutSecrets(t *testing.T) {
	tempDir := t.TempDir()
	metaPath := filepath.Join(tempDir, "meta.yaml")
	legacyMetadata := []byte(`access_policy: accessPolicyValue
port: "8080"
version_creation_timestamp: 2025-01-01T00:00:00Z
client_id: clientIdValue
client_secret: clientSecretValue
app_secret: appSecretValue
automatic_updates_enabled: true
automatic_backups_enabled: true
`)
	assert.Nil(t, os.WriteFile(metaPath, legacyMetadata, 0o600))

	metaCodec := &MetaCodecImpl{
		ClientCredentialsCreator: apps_basic.NewClientCredentialsGeneratorMock(t),
		AuthHelper:               tools.NewAuthHelperMock(t),
	}

	loadedMeta, err := metaCodec.Load(metaPath)
	assert.Nil(t, err)

	assert.Equal(t, "appSecretValue", loadedMeta.AppSecret)
	assert.Equal(t, map[string]string(nil), loadedMeta.Secrets)
}
