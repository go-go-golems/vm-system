//go:build !embed

package web

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

func PublicFS() (fs.FS, error) {
	candidates := []string{}

	if repoRoot, err := findRepoRoot(); err == nil {
		candidates = append(candidates, filepath.Join(repoRoot, "internal", "web", "embed", "public"))
	}

	candidates = append(candidates, filepath.Join("internal", "web", "embed", "public"))

	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)
		candidates = append(candidates, filepath.Join(exeDir, "internal", "web", "embed", "public"))
	}

	for _, candidate := range candidates {
		stat, err := os.Stat(candidate)
		if err != nil || !stat.IsDir() {
			continue
		}

		public := os.DirFS(candidate)
		if err := requireIndexFile(public); err != nil {
			continue
		}
		return public, nil
	}

	return nil, fmt.Errorf("public assets directory not found; looked in %v", candidates)
}

func findRepoRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	current := cwd
	for {
		if _, err := os.Stat(filepath.Join(current, "go.mod")); err == nil {
			return current, nil
		}

		parent := filepath.Dir(current)
		if parent == current {
			return "", fmt.Errorf("go.mod not found from %s upward", cwd)
		}

		current = parent
	}
}
