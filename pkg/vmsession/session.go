package vmsession

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/dop251/goja"
	"github.com/google/uuid"

	"github.com/go-go-golems/vm-system/pkg/vmmodels"
	"github.com/go-go-golems/vm-system/pkg/vmmodules"
	"github.com/go-go-golems/vm-system/pkg/vmpath"
	"github.com/go-go-golems/vm-system/pkg/vmstore"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// SessionManager manages VM sessions
type SessionManager struct {
	store      *vmstore.VMStore
	sessions   map[string]*Session
	sessionsMu sync.RWMutex
	logger     zerolog.Logger
}

// Session represents an active VM session
type Session struct {
	ID            string
	VMID          string
	WorkspaceID   string
	BaseCommitOID string
	WorktreePath  string
	Status        vmmodels.SessionStatus
	Runtime       *goja.Runtime
	ExecutionLock sync.Mutex
	CreatedAt     time.Time
	LastError     string
}

// NewSessionManager creates a new SessionManager
func NewSessionManager(store *vmstore.VMStore) *SessionManager {
	return &SessionManager{
		store:    store,
		sessions: make(map[string]*Session),
		logger:   log.With().Str("component", "session_manager").Logger(),
	}
}

// CreateSession creates a new VM session
func (sm *SessionManager) CreateSession(vmID, workspaceID, baseCommitOID, worktreePath string) (*Session, error) {
	// Verify VM exists
	vm, err := sm.store.GetVM(vmID)
	if err != nil {
		return nil, fmt.Errorf("VM not found: %w", err)
	}

	// Get VM settings
	settings, err := sm.store.GetVMSettings(vmID)
	if err != nil {
		return nil, fmt.Errorf("failed to get VM settings: %w", err)
	}

	// Verify worktree path exists
	if _, err := os.Stat(worktreePath); err != nil {
		return nil, fmt.Errorf("worktree path does not exist: %w", err)
	}

	// Create session record
	sessionID := uuid.New().String()
	session := &Session{
		ID:            sessionID,
		VMID:          vmID,
		WorkspaceID:   workspaceID,
		BaseCommitOID: baseCommitOID,
		WorktreePath:  worktreePath,
		Status:        vmmodels.SessionStarting,
		CreatedAt:     time.Now(),
	}

	// Store session in database
	dbSession := &vmmodels.VMSession{
		ID:            session.ID,
		VMID:          session.VMID,
		WorkspaceID:   session.WorkspaceID,
		BaseCommitOID: session.BaseCommitOID,
		WorktreePath:  session.WorktreePath,
		Status:        string(session.Status),
		CreatedAt:     session.CreatedAt,
	}

	if err := sm.store.CreateSession(dbSession); err != nil {
		return nil, fmt.Errorf("failed to create session in database: %w", err)
	}

	// Initialize goja runtime
	if vm.Engine == "goja" {
		runtime := goja.New()
		session.Runtime = runtime

		// Parse runtime settings
		var runtimeConfig vmmodels.RuntimeConfig
		if err := json.Unmarshal(settings.Runtime, &runtimeConfig); err != nil {
			return nil, fmt.Errorf("failed to parse runtime config: %w", err)
		}

		if err := vmmodules.EnableConfiguredModules(runtime, vm.ExposedModules); err != nil {
			return nil, fmt.Errorf("failed to enable configured modules: %w", err)
		}

		// Set up console if enabled
		if runtimeConfig.Console {
			console := map[string]interface{}{
				"log": func(args ...interface{}) {
					sm.logger.Info().
						Str("session_id", session.ID).
						Interface("args", args).
						Msg("startup console.log")
				},
			}
			runtime.Set("console", console)
		}

		// Load configured libraries into runtime
		if err := sm.loadLibraries(runtime, vm, session.ID); err != nil {
			return nil, fmt.Errorf("failed to load libraries: %w", err)
		}
	}

	// Add to active sessions
	sm.sessionsMu.Lock()
	sm.sessions[sessionID] = session
	sm.sessionsMu.Unlock()

	// Run startup files
	if err := sm.runStartupFiles(session); err != nil {
		session.Status = vmmodels.SessionCrashed
		session.LastError = fmt.Sprintf("startup failed: %v", err)

		dbSession.Status = string(session.Status)
		dbSession.LastError = session.LastError
		sm.store.UpdateSession(dbSession)

		return nil, fmt.Errorf("startup failed: %w", err)
	}

	// Mark session as ready
	session.Status = vmmodels.SessionReady
	dbSession.Status = string(session.Status)
	if err := sm.store.UpdateSession(dbSession); err != nil {
		return nil, fmt.Errorf("failed to update session status: %w", err)
	}

	return session, nil
}

