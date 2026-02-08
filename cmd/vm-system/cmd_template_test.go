package main

import "testing"

func TestTemplateCommandIncludesModuleAndLibrarySubcommands(t *testing.T) {
	cmd := newTemplateCommand()

	expected := map[string]bool{
		"add-module":               true,
		"remove-module":            true,
		"list-modules":             true,
		"add-library":              true,
		"remove-library":           true,
		"list-libraries":           true,
		"list-available-modules":   true,
		"list-available-libraries": true,
	}

	seen := map[string]bool{}
	for _, sub := range cmd.Commands() {
		seen[sub.Name()] = true
	}

	for name := range expected {
		if !seen[name] {
			t.Fatalf("expected template subcommand %q to be registered", name)
		}
	}
}
