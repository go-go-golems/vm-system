package vmexec_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-go-golems/vm-system/pkg/vmcontrol"
	"github.com/go-go-golems/vm-system/pkg/vmexec"
	"github.com/go-go-golems/vm-system/pkg/vmmodels"
	"github.com/go-go-golems/vm-system/pkg/vmsession"
	"github.com/go-go-golems/vm-system/pkg/vmstore"
)

func TestExecuteREPLSuccessPersistsEventOrderAndResult(t *testing.T) {
	fx := newExecutorFixture(t)

	exec, err := fx.executor.ExecuteREPL(fx.sessionID, "console.log('repl'); 20 + 22")
	if err != nil {
		t.Fatalf("execute repl: %v", err)
	}
	if exec.Status != string(vmmodels.ExecOK) {
		t.Fatalf("expected status ok, got %q", exec.Status)
	}
	if exec.EndedAt == nil {
		t.Fatalf("expected ended_at to be set")
	}
	if len(exec.Error) != 0 {
		t.Fatalf("expected empty error payload, got %s", string(exec.Error))
	}

	var valuePayload vmmodels.ValuePayload
	if err := json.Unmarshal(exec.Result, &valuePayload); err != nil {
		t.Fatalf("unmarshal repl result: %v", err)
	}
	if valuePayload.Preview != "42" {
		t.Fatalf("expected repl preview 42, got %q", valuePayload.Preview)
	}

	persisted, err := fx.executor.GetExecution(exec.ID)
	if err != nil {
		t.Fatalf("get persisted execution: %v", err)
	}
	if persisted.Status != string(vmmodels.ExecOK) {
		t.Fatalf("expected persisted status ok, got %q", persisted.Status)
	}

	events, err := fx.executor.GetEvents(exec.ID, 0)
	if err != nil {
		t.Fatalf("get events: %v", err)
	}
	if len(events) != 3 {
		t.Fatalf("expected 3 repl events (input_echo, console, value), got %d", len(events))
	}
	for i, event := range events {
		if event.Seq != i+1 {
			t.Fatalf("expected seq=%d at index=%d, got %d", i+1, i, event.Seq)
		}
	}
	if events[0].Type != string(vmmodels.EventInputEcho) {
		t.Fatalf("expected first event type input_echo, got %q", events[0].Type)
	}
	if events[1].Type != string(vmmodels.EventConsole) {
		t.Fatalf("expected second event type console, got %q", events[1].Type)
	}
	if events[2].Type != string(vmmodels.EventValue) {
		t.Fatalf("expected third event type value, got %q", events[2].Type)
	}
}

