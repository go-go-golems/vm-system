package vmcontrol

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"

	"github.com/go-go-golems/vm-system/pkg/vmmodels"
)

// ExecutionService owns REPL and run-file orchestration.
type ExecutionService struct {
	runtime       ExecutionRuntimePort
	sessionStore  SessionStorePort
	templateStore TemplateStorePort
}

func NewExecutionService(runtime ExecutionRuntimePort, sessionStore SessionStorePort, templateStore TemplateStorePort) *ExecutionService {
	return &ExecutionService{
		runtime:       runtime,
		sessionStore:  sessionStore,
		templateStore: templateStore,
	}
}

func (s *ExecutionService) ExecuteREPL(_ context.Context, input ExecuteREPLInput) (*vmmodels.Execution, error) {
	execution, err := s.runtime.ExecuteREPL(input.SessionID, input.Input)
	if err != nil {
		return nil, err
	}
	if err := s.enforceLimits(input.SessionID, execution.ID); err != nil {
		return nil, err
	}
	return execution, nil
}

func (s *ExecutionService) ExecuteRunFile(_ context.Context, input ExecuteRunFileInput) (*vmmodels.Execution, error) {
	session, err := s.sessionStore.GetSession(input.SessionID)
	if err != nil {
		return nil, err
	}
	safePath, err := normalizeRunFilePath(session.WorktreePath, input.Path)
	if err != nil {
		return nil, err
	}

	execution, err := s.runtime.ExecuteRunFile(input.SessionID, safePath, input.Args, input.Env)
	if err != nil {
		return nil, err
	}
	if err := s.enforceLimits(input.SessionID, execution.ID); err != nil {
		return nil, err
	}
	return execution, nil
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

func normalizeRunFilePath(worktreePath, requestedPath string) (string, error) {
	if filepath.IsAbs(requestedPath) {
		return "", vmmodels.ErrPathTraversal
	}

	cleanRelative := filepath.Clean(requestedPath)
	if cleanRelative == "." {
		return "", vmmodels.ErrFileNotFound
	}
	if strings.HasPrefix(cleanRelative, "..") {
		return "", vmmodels.ErrPathTraversal
	}

	fullPath := filepath.Join(worktreePath, cleanRelative)
	relativeToRoot, err := filepath.Rel(worktreePath, fullPath)
	if err != nil {
		return "", vmmodels.ErrPathTraversal
	}
	if strings.HasPrefix(relativeToRoot, "..") {
		return "", vmmodels.ErrPathTraversal
	}

	return relativeToRoot, nil
}

func (s *ExecutionService) enforceLimits(sessionID, executionID string) error {
	limits, err := s.loadSessionLimits(sessionID)
	if err != nil {
		// Scaffolding is intentionally soft-fail while limit enforcement matures.
		return nil
	}

	events, err := s.runtime.GetEvents(executionID, 0)
	if err != nil {
		return nil
	}

	if limits.MaxEvents > 0 && len(events) > limits.MaxEvents {
		return vmmodels.ErrOutputLimitExceeded
	}

	if limits.MaxOutputKB > 0 {
		var payloadBytes int
		for _, event := range events {
			payloadBytes += len(event.Payload)
		}
		if payloadBytes > limits.MaxOutputKB*1024 {
			return vmmodels.ErrOutputLimitExceeded
		}
	}

	return nil
}

func (s *ExecutionService) loadSessionLimits(sessionID string) (*LimitsConfig, error) {
	session, err := s.sessionStore.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	settings, err := s.templateStore.GetVMSettings(session.VMID)
	if err != nil {
		return nil, err
	}

	limits := &LimitsConfig{}
	if err := json.Unmarshal(settings.Limits, limits); err != nil {
		return nil, err
	}
	return limits, nil
}
