package vmclient

import (
	"context"
	"fmt"

	"github.com/go-go-golems/vm-system/pkg/vmmodels"
)

type ExecuteREPLRequest struct {
	SessionID string `json:"session_id"`
	Input     string `json:"input"`
}

type ExecuteRunFileRequest struct {
	SessionID string                 `json:"session_id"`
	Path      string                 `json:"path"`
	Args      map[string]interface{} `json:"args"`
	Env       map[string]interface{} `json:"env"`
}

func (c *Client) ExecuteREPL(ctx context.Context, request ExecuteREPLRequest) (*vmmodels.Execution, error) {
	var execution vmmodels.Execution
	if err := c.do(ctx, "POST", "/api/v1/executions/repl", request, &execution); err != nil {
		return nil, err
	}
	return &execution, nil
}

func (c *Client) ExecuteRunFile(ctx context.Context, request ExecuteRunFileRequest) (*vmmodels.Execution, error) {
	var execution vmmodels.Execution
	if err := c.do(ctx, "POST", "/api/v1/executions/run-file", request, &execution); err != nil {
		return nil, err
	}
	return &execution, nil
}

func (c *Client) ListExecutions(ctx context.Context, sessionID string, limit int) ([]*vmmodels.Execution, error) {
	path := withQuery("/api/v1/executions", map[string]string{
		"session_id": sessionID,
		"limit":      fmt.Sprintf("%d", limit),
	})

	var executions []*vmmodels.Execution
	if err := c.do(ctx, "GET", path, nil, &executions); err != nil {
		return nil, err
	}
	return executions, nil
}

func (c *Client) GetExecution(ctx context.Context, executionID string) (*vmmodels.Execution, error) {
	var execution vmmodels.Execution
	if err := c.do(ctx, "GET", fmt.Sprintf("/api/v1/executions/%s", executionID), nil, &execution); err != nil {
		return nil, err
	}
	return &execution, nil
}

func (c *Client) GetExecutionEvents(ctx context.Context, executionID string, afterSeq int) ([]*vmmodels.ExecutionEvent, error) {
	path := withQuery(fmt.Sprintf("/api/v1/executions/%s/events", executionID), map[string]string{
		"after_seq": fmt.Sprintf("%d", afterSeq),
	})
	var events []*vmmodels.ExecutionEvent
	if err := c.do(ctx, "GET", path, nil, &events); err != nil {
		return nil, err
	}
	return events, nil
}
