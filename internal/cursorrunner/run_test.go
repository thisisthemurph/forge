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
