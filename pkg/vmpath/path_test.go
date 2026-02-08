package vmpath

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestParseRelWorktreePath(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		wantErr error
		want    string
	}{
		{name: "empty", in: "", wantErr: ErrEmptyRelativePath},
		{name: "dot", in: ".", wantErr: ErrEmptyRelativePath},
		{name: "absolute", in: "/tmp/a.js", wantErr: ErrAbsoluteRelativePath},
		{name: "traversal", in: "../a.js", wantErr: ErrTraversalRelativePath},
		{name: "nested", in: "runtime/startup.js", want: filepath.Clean("runtime/startup.js")},
		{name: "normalized", in: "runtime/../runtime/startup.js", want: filepath.Clean("runtime/startup.js")},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p, err := ParseRelWorktreePath(tc.in)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("expected error %v, got %v", tc.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if p.String() != tc.want {
				t.Fatalf("expected %q, got %q", tc.want, p.String())
			}
		})
	}
}

func TestResolveRejectsSymlinkEscape(t *testing.T) {
	rootDir := t.TempDir()
	outsideDir := t.TempDir()
	outsideFile := filepath.Join(outsideDir, "outside.js")
	if err := os.WriteFile(outsideFile, []byte("42"), 0o644); err != nil {
		t.Fatalf("write outside file: %v", err)
	}

	if err := os.Symlink(outsideFile, filepath.Join(rootDir, "escape.js")); err != nil {
		t.Fatalf("create symlink: %v", err)
	}

	root, err := NewWorktreeRoot(rootDir)
	if err != nil {
		t.Fatalf("new root: %v", err)
	}
	path, err := ParseRelWorktreePath("escape.js")
	if err != nil {
		t.Fatalf("parse path: %v", err)
	}

	_, err = root.Resolve(path)
	if !errors.Is(err, ErrPathEscapesRoot) {
		t.Fatalf("expected ErrPathEscapesRoot, got %v", err)
	}
}

func TestResolveAllowsCanonicalPathInsideRoot(t *testing.T) {
	rootDir := t.TempDir()
	targetDir := filepath.Join(rootDir, "runtime")
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	targetFile := filepath.Join(targetDir, "startup.js")
	if err := os.WriteFile(targetFile, []byte("console.log('ok')"), 0o644); err != nil {
		t.Fatalf("write target: %v", err)
	}

	if err := os.Symlink(targetFile, filepath.Join(rootDir, "startup-link.js")); err != nil {
		t.Fatalf("create symlink: %v", err)
	}

	root, err := NewWorktreeRoot(rootDir)
	if err != nil {
		t.Fatalf("new root: %v", err)
	}
	path, err := ParseRelWorktreePath("startup-link.js")
	if err != nil {
		t.Fatalf("parse path: %v", err)
	}

	resolved, err := root.Resolve(path)
	if err != nil {
		t.Fatalf("resolve path: %v", err)
	}

	wantRel := filepath.Clean("runtime/startup.js")
	if resolved.Relative() != wantRel {
		t.Fatalf("expected relative %q, got %q", wantRel, resolved.Relative())
	}
	if resolved.Absolute() != filepath.Clean(targetFile) {
		t.Fatalf("expected absolute %q, got %q", filepath.Clean(targetFile), resolved.Absolute())
	}
}

func TestNewWorktreeRootRejectsNonDirectory(t *testing.T) {
	rootDir := t.TempDir()
	file := filepath.Join(rootDir, "file.txt")
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	_, err := NewWorktreeRoot(file)
	if !errors.Is(err, ErrRootNotDirectory) {
		t.Fatalf("expected ErrRootNotDirectory, got %v", err)
	}
}
