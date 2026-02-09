package vmstore

// initSchema creates the database schema.
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
