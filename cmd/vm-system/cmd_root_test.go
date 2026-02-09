package main

import (
	"testing"

	"github.com/go-go-golems/glazed/pkg/help"
)

func TestRootCommandRegistersExpectedTopLevelCommands(t *testing.T) {
	root := newRootCommand(help.NewHelpSystem())

	expected := map[string]bool{
		"serve":    true,
		"template": true,
		"session":  true,
		"exec":     true,
		"ops":      true,
		"libs":     true,
	}

	seen := map[string]bool{}
	for _, sub := range root.Commands() {
		seen[sub.Name()] = true
	}

	for name := range expected {
		if !seen[name] {
			t.Fatalf("expected top-level command %q to be registered", name)
		}
	}

	if seen["http"] {
		t.Fatalf("did not expect top-level command %q to be registered", "http")
	}
}
