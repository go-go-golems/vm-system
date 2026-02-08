package vmcontrol

// CreateTemplateInput is the public input model for template creation.
type CreateTemplateInput struct {
	Name   string
	Engine string
}

// CreateSessionInput is the public input model for session creation.
type CreateSessionInput struct {
	TemplateID    string
	WorkspaceID   string
	BaseCommitOID string
	WorktreePath  string
}

// ExecuteREPLInput is the public input model for REPL execution.
type ExecuteREPLInput struct {
	SessionID string
	Input     string
}

// ExecuteRunFileInput is the public input model for file execution.
type ExecuteRunFileInput struct {
	SessionID string
	Path      string
	Args      map[string]interface{}
	Env       map[string]interface{}
}

// RuntimeSummary captures currently active runtime state in daemon memory.
type RuntimeSummary struct {
	ActiveSessions  int      `json:"active_sessions"`
	ActiveSessionID []string `json:"active_session_ids"`
}
