package apps_basic

import (
	"testing"

	"github.com/quollix/common/assert"
	u "github.com/quollix/common/utils"
)

var (
	xwikiOperationDescription       = "starting app 'xwiki'"
	xwikiUpdateOperationDescription = "updating app 'xwiki'"
	vaultwardenOperationDescription = "stopping app 'vaultwarden'"
	postgresOperationDescription    = "backing up 'postgres'"
	shutdownOperationDescription    = "shutdown"
	waitingOperationDescription     = "waiting for certificate"
)

func TestTryBlockAppOperationAllowsDifferentApps(t *testing.T) {
	registry := OperationRegistryImpl{}

	xwikiHandle, err := registry.TryBlockAppOperation("xwiki", xwikiOperationDescription)
	assert.Nil(t, err)

	vaultwardenHandle, err := registry.TryBlockAppOperation("vaultwarden", vaultwardenOperationDescription)
	assert.Nil(t, err)

	assert.Equal(t, []string{xwikiOperationDescription, vaultwardenOperationDescription}, registry.ListOperations())

	xwikiHandle.Done()
	vaultwardenHandle.Done()

	assert.Equal(t, []string{}, registry.ListOperations())
}

func TestTryBlockAppOperationBlocksSameApp(t *testing.T) {
	registry := OperationRegistryImpl{}

	handle, err := registry.TryBlockAppOperation("xwiki", xwikiOperationDescription)
	assert.Nil(t, err)

	_, err = registry.TryBlockAppOperation("xwiki", xwikiUpdateOperationDescription)
	assert.NotNil(t, err)
	assert.Equal(t, concurrentOperationErrorMessage, u.ExtractError(err))

	handle.Done()
}

func TestTryBlockAppOperationTreatsPostgresAsGlobalBlocking(t *testing.T) {
	registry := OperationRegistryImpl{}

	handle, err := registry.TryBlockAppOperation(u.OfficialDatabaseAppName, postgresOperationDescription)
	assert.Nil(t, err)

	_, err = registry.TryBlockAppOperation("xwiki", xwikiOperationDescription)
	assert.NotNil(t, err)
	assert.Equal(t, concurrentOperationErrorMessage, u.ExtractError(err))

	_, err = registry.TryBlockGlobalOperation(shutdownOperationDescription)
	assert.NotNil(t, err)
	assert.Equal(t, concurrentOperationErrorMessage, u.ExtractError(err))

	handle.Done()
}

func TestTryBlockAppOperationReportsFinishedPostgresOperation(t *testing.T) {
	registry := OperationRegistryImpl{}

	handle, err := registry.TryBlockAppOperation(u.OfficialDatabaseAppName, postgresOperationDescription)
	assert.Nil(t, err)

	handle.Done()

	assert.Equal(t, []string{postgresOperationDescription}, registry.ListFinishedAppOperations())
}

func TestTryBlockGlobalOperationBlocksWhileAppBlockingOperationRuns(t *testing.T) {
	registry := OperationRegistryImpl{}

	handle, err := registry.TryBlockAppOperation("xwiki", xwikiOperationDescription)
	assert.Nil(t, err)

	_, err = registry.TryBlockGlobalOperation(shutdownOperationDescription)
	assert.NotNil(t, err)
	assert.Equal(t, concurrentOperationErrorMessage, u.ExtractError(err))

	handle.Done()

	shutdownHandle, err := registry.TryBlockGlobalOperation(shutdownOperationDescription)
	assert.Nil(t, err)
	assert.Equal(t, []string{shutdownOperationDescription}, registry.ListOperations())

	shutdownHandle.Done()
}

func TestRegisterOperationDoesNotBlockBlockingOperations(t *testing.T) {
	registry := OperationRegistryImpl{}

	waitingHandle := registry.RegisterOperation(waitingOperationDescription)

	blockingHandle, err := registry.TryBlockAppOperation("xwiki", xwikiOperationDescription)
	assert.Nil(t, err)

	assert.Equal(t, []string{waitingOperationDescription, xwikiOperationDescription}, registry.ListOperations())

	blockingHandle.Done()
	waitingHandle.Done()

	assert.Equal(t, []string{}, registry.ListOperations())
}

func TestTryBlockGlobalOperationIgnoresNonBlockingOperations(t *testing.T) {
	registry := OperationRegistryImpl{}

	waitingHandle := registry.RegisterOperation(waitingOperationDescription)

	shutdownHandle, err := registry.TryBlockGlobalOperation(shutdownOperationDescription)
	assert.Nil(t, err)
	assert.Equal(t, []string{waitingOperationDescription, shutdownOperationDescription}, registry.ListOperations())

	shutdownHandle.Done()
	waitingHandle.Done()
}

func TestDoneIsIdempotent(t *testing.T) {
	registry := OperationRegistryImpl{}

	handle, err := registry.TryBlockAppOperation("xwiki", xwikiOperationDescription)
	assert.Nil(t, err)

	handle.Done()
	handle.Done()

	assert.Equal(t, []string{}, registry.ListOperations())
}

func TestTryBlockAppOperationEmptyAppNameReturnsError(t *testing.T) {
	registry := OperationRegistryImpl{}

	_, err := registry.TryBlockAppOperation("", xwikiOperationDescription)
	assert.Equal(t, "app name must not be empty", u.ExtractError(err))
}

func TestFinishedAppOperationsIncludeFinishedAppOperations(t *testing.T) {
	registry := OperationRegistryImpl{}

	waitingHandle := registry.RegisterOperation(waitingOperationDescription)

	xwikiHandle, err := registry.TryBlockAppOperation("xwiki", xwikiOperationDescription)
	assert.Nil(t, err)
	xwikiHandle.Done()

	vaultwardenHandle, err := registry.TryBlockAppOperation("vaultwarden", vaultwardenOperationDescription)
	assert.Nil(t, err)
	vaultwardenHandle.Done()

	waitingHandle.Done()

	assert.Equal(t, []string{xwikiOperationDescription, vaultwardenOperationDescription}, registry.ListFinishedAppOperations())
}

func TestClearFinishedAppOperationsRemovesFinishedOperations(t *testing.T) {
	registry := OperationRegistryImpl{}

	handle, err := registry.TryBlockAppOperation("xwiki", xwikiOperationDescription)
	assert.Nil(t, err)

	handle.Done()

	assert.Equal(t, []string{xwikiOperationDescription}, registry.ListFinishedAppOperations())

	registry.ClearFinishedAppOperations()

	assert.Equal(t, []string{}, registry.ListFinishedAppOperations())
}
