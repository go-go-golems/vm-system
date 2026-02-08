package vmhttp_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"testing"
	"time"
)

func TestExecutionContractsStatusEnvelopeAndListGetSemantics(t *testing.T) {
	server, client := newIntegrationTestServer(t)
	defer server.Close()

	worktree := filepath.Join(t.TempDir(), "worktree")
	mustMkdirAll(t, worktree)
	writeFile(t, filepath.Join(worktree, "contract.js"), "console.log('contract-file'); 40 + 2;")

	templateID := createTemplateForTest(t, client, server.URL, "execution-contract-template")
	sessionID := createSessionForTest(t, client, server.URL, templateID, worktree, "ws-execution-contract")

	repl := executionContractResponse{}
	reqJSONStatus(t, client, http.MethodPost, server.URL+"/api/v1/executions/repl", map[string]interface{}{
		"session_id": sessionID,
		"input":      "console.log('repl-contract'); 20 + 22",
	}, http.StatusCreated, &repl)
	assertExecutionEnvelope(t, repl, sessionID, "repl")
	if repl.Input == "" {
		t.Fatalf("expected repl input to be populated in response envelope")
	}
	if len(repl.Result) == 0 {
		t.Fatalf("expected repl result payload in response envelope")
	}

	runFile := executionContractResponse{}
	reqJSONStatus(t, client, http.MethodPost, server.URL+"/api/v1/executions/run-file", map[string]interface{}{
		"session_id": sessionID,
		"path":       "contract.js",
	}, http.StatusCreated, &runFile)
	assertExecutionEnvelope(t, runFile, sessionID, "run_file")
	if runFile.Path != "contract.js" {
		t.Fatalf("expected run-file path contract.js, got %q", runFile.Path)
	}
	if len(runFile.Result) == 0 {
		t.Fatalf("expected run-file result payload in response envelope")
	}

	getRunFile := executionContractResponse{}
	reqJSONStatus(t, client, http.MethodGet, fmt.Sprintf("%s/api/v1/executions/%s", server.URL, runFile.ID), nil, http.StatusOK, &getRunFile)
	if getRunFile.ID != runFile.ID {
		t.Fatalf("expected get execution id %q, got %q", runFile.ID, getRunFile.ID)
	}
	if getRunFile.Kind != "run_file" {
		t.Fatalf("expected get execution kind run_file, got %q", getRunFile.Kind)
	}
	if len(getRunFile.Result) == 0 {
		t.Fatalf("expected run-file result payload in get response")
	}

	list := []executionContractResponse{}
	reqJSONStatus(t, client, http.MethodGet, fmt.Sprintf("%s/api/v1/executions?session_id=%s&limit=2", server.URL, sessionID), nil, http.StatusOK, &list)
	if len(list) != 2 {
		t.Fatalf("expected exactly 2 executions from list, got %d", len(list))
	}
	for _, exec := range list {
		if exec.SessionID != sessionID {
			t.Fatalf("expected listed execution session %q, got %q", sessionID, exec.SessionID)
		}
	}

	events := []executionEventEnvelope{}
	reqJSONStatus(t, client, http.MethodGet, fmt.Sprintf("%s/api/v1/executions/%s/events?after_seq=0", server.URL, repl.ID), nil, http.StatusOK, &events)
	if len(events) == 0 {
		t.Fatalf("expected events for repl execution")
	}
	for i, event := range events {
		if event.ExecutionID != repl.ID {
			t.Fatalf("expected event execution_id %q, got %q", repl.ID, event.ExecutionID)
		}
		if event.Seq != i+1 {
			t.Fatalf("expected contiguous event sequence; index=%d seq=%d", i, event.Seq)
		}
		if event.Type == "" {
			t.Fatalf("expected non-empty event type")
		}
		if event.Ts.IsZero() {
			t.Fatalf("expected non-zero event timestamp")
		}
		var payload interface{}
		if err := json.Unmarshal(event.Payload, &payload); err != nil {
			t.Fatalf("expected valid event payload JSON, seq=%d err=%v", event.Seq, err)
		}
	}

	filtered := []executionEventEnvelope{}
	reqJSONStatus(t, client, http.MethodGet, fmt.Sprintf("%s/api/v1/executions/%s/events?after_seq=%d", server.URL, repl.ID, events[len(events)-1].Seq-1), nil, http.StatusOK, &filtered)
	if len(filtered) != 1 || filtered[0].Seq != events[len(events)-1].Seq {
		t.Fatalf("expected after_seq filtering to return only tail event, got %d entries", len(filtered))
	}

	runFileEvents := []executionEventEnvelope{}
	reqJSONStatus(t, client, http.MethodGet, fmt.Sprintf("%s/api/v1/executions/%s/events?after_seq=0", server.URL, runFile.ID), nil, http.StatusOK, &runFileEvents)
	if len(runFileEvents) < 1 {
		t.Fatalf("expected events for run-file execution")
	}
	last := runFileEvents[len(runFileEvents)-1]
	if last.Type != "value" {
		t.Fatalf("expected run-file to end with value event, got %q", last.Type)
	}
}

type executionContractResponse struct {
	ID        string          `json:"id"`
	SessionID string          `json:"session_id"`
	Kind      string          `json:"kind"`
	Input     string          `json:"input"`
	Path      string          `json:"path"`
	Status    string          `json:"status"`
	StartedAt time.Time       `json:"started_at"`
	EndedAt   *time.Time      `json:"ended_at"`
	Result    json.RawMessage `json:"result"`
	Error     json.RawMessage `json:"error"`
	Metrics   json.RawMessage `json:"metrics"`
}

type executionEventEnvelope struct {
	ExecutionID string          `json:"execution_id"`
	Seq         int             `json:"seq"`
	Ts          time.Time       `json:"ts"`
	Type        string          `json:"type"`
	Payload     json.RawMessage `json:"payload"`
}

func assertExecutionEnvelope(t *testing.T, exec executionContractResponse, expectedSessionID, expectedKind string) {
	t.Helper()

	if exec.ID == "" {
		t.Fatalf("expected non-empty execution id")
	}
	if exec.SessionID != expectedSessionID {
		t.Fatalf("expected session_id %q, got %q", expectedSessionID, exec.SessionID)
	}
	if exec.Kind != expectedKind {
		t.Fatalf("expected kind %q, got %q", expectedKind, exec.Kind)
	}
	if exec.Status != "ok" {
		t.Fatalf("expected status ok, got %q", exec.Status)
	}
	if exec.StartedAt.IsZero() {
		t.Fatalf("expected started_at to be present")
	}
	if exec.EndedAt == nil || exec.EndedAt.IsZero() {
		t.Fatalf("expected ended_at to be present")
	}
	if len(exec.Metrics) == 0 {
		t.Fatalf("expected metrics JSON payload")
	}
}

func reqJSONStatus(t *testing.T, client *http.Client, method, url string, in interface{}, expectedStatus int, out interface{}) {
	t.Helper()

	var bodyReader *bytes.Reader
	if in != nil {
		body, err := json.Marshal(in)
		if err != nil {
			t.Fatalf("marshal request: %v", err)
		}
		bodyReader = bytes.NewReader(body)
	} else {
		bodyReader = bytes.NewReader(nil)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	if in != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != expectedStatus {
		raw, _ := ioReadAll(resp)
		t.Fatalf("expected status %d, got %d (%s)", expectedStatus, resp.StatusCode, string(raw))
	}

	if out != nil {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			t.Fatalf("decode response for %s: %v", url, err)
		}
	}
}
