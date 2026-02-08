package vmclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	stdhttp "net/http"
	"net/url"
	"strings"
	"time"
)

// Client wraps vm-system daemon REST calls for CLI and external consumers.
type Client struct {
	baseURL    string
	httpClient *stdhttp.Client
}

func New(baseURL string, httpClient *stdhttp.Client) *Client {
	if httpClient == nil {
		httpClient = &stdhttp.Client{Timeout: 30 * time.Second}
	}
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: httpClient,
	}
}

type APIError struct {
	StatusCode int
	Code       string
	Message    string
	Details    interface{}
}

func (e *APIError) Error() string {
	if e.Code == "" && e.Message == "" {
		return fmt.Sprintf("api error (status=%d)", e.StatusCode)
	}
	return fmt.Sprintf("api error (status=%d, code=%s): %s", e.StatusCode, e.Code, e.Message)
}

type errorEnvelope struct {
	Error struct {
		Code    string      `json:"code"`
		Message string      `json:"message"`
		Details interface{} `json:"details,omitempty"`
	} `json:"error"`
}

func (c *Client) do(ctx context.Context, method, path string, in interface{}, out interface{}) error {
	var body io.Reader
	if in != nil {
		data, err := json.Marshal(in)
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}
		body = bytes.NewReader(data)
	}

	requestURL := c.baseURL + path
	req, err := stdhttp.NewRequestWithContext(ctx, method, requestURL, body)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	if in != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		apiErr := &APIError{StatusCode: resp.StatusCode}
		var env errorEnvelope
		if err := json.NewDecoder(resp.Body).Decode(&env); err == nil {
			apiErr.Code = env.Error.Code
			apiErr.Message = env.Error.Message
			apiErr.Details = env.Error.Details
		}
		return apiErr
	}

	if out == nil {
		return nil
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}

func withQuery(path string, values map[string]string) string {
	q := url.Values{}
	for k, v := range values {
		if v != "" {
			q.Set(k, v)
		}
	}
	if len(q) == 0 {
		return path
	}
	return path + "?" + q.Encode()
}
