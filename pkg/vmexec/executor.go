package vmexec

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/dop251/goja"
	"github.com/google/uuid"

	"github.com/go-go-golems/vm-system/pkg/vmmodels"
	"github.com/go-go-golems/vm-system/pkg/vmsession"
	"github.com/go-go-golems/vm-system/pkg/vmstore"
)

// Executor executes code in VM sessions
type Executor struct {
	store          *vmstore.VMStore
	sessionManager *vmsession.SessionManager
}

// NewExecutor creates a new Executor
func NewExecutor(store *vmstore.VMStore, sessionManager *vmsession.SessionManager) *Executor {
	return &Executor{
		store:          store,
		sessionManager: sessionManager,
	}
}

// ExecuteREPL executes a REPL snippet
func (e *Executor) ExecuteREPL(sessionID, input string) (*vmmodels.Execution, error) {
	// Get session
	session, err := e.sessionManager.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	// Check session status
	if session.Status != vmmodels.SessionReady {
		return nil, vmmodels.ErrSessionNotReady
	}

	// Acquire execution lock
	if !session.ExecutionLock.TryLock() {
		return nil, vmmodels.ErrSessionBusy
	}
	defer session.ExecutionLock.Unlock()

	// Create execution record
	executionID := uuid.New().String()
	exec := &vmmodels.Execution{
		ID:        executionID,
		SessionID: sessionID,
		Kind:      string(vmmodels.ExecREPL),
		Input:     input,
		Args:      json.RawMessage("[]"),
		Env:       json.RawMessage("{}"),
		Status:    string(vmmodels.ExecRunning),
		StartedAt: time.Now(),
		Metrics:   json.RawMessage("{}"),
	}

	if err := e.store.CreateExecution(exec); err != nil {
		return nil, fmt.Errorf("failed to create execution: %w", err)
	}

	// Capture console output
	var consoleOutput bytes.Buffer
	eventSeq := 1

	// Override console.log to capture output
	console := map[string]interface{}{
		"log": func(args ...interface{}) {
			output := fmt.Sprint(args...)
			consoleOutput.WriteString(output + "\n")
			
			// Create console event
			payload := vmmodels.ConsolePayload{
				Level: "log",
				Text:  output,
			}
			payloadJSON, _ := json.Marshal(payload)
			
			event := &vmmodels.ExecutionEvent{
				ExecutionID: executionID,
				Seq:         eventSeq,
				Ts:          time.Now(),
				Type:        string(vmmodels.EventConsole),
				Payload:     payloadJSON,
			}
			eventSeq++
			
			e.store.AddEvent(event)
		},
	}
	session.Runtime.Set("console", console)

	// Add input echo event
	inputPayload, _ := json.Marshal(map[string]string{"text": input})
	inputEvent := &vmmodels.ExecutionEvent{
		ExecutionID: executionID,
		Seq:         eventSeq,
		Ts:          time.Now(),
		Type:        string(vmmodels.EventInputEcho),
		Payload:     inputPayload,
	}
	eventSeq++
	e.store.AddEvent(inputEvent)

	// Execute code
	value, err := session.Runtime.RunString(input)
	endTime := time.Now()

	if err != nil {
		// Execution failed
		exec.Status = string(vmmodels.ExecError)
		exec.EndedAt = &endTime

		// Create exception event
		exceptionPayload := vmmodels.ExceptionPayload{
			Message: err.Error(),
		}
		if gojaErr, ok := err.(*goja.Exception); ok {
			exceptionPayload.Stack = gojaErr.String()
		}
		exceptionJSON, _ := json.Marshal(exceptionPayload)
		
		exceptionEvent := &vmmodels.ExecutionEvent{
			ExecutionID: executionID,
			Seq:         eventSeq,
			Ts:          time.Now(),
			Type:        string(vmmodels.EventException),
			Payload:     exceptionJSON,
		}
		e.store.AddEvent(exceptionEvent)

		exec.Error = exceptionJSON
		e.store.UpdateExecution(exec)

		return exec, nil
	}

	// Execution succeeded
	exec.Status = string(vmmodels.ExecOK)
	exec.EndedAt = &endTime

	// Create value event
	valuePayload := vmmodels.ValuePayload{
		Type:    value.ExportType().String(),
		Preview: value.String(),
	}
	
	// Try to export as JSON
	if exported := value.Export(); exported != nil {
		if jsonBytes, err := json.Marshal(exported); err == nil {
			valuePayload.JSON = jsonBytes
		}
	}
	
	valueJSON, _ := json.Marshal(valuePayload)
	valueEvent := &vmmodels.ExecutionEvent{
		ExecutionID: executionID,
		Seq:         eventSeq,
		Ts:          time.Now(),
		Type:        string(vmmodels.EventValue),
		Payload:     valueJSON,
	}
	e.store.AddEvent(valueEvent)

	exec.Result = valueJSON
	e.store.UpdateExecution(exec)

	return exec, nil
}

