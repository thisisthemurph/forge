package localgit

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

func runGit(repoRoot string, args ...string) (string, error) {
	cmd := exec.Command("git", append([]string{"-C", repoRoot}, args...)...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = strings.TrimSpace(stdout.String())
		}
		if msg != "" {
			return "", fmt.Errorf("git %v: %w: %s", args, err, msg)
		}
		return "", fmt.Errorf("git %v: %w", args, err)
	}
	return strings.TrimSpace(stdout.String()), nil
}

// CheckMutatingHygiene enforces **Repository hygiene (v1)** before mutating work:
// HEAD must resolve to a branch (not detached) and the index/worktree must be clean.
func CheckMutatingHygiene(repoRoot string) error {
	if _, err := currentBranchShort(repoRoot); err != nil {
		return fmt.Errorf("repository hygiene: %w", err)
	}
	clean, err := isWorktreeClean(repoRoot)
	if err != nil {
		return fmt.Errorf("repository hygiene: %w", err)
	}
	if !clean {
		return fmt.Errorf("repository hygiene: working tree is not clean (commit or stash changes before running Forge)")
	}
	return nil
}

func currentBranchShort(repoRoot string) (string, error) {
	out, err := runGit(repoRoot, "symbolic-ref", "-q", "--short", "HEAD")
	if err != nil {
		return "", fmt.Errorf("detached HEAD (check out a branch before running Forge): %w", err)
	}
	if out == "" {
		return "", fmt.Errorf("detached HEAD (check out a branch before running Forge)")
	}
	return out, nil
}

func isWorktreeClean(repoRoot string) (bool, error) {
	out, err := runGit(repoRoot, "status", "--porcelain")
	if err != nil {
		return false, err
	}
	return out == "", nil
}
