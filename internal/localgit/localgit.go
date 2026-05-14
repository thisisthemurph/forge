package localgit

import (
	"bytes"
	"errors"
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

// FeatureBaseBranchForMutation verifies **Repository hygiene** and returns the
// short branch name checked out at invocation (**Feature base branch**).
func FeatureBaseBranchForMutation(repoRoot string) (string, error) {
	if err := CheckMutatingHygiene(repoRoot); err != nil {
		return "", err
	}
	return currentBranchShort(repoRoot)
}

// EnsureBranchFromExisting creates branchName from the tip of fromBranch when missing,
// then checks it out. If branchName already exists locally, it is checked out only.
// This models cutting the **Feature branch** from the **Feature base branch**.
func EnsureBranchFromExisting(repoRoot, branchName, fromBranch string) error {
	exists, err := localBranchExists(repoRoot, branchName)
	if err != nil {
		return err
	}
	if exists {
		_, err := runGit(repoRoot, "switch", branchName)
		return err
	}
	if _, err := runGit(repoRoot, "branch", branchName, fromBranch); err != nil {
		return err
	}
	_, err = runGit(repoRoot, "switch", branchName)
	return err
}

// EnsureStackedFromParent creates stackedBranch from stackParentBranch when missing,
// then checks it out. If stackedBranch already exists, it is checked out only.
func EnsureStackedFromParent(repoRoot, stackedBranch, stackParentBranch string) error {
	exists, err := localBranchExists(repoRoot, stackedBranch)
	if err != nil {
		return err
	}
	if exists {
		_, err := runGit(repoRoot, "switch", stackedBranch)
		return err
	}
	if _, err := runGit(repoRoot, "switch", stackParentBranch); err != nil {
		return fmt.Errorf("checkout stack parent %q: %w", stackParentBranch, err)
	}
	_, err = runGit(repoRoot, "switch", "-c", stackedBranch)
	return err
}

func localBranchExists(repoRoot, branchName string) (bool, error) {
	err := exec.Command("git", "-C", repoRoot, "show-ref", "--verify", "--quiet", "refs/heads/"+branchName).Run()
	if err == nil {
		return true, nil
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
		return false, nil
	}
	return false, err
}

// PushOriginBranch pushes a local branch to **origin** using `git push -u`
// (**Push remote (v1)**).
func PushOriginBranch(repoRoot, branch string) error {
	if strings.TrimSpace(branch) == "" {
		return fmt.Errorf("push: branch name is required")
	}
	_, err := runGit(repoRoot, "push", "-u", "origin", branch)
	return err
}
