package gitremote

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// GetOriginURL returns `git remote get-url origin` for a git work tree at repoRoot.
func GetOriginURL(repoRoot string) (string, error) {
	cmd := exec.Command("git", "-C", repoRoot, "remote", "get-url", "origin")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git remote get-url origin: %w: %s", err, strings.TrimSpace(stderr.String()))
	}
	return strings.TrimSpace(stdout.String()), nil
}

// FindRepoRoot walks upward from startDir until a .git entry exists (file or directory).
func FindRepoRoot(startDir string) (string, error) {
	dir, err := filepath.Abs(startDir)
	if err != nil {
		return "", err
	}
	for {
		gitEntry := filepath.Join(dir, ".git")
		if _, err := os.Stat(gitEntry); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("not inside a git repository (no .git from %s)", startDir)
		}
		dir = parent
	}
}
