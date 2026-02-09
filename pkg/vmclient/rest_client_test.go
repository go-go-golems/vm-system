package vmclient

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestClientDoMapsAPIErrorEnvelope(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]interface{}{
				"code":    "MODULE_NOT_ALLOWED",
				"message": "module is disallowed",
				"details": map[string]interface{}{"name": "json"},
			},
		})
	}))
	defer server.Close()

	client := New(server.URL, server.Client())
	_, err := client.ListTemplates(context.Background())
	if err == nil {
		t.Fatalf("expected API error")
	}

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T (%v)", err, err)
	}
	if apiErr.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("expected status 422, got %d", apiErr.StatusCode)
	}
	if apiErr.Code != "MODULE_NOT_ALLOWED" {
		t.Fatalf("expected code MODULE_NOT_ALLOWED, got %q", apiErr.Code)
	}
	if apiErr.Message != "module is disallowed" {
		t.Fatalf("expected message from envelope, got %q", apiErr.Message)
	}

	details, ok := apiErr.Details.(map[string]interface{})
	if !ok {
		t.Fatalf("expected details map, got %T", apiErr.Details)
	}
	if details["name"] != "json" {
		t.Fatalf("expected details.name=json, got %#v", details["name"])
	}
}

func TestClientDoReturnsStatusOnlyWhenErrorEnvelopeIsMissing(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("upstream broke"))
	}))
	defer server.Close()

	client := New(server.URL, server.Client())
	_, err := client.Health(context.Background())
	if err == nil {
		t.Fatalf("expected API error")
	}

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T (%v)", err, err)
	}
	if apiErr.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", apiErr.StatusCode)
	}
	if apiErr.Code != "" || apiErr.Message != "" {
		t.Fatalf("expected empty code/message when envelope is not decodable, got code=%q message=%q", apiErr.Code, apiErr.Message)
	}
}

func TestClientDoReturnsDecodeErrorForMalformedSuccessBody(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{not-json"))
	}))
	defer server.Close()

	client := New(server.URL, server.Client())
	_, err := client.RuntimeSummary(context.Background())
	if err == nil {
		t.Fatalf("expected decode error")
	}
	if !strings.Contains(err.Error(), "decode response") {
		t.Fatalf("expected decode response error, got %v", err)
	}
}

func TestWithQuerySkipsEmptyValuesAndEncodesNonEmpty(t *testing.T) {
	t.Parallel()

	got := withQuery("/api/v1/executions", map[string]string{
		"session_id": "abc",
		"limit":      "10",
		"after_seq":  "",
	})

	if !strings.HasPrefix(got, "/api/v1/executions?") {
		t.Fatalf("expected query string suffix, got %q", got)
	}
	if !strings.Contains(got, "session_id=abc") {
		t.Fatalf("expected encoded session_id, got %q", got)
	}
	if !strings.Contains(got, "limit=10") {
		t.Fatalf("expected encoded limit, got %q", got)
	}
	if strings.Contains(got, "after_seq") {
		t.Fatalf("expected empty value to be omitted, got %q", got)
	}
}
