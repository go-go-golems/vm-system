package vmhttp_test

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

func TestExecutionEndpointsLifecycle(t *testing.T) {
	server, client := newIntegrationTestServer(t)
	defer server.Close()

	worktree := filepath.Join(t.TempDir(), "worktree")
	mustMkdirAll(t, worktree)
	writeFile(t, filepath.Join(worktree, "calc.js"), "console.log('calc'); 40 + 2;")

	templateID := createTemplateForTest(t, client, server.URL, "execution-template")
	sessionID := createSessionForTest(t, client, server.URL, templateID, worktree, "ws-exec")

	replExec := struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}{}
	postJSON(t, client, server.URL+"/api/v1/executions/repl", map[string]interface{}{
		"session_id": sessionID,
		"input":      "21 * 2",
	}, &replExec)
	if replExec.ID == "" || replExec.Status != "ok" {
		t.Fatalf("expected successful repl execution, got id=%q status=%q", replExec.ID, replExec.Status)
	}

	runFileExec := struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}{}
	postJSON(t, client, server.URL+"/api/v1/executions/run-file", map[string]interface{}{
		"session_id": sessionID,
		"path":       "calc.js",
	}, &runFileExec)
	if runFileExec.ID == "" || runFileExec.Status != "ok" {
		t.Fatalf("expected successful run-file execution, got id=%q status=%q", runFileExec.ID, runFileExec.Status)
	}

	getExec := struct {
		ID     string `json:"id"`
		Status string `json:"status"`
		Kind   string `json:"kind"`
	}{}
	getJSON(t, client, fmt.Sprintf("%s/api/v1/executions/%s", server.URL, replExec.ID), &getExec)
	if getExec.ID != replExec.ID || getExec.Kind != "repl" {
		t.Fatalf("expected repl execution lookup, got id=%q kind=%q", getExec.ID, getExec.Kind)
	}

	listExec := []struct {
		ID string `json:"id"`
	}{}
	getJSON(t, client, fmt.Sprintf("%s/api/v1/executions?session_id=%s&limit=1", server.URL, sessionID), &listExec)
	if len(listExec) != 1 {
		t.Fatalf("expected one execution with limit=1, got %d", len(listExec))
	}

	events := []struct {
		Seq int `json:"seq"`
	}{}
	getJSON(t, client, fmt.Sprintf("%s/api/v1/executions/%s/events?after_seq=0", server.URL, replExec.ID), &events)
	if len(events) == 0 {
		t.Fatalf("expected events for repl execution")
	}

	eventsAfterFirst := []struct {
		Seq int `json:"seq"`
	}{}
	getJSON(t, client, fmt.Sprintf("%s/api/v1/executions/%s/events?after_seq=1", server.URL, replExec.ID), &eventsAfterFirst)
	if len(eventsAfterFirst) >= len(events) {
		t.Fatalf("expected fewer events after_seq filter; got all=%d filtered=%d", len(events), len(eventsAfterFirst))
	}

	doRequest(t, client, http.MethodGet, fmt.Sprintf("%s/api/v1/executions/%s", server.URL, "does-not-exist"), nil, http.StatusNotFound, map[string]string{
		"code": "EXECUTION_NOT_FOUND",
	})
}

func createSessionForTest(t *testing.T, client *http.Client, baseURL, templateID, worktree, workspaceID string) string {
	t.Helper()
	session := struct {
		ID string `json:"id"`
	}{}
	postJSON(t, client, baseURL+"/api/v1/sessions", map[string]interface{}{
		"template_id":     templateID,
		"workspace_id":    workspaceID,
		"base_commit_oid": "deadbeef",
		"worktree_path":   worktree,
	}, &session)
	if session.ID == "" {
		t.Fatalf("expected session id")
	}
	return session.ID
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write file %s: %v", path, err)
	}
}
