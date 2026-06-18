package apps_basic

import (
	"server/tools"
	"sort"
	"sync"

	u "github.com/quollix/common/utils"
)

const concurrentOperationErrorMessage = "another app operation is ongoing"

type OperationRegistry interface {
	TryBlockAppOperation(appName, description string) (OperationHandle, error)
	TryBlockGlobalOperation(description string) (OperationHandle, error)
	RegisterOperation(description string) OperationHandle
	ListOperations() []string
	ListFinishedAppOperations() []string
	ClearFinishedAppOperations()
}

type OperationHandle interface {
	Done()
}

type operationScope string

const (
	operationScopeNonBlocking    operationScope = "non_blocking"
	operationScopeAppBlocking    operationScope = "app_blocking"
	operationScopeGlobalBlocking operationScope = "global_blocking"
)

type operationEntry struct {
	Id                     int
	AppName                string
	Description            string
	Scope                  operationScope
	IsReportedAppOperation bool
}

type OperationRegistryImpl struct {
	mutex                 sync.Mutex
	nextId                int
	operations            map[int]operationEntry
	finishedAppOperations []string
}

type operationHandle struct {
	registry *OperationRegistryImpl
	id       int
	once     sync.Once
}

func (o *OperationRegistryImpl) TryBlockAppOperation(appName, description string) (OperationHandle, error) {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	if appName == "" {
		return nil, u.Logger.NewError("app name must not be empty")
	}

	// The official database app is a deliberate special case: any blocking postgres
	// operation must run exclusively because it can affect the whole system.
	if appName == u.OfficialDatabaseAppName {
		return o.tryRegisterBlockingOperationLocked(appName, description, operationScopeGlobalBlocking)
	}

	for _, operation := range o.operations {
		if operation.Scope == operationScopeGlobalBlocking {
			return nil, newConcurrentOperationError(description, operation.Description)
		}
		if operation.Scope == operationScopeAppBlocking && operation.AppName == appName {
			return nil, newConcurrentOperationError(description, operation.Description)
		}
	}

	return o.registerOperationLocked(appName, description, operationScopeAppBlocking), nil
}

func (o *OperationRegistryImpl) TryBlockGlobalOperation(description string) (OperationHandle, error) {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	return o.tryRegisterBlockingOperationLocked("", description, operationScopeGlobalBlocking)
}

func (o *OperationRegistryImpl) RegisterOperation(description string) OperationHandle {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	return o.registerOperationLocked("", description, operationScopeNonBlocking)
}

func (o *OperationRegistryImpl) ListOperations() []string {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	orderedIds := make([]int, 0, len(o.operations))
	for id := range o.operations {
		orderedIds = append(orderedIds, id)
	}
	sort.Ints(orderedIds)

	result := make([]string, 0, len(orderedIds))
	for _, id := range orderedIds {
		result = append(result, o.operations[id].Description)
	}
	return result
}

func (o *OperationRegistryImpl) ListFinishedAppOperations() []string {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	if len(o.finishedAppOperations) == 0 {
		return []string{}
	}

	return append([]string(nil), o.finishedAppOperations...)
}

func (o *OperationRegistryImpl) ClearFinishedAppOperations() {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	o.finishedAppOperations = nil
}

func (o *OperationRegistryImpl) tryRegisterBlockingOperationLocked(appName, description string, scope operationScope) (OperationHandle, error) {
	for _, operation := range o.operations {
		if operation.Scope != operationScopeNonBlocking {
			return nil, newConcurrentOperationError(description, operation.Description)
		}
	}

	return o.registerOperationLocked(appName, description, scope), nil
}

func (o *OperationRegistryImpl) registerOperationLocked(appName, description string, scope operationScope) OperationHandle {
	if o.operations == nil {
		o.operations = make(map[int]operationEntry)
	}

	o.nextId++
	entry := operationEntry{
		Id:                     o.nextId,
		AppName:                appName,
		Description:            description,
		Scope:                  scope,
		IsReportedAppOperation: appName != "",
	}
	o.operations[entry.Id] = entry

	return &operationHandle{
		registry: o,
		id:       entry.Id,
	}
}

func (o *OperationRegistryImpl) unregister(id int) {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	entry, ok := o.operations[id]
	delete(o.operations, id)
	if ok && entry.IsReportedAppOperation {
		o.finishedAppOperations = append(o.finishedAppOperations, entry.Description)
	}
}

func (o *operationHandle) Done() {
	if o == nil || o.registry == nil {
		return
	}

	o.once.Do(func() {
		o.registry.unregister(o.id)
	})
}

func newConcurrentOperationError(attemptedOperation, ongoingOperation string) error {
	return u.Logger.NewError(concurrentOperationErrorMessage, tools.AttemptedOperationField, attemptedOperation, tools.OngoingOperationField, ongoingOperation)
}