func TestExecuteREPLErrorPersistsExceptionAndExecutionError(t *testing.T) {
	fx := newExecutorFixture(t)

	exec, err := fx.executor.ExecuteREPL(fx.sessionID, "throw new Error('boom')")
	if err != nil {
		t.Fatalf("execute repl error case: %v", err)
	}
	if exec.Status != string(vmmodels.ExecError) {
		t.Fatalf("expected error status, got %q", exec.Status)
	}
	if len(exec.Error) == 0 {
		t.Fatalf("expected execution error payload")
	}
	if len(exec.Result) != 0 {
		t.Fatalf("expected no result payload on error, got %s", string(exec.Result))
	}

	var exceptionPayload vmmodels.ExceptionPayload
	if err := json.Unmarshal(exec.Error, &exceptionPayload); err != nil {
		t.Fatalf("unmarshal repl error payload: %v", err)
	}
	if exceptionPayload.Message == "" {
		t.Fatalf("expected error message in exception payload")
	}

	events, err := fx.executor.GetEvents(exec.ID, 0)
	if err != nil {
		t.Fatalf("get events: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("expected 2 repl error events (input_echo, exception), got %d", len(events))
	}
	if events[0].Type != string(vmmodels.EventInputEcho) || events[1].Type != string(vmmodels.EventException) {
		t.Fatalf("unexpected repl error event types: [%s, %s]", events[0].Type, events[1].Type)
	}
}

func TestExecuteRunFilePersistsResultAndValueEvent(t *testing.T) {
	fx := newExecutorFixture(t)

	if err := os.WriteFile(filepath.Join(fx.worktree, "script.js"), []byte("console.log('file'); 40 + 2;"), 0o644); err != nil {
		t.Fatalf("write script file: %v", err)
	}

	exec, err := fx.executor.ExecuteRunFile(fx.sessionID, "script.js", map[string]interface{}{"a": 1}, map[string]interface{}{"env": "test"})
	if err != nil {
		t.Fatalf("execute run-file: %v", err)
	}
	if exec.Kind != string(vmmodels.ExecRunFile) {
		t.Fatalf("expected run_file kind, got %q", exec.Kind)
	}
	if exec.Status != string(vmmodels.ExecOK) {
		t.Fatalf("expected status ok, got %q", exec.Status)
	}
	if exec.Path != "script.js" {
		t.Fatalf("expected stored path script.js, got %q", exec.Path)
	}
	if len(exec.Result) == 0 {
		t.Fatalf("expected result payload for run-file contract")
	}
	var runFileResult vmmodels.ValuePayload
	if err := json.Unmarshal(exec.Result, &runFileResult); err != nil {
		t.Fatalf("unmarshal run-file result: %v", err)
	}
	if runFileResult.Preview != "42" {
		t.Fatalf("expected run-file preview 42, got %q", runFileResult.Preview)
	}

	events, err := fx.executor.GetEvents(exec.ID, 0)
	if err != nil {
		t.Fatalf("get events: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("expected 2 run-file events (console, value), got %d", len(events))
	}
	if events[0].Type != string(vmmodels.EventConsole) {
		t.Fatalf("expected console event, got %q", events[0].Type)
	}
	if events[1].Type != string(vmmodels.EventValue) {
		t.Fatalf("expected value event, got %q", events[1].Type)
	}

	persisted, err := fx.executor.GetExecution(exec.ID)
	if err != nil {
		t.Fatalf("get persisted run-file execution: %v", err)
	}
	if persisted.Status != string(vmmodels.ExecOK) {
		t.Fatalf("expected persisted status ok, got %q", persisted.Status)
	}
	if len(persisted.Result) == 0 {
		t.Fatalf("expected persisted result payload for run-file")
	}
	if !jsonContainsKV(t, persisted.Args, "a") {
		t.Fatalf("expected args payload to contain key 'a': %s", string(persisted.Args))
	}
	if !jsonContainsKV(t, persisted.Env, "env") {
		t.Fatalf("expected env payload to contain key 'env': %s", string(persisted.Env))
	}
}

type executorFixture struct {
	executor  *vmexec.Executor
	sessionID string
	worktree  string
}

func newExecutorFixture(t *testing.T) executorFixture {
	t.Helper()

	tmp := t.TempDir()
	dbPath := filepath.Join(tmp, "vm-system.db")
	worktree := filepath.Join(tmp, "worktree")
	if err := os.MkdirAll(worktree, 0o755); err != nil {
		t.Fatalf("mkdir worktree: %v", err)
	}

	store, err := vmstore.NewVMStore(dbPath)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })

	templateService := vmcontrol.NewTemplateService(store)
	vm, err := templateService.Create(context.Background(), vmcontrol.CreateTemplateInput{Name: "vmexec-regression-template", Engine: "goja"})
	if err != nil {
		t.Fatalf("create template: %v", err)
	}

	sessionManager := vmsession.NewSessionManager(store)
	session, err := sessionManager.CreateSession(vm.ID, "workspace-vmexec", "deadbeef", worktree)
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	return executorFixture{
		executor:  vmexec.NewExecutor(store, sessionManager),
		sessionID: session.ID,
		worktree:  worktree,
	}
}

func jsonContainsKV(t *testing.T, raw json.RawMessage, key string) bool {
	t.Helper()
	if len(raw) == 0 {
		return false
	}
	m := map[string]interface{}{}
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatalf("unmarshal json map: %v", err)
	}
	_, ok := m[key]
	return ok
}
