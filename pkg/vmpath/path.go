package vmpath

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var (
	ErrInvalidRoot           = errors.New("invalid worktree root")
	ErrRootNotDirectory      = errors.New("worktree root is not a directory")
	ErrEmptyRelativePath     = errors.New("empty relative path")
	ErrAbsoluteRelativePath  = errors.New("absolute path is not allowed")
	ErrTraversalRelativePath = errors.New("path traversal is not allowed")
	ErrPathEscapesRoot       = errors.New("resolved path escapes worktree root")
)

// WorktreeRoot is a canonicalized worktree root directory.
type WorktreeRoot struct {
	canonical string
}

// RelWorktreePath is a validated relative path within a worktree root.
type RelWorktreePath struct {
	value string
}

// ResolvedWorktreePath is a path resolved against a worktree root.
type ResolvedWorktreePath struct {
	root     WorktreeRoot
	relative string
	absolute string
}

func NewWorktreeRoot(path string) (WorktreeRoot, error) {
	if strings.TrimSpace(path) == "" {
		return WorktreeRoot{}, ErrInvalidRoot
	}

	abs, err := filepath.Abs(path)
	if err != nil {
		return WorktreeRoot{}, fmt.Errorf("%w: %v", ErrInvalidRoot, err)
	}
	abs = filepath.Clean(abs)

	info, err := os.Stat(abs)
	if err != nil {
		return WorktreeRoot{}, fmt.Errorf("%w: %v", ErrInvalidRoot, err)
	}
	if !info.IsDir() {
		return WorktreeRoot{}, ErrRootNotDirectory
	}

	canonical, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return WorktreeRoot{}, fmt.Errorf("%w: %v", ErrInvalidRoot, err)
	}

	return WorktreeRoot{canonical: filepath.Clean(canonical)}, nil
}

func (w WorktreeRoot) Canonical() string {
	return w.canonical
}

func ParseRelWorktreePath(path string) (RelWorktreePath, error) {
	if strings.TrimSpace(path) == "" {
		return RelWorktreePath{}, ErrEmptyRelativePath
	}

	if filepath.IsAbs(path) {
		return RelWorktreePath{}, ErrAbsoluteRelativePath
	}

	clean := filepath.Clean(path)
	if clean == "." {
		return RelWorktreePath{}, ErrEmptyRelativePath
	}
	if clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return RelWorktreePath{}, ErrTraversalRelativePath
	}

	return RelWorktreePath{value: clean}, nil
}

func (r RelWorktreePath) String() string {
	return r.value
}

func (w WorktreeRoot) Resolve(path RelWorktreePath) (ResolvedWorktreePath, error) {
	candidate := filepath.Join(w.canonical, path.value)
	resolved := candidate

	if eval, err := filepath.EvalSymlinks(candidate); err == nil {
		resolved = eval
	} else if !os.IsNotExist(err) {
		return ResolvedWorktreePath{}, fmt.Errorf("resolve symlinks for %q: %w", path.value, err)
	}

	resolvedAbs, err := filepath.Abs(resolved)
	if err != nil {
		return ResolvedWorktreePath{}, fmt.Errorf("resolve absolute path for %q: %w", path.value, err)
	}
	resolvedAbs = filepath.Clean(resolvedAbs)

	relToRoot, err := filepath.Rel(w.canonical, resolvedAbs)
	if err != nil {
		return ResolvedWorktreePath{}, fmt.Errorf("resolve root relation for %q: %w", path.value, err)
	}
	if relToRoot == ".." || strings.HasPrefix(relToRoot, ".."+string(filepath.Separator)) {
		return ResolvedWorktreePath{}, ErrPathEscapesRoot
	}

	return ResolvedWorktreePath{
		root:     w,
		relative: filepath.Clean(relToRoot),
		absolute: resolvedAbs,
	}, nil
}

func (r ResolvedWorktreePath) Root() WorktreeRoot {
	return r.root
}

func (r ResolvedWorktreePath) Relative() string {
	return r.relative
}

func (r ResolvedWorktreePath) Absolute() string {
	return r.absolute
}
