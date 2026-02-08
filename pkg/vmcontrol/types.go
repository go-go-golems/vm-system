package vmcontrol

import "encoding/json"

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

// LimitsConfig mirrors public JSON payload shape for template settings.
type LimitsConfig struct {
	CPUMs       int `json:"cpu_ms"`
	WallMs      int `json:"wall_ms"`
	MemMB       int `json:"mem_mb"`
	MaxEvents   int `json:"max_events"`
	MaxOutputKB int `json:"max_output_kb"`
}

// ResolverConfig mirrors public JSON payload shape for template settings.
type ResolverConfig struct {
	Roots                    []string `json:"roots"`
	Extensions               []string `json:"extensions"`
	AllowAbsoluteRepoImports bool     `json:"allow_absolute_repo_imports"`
}

// RuntimeConfig mirrors public JSON payload shape for template settings.
type RuntimeConfig struct {
	ESM     bool `json:"esm"`
	Strict  bool `json:"strict"`
	Console bool `json:"console"`
}

func mustMarshalJSON(v interface{}, fallback string) json.RawMessage {
	data, err := json.Marshal(v)
	if err != nil {
		return json.RawMessage(fallback)
	}
	return data
}
