package vmcontrol

import (
	"context"

	"github.com/go-go-golems/vm-system/pkg/vmmodels"
)

// ExecutionService owns REPL and run-file orchestration.
type ExecutionService struct {
	runtime ExecutionRuntimePort
}

func NewExecutionService(runtime ExecutionRuntimePort) *ExecutionService {
	return &ExecutionService{runtime: runtime}
}

func (s *ExecutionService) ExecuteREPL(_ context.Context, input ExecuteREPLInput) (*vmmodels.Execution, error) {
	return s.runtime.ExecuteREPL(input.SessionID, input.Input)
}

func (s *ExecutionService) ExecuteRunFile(_ context.Context, input ExecuteRunFileInput) (*vmmodels.Execution, error) {
	return s.runtime.ExecuteRunFile(input.SessionID, input.Path, input.Args, input.Env)
}

func (s *ExecutionService) Get(_ context.Context, executionID string) (*vmmodels.Execution, error) {
	return s.runtime.GetExecution(executionID)
}

func (s *ExecutionService) List(_ context.Context, sessionID string, limit int) ([]*vmmodels.Execution, error) {
	return s.runtime.ListExecutions(sessionID, limit)
}

func (s *ExecutionService) Events(_ context.Context, executionID string, afterSeq int) ([]*vmmodels.ExecutionEvent, error) {
	return s.runtime.GetEvents(executionID, afterSeq)
}
