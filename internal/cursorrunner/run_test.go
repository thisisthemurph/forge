package cursorrunner

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// writeFakeCursor writes a shell script that acts as a stand-in for `cursor`.
// The script body is appended after `#!/bin/sh` so tests can express custom
// exit codes, side-effects (writing pwd / env to a file), etc.
func writeFakeCursor(t *testing.T, body string) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("cursor runner integration tests rely on a POSIX shell")
	}
	dir := t.TempDir()
	bin := filepath.Join(dir, "cursor")
	contents := "#!/bin/sh\n" + body + "\n"
	if err := os.WriteFile(bin, []byte(contents), 0o755); err != nil {
		t.Fatal(err)
	}
	return bin
}

func TestRun_successExitsNil(t *testing.T) {
	bin := writeFakeCursor(t, "exit 0")
	err := Run(context.Background(), Config{
		Bin:      bin,
		RepoRoot: t.TempDir(),
		Env:      []string{},
		Stdin:    strings.NewReader(""),
		Stdout:   io.Discard,
		Stderr:   io.Discard,
	})
	if err != nil {
		t.Fatalf("expected nil error on exit 0, got %v", err)
	}
}

func TestRun_setsWorkingDirectoryToRepoRoot(t *testing.T) {
	repoRoot := t.TempDir()
	resolvedRoot, err := filepath.EvalSymlinks(repoRoot)
	if err != nil {
		t.Fatal(err)
	}
	outFile := filepath.Join(t.TempDir(), "pwd.txt")
	bin := writeFakeCursor(t, "pwd > "+outFile)

	if err := Run(context.Background(), Config{
		Bin:      bin,
		RepoRoot: repoRoot,
		Env:      []string{"PATH=" + os.Getenv("PATH")},
		Stdin:    strings.NewReader(""),
		Stdout:   io.Discard,
		Stderr:   io.Discard,
	}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatal(err)
	}
	got := strings.TrimSpace(string(data))
	if got != resolvedRoot {
		t.Fatalf("expected cwd %q, got %q", resolvedRoot, got)
	}
}

func TestRun_doesNotPassForgePrefixedEnvToChild(t *testing.T) {
	repoRoot := t.TempDir()
	outFile := filepath.Join(t.TempDir(), "env.txt")
	bin := writeFakeCursor(t, "env > "+outFile)

	parentEnv := []string{
		"PATH=" + os.Getenv("PATH"),
		"FORGE_FEATURE=11",
		"FORGE_CURSOR_BIN=/should/not/leak",
		"HOME=" + os.Getenv("HOME"),
	}

	if err := Run(context.Background(), Config{
		Bin:      bin,
		RepoRoot: repoRoot,
		Env:      FilterEnv(parentEnv),
		Stdin:    strings.NewReader(""),
		Stdout:   io.Discard,
		Stderr:   io.Discard,
	}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "FORGE_") {
			t.Fatalf("child env contained forge-prefixed entry: %q", line)
		}
	}
}

func TestRun_nonZeroExitReturnsError(t *testing.T) {
	bin := writeFakeCursor(t, "exit 7")
	err := Run(context.Background(), Config{
		Bin:      bin,
		RepoRoot: t.TempDir(),
		Env:      []string{},
		Stdin:    strings.NewReader(""),
		Stdout:   io.Discard,
		Stderr:   io.Discard,
	})
	if err == nil {
		t.Fatal("expected error on non-zero exit")
	}
	var runErr *ExitError
	if !errors.As(err, &runErr) {
		t.Fatalf("expected *cursorrunner.ExitError, got %T: %v", err, err)
	}
	if runErr.ExitCode != 7 {
		t.Fatalf("expected exit code 7, got %d", runErr.ExitCode)
	}
}
