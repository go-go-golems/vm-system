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
	store          executionStore
	sessionManager *vmsession.SessionManager
}

type executionStore interface {
	CreateExecution(exec *vmmodels.Execution) error
	UpdateExecution(exec *vmmodels.Execution) error
	AddEvent(event *vmmodels.ExecutionEvent) error
	GetExecution(id string) (*vmmodels.Execution, error)
	GetEvents(executionID string, afterSeq int) ([]*vmmodels.ExecutionEvent, error)
	ListExecutions(sessionID string, limit int) ([]*vmmodels.Execution, error)
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
	store       executionStore
	executionID string
	nextSeq     int
	err         error
}

type executionPipelineConfig struct {
	sessionID     string
	recordInput   executionRecordInput
	setupRuntime  func(*vmsession.Session, *eventRecorder) error
	run           func(*vmsession.Session, *eventRecorder) (goja.Value, error)
	handleSuccess func(*vmmodels.Execution, *eventRecorder, goja.Value, time.Time) error
	handleError   func(*vmmodels.Execution, *eventRecorder, error, time.Time) error
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

func newEventRecorder(store executionStore, executionID string) *eventRecorder {
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

func (e *Executor) finalizeExecutionSuccess(exec *vmmodels.Execution, endedAt time.Time, result json.RawMessage) error {
	exec.Status = string(vmmodels.ExecOK)
	exec.EndedAt = &endedAt
	exec.Result = result
	if err := e.store.UpdateExecution(exec); err != nil {
		return fmt.Errorf("failed to persist successful execution %s: %w", exec.ID, err)
	}
	return nil
}

func (e *Executor) finalizeExecutionError(exec *vmmodels.Execution, endedAt time.Time, exception json.RawMessage) error {
	exec.Status = string(vmmodels.ExecError)
	exec.EndedAt = &endedAt
	exec.Error = exception
	if err := e.store.UpdateExecution(exec); err != nil {
		return fmt.Errorf("failed to persist failed execution %s: %w", exec.ID, err)
	}
	return nil
}

func (e *Executor) installConsoleRecorder(session *vmsession.Session, recorder *eventRecorder) {
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
}

func exceptionPayloadJSON(runErr error) json.RawMessage {
	exceptionPayload := vmmodels.ExceptionPayload{
		Message: runErr.Error(),
	}
	if gojaErr, ok := runErr.(*goja.Exception); ok {
		exceptionPayload.Stack = gojaErr.String()
	}
	exceptionJSON, _ := json.Marshal(exceptionPayload)
	return exceptionJSON
}

func valuePayloadJSON(value goja.Value) json.RawMessage {
	valuePayload := vmmodels.ValuePayload{
		Type:    value.ExportType().String(),
		Preview: value.String(),
	}
	if exported := value.Export(); exported != nil {
		if jsonBytes, err := json.Marshal(exported); err == nil {
			valuePayload.JSON = jsonBytes
		}
	}
	valueJSON, _ := json.Marshal(valuePayload)
	return valueJSON
}

func (e *Executor) runExecutionPipeline(cfg executionPipelineConfig) (*vmmodels.Execution, error) {
	session, unlock, err := e.prepareSession(cfg.sessionID)
	if err != nil {
		return nil, err
	}
	defer unlock()

	recordInput := cfg.recordInput
	recordInput.sessionID = cfg.sessionID

	exec := e.newExecutionRecord(recordInput)
	if err := e.store.CreateExecution(exec); err != nil {
		return nil, fmt.Errorf("failed to create execution: %w", err)
	}

	recorder := newEventRecorder(e.store, exec.ID)
	if cfg.setupRuntime != nil {
		if err := cfg.setupRuntime(session, recorder); err != nil {
			return nil, err
		}
	}

	value, runErr := cfg.run(session, recorder)
	endedAt := time.Now()
	if recorder.Err() != nil {
		return nil, recorder.Err()
	}

	if runErr != nil {
		if cfg.handleError != nil {
			if err := cfg.handleError(exec, recorder, runErr, endedAt); err != nil {
				return nil, err
			}
			return exec, nil
		}
		return nil, runErr
	}

	if cfg.handleSuccess != nil {
		if err := cfg.handleSuccess(exec, recorder, value, endedAt); err != nil {
			return nil, err
		}
	}

	return exec, nil
}

// ExecuteREPL executes a REPL snippet
func (e *Executor) ExecuteREPL(sessionID, input string) (*vmmodels.Execution, error) {
	return e.runExecutionPipeline(executionPipelineConfig{
		sessionID: sessionID,
		recordInput: executionRecordInput{
			kind:  vmmodels.ExecREPL,
			input: input,
		},
		setupRuntime: func(session *vmsession.Session, recorder *eventRecorder) error {
			e.installConsoleRecorder(session, recorder)
			return recorder.emit(vmmodels.EventInputEcho, map[string]string{"text": input})
		},
		run: func(session *vmsession.Session, _ *eventRecorder) (goja.Value, error) {
			return session.Runtime.RunString(input)
		},
		handleError: func(exec *vmmodels.Execution, recorder *eventRecorder, runErr error, endedAt time.Time) error {
			exceptionJSON := exceptionPayloadJSON(runErr)
			if err := recorder.emitRaw(vmmodels.EventException, exceptionJSON); err != nil {
				return err
			}
			return e.finalizeExecutionError(exec, endedAt, exceptionJSON)
		},
		handleSuccess: func(exec *vmmodels.Execution, recorder *eventRecorder, value goja.Value, endedAt time.Time) error {
			valueJSON := valuePayloadJSON(value)
			if err := recorder.emitRaw(vmmodels.EventValue, valueJSON); err != nil {
				return err
			}
			return e.finalizeExecutionSuccess(exec, endedAt, valueJSON)
		},
	})
}

// ExecuteRunFile executes a file
func (e *Executor) ExecuteRunFile(sessionID, path string, args, env map[string]interface{}) (*vmmodels.Execution, error) {
	argsJSON, _ := json.Marshal(args)
	envJSON, _ := json.Marshal(env)
	var fileContent []byte
	return e.runExecutionPipeline(executionPipelineConfig{
		sessionID: sessionID,
		recordInput: executionRecordInput{
			kind:     vmmodels.ExecRunFile,
			path:     path,
			argsJSON: argsJSON,
			envJSON:  envJSON,
		},
		setupRuntime: func(session *vmsession.Session, recorder *eventRecorder) error {
			filePath := filepath.Join(session.WorktreePath, path)
			if _, err := os.Stat(filePath); err != nil {
				return fmt.Errorf("%w: %s", vmmodels.ErrFileNotFound, path)
			}

			content, err := os.ReadFile(filePath)
			if err != nil {
				return fmt.Errorf("failed to read file: %w", err)
			}
			fileContent = content

			e.installConsoleRecorder(session, recorder)
			session.Runtime.Set("__ARGS__", args)
			return nil
		},
		run: func(session *vmsession.Session, _ *eventRecorder) (goja.Value, error) {
			return session.Runtime.RunString(string(fileContent))
		},
		handleError: func(exec *vmmodels.Execution, recorder *eventRecorder, runErr error, endedAt time.Time) error {
			exceptionJSON := exceptionPayloadJSON(runErr)
			if err := recorder.emitRaw(vmmodels.EventException, exceptionJSON); err != nil {
				return err
			}
			return e.finalizeExecutionError(exec, endedAt, exceptionJSON)
		},
		handleSuccess: func(exec *vmmodels.Execution, recorder *eventRecorder, value goja.Value, endedAt time.Time) error {
			valueJSON := valuePayloadJSON(value)
			if err := recorder.emitRaw(vmmodels.EventValue, valueJSON); err != nil {
				return err
			}
			return e.finalizeExecutionSuccess(exec, endedAt, valueJSON)
		},
	})
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
