package vmhttp_test

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
)

func TestNativeModulesRequireAndJSONBuiltinSemantics(t *testing.T) {
	server, client := newIntegrationTestServer(t)
	defer server.Close()

	worktree := filepath.Join(t.TempDir(), "worktree")
	mustMkdirAll(t, worktree)
	nativeFile := filepath.Join(worktree, "native.txt")
	writeFile(t, nativeFile, "hello-native-module")

	templateID := createTemplateForTest(t, client, server.URL, "native-module-template")

	// JSON is a JS built-in and should work without template module configuration.
	sessionNoModules := createSessionForTest(t, client, server.URL, templateID, worktree, "ws-native-0")
	jsonExec := struct {
		ID     string          `json:"id"`
		Status string          `json:"status"`
		Result json.RawMessage `json:"result"`
	}{}
	postJSON(t, client, server.URL+"/api/v1/executions/repl", map[string]interface{}{
		"session_id": sessionNoModules,
		"input":      `JSON.stringify({answer: 42})`,
	}, &jsonExec)
	if jsonExec.Status != "ok" {
		t.Fatalf("expected JSON.stringify execution status ok, got %q", jsonExec.Status)
	}

	// Built-ins cannot be configured per template anymore.
	doRequest(t, client, "POST", fmt.Sprintf("%s/api/v1/templates/%s/modules", server.URL, templateID), map[string]interface{}{
		"name": "json",
	}, 422, map[string]string{
		"code": "MODULE_NOT_ALLOWED",
	})

	// Configuring a registered native module should enable require(\"fs\") usage.
	postJSON(t, client, fmt.Sprintf("%s/api/v1/templates/%s/modules", server.URL, templateID), map[string]interface{}{
		"name": "fs",
	}, &map[string]interface{}{})

	sessionWithFS := createSessionForTest(t, client, server.URL, templateID, worktree, "ws-native-1")
	replExec := struct {
		ID     string          `json:"id"`
		Status string          `json:"status"`
		Result json.RawMessage `json:"result"`
	}{}
	postJSON(t, client, server.URL+"/api/v1/executions/repl", map[string]interface{}{
		"session_id": sessionWithFS,
		"input":      fmt.Sprintf(`const fs = require("fs"); fs.readFileSync(%q);`, nativeFile),
	}, &replExec)
	if replExec.Status != "ok" {
		t.Fatalf("expected native module execution status ok, got %q", replExec.Status)
	}

	resultPayload := struct {
		Preview string `json:"preview"`
	}{}
	if err := json.Unmarshal(replExec.Result, &resultPayload); err != nil {
		t.Fatalf("unmarshal native module result payload: %v", err)
	}
	if resultPayload.Preview != "hello-native-module" {
		t.Fatalf("expected fs.readFileSync preview hello-native-module, got %q", resultPayload.Preview)
	}
}

func TestNativeModulesDatabaseAndExecConfiguredVsUnconfigured(t *testing.T) {
	server, client := newIntegrationTestServer(t)
	defer server.Close()

	worktree := filepath.Join(t.TempDir(), "worktree")
	mustMkdirAll(t, worktree)

	templateNoModules := createTemplateForTest(t, client, server.URL, "native-module-disabled-template")
	sessionNoModules := createSessionForTest(t, client, server.URL, templateNoModules, worktree, "ws-native-disabled")

	requireDatabase := executionResponse{}
	postJSON(t, client, server.URL+"/api/v1/executions/repl", map[string]interface{}{
		"session_id": sessionNoModules,
		"input":      `require("database")`,
	}, &requireDatabase)
	if requireDatabase.Status != "error" {
		t.Fatalf("expected require(database) to fail when module is not configured, got %q", requireDatabase.Status)
	}
	if !strings.Contains(requireDatabase.Error.Message, "Invalid module") {
		t.Fatalf("expected require(database) error message to mention invalid module, got %q", requireDatabase.Error.Message)
	}

	requireExec := executionResponse{}
	postJSON(t, client, server.URL+"/api/v1/executions/repl", map[string]interface{}{
		"session_id": sessionNoModules,
		"input":      `require("exec")`,
	}, &requireExec)
	if requireExec.Status != "error" {
		t.Fatalf("expected require(exec) to fail when module is not configured, got %q", requireExec.Status)
	}
	if !strings.Contains(requireExec.Error.Message, "Invalid module") {
		t.Fatalf("expected require(exec) error message to mention invalid module, got %q", requireExec.Error.Message)
	}

	templateWithModules := createTemplateForTest(t, client, server.URL, "native-module-enabled-template")
	postJSON(t, client, fmt.Sprintf("%s/api/v1/templates/%s/modules", server.URL, templateWithModules), map[string]interface{}{
		"name": "database",
	}, &map[string]interface{}{})
	postJSON(t, client, fmt.Sprintf("%s/api/v1/templates/%s/modules", server.URL, templateWithModules), map[string]interface{}{
		"name": "exec",
	}, &map[string]interface{}{})

	sessionWithModules := createSessionForTest(t, client, server.URL, templateWithModules, worktree, "ws-native-enabled")

	execModuleRun := executionResponse{}
	postJSON(t, client, server.URL+"/api/v1/executions/repl", map[string]interface{}{
		"session_id": sessionWithModules,
		"input":      `(() => { const execModule = require("exec"); return execModule.run("/bin/echo", ["exec-module-ok"]).trim(); })()`,
	}, &execModuleRun)
	if execModuleRun.Status != "ok" {
		t.Fatalf("expected configured exec module to run successfully, got status=%q error=%q", execModuleRun.Status, execModuleRun.Error.Message)
	}
	if resultPreview(t, execModuleRun.Result) != "exec-module-ok" {
		t.Fatalf("expected exec module preview exec-module-ok, got %q", resultPreview(t, execModuleRun.Result))
	}

	databaseModuleRun := executionResponse{}
	postJSON(t, client, server.URL+"/api/v1/executions/repl", map[string]interface{}{
		"session_id": sessionWithModules,
		"input":      `(() => { const databaseModule = require("database"); databaseModule.configure("sqlite3", ":memory:"); databaseModule.exec("CREATE TABLE t (id INTEGER PRIMARY KEY, name TEXT)"); databaseModule.exec("INSERT INTO t(name) VALUES (?)", "alice"); return databaseModule.query("SELECT name FROM t WHERE id = ?", 1)[0].name; })()`,
	}, &databaseModuleRun)
	if databaseModuleRun.Status != "ok" {
		t.Fatalf("expected configured database module to run successfully, got status=%q error=%q", databaseModuleRun.Status, databaseModuleRun.Error.Message)
	}
	if resultPreview(t, databaseModuleRun.Result) != "alice" {
		t.Fatalf("expected database module preview alice, got %q", resultPreview(t, databaseModuleRun.Result))
	}
}

type executionResponse struct {
	ID     string          `json:"id"`
	Status string          `json:"status"`
	Result json.RawMessage `json:"result"`
	Error  struct {
		Message string `json:"message"`
	} `json:"error"`
}

func resultPreview(t *testing.T, raw json.RawMessage) string {
	t.Helper()
	var payload struct {
		Preview string `json:"preview"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("unmarshal result preview: %v", err)
	}
	return payload.Preview
}
