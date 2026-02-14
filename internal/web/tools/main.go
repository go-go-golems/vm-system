package main

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
)

const (
	controlGitIgnoreName  = ".gitignore"
	controlPlaceholder    = "README_DO_NOT_DELETE.txt"
	controlGitIgnoreBody  = "*\n!.gitignore\n!README_DO_NOT_DELETE.txt\n"
	controlPlaceholderTxt = "Placeholder file to keep this directory embeddable before generated UI assets exist.\nGenerated assets are intentionally git-ignored.\n"
)

func main() {
	repoRoot, err := findRepoRoot()
	if err != nil {
		fatalf("find repo root: %v", err)
	}

	frontendDir := filepath.Join(repoRoot, "ui")
	frontendOut := filepath.Join(frontendDir, "dist", "public")
	targetDir := filepath.Join(repoRoot, "internal", "web", "embed", "public")

	fmt.Printf("[webgen] repo root: %s\n", repoRoot)
	fmt.Printf("[webgen] building frontend: %s\n", frontendDir)

	if err := run("pnpm", "-C", frontendDir, "run", "build"); err != nil {
		fatalf("build frontend: %v", err)
	}

	if err := os.RemoveAll(targetDir); err != nil {
		fatalf("clean target dir: %v", err)
	}
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		fatalf("create target dir: %v", err)
	}

	fmt.Printf("[webgen] copying %s -> %s\n", frontendOut, targetDir)
	if err := copyDir(frontendOut, targetDir); err != nil {
		fatalf("copy frontend output: %v", err)
	}

	if err := writeControlFiles(targetDir); err != nil {
		fatalf("write control files: %v", err)
	}

	fmt.Println("[webgen] done")
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

func run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func copyDir(srcDir, dstDir string) error {
	return filepath.WalkDir(srcDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		rel, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}

		dstPath := filepath.Join(dstDir, rel)

		if d.IsDir() {
			return os.MkdirAll(dstPath, 0o755)
		}

		info, err := d.Info()
		if err != nil {
			return err
		}

		if err := copyFile(path, dstPath, info.Mode()); err != nil {
			return err
		}

		return nil
	})
}

func copyFile(srcPath, dstPath string, mode fs.FileMode) error {
	src, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer src.Close()

	if err := os.MkdirAll(filepath.Dir(dstPath), 0o755); err != nil {
		return err
	}

	dst, err := os.OpenFile(dstPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode.Perm())
	if err != nil {
		return err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return err
	}

	return nil
}

func writeControlFiles(targetDir string) error {
	gitIgnorePath := filepath.Join(targetDir, controlGitIgnoreName)
	if err := os.WriteFile(gitIgnorePath, []byte(controlGitIgnoreBody), 0o644); err != nil {
		return err
	}

	placeholderPath := filepath.Join(targetDir, controlPlaceholder)
	if err := os.WriteFile(placeholderPath, []byte(controlPlaceholderTxt), 0o644); err != nil {
		return err
	}

	return nil
}

func fatalf(format string, args ...any) {
	_, _ = fmt.Fprintf(os.Stderr, "[webgen] ERROR: "+format+"\n", args...)
	os.Exit(1)
}
