package vmmodels

import (
	"encoding/json"
	"errors"
	"time"
)

// Common errors
var (
	ErrVMNotFound             = errors.New("VM not found")
	ErrSessionNotFound        = errors.New("session not found")
	ErrExecutionNotFound      = errors.New("execution not found")
	ErrSessionNotReady        = errors.New("session not ready")
	ErrSessionBusy            = errors.New("session busy")
	ErrPathTraversal          = errors.New("path traversal is not allowed")
	ErrStartupModeUnsupported = errors.New("startup mode is not supported")
	ErrModuleNotAllowed       = errors.New("module not allowed")
	ErrFileNotFound           = errors.New("file not found")
	ErrImportResolutionFailed = errors.New("import resolution failed")
	ErrStartupFailed          = errors.New("startup failed")
	ErrExecTimeout            = errors.New("execution timeout")
	ErrOutputLimitExceeded    = errors.New("output limit exceeded")
	ErrInternalVMError        = errors.New("internal VM error")
)

// VM represents a VM profile (configuration template)
type VM struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Engine         string    `json:"engine"` // quickjs, goja, node, custom
	IsActive       bool      `json:"is_active"`
	ExposedModules []string  `json:"exposed_modules"` // IDs of exposed modules
	Libraries      []string  `json:"libraries"`       // IDs of loaded libraries
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// VMSettings contains VM configuration settings
type VMSettings struct {
	VMID     string          `json:"vm_id"`
	Limits   json.RawMessage `json:"limits"`   // LimitsConfig as JSON
	Resolver json.RawMessage `json:"resolver"` // ResolverConfig as JSON
	Runtime  json.RawMessage `json:"runtime"`  // RuntimeConfig as JSON
}

// LimitsConfig defines resource limits

type LimitsConfig struct {
	CPUMs       int `json:"cpu_ms"`
	WallMs      int `json:"wall_ms"`
	MemMB       int `json:"mem_mb"`
	MaxEvents   int `json:"max_events"`
	MaxOutputKB int `json:"max_output_kb"`
}

// ResolverConfig defines module resolution settings
type ResolverConfig struct {
	Roots                    []string `json:"roots"`
	Extensions               []string `json:"extensions"`
	AllowAbsoluteRepoImports bool     `json:"allow_absolute_repo_imports"`
}

// RuntimeConfig defines runtime settings
type RuntimeConfig struct {
	ESM     bool `json:"esm"`
	Strict  bool `json:"strict"`
	Console bool `json:"console"`
}

// VMCapability represents a module or global exposure
type VMCapability struct {
	ID      string          `json:"id"`
	VMID    string          `json:"vm_id"`
	Kind    string          `json:"kind"` // module, global, fs, net, env
	Name    string          `json:"name"` // import specifier or global name
	Enabled bool            `json:"enabled"`
	Config  json.RawMessage `json:"config"` // capability-specific config
}

// VMStartupFile represents a file to run at session startup
type VMStartupFile struct {
	ID         string `json:"id"`
	VMID       string `json:"vm_id"`
	Path       string `json:"path"` // repo path
	OrderIndex int    `json:"order_index"`
	Mode       string `json:"mode"` // eval (import is currently unsupported)
}

// VMSession represents a VM runtime instance
type VMSession struct {
	ID            string          `json:"id"`
	VMID          string          `json:"vm_id"`
	WorkspaceID   string          `json:"workspace_id"`
	BaseCommitOID string          `json:"base_commit_oid"`
	WorktreePath  string          `json:"worktree_path"`
	Status        string          `json:"status"` // starting, ready, crashed, closed
	CreatedAt     time.Time       `json:"created_at"`
	ClosedAt      *time.Time      `json:"closed_at,omitempty"`
	LastError     string          `json:"last_error,omitempty"`
	RuntimeMeta   json.RawMessage `json:"runtime_meta,omitempty"`
}

// SessionStatus represents session states
type SessionStatus string

const (
	SessionStarting SessionStatus = "starting"
	SessionReady    SessionStatus = "ready"
	SessionCrashed  SessionStatus = "crashed"
	SessionClosed   SessionStatus = "closed"
)

// Execution represents a discrete code execution
type Execution struct {
	ID        string          `json:"id"`
	SessionID string          `json:"session_id"`
	Kind      string          `json:"kind"`            // startup, run_file, repl
	Input     string          `json:"input,omitempty"` // snippet for repl
	Path      string          `json:"path,omitempty"`  // entry path for run_file/startup
	Args      json.RawMessage `json:"args"`
	Env       json.RawMessage `json:"env"`
	Status    string          `json:"status"` // running, ok, error, timeout, cancelled
	StartedAt time.Time       `json:"started_at"`
	EndedAt   *time.Time      `json:"ended_at,omitempty"`
	Result    json.RawMessage `json:"result,omitempty"`
	Error     json.RawMessage `json:"error,omitempty"`
	Metrics   json.RawMessage `json:"metrics"`
}

// ExecutionKind represents execution types
type ExecutionKind string

const (
	ExecStartup ExecutionKind = "startup"
	ExecRunFile ExecutionKind = "run_file"
	ExecREPL    ExecutionKind = "repl"
)

// ExecutionStatus represents execution states
type ExecutionStatus string

const (
	ExecRunning   ExecutionStatus = "running"
	ExecOK        ExecutionStatus = "ok"
	ExecError     ExecutionStatus = "error"
	ExecTimeout   ExecutionStatus = "timeout"
	ExecCancelled ExecutionStatus = "cancelled"
)

// ExecutionEvent represents an event during execution
type ExecutionEvent struct {
	ExecutionID string          `json:"execution_id"`
	Seq         int             `json:"seq"`
	Ts          time.Time       `json:"ts"`
	Type        string          `json:"type"` // stdout, stderr, console, value, exception, system, input_echo
	Payload     json.RawMessage `json:"payload"`
}

// EventType represents event types
type EventType string

const (
	EventStdout    EventType = "stdout"
	EventStderr    EventType = "stderr"
	EventConsole   EventType = "console"
	EventValue     EventType = "value"
	EventException EventType = "exception"
	EventSystem    EventType = "system"
	EventInputEcho EventType = "input_echo"
)

// ConsolePayload represents console event payload
type ConsolePayload struct {
	Level string `json:"level"` // log, warn, error, info, debug
	Text  string `json:"text"`
}

// ValuePayload represents value event payload
type ValuePayload struct {
	Type    string          `json:"type"`
	Preview string          `json:"preview"`
	JSON    json.RawMessage `json:"json,omitempty"`
}

// ExceptionPayload represents exception event payload
type ExceptionPayload struct {
	Message string `json:"message"`
	Stack   string `json:"stack,omitempty"`
}

// SystemPayload represents system event payload
type SystemPayload struct {
	Message string `json:"message"`
	Level   string `json:"level"` // info, warn, error
}
