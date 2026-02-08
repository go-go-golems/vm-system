package vmmodels

import (
	"encoding/json"
	"testing"
)

func TestMarshalJSONWithFallbackSuccess(t *testing.T) {
	got := MarshalJSONWithFallback(map[string]int{"n": 42}, json.RawMessage(`{"fallback":true}`))
	if string(got) != `{"n":42}` {
		t.Fatalf("expected marshaled JSON, got %s", string(got))
	}
}

func TestMarshalJSONWithFallbackFailureUsesFallback(t *testing.T) {
	type bad struct {
		Fn func()
	}

	got := MarshalJSONWithFallback(bad{Fn: func() {}}, json.RawMessage(`{"fallback":true}`))
	if string(got) != `{"fallback":true}` {
		t.Fatalf("expected fallback JSON on marshal failure, got %s", string(got))
	}
}

func TestMarshalJSONWithFallbackFailureEmptyFallbackBecomesNull(t *testing.T) {
	type bad struct {
		Fn func()
	}

	got := MarshalJSONWithFallback(bad{Fn: func() {}}, nil)
	if string(got) != "null" {
		t.Fatalf("expected null fallback when fallback is empty, got %s", string(got))
	}
}
