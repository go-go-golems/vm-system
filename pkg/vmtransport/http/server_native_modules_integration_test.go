package vmhttp_test

import (
	"encoding/json"
	"fmt"
	"path/filepath"
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
