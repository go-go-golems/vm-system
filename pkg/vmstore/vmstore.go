package vmstore

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/go-go-golems/vm-system/pkg/vmmodels"
)

// VMStore manages VM-related data in SQLite
type VMStore struct {
	db *sql.DB
}

// mustMarshalJSON marshals a value to JSON, returning empty array on error
func mustMarshalJSON(v interface{}) string {
	data, err := json.Marshal(v)
	if err != nil {
		return "[]"
	}
	return string(data)
}

// NewVMStore creates a new VMStore
func NewVMStore(dbPath string) (*VMStore, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	store := &VMStore{db: db}

	// Initialize schema
	if err := store.initSchema(); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return store, nil
}

// Close closes the database connection
func (s *VMStore) Close() error {
	return s.db.Close()
}

// initSchema creates the database schema
func (s *VMStore) initSchema() error {
	schema := `
	-- VM profiles
	CREATE TABLE IF NOT EXISTS vm (
		id TEXT PRIMARY KEY,
		name TEXT UNIQUE NOT NULL,
		engine TEXT NOT NULL,
		is_active INTEGER NOT NULL DEFAULT 1,
		exposed_modules_json TEXT NOT NULL DEFAULT '[]',
		libraries_json TEXT NOT NULL DEFAULT '[]',
		created_at INTEGER NOT NULL,
		updated_at INTEGER NOT NULL
	);

	-- VM settings
	CREATE TABLE IF NOT EXISTS vm_settings (
		vm_id TEXT PRIMARY KEY REFERENCES vm(id) ON DELETE CASCADE,
		limits_json TEXT NOT NULL,
		resolver_json TEXT NOT NULL,
		runtime_json TEXT NOT NULL
	);

	-- VM capabilities (module exposure allowlist)
	CREATE TABLE IF NOT EXISTS vm_capability (
		id TEXT PRIMARY KEY,
		vm_id TEXT NOT NULL REFERENCES vm(id) ON DELETE CASCADE,
		kind TEXT NOT NULL,
		name TEXT NOT NULL,
		enabled INTEGER NOT NULL DEFAULT 1,
		config_json TEXT NOT NULL DEFAULT '{}',
		UNIQUE(vm_id, kind, name)
	);

	-- VM startup files
	CREATE TABLE IF NOT EXISTS vm_startup_file (
		id TEXT PRIMARY KEY,
		vm_id TEXT NOT NULL REFERENCES vm(id) ON DELETE CASCADE,
		path TEXT NOT NULL,
		order_index INTEGER NOT NULL,
		mode TEXT NOT NULL,
		UNIQUE(vm_id, path)
	);

	CREATE INDEX IF NOT EXISTS idx_startup_file_order ON vm_startup_file(vm_id, order_index);

	-- VM sessions
	CREATE TABLE IF NOT EXISTS vm_session (
		id TEXT PRIMARY KEY,
		vm_id TEXT NOT NULL REFERENCES vm(id),
		workspace_id TEXT NOT NULL,
		base_commit_oid TEXT NOT NULL,
		worktree_path TEXT NOT NULL,
		status TEXT NOT NULL,
		created_at INTEGER NOT NULL,
		closed_at INTEGER,
		last_error_json TEXT,
		runtime_meta_json TEXT
	);

	-- Executions
	CREATE TABLE IF NOT EXISTS execution (
		id TEXT PRIMARY KEY,
		session_id TEXT NOT NULL REFERENCES vm_session(id) ON DELETE CASCADE,
		kind TEXT NOT NULL,
		input TEXT,
		path TEXT,
		args_json TEXT NOT NULL DEFAULT '[]',
		env_json TEXT NOT NULL DEFAULT '{}',
		status TEXT NOT NULL,
		started_at INTEGER NOT NULL,
		ended_at INTEGER,
		result_json TEXT,
		error_json TEXT,
		metrics_json TEXT NOT NULL DEFAULT '{}'
	);

	CREATE INDEX IF NOT EXISTS idx_execution_session ON execution(session_id, started_at);

	-- Execution events
	CREATE TABLE IF NOT EXISTS execution_event (
		execution_id TEXT NOT NULL REFERENCES execution(id) ON DELETE CASCADE,
		seq INTEGER NOT NULL,
		ts INTEGER NOT NULL,
		type TEXT NOT NULL,
		payload_json TEXT NOT NULL,
		PRIMARY KEY(execution_id, seq)
	);
	`

	_, err := s.db.Exec(schema)
	return err
}

