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

func TestFeatureBaseBranchForMutation_returnsCurrentBranch(t *testing.T) {
	repo := initRepoWithMain(t)
	base, err := FeatureBaseBranchForMutation(repo)
	if err != nil {
		t.Fatal(err)
	}
	if base != "main" {
		t.Fatalf("expected main, got %q", base)
	}
}

func TestEnsureBranchFromExisting_createsFromFeatureBase(t *testing.T) {
	repo := initRepoWithMain(t)
	mainTip := gitCmd(t, repo, "rev-parse", "HEAD")

	if _, err := FeatureBaseBranchForMutation(repo); err != nil {
		t.Fatal(err)
	}
	const featureBranch = "forge/feature/42/base"
	if err := EnsureBranchFromExisting(repo, featureBranch, "main"); err != nil {
		t.Fatal(err)
	}
	if got := gitCmd(t, repo, "branch", "--show-current"); got != featureBranch {
		t.Fatalf("expected checkout %s, on %q", featureBranch, got)
	}
	if got := gitCmd(t, repo, "rev-parse", "HEAD"); got != mainTip {
		t.Fatalf("feature branch tip should match main tip at creation, main=%s got=%s", mainTip, got)
	}

	if err := EnsureBranchFromExisting(repo, featureBranch, "main"); err != nil {
		t.Fatalf("second call should switch to existing branch: %v", err)
	}
	if got := gitCmd(t, repo, "branch", "--show-current"); got != featureBranch {
		t.Fatalf("expected still on %s, on %q", featureBranch, got)
	}
}

func TestCheckMutatingHygiene_stagedChanges(t *testing.T) {
	repo := initRepoWithMain(t)
	readme := filepath.Join(repo, "README.md")
	if err := os.WriteFile(readme, []byte("staged\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitCmd(t, repo, "add", "README.md")

	err := CheckMutatingHygiene(repo)
	if err == nil {
		t.Fatal("expected error when index has staged changes")
	}
}

func TestEnsureStackedFromParent_branchesFromStackParent(t *testing.T) {
	repo := initRepoWithMain(t)
	if _, err := FeatureBaseBranchForMutation(repo); err != nil {
		t.Fatal(err)
	}
	const featureBranch = "forge/feature/7/base"
	const stackedBranch = "forge/feature/7/issue/99"
	if err := EnsureBranchFromExisting(repo, featureBranch, "main"); err != nil {
		t.Fatal(err)
	}
	readme := filepath.Join(repo, "README.md")
	if err := os.WriteFile(readme, []byte("on feature\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitCmd(t, repo, "add", "README.md")
	gitCmd(t, repo, "commit", "-m", "feature work")
	featureTip := gitCmd(t, repo, "rev-parse", "HEAD")

	if err := EnsureStackedFromParent(repo, stackedBranch, featureBranch); err != nil {
		t.Fatal(err)
	}
	if got := gitCmd(t, repo, "branch", "--show-current"); got != stackedBranch {
		t.Fatalf("expected checkout %s, on %q", stackedBranch, got)
	}
	if got := gitCmd(t, repo, "rev-parse", "HEAD"); got != featureTip {
		t.Fatalf("stacked branch should start at feature tip %s, got %s", featureTip, got)
	}
}

func TestPushOriginBranch_pushesToBareRemote(t *testing.T) {
	repo := initRepoWithMain(t)
	parent := t.TempDir()
	bare := filepath.Join(parent, "origin.git")
	gitCmd(t, parent, "init", "--bare", "origin.git")

	gitCmd(t, repo, "remote", "add", "origin", bare)
	gitCmd(t, repo, "push", "-u", "origin", "main")

	readme := filepath.Join(repo, "README.md")
	if err := os.WriteFile(readme, []byte("topic work\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitCmd(t, repo, "checkout", "-b", "topic")
	gitCmd(t, repo, "add", "README.md")
	gitCmd(t, repo, "commit", "-m", "topic")

	if err := PushOriginBranch(repo, "topic"); err != nil {
		t.Fatal(err)
	}
	gitCmd(t, bare, "show-ref", "--verify", "refs/heads/topic")
}
