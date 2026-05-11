package localgit

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func gitCmd(t *testing.T, repoRoot string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", repoRoot}, args...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

func initRepoWithMain(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	gitCmd(t, dir, "init", "--initial-branch=main")
	gitCmd(t, dir, "config", "user.email", "forge-test@example.com")
	gitCmd(t, dir, "config", "user.name", "Forge Test")
	readme := filepath.Join(dir, "README.md")
	if err := os.WriteFile(readme, []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitCmd(t, dir, "add", "README.md")
	gitCmd(t, dir, "commit", "-m", "init")
	return dir
}

func TestCheckMutatingHygiene_detachedHEAD(t *testing.T) {
	repo := initRepoWithMain(t)
	head := gitCmd(t, repo, "rev-parse", "HEAD")
	gitCmd(t, repo, "checkout", "--detach", head)

	err := CheckMutatingHygiene(repo)
	if err == nil {
		t.Fatal("expected error for detached HEAD")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "detached") {
		t.Fatalf("error should mention detached HEAD, got: %v", err)
	}
}

func TestCheckMutatingHygiene_dirtyWorktree(t *testing.T) {
	repo := initRepoWithMain(t)
	if err := os.WriteFile(filepath.Join(repo, "README.md"), []byte("dirty\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	err := CheckMutatingHygiene(repo)
	if err == nil {
		t.Fatal("expected error for dirty worktree")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "clean") {
		t.Fatalf("error should mention clean worktree, got: %v", err)
	}
}

func TestCheckMutatingHygiene_cleanBranch(t *testing.T) {
	repo := initRepoWithMain(t)
	if err := CheckMutatingHygiene(repo); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}
