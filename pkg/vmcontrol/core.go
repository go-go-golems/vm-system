package vmcontrol

import (
	"github.com/go-go-golems/vm-system/pkg/vmexec"
	"github.com/go-go-golems/vm-system/pkg/vmsession"
	"github.com/go-go-golems/vm-system/pkg/vmstore"
)

// Core is the transport-agnostic orchestration entrypoint used by daemon, CLI clients, and in-process consumers.
type Core struct {
	Templates  *TemplateService
	Sessions   *SessionService
	Executions *ExecutionService
	Registry   *RuntimeRegistry
}

// NewCore builds the standard core wiring from concrete store + runtime implementations.
func NewCore(store *vmstore.VMStore) *Core {
	sessionRuntime := vmsession.NewSessionManager(store)
	executionRuntime := vmexec.NewExecutor(store, sessionRuntime)
	return NewCoreWithPorts(store, sessionRuntime, executionRuntime)
}

// NewCoreWithPorts allows tests and non-daemon embeddings to provide custom adapters.
func NewCoreWithPorts(store StorePort, sessionRuntime SessionRuntimePort, executionRuntime ExecutionRuntimePort) *Core {
	return &Core{
		Templates:  NewTemplateService(store),
		Sessions:   NewSessionService(store, sessionRuntime),
		Executions: NewExecutionService(executionRuntime, store, store),
		Registry:   NewRuntimeRegistry(sessionRuntime),
	}
}