// GetSession retrieves an active session
func (sm *SessionManager) GetSession(sessionID string) (*Session, error) {
	sm.sessionsMu.RLock()
	session, ok := sm.sessions[sessionID]
	sm.sessionsMu.RUnlock()

	if !ok {
		return nil, vmmodels.ErrSessionNotFound
	}

	return session, nil
}

// CloseSession closes a session and releases resources
func (sm *SessionManager) CloseSession(sessionID string) error {
	sm.sessionsMu.Lock()
	_, ok := sm.sessions[sessionID]
	if ok {
		delete(sm.sessions, sessionID)
	}
	sm.sessionsMu.Unlock()

	if !ok {
		return vmmodels.ErrSessionNotFound
	}

	// Update database
	dbSession, err := sm.store.GetSession(sessionID)
	if err != nil {
		return err
	}

	now := time.Now()
	dbSession.Status = string(vmmodels.SessionClosed)
	dbSession.ClosedAt = &now

	return sm.store.UpdateSession(dbSession)
}

// runStartupFiles executes startup files for a session
func (sm *SessionManager) runStartupFiles(session *Session) error {
	root, err := vmpath.NewWorktreeRoot(session.WorktreePath)
	if err != nil {
		return fmt.Errorf("invalid worktree root: %w", err)
	}

	// Get startup files
	startupFiles, err := sm.store.ListStartupFiles(session.VMID)
	if err != nil {
		return err
	}

	// Execute each startup file
	for _, file := range startupFiles {
		relPath, err := vmpath.ParseRelWorktreePath(file.Path)
		if err != nil {
			switch {
			case errors.Is(err, vmpath.ErrAbsoluteRelativePath), errors.Is(err, vmpath.ErrTraversalRelativePath), errors.Is(err, vmpath.ErrEmptyRelativePath):
				return fmt.Errorf("%w: startup file path %q", vmmodels.ErrPathTraversal, file.Path)
			default:
				return fmt.Errorf("invalid startup file path %q: %w", file.Path, err)
			}
		}

		resolvedPath, err := root.Resolve(relPath)
		if err != nil {
			if errors.Is(err, vmpath.ErrPathEscapesRoot) {
				return fmt.Errorf("%w: startup file path %q", vmmodels.ErrPathTraversal, file.Path)
			}
			return fmt.Errorf("resolve startup path %q: %w", file.Path, err)
		}
		filePath := resolvedPath.Absolute()

		// Check if file exists
		if _, err := os.Stat(filePath); err != nil {
			return fmt.Errorf("startup file not found: %s: %w", file.Path, err)
		}

		// Read file content
		content, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read startup file %s: %w", file.Path, err)
		}

		switch file.Mode {
		case "", "eval":
			if _, err := session.Runtime.RunString(string(content)); err != nil {
				return fmt.Errorf("failed to execute startup file %s: %w", file.Path, err)
			}
		default:
			return fmt.Errorf("%w: %s", vmmodels.ErrStartupModeUnsupported, file.Mode)
		}
	}

	return nil
}

// ListSessions lists all active sessions
func (sm *SessionManager) ListSessions() []*Session {
	sm.sessionsMu.RLock()
	defer sm.sessionsMu.RUnlock()

	sessions := make([]*Session, 0, len(sm.sessions))
	for _, session := range sm.sessions {
		sessions = append(sessions, session)
	}

	return sessions
}

// loadLibraries loads configured JavaScript libraries into the goja runtime
func (sm *SessionManager) loadLibraries(runtime *goja.Runtime, vm *vmmodels.VM, sessionID string) error {
	if len(vm.Libraries) == 0 {
		return nil // No libraries to load
	}

	// Get library cache directory
	cacheDir := filepath.Join(".vm-cache", "libraries")

	// Load each configured library
	for _, libName := range vm.Libraries {
		libPath := filepath.Join(cacheDir, libName+".js")

		// Check if library file exists
		if _, err := os.Stat(libPath); err != nil {
			return fmt.Errorf("library %s not found in cache (run 'vm-system libs download' first): %w", libName, err)
		}

		// Read library content
		content, err := os.ReadFile(libPath)
		if err != nil {
			return fmt.Errorf("failed to read library %s: %w", libName, err)
		}

		// Execute library code in runtime
		if _, err := runtime.RunString(string(content)); err != nil {
			return fmt.Errorf("failed to load library %s: %w", libName, err)
		}

		sm.logger.Info().
			Str("session_id", sessionID).
			Str("template_id", vm.ID).
			Str("library", libName).
			Msg("loaded library into runtime session")
	}

	return nil
}
