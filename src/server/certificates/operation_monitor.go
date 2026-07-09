package certificates

import (
	"sync"

	u "github.com/quollix/common/utils"
)

type OperationState string

const (
	OperationStateIdle    OperationState = "idle"
	OperationStateRunning OperationState = "running"
	OperationStateSuccess OperationState = "success"
	OperationStateError   OperationState = "error"
)

type OperationMonitor interface {
	BeginRun(initialOperation string) int
	SetOperation(runId int, operationName string)
	EndRun(runId int, wasSuccessful bool, finalMessage string)
	Clear(runId int)
	GetStatus() (state OperationState, operation string)
}

type OperationMonitorImpl struct {
	mutex            sync.Mutex
	currentRunId     int
	currentState     OperationState
	currentOperation string
}

func (o *OperationMonitorImpl) BeginRun(initialOperation string) int {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	o.currentRunId++
	o.currentState = OperationStateRunning
	o.currentOperation = initialOperation
	u.Logger.Info("Starting certificate operation", "operation_name", initialOperation, "run_id", o.currentRunId)
	return o.currentRunId
}

func (o *OperationMonitorImpl) SetOperation(runId int, operationName string) {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	if runId != o.currentRunId || o.currentState != OperationStateRunning {
		return
	}
	o.currentOperation = operationName
	u.Logger.Info("Setting certificate operation", "operation_name", operationName, "run_id", runId)
}

func (o *OperationMonitorImpl) EndRun(runId int, wasSuccessful bool, finalMessage string) {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	if runId != o.currentRunId {
		return
	}

	if wasSuccessful {
		u.Logger.Info("Certificate operation finished successfully")
		o.currentState = OperationStateSuccess
	} else {
		u.Logger.Info("Certificate operation finished with an error", "error_message", finalMessage)
		o.currentState = OperationStateError
	}

	o.currentOperation = finalMessage
	u.Logger.Info("Ending certificate operation", "final_message", finalMessage, "run_id", runId, "success", wasSuccessful)
}

func (o *OperationMonitorImpl) Clear(runId int) {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	if runId != o.currentRunId {
		return
	}
	o.currentState = OperationStateIdle
	o.currentOperation = ""
	u.Logger.Info("Clearing certificate operation", "run_id", runId)
}

func (o *OperationMonitorImpl) GetStatus() (OperationState, string) {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	return o.currentState, o.currentOperation
}
