package certificates

import (
	"testing"

	"github.com/quollix/common/assert"
)

func TestOperationMonitor_BeginRun_SetsRunningStateAndReturnsIncrementedRunId(t *testing.T) {
	operationMonitor := &OperationMonitorImpl{}

	firstRunId := operationMonitor.BeginRun("starting")
	assert.Equal(t, 1, firstRunId)

	firstState, firstOperation := operationMonitor.GetStatus()
	assert.Equal(t, OperationStateRunning, firstState)
	assert.Equal(t, "starting", firstOperation)

	secondRunId := operationMonitor.BeginRun("starting second")
	assert.Equal(t, 2, secondRunId)

	secondState, secondOperation := operationMonitor.GetStatus()
	assert.Equal(t, OperationStateRunning, secondState)
	assert.Equal(t, "starting second", secondOperation)
}

func TestOperationMonitor_SetOperation_UpdatesOnlyForCurrentRunAndOnlyWhileRunning(t *testing.T) {
	operationMonitor := &OperationMonitorImpl{}

	firstRunId := operationMonitor.BeginRun("run1-start")
	operationMonitor.SetOperation(firstRunId, "run1-step1")

	firstState, firstOperation := operationMonitor.GetStatus()
	assert.Equal(t, OperationStateRunning, firstState)
	assert.Equal(t, "run1-step1", firstOperation)

	secondRunId := operationMonitor.BeginRun("run2-start")
	secondStateAfterBegin, secondOperationAfterBegin := operationMonitor.GetStatus()
	assert.Equal(t, OperationStateRunning, secondStateAfterBegin)
	assert.Equal(t, "run2-start", secondOperationAfterBegin)

	operationMonitor.SetOperation(firstRunId, "run1-should-be-ignored")
	stateAfterIgnored, operationAfterIgnored := operationMonitor.GetStatus()
	assert.Equal(t, OperationStateRunning, stateAfterIgnored)
	assert.Equal(t, "run2-start", operationAfterIgnored)

	operationMonitor.SetOperation(secondRunId, "run2-step1")
	stateAfterUpdate, operationAfterUpdate := operationMonitor.GetStatus()
	assert.Equal(t, OperationStateRunning, stateAfterUpdate)
	assert.Equal(t, "run2-step1", operationAfterUpdate)

	operationMonitor.EndRun(secondRunId, true, "done")
	operationMonitor.SetOperation(secondRunId, "should-not-apply-after-end")
	stateAfterEndSet, operationAfterEndSet := operationMonitor.GetStatus()
	assert.Equal(t, OperationStateSuccess, stateAfterEndSet)
	assert.Equal(t, "done", operationAfterEndSet)
}

func TestOperationMonitor_EndRun_SetsSuccessOrErrorAndAppliesOnlyForCurrentRun(t *testing.T) {
	operationMonitor := &OperationMonitorImpl{}

	firstRunId := operationMonitor.BeginRun("run1-start")
	secondRunId := operationMonitor.BeginRun("run2-start")

	operationMonitor.EndRun(firstRunId, true, "run1-done-ignored")
	stateAfterIgnoredEnd, operationAfterIgnoredEnd := operationMonitor.GetStatus()
	assert.Equal(t, OperationStateRunning, stateAfterIgnoredEnd)
	assert.Equal(t, "run2-start", operationAfterIgnoredEnd)

	operationMonitor.EndRun(secondRunId, true, "run2-success")
	stateAfterSuccess, operationAfterSuccess := operationMonitor.GetStatus()
	assert.Equal(t, OperationStateSuccess, stateAfterSuccess)
	assert.Equal(t, "run2-success", operationAfterSuccess)

	thirdRunId := operationMonitor.BeginRun("run3-start")
	operationMonitor.EndRun(thirdRunId, false, "run3-error")
	stateAfterError, operationAfterError := operationMonitor.GetStatus()
	assert.Equal(t, OperationStateError, stateAfterError)
	assert.Equal(t, "run3-error", operationAfterError)
}

func TestOperationMonitor_Clear_ResetsOnlyForCurrentRun(t *testing.T) {
	operationMonitor := &OperationMonitorImpl{}

	firstRunId := operationMonitor.BeginRun("run1-start")
	secondRunId := operationMonitor.BeginRun("run2-start")

	operationMonitor.EndRun(secondRunId, true, "run2-success")
	stateAfterEnd, operationAfterEnd := operationMonitor.GetStatus()
	assert.Equal(t, OperationStateSuccess, stateAfterEnd)
	assert.Equal(t, "run2-success", operationAfterEnd)

	operationMonitor.Clear(firstRunId)
	stateAfterWrongClear, operationAfterWrongClear := operationMonitor.GetStatus()
	assert.Equal(t, OperationStateSuccess, stateAfterWrongClear)
	assert.Equal(t, "run2-success", operationAfterWrongClear)

	operationMonitor.Clear(secondRunId)
	stateAfterClear, operationAfterClear := operationMonitor.GetStatus()
	assert.Equal(t, OperationStateIdle, stateAfterClear)
	assert.Equal(t, "", operationAfterClear)
}

func TestOperationMonitor_ZeroValue_IsIdleWithEmptyOperation(t *testing.T) {
	operationMonitor := &OperationMonitorImpl{}

	state, operation := operationMonitor.GetStatus()
	assert.Equal(t, OperationState(""), state)
	assert.Equal(t, "", operation)
}
