package vmexec

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/go-go-golems/vm-system/pkg/vmmodels"
	"github.com/go-go-golems/vm-system/pkg/vmsession"
	"github.com/go-go-golems/vm-system/pkg/vmstore"
)

func TestExecuteREPLCreateExecutionFailureReturnsDeterministicError(t *testing.T) {
	const forcedMsg = "forced create failure"
	executor, sessionID := newPersistenceFailureFixture(t, func(s *failingExecutionStore) {
		s.createErr = errors.New(forcedMsg)
	})

	exec, err := executor.ExecuteREPL(sessionID, "1+1")
	if err == nil {
		t.Fatalf("expected create execution failure")
	}
	if exec != nil {
		t.Fatalf("expected nil execution when create fails")
	}
	if !strings.Contains(err.Error(), "failed to create execution") || !strings.Contains(err.Error(), forcedMsg) {
		t.Fatalf("expected wrapped create error, got: %v", err)
	}
}

func TestExecuteREPLAddEventFailureReturnsDeterministicError(t *testing.T) {
	const forcedMsg = "forced add_event failure"
	executor, sessionID := newPersistenceFailureFixture(t, func(s *failingExecutionStore) {
		s.addEventErr = errors.New(forcedMsg)
	})

	exec, err := executor.ExecuteREPL(sessionID, "1+1")
	if err == nil {
		t.Fatalf("expected add event failure")
	}
	if exec != nil {
		t.Fatalf("expected nil execution when add event fails early")
	}
	if !strings.Contains(err.Error(), "failed to persist event") || !strings.Contains(err.Error(), forcedMsg) {
		t.Fatalf("expected wrapped add event error, got: %v", err)
	}
}

func TestExecuteREPLUpdateExecutionFailureReturnsDeterministicError(t *testing.T) {
	const forcedMsg = "forced update failure"
	executor, sessionID := newPersistenceFailureFixture(t, func(s *failingExecutionStore) {
		s.updateErr = errors.New(forcedMsg)
	})

	exec, err := executor.ExecuteREPL(sessionID, "1+1")
	if err == nil {
		t.Fatalf("expected update execution failure")
	}
	if exec != nil {
		t.Fatalf("expected nil execution when finalize update fails")
	}
	if !strings.Contains(err.Error(), "failed to persist successful execution") || !strings.Contains(err.Error(), forcedMsg) {
		t.Fatalf("expected wrapped update error, got: %v", err)
	}
}

type failingExecutionStore struct {
	base        *vmstore.VMStore
	createErr   error
	addEventErr error
	updateErr   error
}

func (s *failingExecutionStore) CreateExecution(exec *vmmodels.Execution) error {
	if s.createErr != nil {
		return s.createErr
	}
	return s.base.CreateExecution(exec)
}

func (s *failingExecutionStore) UpdateExecution(exec *vmmodels.Execution) error {
	if s.updateErr != nil {
		return s.updateErr
	}
	return s.base.UpdateExecution(exec)
}

func (s *failingExecutionStore) AddEvent(event *vmmodels.ExecutionEvent) error {
	if s.addEventErr != nil {
		return s.addEventErr
	}
	return s.base.AddEvent(event)
}

func (s *failingExecutionStore) GetExecution(id string) (*vmmodels.Execution, error) {
	return s.base.GetExecution(id)
}

func (s *failingExecutionStore) GetEvents(executionID string, afterSeq int) ([]*vmmodels.ExecutionEvent, error) {
	return s.base.GetEvents(executionID, afterSeq)
}

func (s *failingExecutionStore) ListExecutions(sessionID string, limit int) ([]*vmmodels.Execution, error) {
	return s.base.ListExecutions(sessionID, limit)
}

func newPersistenceFailureFixture(t *testing.T, configure func(*failingExecutionStore)) (*Executor, string) {
	t.Helper()

	tmp := t.TempDir()
	dbPath := filepath.Join(tmp, "vm-system.db")
	worktree := filepath.Join(tmp, "worktree")
	if err := os.MkdirAll(worktree, 0o755); err != nil {
		t.Fatalf("mkdir worktree: %v", err)
	}

	baseStore, err := vmstore.NewVMStore(dbPath)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	t.Cleanup(func() { _ = baseStore.Close() })

	now := time.Now()
	vm := &vmmodels.VM{
		ID:        uuid.NewString(),
		Name:      "vmexec-persistence-failures",
		Engine:    "goja",
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := baseStore.CreateVM(vm); err != nil {
		t.Fatalf("create vm: %v", err)
	}
	if err := baseStore.SetVMSettings(&vmmodels.VMSettings{
		VMID: vm.ID,
		Limits: json.RawMessage(`{
      "cpu_ms": 2000,
      "wall_ms": 5000,
      "mem_mb": 128,
      "max_events": 50000,
      "max_output_kb": 256
    }`),
		Resolver: json.RawMessage(`{
      "roots": ["."],
      "extensions": [".js", ".mjs"],
      "allow_absolute_repo_imports": true
    }`),
		Runtime: json.RawMessage(`{
      "esm": true,
      "strict": true,
      "console": true
    }`),
	}); err != nil {
		t.Fatalf("set vm settings: %v", err)
	}

	sessionManager := vmsession.NewSessionManager(baseStore)
	session, err := sessionManager.CreateSession(vm.ID, "workspace-vmexec", "deadbeef", worktree)
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	failingStore := &failingExecutionStore{base: baseStore}
	if configure != nil {
		configure(failingStore)
	}

	return &Executor{
		store:          failingStore,
		sessionManager: sessionManager,
	}, session.ID
}
