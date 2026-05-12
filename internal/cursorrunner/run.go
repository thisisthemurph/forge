package cursorrunner

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
)

// Config describes one invocation of the Cursor CLI by Forge.
//
// Bin is the absolute path to the Cursor executable (see ResolveBin).
// RepoRoot is the working directory for the child process (per **Repository
// root** in CONTEXT.md).
// Env is the full child environment; callers should pass it through FilterEnv
// so **Cursor environment (v1)** holds (no FORGE_* leakage).
// Stdin/Stdout/Stderr are wired straight to the child (typically os.Stdin /
// os.Stdout / os.Stderr in production for inherited stdio).
type Config struct {
	Bin      string
	RepoRoot string
	Env      []string
	Stdin    io.Reader
	Stdout   io.Writer
	Stderr   io.Writer
}

// ExitError reports a non-zero Cursor exit. **Agent failure hygiene (v1)**:
// Forge never resets or cleans the working tree on agent failure; the caller
// receives this error and exits non-zero with the tree as-is for inspection.
type ExitError struct {
	ExitCode int
	Err      error
}

func (e *ExitError) Error() string {
	return fmt.Sprintf("cursor exited with status %d: %v", e.ExitCode, e.Err)
}

func (e *ExitError) Unwrap() error { return e.Err }

// Run executes the Cursor CLI per cfg and maps its exit status to a Go error.
// Exit 0 returns nil; any non-zero exit is wrapped in *ExitError.
//
// Run performs no cleanup or reset of cfg.RepoRoot on failure — per **Agent
// failure hygiene (v1)** the checkout is left as-is for the operator to
// inspect.
func Run(ctx context.Context, cfg Config, args ...string) error {
	cmd := exec.CommandContext(ctx, cfg.Bin, args...)
	cmd.Dir = cfg.RepoRoot
	cmd.Env = cfg.Env
	cmd.Stdin = cfg.Stdin
	cmd.Stdout = cfg.Stdout
	cmd.Stderr = cfg.Stderr

	err := cmd.Run()
	if err == nil {
		return nil
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return &ExitError{ExitCode: exitErr.ExitCode(), Err: err}
	}
	return fmt.Errorf("invoke cursor %q: %w", cfg.Bin, err)
}
