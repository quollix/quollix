//go:build integration

package backups

import (
	"os"
	"path/filepath"
	"server/apps_basic"
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
		"accessPolicyValue",
		"8080",
		versionCreationTimestamp,
		true,
		true,
	)

	clientCredentialsGeneratorMock := apps_basic.NewClientCredentialsGeneratorMock(t)
	metaCodec := &MetaCodecImpl{
		ClientCredentialsCreator: clientCredentialsGeneratorMock,
	}

	assert.Nil(t, metaCodec.Save(metaPath, metaToSave))

	loadedMeta, err := metaCodec.Load(metaPath)
	assert.Nil(t, err)

	assert.Equal(t, metaToSave.AccessPolicy, loadedMeta.AccessPolicy)
	assert.Equal(t, metaToSave.Port, loadedMeta.Port)
	assert.True(t, metaToSave.VersionCreationTimestamp.Equal(loadedMeta.VersionCreationTimestamp))
	assert.Equal(t, metaToSave.ClientId, loadedMeta.ClientId)
	assert.Equal(t, metaToSave.ClientSecret, loadedMeta.ClientSecret)

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
		"accessPolicyValue",
		"8080",
		versionCreationTimestamp,
		true,
		true,
	)

	clientCredentialsGeneratorMock := apps_basic.NewClientCredentialsGeneratorMock(t)
	clientCredentialsGeneratorMock.EXPECT().Generate().Return("generatedClientId", "generatedClientSecret", nil)

	metaCodec := &MetaCodecImpl{
		ClientCredentialsCreator: clientCredentialsGeneratorMock,
	}

	assert.Nil(t, metaCodec.Save(metaPath, metaWithoutCredentials))

	loadedMeta, err := metaCodec.Load(metaPath)
	assert.Nil(t, err)

	assert.Equal(t, "generatedClientId", loadedMeta.ClientId)
	assert.Equal(t, "generatedClientSecret", loadedMeta.ClientSecret)

	clientCredentialsGeneratorMock.AssertExpectations(t)
}
