package cursorrunner

import "strings"

// FilterEnv returns env with every "FORGE_*" entry removed, preserving order.
// It enforces the **Cursor environment (v1)** rule that Forge does not inject
// FORGE_-prefixed variables into the Cursor child process.
func FilterEnv(env []string) []string {
	out := make([]string, 0, len(env))
	for _, kv := range env {
		if strings.HasPrefix(kv, "FORGE_") {
			continue
		}
		out = append(out, kv)
	}
	return out
}
