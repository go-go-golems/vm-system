package vmhttp_test

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

func TestSessionLifecycleEndpoints(t *testing.T) {
	server, client := newIntegrationTestServer(t)
	defer server.Close()

	worktree := filepath.Join(t.TempDir(), "worktree")
	mustMkdirAll(t, worktree)

	templateID := createTemplateForTest(t, client, server.URL, "session-lifecycle-template")

	createSession := func(workspace string) string {
		session := struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		}{}
		postJSON(t, client, server.URL+"/api/v1/sessions", map[string]interface{}{
			"template_id":     templateID,
			"workspace_id":    workspace,
			"base_commit_oid": "deadbeef",
			"worktree_path":   worktree,
		}, &session)
		if session.ID == "" {
			t.Fatalf("expected created session id")
		}
		if session.Status != "ready" {
			t.Fatalf("expected created session status ready, got %q", session.Status)
		}
		return session.ID
	}

	sessionA := createSession("ws-a")
	sessionB := createSession("ws-b")

	summary := struct {
		ActiveSessions int `json:"active_sessions"`
	}{}
	getJSON(t, client, server.URL+"/api/v1/runtime/summary", &summary)
	if summary.ActiveSessions != 2 {
		t.Fatalf("expected 2 active sessions after creation, got %d", summary.ActiveSessions)
	}

	listAll := []struct {
		ID string `json:"id"`
	}{}
	getJSON(t, client, server.URL+"/api/v1/sessions", &listAll)
	if len(listAll) < 2 {
		t.Fatalf("expected at least 2 sessions in list, got %d", len(listAll))
	}

	sessionGet := struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}{}
	getJSON(t, client, fmt.Sprintf("%s/api/v1/sessions/%s", server.URL, sessionA), &sessionGet)
	if sessionGet.ID != sessionA {
		t.Fatalf("expected session id %s, got %s", sessionA, sessionGet.ID)
	}

	readySessions := []struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}{}
	getJSON(t, client, server.URL+"/api/v1/sessions?status=ready", &readySessions)
	if len(readySessions) < 2 {
		t.Fatalf("expected at least 2 ready sessions, got %d", len(readySessions))
	}

	closed := struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}{}
	doRequest(t, client, "POST", fmt.Sprintf("%s/api/v1/sessions/%s/close", server.URL, sessionA), map[string]string{}, 200, nil)
	getJSON(t, client, fmt.Sprintf("%s/api/v1/sessions/%s", server.URL, sessionA), &closed)
	if closed.Status != "closed" {
		t.Fatalf("expected closed status after close, got %q", closed.Status)
	}
	getJSON(t, client, server.URL+"/api/v1/runtime/summary", &summary)
	if summary.ActiveSessions != 1 {
		t.Fatalf("expected 1 active session after closing one, got %d", summary.ActiveSessions)
	}

	deleted := struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}{}
	doRequest(t, client, "DELETE", fmt.Sprintf("%s/api/v1/sessions/%s", server.URL, sessionB), nil, 200, nil)
	getJSON(t, client, fmt.Sprintf("%s/api/v1/sessions/%s", server.URL, sessionB), &deleted)
	if deleted.Status != "closed" {
		t.Fatalf("expected closed status after delete alias, got %q", deleted.Status)
	}
	getJSON(t, client, server.URL+"/api/v1/runtime/summary", &summary)
	if summary.ActiveSessions != 0 {
		t.Fatalf("expected 0 active sessions after closing all, got %d", summary.ActiveSessions)
	}

	closedSessions := []struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}{}
	getJSON(t, client, server.URL+"/api/v1/sessions?status=closed", &closedSessions)
	if len(closedSessions) < 2 {
		t.Fatalf("expected at least 2 closed sessions, got %d", len(closedSessions))
	}

	doRequest(t, client, "GET", fmt.Sprintf("%s/api/v1/sessions/%s", server.URL, "does-not-exist"), nil, 404, map[string]string{
		"code": "SESSION_NOT_FOUND",
	})
}

func createTemplateForTest(t *testing.T, client *http.Client, baseURL, name string) string {
	t.Helper()
	out := struct {
		ID string `json:"id"`
	}{}
	postJSON(t, client, baseURL+"/api/v1/templates", map[string]interface{}{"name": name}, &out)
	if out.ID == "" {
		t.Fatalf("expected template id")
	}
	return out.ID
}

func mustMkdirAll(t *testing.T, p string) {
	t.Helper()
	if err := os.MkdirAll(p, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", p, err)
	}
}
