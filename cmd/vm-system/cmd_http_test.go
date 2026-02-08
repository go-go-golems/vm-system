package main

import "testing"

func TestHTTPCommandIncludesTemplateSessionExecSubcommands(t *testing.T) {
	cmd := newHTTPCommand()

	expected := map[string]bool{
		"template": true,
		"session":  true,
		"exec":     true,
	}

	seen := map[string]bool{}
	for _, sub := range cmd.Commands() {
		seen[sub.Name()] = true
	}

	for name := range expected {
		if !seen[name] {
			t.Fatalf("expected http subcommand %q to be registered", name)
		}
	}
}
