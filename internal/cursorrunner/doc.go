// Package cursorrunner is Forge's **Agent runner (v1)** integration for the
// **Cursor CLI**. It is intentionally narrow: it knows how to find the Cursor
// binary, scrub the environment per **Cursor environment (v1)**, execute
// Cursor in **Repository root** with inherited stdio, and translate its exit
// status into a Go error per **Agent failure hygiene (v1)**.
//
// # Module surface
//
//   - ResolveBin — implements **Cursor binary resolution (v1)**: honours
//     FORGE_CURSOR_BIN (must be absolute) or falls back to "cursor" on PATH.
//   - FilterEnv — strips every FORGE_* entry so the Cursor child process
//     never sees Forge-private variables.
//   - Run — executes the Cursor CLI with the caller's Config and returns
//     nil on exit 0 or a wrapped *ExitError on any non-zero exit. Run does
//     not reset, clean, or otherwise touch the working tree on failure.
//
// # Smoke path (manual)
//
// The exit-status and environment-filtering behaviour is exercised by the
// package's automated tests using a temporary shell script as a stand-in for
// `cursor` (see writeFakeCursor in run_test.go). Those tests do not require
// a real Cursor installation and run by default.
//
// To smoke-test against the actual Cursor binary on a developer machine:
//
//  1. Ensure `cursor --version` works (or set FORGE_CURSOR_BIN to an
//     absolute path).
//  2. From a clean checkout, call cursorrunner.ResolveBin with os.Getenv
//     and exec.LookPath and confirm it returns the expected path.
//  3. Call cursorrunner.Run with cwd=<your repo root> and a harmless
//     argument (for example, `--help`) and confirm exit 0 produces no
//     error.
//
// When `cursor` is not installed and FORGE_CURSOR_BIN is unset, callers
// should skip Run-based smoke checks; ResolveBin returns an actionable
// error in that case which Forge surfaces to the operator.
package cursorrunner
