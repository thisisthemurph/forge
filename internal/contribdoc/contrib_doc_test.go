package contribdoc

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func moduleRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	dir := filepath.Dir(file)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found when walking up from test file")
		}
		dir = parent
	}
}

func contributorForgeCLIDoc(t *testing.T) string {
	t.Helper()
	root := moduleRoot(t)
	path := filepath.Join(root, "docs", "contributors", "forge-cli.md")
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read contributor doc %s: %v", path, err)
	}
	return string(b)
}

// Issue #9 acceptance: agent runner scope (Cursor-only v1), operator responsibilities.
func TestContributorForgeCLIDoc_AgentRunnerScope(t *testing.T) {
	doc := contributorForgeCLIDoc(t)
	for _, needle := range []string{
		"Cursor CLI",
		"FORGE_CURSOR_BIN",
		"Human merge responsibility",
		"Concurrency",
		"tokens",
	} {
		if !strings.Contains(doc, needle) {
			t.Errorf("docs/contributors/forge-cli.md should mention %q (agent runner / operator expectations)", needle)
		}
	}
}
