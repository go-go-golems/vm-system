package vmhttp_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLibrariesLodashConfiguredAndCachedVsUnconfigured(t *testing.T) {
	testRoot := t.TempDir()
	chdirForTest(t, testRoot)

	cacheDir := filepath.Join(".vm-cache", "libraries")
	mustMkdirAll(t, cacheDir)
	writeFile(t, filepath.Join(cacheDir, "lodash-4.17.21.js"), lodashFixture())

	server, client := newIntegrationTestServer(t)
	defer server.Close()

	worktree := filepath.Join(testRoot, "worktree")
	mustMkdirAll(t, worktree)

	templateWithLodash := createTemplateForTest(t, client, server.URL, "lodash-enabled-template")
	postJSON(t, client, fmt.Sprintf("%s/api/v1/templates/%s/libraries", server.URL, templateWithLodash), map[string]interface{}{
		"name": "lodash-4.17.21",
	}, &map[string]interface{}{})

	sessionWithLodash := createSessionForTest(t, client, server.URL, templateWithLodash, worktree, "ws-lodash-on")
	chunkCount := executionResponse{}
	postJSON(t, client, server.URL+"/api/v1/executions/repl", map[string]interface{}{
		"session_id": sessionWithLodash,
		"input":      `_.chunk([1,2,3,4], 2).length`,
	}, &chunkCount)
	if chunkCount.Status != "ok" {
		t.Fatalf("expected lodash chunk execution to succeed, got status=%q error=%q", chunkCount.Status, chunkCount.Error.Message)
	}
	if resultPreview(t, chunkCount.Result) != "2" {
		t.Fatalf("expected lodash chunk length preview 2, got %q", resultPreview(t, chunkCount.Result))
	}

	templateWithoutLodash := createTemplateForTest(t, client, server.URL, "lodash-disabled-template")
	sessionWithoutLodash := createSessionForTest(t, client, server.URL, templateWithoutLodash, worktree, "ws-lodash-off")
	missingUnderscore := executionResponse{}
	postJSON(t, client, server.URL+"/api/v1/executions/repl", map[string]interface{}{
		"session_id": sessionWithoutLodash,
		"input":      `_.chunk([1,2], 1)`,
	}, &missingUnderscore)
	if missingUnderscore.Status != "error" {
		t.Fatalf("expected lodash call to fail when not configured, got status=%q", missingUnderscore.Status)
	}
	if !strings.Contains(missingUnderscore.Error.Message, "is not defined") {
		t.Fatalf("expected missing lodash error to mention undefined underscore, got %q", missingUnderscore.Error.Message)
	}
}

func TestSessionCreateConfiguredLibraryMissingCacheFails(t *testing.T) {
	testRoot := t.TempDir()
	chdirForTest(t, testRoot)

	mustMkdirAll(t, filepath.Join(".vm-cache", "libraries"))

	server, client := newIntegrationTestServer(t)
	defer server.Close()

	worktree := filepath.Join(testRoot, "worktree")
	mustMkdirAll(t, worktree)

	templateWithLodash := createTemplateForTest(t, client, server.URL, "lodash-missing-cache-template")
	postJSON(t, client, fmt.Sprintf("%s/api/v1/templates/%s/libraries", server.URL, templateWithLodash), map[string]interface{}{
		"name": "lodash-4.17.21",
	}, &map[string]interface{}{})

	reqBody, err := json.Marshal(map[string]interface{}{
		"template_id":     templateWithLodash,
		"workspace_id":    "ws-lodash-missing-cache",
		"base_commit_oid": "deadbeef",
		"worktree_path":   worktree,
	})
	if err != nil {
		t.Fatalf("marshal create session body: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, server.URL+"/api/v1/sessions", bytes.NewReader(reqBody))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("create session request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		raw, _ := ioReadAll(resp)
		t.Fatalf("expected status 500 when configured library cache is missing, got %d (%s)", resp.StatusCode, string(raw))
	}

	env := struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}{}
	if err := json.NewDecoder(resp.Body).Decode(&env); err != nil {
		t.Fatalf("decode error envelope: %v", err)
	}
	if env.Error.Code != "INTERNAL" {
		t.Fatalf("expected INTERNAL error code, got %q", env.Error.Code)
	}
	if !strings.Contains(env.Error.Message, "failed to load libraries") {
		t.Fatalf("expected missing cache message to mention failed library loading, got %q", env.Error.Message)
	}

	sessions := []struct {
		WorkspaceID string `json:"workspace_id"`
		Status      string `json:"status"`
		LastError   string `json:"last_error"`
	}{}
	getJSON(t, client, server.URL+"/api/v1/sessions", &sessions)

	found := false
	for _, s := range sessions {
		if s.WorkspaceID != "ws-lodash-missing-cache" {
			continue
		}
		found = true
		if s.Status != "crashed" {
			t.Fatalf("expected failed create session to be marked crashed, got status=%q", s.Status)
		}
		if !strings.Contains(s.LastError, "failed to load libraries") {
			t.Fatalf("expected last_error to include library load failure, got %q", s.LastError)
		}
		break
	}
	if !found {
		t.Fatalf("expected failed session record to exist in session list")
	}
}

func chdirForTest(t *testing.T, dir string) {
	t.Helper()
	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir to %s: %v", dir, err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(oldWD); err != nil {
			t.Fatalf("restore cwd to %s: %v", oldWD, err)
		}
	})
}

func lodashFixture() string {
	return `var _ = {
  chunk: function(arr, size) {
    var out = [];
    for (var i = 0; i < arr.length; i += size) {
      out.push(arr.slice(i, i + size));
    }
    return out;
  }
};`
}