// CreateVM creates a new VM profile
func (s *VMStore) CreateVM(vm *vmmodels.VM) error {
	_, err := s.db.Exec(`
		INSERT INTO vm (id, name, engine, is_active, exposed_modules_json, libraries_json, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, vm.ID, vm.Name, vm.Engine, vm.IsActive, mustMarshalJSON(vm.ExposedModules), mustMarshalJSON(vm.Libraries), vm.CreatedAt.Unix(), vm.UpdatedAt.Unix())
	return err
}

// GetVM retrieves a VM by ID
func (s *VMStore) GetVM(id string) (*vmmodels.VM, error) {
	var vm vmmodels.VM
	var createdAt, updatedAt int64
	var exposedModulesJSON, librariesJSON string

	err := s.db.QueryRow(`
		SELECT id, name, engine, is_active, exposed_modules_json, libraries_json, created_at, updated_at
		FROM vm WHERE id = ?
	`, id).Scan(&vm.ID, &vm.Name, &vm.Engine, &vm.IsActive, &exposedModulesJSON, &librariesJSON, &createdAt, &updatedAt)

	if err == sql.ErrNoRows {
		return nil, vmmodels.ErrVMNotFound
	}
	if err != nil {
		return nil, err
	}

	vm.CreatedAt = time.Unix(createdAt, 0)
	vm.UpdatedAt = time.Unix(updatedAt, 0)

	// Unmarshal JSON arrays
	json.Unmarshal([]byte(exposedModulesJSON), &vm.ExposedModules)
	json.Unmarshal([]byte(librariesJSON), &vm.Libraries)

	return &vm, nil
}

// ListVMs lists all VMs
func (s *VMStore) ListVMs() ([]*vmmodels.VM, error) {
	rows, err := s.db.Query(`
		SELECT id, name, engine, is_active, exposed_modules_json, libraries_json, created_at, updated_at
		FROM vm ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var vms []*vmmodels.VM
	for rows.Next() {
		var vm vmmodels.VM
		var createdAt, updatedAt int64

		var exposedModulesJSON, librariesJSON string
		if err := rows.Scan(&vm.ID, &vm.Name, &vm.Engine, &vm.IsActive, &exposedModulesJSON, &librariesJSON, &createdAt, &updatedAt); err != nil {
			return nil, err
		}

		vm.CreatedAt = time.Unix(createdAt, 0)
		vm.UpdatedAt = time.Unix(updatedAt, 0)
		json.Unmarshal([]byte(exposedModulesJSON), &vm.ExposedModules)
		json.Unmarshal([]byte(librariesJSON), &vm.Libraries)
		vms = append(vms, &vm)
	}

	return vms, rows.Err()
}

// UpdateVM updates a VM profile
func (s *VMStore) UpdateVM(vm *vmmodels.VM) error {
	vm.UpdatedAt = time.Now()
	_, err := s.db.Exec(`
		UPDATE vm SET name = ?, engine = ?, is_active = ?, exposed_modules_json = ?, libraries_json = ?, updated_at = ?
		WHERE id = ?
	`, vm.Name, vm.Engine, vm.IsActive, mustMarshalJSON(vm.ExposedModules), mustMarshalJSON(vm.Libraries), vm.UpdatedAt.Unix(), vm.ID)
	return err
}

// DeleteVM deletes a VM profile
func (s *VMStore) DeleteVM(id string) error {
	_, err := s.db.Exec("DELETE FROM vm WHERE id = ?", id)
	return err
}

// SetVMSettings sets VM settings
func (s *VMStore) SetVMSettings(settings *vmmodels.VMSettings) error {
	_, err := s.db.Exec(`
		INSERT OR REPLACE INTO vm_settings (vm_id, limits_json, resolver_json, runtime_json)
		VALUES (?, ?, ?, ?)
	`, settings.VMID, settings.Limits, settings.Resolver, settings.Runtime)
	return err
}

// GetVMSettings retrieves VM settings
func (s *VMStore) GetVMSettings(vmID string) (*vmmodels.VMSettings, error) {
	var settings vmmodels.VMSettings
	err := s.db.QueryRow(`
		SELECT vm_id, limits_json, resolver_json, runtime_json
		FROM vm_settings WHERE vm_id = ?
	`, vmID).Scan(&settings.VMID, &settings.Limits, &settings.Resolver, &settings.Runtime)

	if err == sql.ErrNoRows {
		return nil, vmmodels.ErrVMNotFound
	}
	return &settings, err
}

// AddCapability adds a capability to a VM
func (s *VMStore) AddCapability(cap *vmmodels.VMCapability) error {
	_, err := s.db.Exec(`
		INSERT INTO vm_capability (id, vm_id, kind, name, enabled, config_json)
		VALUES (?, ?, ?, ?, ?, ?)
	`, cap.ID, cap.VMID, cap.Kind, cap.Name, cap.Enabled, cap.Config)
	return err
}

// ListCapabilities lists all capabilities for a VM
func (s *VMStore) ListCapabilities(vmID string) ([]*vmmodels.VMCapability, error) {
	rows, err := s.db.Query(`
		SELECT id, vm_id, kind, name, enabled, config_json
		FROM vm_capability WHERE vm_id = ?
	`, vmID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var caps []*vmmodels.VMCapability
	for rows.Next() {
		var cap vmmodels.VMCapability
		if err := rows.Scan(&cap.ID, &cap.VMID, &cap.Kind, &cap.Name, &cap.Enabled, &cap.Config); err != nil {
			return nil, err
		}
		caps = append(caps, &cap)
	}

	return caps, rows.Err()
}

// GetCapability retrieves a specific capability
func (s *VMStore) GetCapability(vmID, kind, name string) (*vmmodels.VMCapability, error) {
	var cap vmmodels.VMCapability
	err := s.db.QueryRow(`
		SELECT id, vm_id, kind, name, enabled, config_json
		FROM vm_capability WHERE vm_id = ? AND kind = ? AND name = ?
	`, vmID, kind, name).Scan(&cap.ID, &cap.VMID, &cap.Kind, &cap.Name, &cap.Enabled, &cap.Config)

	if err == sql.ErrNoRows {
		return nil, vmmodels.ErrModuleNotAllowed
	}
	return &cap, err
}

// DeleteCapability deletes a capability
func (s *VMStore) DeleteCapability(id string) error {
	_, err := s.db.Exec("DELETE FROM vm_capability WHERE id = ?", id)
	return err
}

// AddStartupFile adds a startup file to a VM
func (s *VMStore) AddStartupFile(file *vmmodels.VMStartupFile) error {
	_, err := s.db.Exec(`
		INSERT INTO vm_startup_file (id, vm_id, path, order_index, mode)
		VALUES (?, ?, ?, ?, ?)
	`, file.ID, file.VMID, file.Path, file.OrderIndex, file.Mode)
	return err
}

// ListStartupFiles lists all startup files for a VM
func (s *VMStore) ListStartupFiles(vmID string) ([]*vmmodels.VMStartupFile, error) {
	rows, err := s.db.Query(`
		SELECT id, vm_id, path, order_index, mode
		FROM vm_startup_file WHERE vm_id = ? ORDER BY order_index
	`, vmID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []*vmmodels.VMStartupFile
	for rows.Next() {
		var file vmmodels.VMStartupFile
		if err := rows.Scan(&file.ID, &file.VMID, &file.Path, &file.OrderIndex, &file.Mode); err != nil {
			return nil, err
		}
		files = append(files, &file)
	}

	return files, rows.Err()
}

// DeleteStartupFile deletes a startup file
func (s *VMStore) DeleteStartupFile(id string) error {
	_, err := s.db.Exec("DELETE FROM vm_startup_file WHERE id = ?", id)
	return err
}

// CreateSession creates a new VM session
func (s *VMStore) CreateSession(session *vmmodels.VMSession) error {
	_, err := s.db.Exec(`
		INSERT INTO vm_session (id, vm_id, workspace_id, base_commit_oid, worktree_path, status, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, session.ID, session.VMID, session.WorkspaceID, session.BaseCommitOID, session.WorktreePath, session.Status, session.CreatedAt.Unix())
	return err
}

// GetSession retrieves a session by ID
func (s *VMStore) GetSession(id string) (*vmmodels.VMSession, error) {
	var session vmmodels.VMSession
	var createdAt int64
	var closedAt sql.NullInt64
	var lastError sql.NullString
	var runtimeMeta sql.NullString

	err := s.db.QueryRow(`
		SELECT id, vm_id, workspace_id, base_commit_oid, worktree_path, status, created_at, closed_at, last_error_json, runtime_meta_json
		FROM vm_session WHERE id = ?
	`, id).Scan(&session.ID, &session.VMID, &session.WorkspaceID, &session.BaseCommitOID, &session.WorktreePath, &session.Status, &createdAt, &closedAt, &lastError, &runtimeMeta)

	if err == sql.ErrNoRows {
		return nil, vmmodels.ErrSessionNotFound
	}
	if err != nil {
		return nil, err
	}

	session.CreatedAt = time.Unix(createdAt, 0)
	if closedAt.Valid {
		t := time.Unix(closedAt.Int64, 0)
		session.ClosedAt = &t
	}
	if lastError.Valid {
		session.LastError = lastError.String
	}
	if runtimeMeta.Valid {
		session.RuntimeMeta = json.RawMessage(runtimeMeta.String)
	}

	return &session, nil
}

// UpdateSession updates a session
func (s *VMStore) UpdateSession(session *vmmodels.VMSession) error {
	var closedAt interface{}
	if session.ClosedAt != nil {
		closedAt = session.ClosedAt.Unix()
	}

	_, err := s.db.Exec(`
		UPDATE vm_session SET status = ?, closed_at = ?, last_error_json = ?, runtime_meta_json = ?
		WHERE id = ?
	`, session.Status, closedAt, session.LastError, session.RuntimeMeta, session.ID)
	return err
}

// ListSessions lists sessions with optional status filter
func (s *VMStore) ListSessions(status string) ([]*vmmodels.VMSession, error) {
	query := "SELECT id, vm_id, workspace_id, base_commit_oid, worktree_path, status, created_at, closed_at, last_error_json, runtime_meta_json FROM vm_session"
	args := []interface{}{}

	if status != "" {
		query += " WHERE status = ?"
		args = append(args, status)
	}

	query += " ORDER BY created_at DESC"

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*vmmodels.VMSession
	for rows.Next() {
		var session vmmodels.VMSession
		var createdAt int64
		var closedAt sql.NullInt64
		var lastError sql.NullString
		var runtimeMeta sql.NullString

		if err := rows.Scan(&session.ID, &session.VMID, &session.WorkspaceID, &session.BaseCommitOID, &session.WorktreePath, &session.Status, &createdAt, &closedAt, &lastError, &runtimeMeta); err != nil {
			return nil, err
		}

		session.CreatedAt = time.Unix(createdAt, 0)
		if closedAt.Valid {
			t := time.Unix(closedAt.Int64, 0)
			session.ClosedAt = &t
		}
		if lastError.Valid {
			session.LastError = lastError.String
		}
		if runtimeMeta.Valid {
			session.RuntimeMeta = json.RawMessage(runtimeMeta.String)
		}

		sessions = append(sessions, &session)
	}

	return sessions, rows.Err()
}

// CreateExecution creates a new execution
func (s *VMStore) CreateExecution(exec *vmmodels.Execution) error {
	_, err := s.db.Exec(`
		INSERT INTO execution (id, session_id, kind, input, path, args_json, env_json, status, started_at, metrics_json)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, exec.ID, exec.SessionID, exec.Kind, exec.Input, exec.Path, exec.Args, exec.Env, exec.Status, exec.StartedAt.Unix(), exec.Metrics)
	return err
}

// UpdateExecution updates an execution
func (s *VMStore) UpdateExecution(exec *vmmodels.Execution) error {
	var endedAt interface{}
	if exec.EndedAt != nil {
		endedAt = exec.EndedAt.Unix()
	}

	_, err := s.db.Exec(`
		UPDATE execution SET status = ?, ended_at = ?, result_json = ?, error_json = ?, metrics_json = ?
		WHERE id = ?
	`, exec.Status, endedAt, exec.Result, exec.Error, exec.Metrics, exec.ID)
	return err
}

// GetExecution retrieves an execution by ID
func (s *VMStore) GetExecution(id string) (*vmmodels.Execution, error) {
	var exec vmmodels.Execution
	var startedAt int64
	var endedAt sql.NullInt64
	var input, path sql.NullString
	var result, errorJSON sql.NullString

	err := s.db.QueryRow(`
		SELECT id, session_id, kind, input, path, args_json, env_json, status, started_at, ended_at, result_json, error_json, metrics_json
		FROM execution WHERE id = ?
	`, id).Scan(&exec.ID, &exec.SessionID, &exec.Kind, &input, &path, &exec.Args, &exec.Env, &exec.Status, &startedAt, &endedAt, &result, &errorJSON, &exec.Metrics)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("execution not found")
	}
	if err != nil {
		return nil, err
	}

	exec.StartedAt = time.Unix(startedAt, 0)
	if endedAt.Valid {
		t := time.Unix(endedAt.Int64, 0)
		exec.EndedAt = &t
	}
	if input.Valid {
		exec.Input = input.String
	}
	if path.Valid {
		exec.Path = path.String
	}
	if result.Valid {
		exec.Result = json.RawMessage(result.String)
	}
	if errorJSON.Valid {
		exec.Error = json.RawMessage(errorJSON.String)
	}

	return &exec, nil
}

// ListExecutions lists executions for a session
func (s *VMStore) ListExecutions(sessionID string, limit int) ([]*vmmodels.Execution, error) {
	rows, err := s.db.Query(`
		SELECT id, session_id, kind, input, path, args_json, env_json, status, started_at, ended_at, result_json, error_json, metrics_json
		FROM execution WHERE session_id = ? ORDER BY started_at DESC LIMIT ?
	`, sessionID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var execs []*vmmodels.Execution
	for rows.Next() {
		var exec vmmodels.Execution
		var startedAt int64
		var endedAt sql.NullInt64
		var input, path sql.NullString
		var result, errorJSON sql.NullString

		if err := rows.Scan(&exec.ID, &exec.SessionID, &exec.Kind, &input, &path, &exec.Args, &exec.Env, &exec.Status, &startedAt, &endedAt, &result, &errorJSON, &exec.Metrics); err != nil {
			return nil, err
		}

		exec.StartedAt = time.Unix(startedAt, 0)
		if endedAt.Valid {
			t := time.Unix(endedAt.Int64, 0)
			exec.EndedAt = &t
		}
		if input.Valid {
			exec.Input = input.String
		}
		if path.Valid {
			exec.Path = path.String
		}
		if result.Valid {
			exec.Result = json.RawMessage(result.String)
		}
		if errorJSON.Valid {
			exec.Error = json.RawMessage(errorJSON.String)
		}

		execs = append(execs, &exec)
	}

	return execs, rows.Err()
}

// AddEvent adds an event to an execution
func (s *VMStore) AddEvent(event *vmmodels.ExecutionEvent) error {
	_, err := s.db.Exec(`
		INSERT INTO execution_event (execution_id, seq, ts, type, payload_json)
		VALUES (?, ?, ?, ?, ?)
	`, event.ExecutionID, event.Seq, event.Ts.Unix(), event.Type, event.Payload)
	return err
}

// GetEvents retrieves events for an execution
func (s *VMStore) GetEvents(executionID string, afterSeq int) ([]*vmmodels.ExecutionEvent, error) {
	rows, err := s.db.Query(`
		SELECT execution_id, seq, ts, type, payload_json
		FROM execution_event WHERE execution_id = ? AND seq > ? ORDER BY seq
	`, executionID, afterSeq)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*vmmodels.ExecutionEvent
	for rows.Next() {
		var event vmmodels.ExecutionEvent
		var ts int64

		if err := rows.Scan(&event.ExecutionID, &event.Seq, &ts, &event.Type, &event.Payload); err != nil {
			return nil, err
		}

		event.Ts = time.Unix(ts, 0)
		events = append(events, &event)
	}

	return events, rows.Err()
}
