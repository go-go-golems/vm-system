package vmcontrol

import (
	"context"
	"fmt"

	"github.com/go-go-golems/vm-system/pkg/vmmodels"
)

// SessionService owns lifecycle operations for runtime sessions.
type SessionService struct {
	store   SessionStorePort
	runtime SessionRuntimePort
}

func NewSessionService(store SessionStorePort, runtime SessionRuntimePort) *SessionService {
	return &SessionService{
		store:   store,
		runtime: runtime,
	}
}

func (s *SessionService) Create(_ context.Context, input CreateSessionInput) (*vmmodels.VMSession, error) {
	session, err := s.runtime.CreateSession(
		input.TemplateID,
		input.WorkspaceID,
		input.BaseCommitOID,
		input.WorktreePath,
	)
	if err != nil {
		return nil, err
	}

	out, err := s.store.GetSession(session.ID)
	if err != nil {
		return nil, fmt.Errorf("session created but could not be loaded from store: %w", err)
	}
	return out, nil
}

func (s *SessionService) Get(_ context.Context, sessionID string) (*vmmodels.VMSession, error) {
	return s.store.GetSession(sessionID)
}

func (s *SessionService) List(_ context.Context, status string) ([]*vmmodels.VMSession, error) {
	return s.store.ListSessions(status)
}

func (s *SessionService) Close(_ context.Context, sessionID string) (*vmmodels.VMSession, error) {
	if err := s.runtime.CloseSession(sessionID); err != nil {
		return nil, err
	}
	return s.store.GetSession(sessionID)
}
