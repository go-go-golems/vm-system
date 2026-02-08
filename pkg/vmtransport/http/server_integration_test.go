package vmhttp_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-go-golems/vm-system/pkg/vmcontrol"
	"github.com/go-go-golems/vm-system/pkg/vmstore"
	vmhttp "github.com/go-go-golems/vm-system/pkg/vmtransport/http"
)

func TestSessionContinuityAcrossAPIRequests(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "vm-system.db")
	worktree := filepath.Join(tmpDir, "worktree")
	if err := os.MkdirAll(worktree, 0o755); err != nil {
		t.Fatalf("mkdir worktree: %v", err)
	}

	store, err := vmstore.NewVMStore(dbPath)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	defer store.Close()

	core := vmcontrol.NewCore(store)
	server := httptest.NewServer(vmhttp.NewHandler(core))
	defer server.Close()

	templateResp := struct {
		ID string `json:"id"`
	}{}
	postJSON(t, server.Client(), server.URL+"/api/v1/templates", map[string]interface{}{
		"name": "continuity-template",
	}, &templateResp)
	if templateResp.ID == "" {
		t.Fatalf("expected template id")
	}

	sessionResp := struct {
		ID string `json:"id"`
	}{}
	postJSON(t, server.Client(), server.URL+"/api/v1/sessions", map[string]interface{}{
		"template_id":     templateResp.ID,
		"workspace_id":    "workspace-continuity",
		"base_commit_oid": "deadbeef",
		"worktree_path":   worktree,
	}, &sessionResp)
	if sessionResp.ID == "" {
		t.Fatalf("expected session id")
	}

	// First request seeds runtime state.
	firstExecution := struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}{}
	postJSON(t, server.Client(), server.URL+"/api/v1/executions/repl", map[string]interface{}{
		"session_id": sessionResp.ID,
		"input":      "var persisted = 20; persisted",
	}, &firstExecution)
	if firstExecution.Status != "ok" {
		t.Fatalf("expected first execution to succeed, got status=%s", firstExecution.Status)
	}

	// Second independent request verifies continuity in same session runtime.
	secondExecution := struct {
		ID     string          `json:"id"`
		Status string          `json:"status"`
		Result json.RawMessage `json:"result"`
	}{}
	postJSON(t, server.Client(), server.URL+"/api/v1/executions/repl", map[string]interface{}{
		"session_id": sessionResp.ID,
		"input":      "persisted + 22",
	}, &secondExecution)
	if secondExecution.Status != "ok" {
		t.Fatalf("expected second execution to succeed, got status=%s", secondExecution.Status)
	}

	resultPayload := struct {
		Preview string `json:"preview"`
	}{}
	if err := json.Unmarshal(secondExecution.Result, &resultPayload); err != nil {
		t.Fatalf("unmarshal execution result: %v", err)
	}
	if resultPayload.Preview != "42" {
		t.Fatalf("expected continuity result preview=42, got %q", resultPayload.Preview)
	}

	eventsResp := []map[string]interface{}{}
	getJSON(t, server.Client(), fmt.Sprintf("%s/api/v1/executions/%s/events?after_seq=0", server.URL, secondExecution.ID), &eventsResp)
	if len(eventsResp) == 0 {
		t.Fatalf("expected events for second execution")
	}

	summaryResp := struct {
		ActiveSessions int `json:"active_sessions"`
	}{}
	getJSON(t, server.Client(), server.URL+"/api/v1/runtime/summary", &summaryResp)
	if summaryResp.ActiveSessions != 1 {
		t.Fatalf("expected one active session, got %d", summaryResp.ActiveSessions)
	}
}

func postJSON(t *testing.T, client *http.Client, url string, in interface{}, out interface{}) {
	t.Helper()

	body, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	resp, err := client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("post %s: %v", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		raw, _ := ioReadAll(resp)
		t.Fatalf("unexpected status code %d for %s: %s", resp.StatusCode, url, string(raw))
	}

	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		t.Fatalf("decode response for %s: %v", url, err)
	}
}

func getJSON(t *testing.T, client *http.Client, url string, out interface{}) {
	t.Helper()

	resp, err := client.Get(url)
	if err != nil {
		t.Fatalf("get %s: %v", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		raw, _ := ioReadAll(resp)
		t.Fatalf("unexpected status code %d for %s: %s", resp.StatusCode, url, string(raw))
	}

	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		t.Fatalf("decode response for %s: %v", url, err)
	}
}

func ioReadAll(resp *http.Response) ([]byte, error) {
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(resp.Body)
	return buf.Bytes(), err
}
