package assets

import (
	"fmt"
	"path/filepath"
	"server/tools"
	"sync"
)

type AssetStore interface {
	Put(path string, content []byte)
	Get(path string) ([]byte, bool)
	Has(path string) bool
	Clear()
	GetVersionedInjectedWebResourcePath(path, fileNameWithoutExtension, fileExtension string) string
}

type AssetStoreImpl struct {
	mutex      sync.RWMutex `wire:"-"`
	AssetBytes map[string][]byte
}

func (store *AssetStoreImpl) Put(path string, content []byte) {
	store.mutex.Lock()
	defer store.mutex.Unlock()
	store.AssetBytes[path] = content
}

func (store *AssetStoreImpl) Get(path string) ([]byte, bool) {
	store.mutex.RLock()
	defer store.mutex.RUnlock()
	content, ok := store.AssetBytes[path]
	return content, ok
}

func (store *AssetStoreImpl) Has(path string) bool {
	store.mutex.RLock()
	defer store.mutex.RUnlock()
	_, ok := store.AssetBytes[path]
	return ok
}

func (store *AssetStoreImpl) Clear() {
	store.mutex.Lock()
	defer store.mutex.Unlock()
	store.AssetBytes = map[string][]byte{}
}

func (store *AssetStoreImpl) GetVersionedInjectedWebResourcePath(path, fileNameWithoutExtension, fileExtension string) string {
	fileName := fmt.Sprintf("%s.%s.%s", fileNameWithoutExtension, tools.ApplicationVersion, fileExtension)
	return filepath.Join(path, fileName)
}
