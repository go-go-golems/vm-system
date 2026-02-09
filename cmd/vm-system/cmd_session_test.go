package main

import "testing"

func TestSessionCommandIncludesCloseNotDelete(t *testing.T) {
	cmd := newSessionCommand()

	expected := map[string]bool{
		"create": true,
		"list":   true,
		"get":    true,
		"close":  true,
	}

	seen := map[string]bool{}
	for _, sub := range cmd.Commands() {
		seen[sub.Name()] = true
	}

	for name := range expected {
		if !seen[name] {
			t.Fatalf("expected session subcommand %q to be registered", name)
		}
	}

	if seen["delete"] {
		t.Fatalf("did not expect session subcommand %q to be registered", "delete")
	}
}
