package vmcontrol

import (
	"github.com/go-go-golems/vm-system/pkg/vmexec"
	"github.com/go-go-golems/vm-system/pkg/vmmodels"
	"github.com/go-go-golems/vm-system/pkg/vmsession"
)

// TemplateStorePort defines persistent template operations used by the core.
type TemplateStorePort interface {
	CreateVM(vm *vmmodels.VM) error
	GetVM(id string) (*vmmodels.VM, error)
	ListVMs() ([]*vmmodels.VM, error)
	DeleteVM(id string) error
	SetVMSettings(settings *vmmodels.VMSettings) error
	GetVMSettings(vmID string) (*vmmodels.VMSettings, error)
	AddCapability(cap *vmmodels.VMCapability) error
	ListCapabilities(vmID string) ([]*vmmodels.VMCapability, error)
	AddStartupFile(file *vmmodels.VMStartupFile) error
	ListStartupFiles(vmID string) ([]*vmmodels.VMStartupFile, error)
}

// SessionStorePort defines persistent session operations used by the core.
type SessionStorePort interface {
	GetSession(id string) (*vmmodels.VMSession, error)
	ListSessions(status string) ([]*vmmodels.VMSession, error)
}

// StorePort combines template and session storage capabilities.
type StorePort interface {
	TemplateStorePort
	SessionStorePort
}

// SessionRuntimePort defines runtime session orchestration operations.
type SessionRuntimePort interface {
	CreateSession(vmID, workspaceID, baseCommitOID, worktreePath string) (*vmsession.Session, error)
	GetSession(sessionID string) (*vmsession.Session, error)
	CloseSession(sessionID string) error
	ListSessions() []*vmsession.Session
}

// ExecutionRuntimePort defines runtime execution orchestration operations.
type ExecutionRuntimePort interface {
	ExecuteREPL(sessionID, input string) (*vmmodels.Execution, error)
	ExecuteRunFile(sessionID, path string, args, env map[string]interface{}) (*vmmodels.Execution, error)
	ListExecutions(sessionID string, limit int) ([]*vmmodels.Execution, error)
	GetExecution(executionID string) (*vmmodels.Execution, error)
	GetEvents(executionID string, afterSeq int) ([]*vmmodels.ExecutionEvent, error)
}

var (
	_ SessionRuntimePort   = (*vmsession.SessionManager)(nil)
	_ ExecutionRuntimePort = (*vmexec.Executor)(nil)
)
