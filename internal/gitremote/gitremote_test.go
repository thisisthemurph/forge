package gitremote

import (
	"os/exec"
	"path/filepath"
	"testing"
)

func TestGetOriginURL(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	run := func(name string, args ...string) {
		t.Helper()
		cmd := exec.Command(name, args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%v %v: %v\n%s", name, args, err, out)
		}
	}
	run("git", "init")
	run("git", "remote", "add", "origin", "https://github.com/testorg/testrepo.git")

	got, err := GetOriginURL(dir)
	if err != nil {
		t.Fatal(err)
	}
	if want := "https://github.com/testorg/testrepo.git"; got != want {
		t.Fatalf("GetOriginURL = %q, want %q", got, want)
	}
}

func TestGetOriginURLMissing(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	cmd := exec.Command("git", "init")
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}
	_, err := GetOriginURL(dir)
	if err == nil {
		t.Fatal("expected error without origin remote")
	}
}

func TestFindRepoRoot(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	sub := filepath.Join(root, "a", "b")
	if err := exec.Command("mkdir", "-p", sub).Run(); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("git", "init")
	cmd.Dir = root
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}
	found, err := FindRepoRoot(sub)
	if err != nil {
		t.Fatal(err)
	}
	if found != root {
		t.Fatalf("FindRepoRoot = %q, want %q", found, root)
	}
}
