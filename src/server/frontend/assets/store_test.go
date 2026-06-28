package assets

import (
	"fmt"
	"server/tools"
	"testing"

	"github.com/quollix/common/assert"
)

func TestAssetStoreImpl_Workflow(t *testing.T) {
	assetStore := &AssetStoreImpl{
		AssetBytes: map[string][]byte{},
	}

	assetPath := "global/frame.abc123.js"
	expectedBytes := []byte("hello")

	assert.False(t, assetStore.Has(assetPath))
	actualBytes, ok := assetStore.Get(assetPath)
	assert.False(t, ok)
	assert.Nil(t, actualBytes)

	assetStore.Put(assetPath, expectedBytes)

	assert.True(t, assetStore.Has(assetPath))
	actualBytes, ok = assetStore.Get(assetPath)
	assert.True(t, ok)
	assert.Equal(t, expectedBytes, actualBytes)

	assetStore.Clear()

	assert.False(t, assetStore.Has(assetPath))
	actualBytes, ok = assetStore.Get(assetPath)
	assert.False(t, ok)
	assert.Nil(t, actualBytes)
}

func TestAssetStoreImpl_GetVersionedInjectedWebResourcePath(t *testing.T) {
	assetStore := &AssetStoreImpl{}

	actualPath := assetStore.GetVersionedInjectedWebResourcePath("/frontend/resources", "global/frame", "js")

	expectedPath := fmt.Sprintf("/frontend/resources/global/frame.%s.js", tools.ApplicationVersion)
	assert.Equal(t, expectedPath, actualPath)
}
