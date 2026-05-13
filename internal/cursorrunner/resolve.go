package cursorrunner

import (
	"fmt"
	"path/filepath"
)

// ResolveBin implements **Cursor binary resolution (v1)**:
//
//   - If FORGE_CURSOR_BIN is set, it must be an absolute path to the Cursor
//     executable; PATH is ignored entirely.
//   - Otherwise, `cursor` is looked up on PATH (via lookPath, typically
//     exec.LookPath).
//
// getenv and lookPath are injected so callers can wire os.Getenv /
// exec.LookPath in production and stub them in tests.
func ResolveBin(getenv func(string) string, lookPath func(string) (string, error)) (string, error) {
	if override := getenv("FORGE_CURSOR_BIN"); override != "" {
		if !filepath.IsAbs(override) {
			return "", fmt.Errorf("FORGE_CURSOR_BIN must be an absolute path to the Cursor executable, got %q", override)
		}
		return override, nil
	}
	path, err := lookPath("cursor")
	if err != nil {
		return "", fmt.Errorf("cursor binary not found on PATH (install Cursor CLI or set FORGE_CURSOR_BIN to an absolute path): %w", err)
	}
	return path, nil
}