// ExecuteRunFile executes a file
func (e *Executor) ExecuteRunFile(sessionID, path string, args, env map[string]interface{}) (*vmmodels.Execution, error) {
	// Get session
	session, err := e.sessionManager.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	// Check session status
	if session.Status != vmmodels.SessionReady {
		return nil, vmmodels.ErrSessionNotReady
	}

	// Acquire execution lock
	if !session.ExecutionLock.TryLock() {
		return nil, vmmodels.ErrSessionBusy
	}
	defer session.ExecutionLock.Unlock()

	// Resolve file path
	filePath := filepath.Join(session.WorktreePath, path)
	if _, err := os.Stat(filePath); err != nil {
		return nil, fmt.Errorf("%w: %s", vmmodels.ErrFileNotFound, path)
	}

	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Create execution record
	executionID := uuid.New().String()
	argsJSON, _ := json.Marshal(args)
	envJSON, _ := json.Marshal(env)
	
	exec := &vmmodels.Execution{
		ID:        executionID,
		SessionID: sessionID,
		Kind:      string(vmmodels.ExecRunFile),
		Path:      path,
		Args:      argsJSON,
		Env:       envJSON,
		Status:    string(vmmodels.ExecRunning),
		StartedAt: time.Now(),
		Metrics:   json.RawMessage("{}"),
	}

	if err := e.store.CreateExecution(exec); err != nil {
		return nil, fmt.Errorf("failed to create execution: %w", err)
	}

	eventSeq := 1

	// Capture console output
	console := map[string]interface{}{
		"log": func(args ...interface{}) {
			output := fmt.Sprint(args...)
			
			payload := vmmodels.ConsolePayload{
				Level: "log",
				Text:  output,
			}
			payloadJSON, _ := json.Marshal(payload)
			
			event := &vmmodels.ExecutionEvent{
				ExecutionID: executionID,
				Seq:         eventSeq,
				Ts:          time.Now(),
				Type:        string(vmmodels.EventConsole),
				Payload:     payloadJSON,
			}
			eventSeq++
			
			e.store.AddEvent(event)
		},
	}
	session.Runtime.Set("console", console)

	// Set __ARGS__ global
	session.Runtime.Set("__ARGS__", args)

	// Execute file
	_, err = session.Runtime.RunString(string(content))
	endTime := time.Now()

	if err != nil {
		// Execution failed
		exec.Status = string(vmmodels.ExecError)
		exec.EndedAt = &endTime

		exceptionPayload := vmmodels.ExceptionPayload{
			Message: err.Error(),
		}
		if gojaErr, ok := err.(*goja.Exception); ok {
			exceptionPayload.Stack = gojaErr.String()
		}
		exceptionJSON, _ := json.Marshal(exceptionPayload)
		
		exceptionEvent := &vmmodels.ExecutionEvent{
			ExecutionID: executionID,
			Seq:         eventSeq,
			Ts:          time.Now(),
			Type:        string(vmmodels.EventException),
			Payload:     exceptionJSON,
		}
		e.store.AddEvent(exceptionEvent)

		exec.Error = exceptionJSON
		e.store.UpdateExecution(exec)

		return exec, nil
	}

	// Execution succeeded
	exec.Status = string(vmmodels.ExecOK)
	exec.EndedAt = &endTime
	e.store.UpdateExecution(exec)

	return exec, nil
}

// GetExecution retrieves an execution
func (e *Executor) GetExecution(executionID string) (*vmmodels.Execution, error) {
	return e.store.GetExecution(executionID)
}

// GetEvents retrieves events for an execution
func (e *Executor) GetEvents(executionID string, afterSeq int) ([]*vmmodels.ExecutionEvent, error) {
	return e.store.GetEvents(executionID, afterSeq)
}

// ListExecutions lists executions for a session
func (e *Executor) ListExecutions(sessionID string, limit int) ([]*vmmodels.Execution, error) {
	return e.store.ListExecutions(sessionID, limit)
}
