package vmdaemon

import (
	"net/http"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-go-golems/vm-system/pkg/vmmodels"
	"github.com/go-go-golems/vm-system/pkg/vmstore"
)

func TestNewClosesStaleSessionsOnStartup(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "vm-system.db")
	store := mustNewStore(t, dbPath)

	vmID := "vm-gc-test"
	now := time.Now().Add(-2 * time.Minute)
	if err := store.CreateVM(&vmmodels.VM{
		ID:        vmID,
		Name:      "gc-template",
		Engine:    "goja",
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}); err != nil {
		t.Fatalf("create vm: %v", err)
	}

	createSession := func(id, workspaceID, status, lastError string, closedAt *time.Time) {
		t.Helper()
		session := &vmmodels.VMSession{
			ID:            id,
			VMID:          vmID,
			WorkspaceID:   workspaceID,
			BaseCommitOID: "deadbeef",
			WorktreePath:  "/tmp/worktree",
			Status:        status,
			CreatedAt:     now,
			LastError:     lastError,
			ClosedAt:      closedAt,
		}
		if err := store.CreateSession(session); err != nil {
			t.Fatalf("create session %s: %v", id, err)
		}
		if session.LastError != "" || session.ClosedAt != nil || session.Status == string(vmmodels.SessionClosed) {
			if err := store.UpdateSession(session); err != nil {
				t.Fatalf("update session %s: %v", id, err)
			}
		}
	}

	closedAt := now.Add(15 * time.Second)
	createSession("session-starting", "ws-starting", string(vmmodels.SessionStarting), "", nil)
	createSession("session-ready", "ws-ready", string(vmmodels.SessionReady), "", nil)
	createSession("session-crashed", "ws-crashed", string(vmmodels.SessionCrashed), "startup failed: boom", nil)
	createSession("session-closed", "ws-closed", string(vmmodels.SessionClosed), "already closed", &closedAt)

	if err := store.Close(); err != nil {
		t.Fatalf("close seed store: %v", err)
	}

	app, err := New(DefaultConfig(dbPath), http.NewServeMux())
	if err != nil {
		t.Fatalf("new daemon app: %v", err)
	}
	defer app.Close()

	sessions, err := app.store.ListSessions("")
	if err != nil {
		t.Fatalf("list sessions after startup: %v", err)
	}

	byID := map[string]*vmmodels.VMSession{}
	for _, session := range sessions {
		byID[session.ID] = session
	}

	assertClosedWithGCReason(t, byID["session-starting"])
	assertClosedWithGCReason(t, byID["session-ready"])

	if got := byID["session-crashed"]; got.Status != string(vmmodels.SessionCrashed) || got.LastError != "startup failed: boom" {
		t.Fatalf("crashed session unexpectedly changed: status=%q last_error=%q", got.Status, got.LastError)
	}
	if got := byID["session-closed"]; got.Status != string(vmmodels.SessionClosed) || got.LastError != "already closed" {
		t.Fatalf("closed session unexpectedly changed: status=%q last_error=%q", got.Status, got.LastError)
	}
}

func mustNewStore(t *testing.T, dbPath string) *vmstore.VMStore {
	t.Helper()
	store, err := vmstore.NewVMStore(dbPath)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	return store
}

func assertClosedWithGCReason(t *testing.T, session *vmmodels.VMSession) {
	t.Helper()
	if session == nil {
		t.Fatalf("expected session record to exist")
	}
	if session.Status != string(vmmodels.SessionClosed) {
		t.Fatalf("expected session %s to be closed, got %q", session.ID, session.Status)
	}
	if session.ClosedAt == nil {
		t.Fatalf("expected session %s to have closed_at set", session.ID)
	}
	if session.LastError != sessionStartupGCMessage {
		t.Fatalf("expected session %s last_error %q, got %q", session.ID, sessionStartupGCMessage, session.LastError)
	}
}
