package vmclient

import (
	"context"
	"fmt"

	"github.com/go-go-golems/vm-system/pkg/vmmodels"
)

type CreateSessionRequest struct {
	TemplateID    string `json:"template_id"`
	WorkspaceID   string `json:"workspace_id"`
	BaseCommitOID string `json:"base_commit_oid"`
	WorktreePath  string `json:"worktree_path"`
}

func (c *Client) CreateSession(ctx context.Context, request CreateSessionRequest) (*vmmodels.VMSession, error) {
	var session vmmodels.VMSession
	if err := c.do(ctx, "POST", "/api/v1/sessions", request, &session); err != nil {
		return nil, err
	}
	return &session, nil
}

func (c *Client) ListSessions(ctx context.Context, status string) ([]*vmmodels.VMSession, error) {
	path := withQuery("/api/v1/sessions", map[string]string{"status": status})
	var sessions []*vmmodels.VMSession
	if err := c.do(ctx, "GET", path, nil, &sessions); err != nil {
		return nil, err
	}
	return sessions, nil
}

func (c *Client) GetSession(ctx context.Context, sessionID string) (*vmmodels.VMSession, error) {
	var session vmmodels.VMSession
	if err := c.do(ctx, "GET", fmt.Sprintf("/api/v1/sessions/%s", sessionID), nil, &session); err != nil {
		return nil, err
	}
	return &session, nil
}

func (c *Client) CloseSession(ctx context.Context, sessionID string) (*vmmodels.VMSession, error) {
	var session vmmodels.VMSession
	if err := c.do(ctx, "POST", fmt.Sprintf("/api/v1/sessions/%s/close", sessionID), map[string]string{}, &session); err != nil {
		return nil, err
	}
	return &session, nil
}
