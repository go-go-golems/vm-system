package vmhttp_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-go-golems/vm-system/pkg/vmcontrol"
	"github.com/go-go-golems/vm-system/pkg/vmmodels"
	"github.com/go-go-golems/vm-system/pkg/vmstore"
	vmhttp "github.com/go-go-golems/vm-system/pkg/vmtransport/http"
)

func TestSafetyPathTraversalAndOutputLimitEnforcement(t *testing.T) {
	server, client, store := newIntegrationServerWithStore(t)
	defer server.Close()

	worktree := filepath.Join(t.TempDir(), "worktree")
	mustMkdirAll(t, worktree)

	templateID := createTemplateForTest(t, client, server.URL, "safety-template")
	setTightLimitsForTemplate(t, store, templateID)
	sessionID := createSessionForTest(t, client, server.URL, templateID, worktree, "ws-safety")

	doRequest(t, client, http.MethodPost, server.URL+"/api/v1/executions/run-file", map[string]interface{}{
		"session_id": sessionID,
		"path":       "../etc/passwd",
	}, http.StatusUnprocessableEntity, map[string]string{
		"code": "INVALID_PATH",
	})

	outsideDir := t.TempDir()
	outsideFile := filepath.Join(outsideDir, "outside.js")
	if err := os.WriteFile(outsideFile, []byte("40 + 2"), 0o644); err != nil {
		t.Fatalf("write outside file: %v", err)
	}
	if err := os.Symlink(outsideFile, filepath.Join(worktree, "escape-link.js")); err != nil {
		t.Fatalf("create symlink: %v", err)
	}
	doRequest(t, client, http.MethodPost, server.URL+"/api/v1/executions/run-file", map[string]interface{}{
		"session_id": sessionID,
		"path":       "escape-link.js",
	}, http.StatusUnprocessableEntity, map[string]string{
		"code": "INVALID_PATH",
	})

	doRequest(t, client, http.MethodPost, server.URL+"/api/v1/executions/repl", map[string]interface{}{
		"session_id": sessionID,
		"input":      "1+1",
	}, http.StatusUnprocessableEntity, map[string]string{
		"code": "OUTPUT_LIMIT_EXCEEDED",
	})
}

func newIntegrationServerWithStore(t *testing.T) (*httptest.Server, *http.Client, *vmstore.VMStore) {
	t.Helper()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "vm-system.db")
	store, err := vmstore.NewVMStore(dbPath)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })

	core := vmcontrol.NewCore(store)
	server := httptest.NewServer(vmhttp.NewHandler(core))
	return server, server.Client(), store
}

func setTightLimitsForTemplate(t *testing.T, store *vmstore.VMStore, templateID string) {
	t.Helper()

	settings := &vmmodels.VMSettings{
		VMID: templateID,
		Limits: json.RawMessage(`{
      "cpu_ms": 2000,
      "wall_ms": 5000,
      "mem_mb": 128,
      "max_events": 1,
      "max_output_kb": 1
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
	}
	if err := store.SetVMSettings(settings); err != nil {
		t.Fatalf("set tight limits: %v", err)
	}
}
