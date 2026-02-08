package vmexec

import (
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

type executionRecordInput struct {
	sessionID string
	kind      vmmodels.ExecutionKind
	input     string
	path      string
	argsJSON  json.RawMessage
	envJSON   json.RawMessage
}

type eventRecorder struct {
	store       *vmstore.VMStore
	executionID string
	nextSeq     int
	err         error
}

// NewExecutor creates a new Executor
func NewExecutor(store *vmstore.VMStore, sessionManager *vmsession.SessionManager) *Executor {
	return &Executor{
		store:          store,
		sessionManager: sessionManager,
	}
}

func (e *Executor) prepareSession(sessionID string) (*vmsession.Session, func(), error) {
	session, err := e.sessionManager.GetSession(sessionID)
	if err != nil {
		return nil, nil, err
	}
	if session.Status != vmmodels.SessionReady {
		return nil, nil, vmmodels.ErrSessionNotReady
	}
	if !session.ExecutionLock.TryLock() {
		return nil, nil, vmmodels.ErrSessionBusy
	}
	return session, session.ExecutionLock.Unlock, nil
}

func (e *Executor) newExecutionRecord(in executionRecordInput) *vmmodels.Execution {
	argsJSON := in.argsJSON
	if len(argsJSON) == 0 {
		argsJSON = json.RawMessage("[]")
	}
	envJSON := in.envJSON
	if len(envJSON) == 0 {
		envJSON = json.RawMessage("{}")
	}

	return &vmmodels.Execution{
		ID:        uuid.New().String(),
		SessionID: in.sessionID,
		Kind:      string(in.kind),
		Input:     in.input,
		Path:      in.path,
		Args:      argsJSON,
		Env:       envJSON,
		Status:    string(vmmodels.ExecRunning),
		StartedAt: time.Now(),
		Metrics:   json.RawMessage("{}"),
	}
}

func newEventRecorder(store *vmstore.VMStore, executionID string) *eventRecorder {
	return &eventRecorder{
		store:       store,
		executionID: executionID,
		nextSeq:     1,
	}
}

func (r *eventRecorder) recordError(err error) {
	if err != nil && r.err == nil {
		r.err = err
	}
}

func (r *eventRecorder) Err() error {
	return r.err
}

func (r *eventRecorder) emit(eventType vmmodels.EventType, payload interface{}) error {
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal %s event payload: %w", eventType, err)
	}
	return r.emitRaw(eventType, payloadJSON)
}

func (r *eventRecorder) emitRaw(eventType vmmodels.EventType, payload json.RawMessage) error {
	event := &vmmodels.ExecutionEvent{
		ExecutionID: r.executionID,
		Seq:         r.nextSeq,
		Ts:          time.Now(),
		Type:        string(eventType),
		Payload:     payload,
	}
	if err := r.store.AddEvent(event); err != nil {
		return fmt.Errorf("failed to persist event %s seq=%d: %w", eventType, event.Seq, err)
	}
	r.nextSeq++
	return nil
}

// ExecuteREPL executes a REPL snippet
func (e *Executor) ExecuteREPL(sessionID, input string) (*vmmodels.Execution, error) {
	session, unlock, err := e.prepareSession(sessionID)
	if err != nil {
		return nil, err
	}
	defer unlock()

	exec := e.newExecutionRecord(executionRecordInput{
		sessionID: sessionID,
		kind:      vmmodels.ExecREPL,
		input:     input,
	})
	executionID := exec.ID

	if err := e.store.CreateExecution(exec); err != nil {
		return nil, fmt.Errorf("failed to create execution: %w", err)
	}

	recorder := newEventRecorder(e.store, executionID)

	// Override console.log to capture output
	console := map[string]interface{}{
		"log": func(args ...interface{}) {
			output := fmt.Sprint(args...)
			payload := vmmodels.ConsolePayload{
				Level: "log",
				Text:  output,
			}
			recorder.recordError(recorder.emit(vmmodels.EventConsole, payload))
		},
	}
	session.Runtime.Set("console", console)

	// Add input echo event
	if err := recorder.emit(vmmodels.EventInputEcho, map[string]string{"text": input}); err != nil {
		return nil, err
	}

	// Execute code
	value, err := session.Runtime.RunString(input)
	endTime := time.Now()
	if recorder.Err() != nil {
		return nil, recorder.Err()
	}

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
		if err := recorder.emitRaw(vmmodels.EventException, exceptionJSON); err != nil {
			return nil, err
		}

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
	if err := recorder.emitRaw(vmmodels.EventValue, valueJSON); err != nil {
		return nil, err
	}

	exec.Result = valueJSON
	e.store.UpdateExecution(exec)

	return exec, nil
}

// ExecuteRunFile executes a file
func (e *Executor) ExecuteRunFile(sessionID, path string, args, env map[string]interface{}) (*vmmodels.Execution, error) {
	session, unlock, err := e.prepareSession(sessionID)
	if err != nil {
		return nil, err
	}
	defer unlock()

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

	argsJSON, _ := json.Marshal(args)
	envJSON, _ := json.Marshal(env)
	exec := e.newExecutionRecord(executionRecordInput{
		sessionID: sessionID,
		kind:      vmmodels.ExecRunFile,
		path:      path,
		argsJSON:  argsJSON,
		envJSON:   envJSON,
	})
	executionID := exec.ID

	if err := e.store.CreateExecution(exec); err != nil {
		return nil, fmt.Errorf("failed to create execution: %w", err)
	}

	recorder := newEventRecorder(e.store, executionID)

	// Capture console output
	console := map[string]interface{}{
		"log": func(args ...interface{}) {
			output := fmt.Sprint(args...)
			payload := vmmodels.ConsolePayload{
				Level: "log",
				Text:  output,
			}
			recorder.recordError(recorder.emit(vmmodels.EventConsole, payload))
		},
	}
	session.Runtime.Set("console", console)

	// Set __ARGS__ global
	session.Runtime.Set("__ARGS__", args)

	// Execute file
	_, err = session.Runtime.RunString(string(content))
	endTime := time.Now()
	if recorder.Err() != nil {
		return nil, recorder.Err()
	}

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
		if err := recorder.emitRaw(vmmodels.EventException, exceptionJSON); err != nil {
			return nil, err
		}

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
