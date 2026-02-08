package vmmodels

import (
	"errors"
	"testing"
)

func TestParseTemplateID(t *testing.T) {
	valid := "6f03fdb4-8b28-4d53-98f7-5a6d5fef2df8"

	parsed, err := ParseTemplateID(valid)
	if err != nil {
		t.Fatalf("expected valid template id, got error: %v", err)
	}
	if parsed.String() != valid {
		t.Fatalf("expected normalized id %q, got %q", valid, parsed.String())
	}

	_, err = ParseTemplateID("not-a-uuid")
	if !errors.Is(err, ErrInvalidTemplateID) {
		t.Fatalf("expected ErrInvalidTemplateID, got %v", err)
	}
}

func TestParseSessionID(t *testing.T) {
	valid := "47d7c945-4c95-43ba-8706-6bfac52c957c"

	parsed, err := ParseSessionID(valid)
	if err != nil {
		t.Fatalf("expected valid session id, got error: %v", err)
	}
	if parsed.String() != valid {
		t.Fatalf("expected normalized id %q, got %q", valid, parsed.String())
	}

	_, err = ParseSessionID("")
	if !errors.Is(err, ErrInvalidSessionID) {
		t.Fatalf("expected ErrInvalidSessionID, got %v", err)
	}
}

func TestParseExecutionID(t *testing.T) {
	valid := "dd0f93bb-66e9-42e4-9785-06be3e34f1fe"

	parsed, err := ParseExecutionID(valid)
	if err != nil {
		t.Fatalf("expected valid execution id, got error: %v", err)
	}
	if parsed.String() != valid {
		t.Fatalf("expected normalized id %q, got %q", valid, parsed.String())
	}

	_, err = ParseExecutionID("   ")
	if !errors.Is(err, ErrInvalidExecutionID) {
		t.Fatalf("expected ErrInvalidExecutionID, got %v", err)
	}
}

func TestMustIDHelpersPanicOnInvalid(t *testing.T) {
	assertPanics(t, func() { _ = MustTemplateID("invalid") })
	assertPanics(t, func() { _ = MustSessionID("invalid") })
	assertPanics(t, func() { _ = MustExecutionID("invalid") })
}

func assertPanics(t *testing.T, fn func()) {
	t.Helper()
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic")
		}
	}()
	fn()
}
