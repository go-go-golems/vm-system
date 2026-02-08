package vmhttp_test

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/go-go-golems/vm-system/pkg/vmcontrol"
	"github.com/go-go-golems/vm-system/pkg/vmexec"
	"github.com/go-go-golems/vm-system/pkg/vmsession"
	"github.com/go-go-golems/vm-system/pkg/vmstore"
	vmhttp "github.com/go-go-golems/vm-system/pkg/vmtransport/http"
)

func TestAPIErrorContractsValidationNotFoundConflictAndUnprocessable(t *testing.T) {
	server, client, sessionManager := newIntegrationServerWithSessionManager(t)
	defer server.Close()

	tests := []struct {
		name         string
		method       string
		path         string
		body         interface{}
		status       int
		expectedCode string
	}{
		{
			name:         "missing template name",
			method:       http.MethodPost,
			path:         "/api/v1/templates",
			body:         map[string]interface{}{},
			status:       http.StatusBadRequest,
			expectedCode: "VALIDATION_ERROR",
		},
		{
			name:         "executions list missing session_id",
			method:       http.MethodGet,
			path:         "/api/v1/executions?limit=3",
			body:         nil,
			status:       http.StatusBadRequest,
			expectedCode: "VALIDATION_ERROR",
		},
		{
			name:         "events invalid after_seq",
			method:       http.MethodGet,
			path:         "/api/v1/executions/abc/events?after_seq=-1",
			body:         nil,
			status:       http.StatusBadRequest,
			expectedCode: "VALIDATION_ERROR",
		},
		{
			name:         "template not found",
			method:       http.MethodGet,
			path:         "/api/v1/templates/does-not-exist",
			body:         nil,
			status:       http.StatusNotFound,
			expectedCode: "TEMPLATE_NOT_FOUND",
		},
		{
			name:         "session not found",
			method:       http.MethodGet,
			path:         "/api/v1/sessions/does-not-exist",
			body:         nil,
			status:       http.StatusNotFound,
			expectedCode: "SESSION_NOT_FOUND",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			doRequest(t, client, tc.method, server.URL+tc.path, tc.body, tc.status, map[string]string{
				"code": tc.expectedCode,
			})
		})
	}

	worktree := filepath.Join(t.TempDir(), "worktree")
	mustMkdirAll(t, worktree)
	templateID := createTemplateForTest(t, client, server.URL, "error-contract-template")
	sessionID := createSessionForTest(t, client, server.URL, templateID, worktree, "ws-errors")

	session, err := sessionManager.GetSession(sessionID)
	if err != nil {
		t.Fatalf("get session from manager: %v", err)
	}
	session.ExecutionLock.Lock()
	doRequest(t, client, http.MethodPost, server.URL+"/api/v1/executions/repl", map[string]interface{}{
		"session_id": sessionID,
		"input":      "1+1",
	}, http.StatusConflict, map[string]string{
		"code": "SESSION_BUSY",
	})
	session.ExecutionLock.Unlock()

	doRequest(t, client, http.MethodPost, server.URL+"/api/v1/executions/run-file", map[string]interface{}{
		"session_id": sessionID,
		"path":       "../etc/passwd",
	}, http.StatusUnprocessableEntity, map[string]string{
		"code": "INVALID_PATH",
	})
}

func newIntegrationServerWithSessionManager(t *testing.T) (*httptest.Server, *http.Client, *vmsession.SessionManager) {
	t.Helper()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "vm-system.db")
	store, err := vmstore.NewVMStore(dbPath)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })

	sessionManager := vmsession.NewSessionManager(store)
	executor := vmexec.NewExecutor(store, sessionManager)
	core := vmcontrol.NewCoreWithPorts(store, sessionManager, executor)
	server := httptest.NewServer(vmhttp.NewHandler(core))
	return server, server.Client(), sessionManager
}
