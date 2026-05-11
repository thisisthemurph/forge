package token

import (
	"fmt"
	"strings"
)

// Resolve returns a GitHub token using PRD order: GH_TOKEN, GITHUB_TOKEN, then ghAuthToken.
// ghAuthToken typically runs `gh auth token`; it is only called when both env vars are unset or whitespace.
func Resolve(getenv func(string) string, ghAuthToken func() (string, error)) (string, error) {
	if t := strings.TrimSpace(getenv("GH_TOKEN")); t != "" {
		return t, nil
	}
	if t := strings.TrimSpace(getenv("GITHUB_TOKEN")); t != "" {
		return t, nil
	}
	if ghAuthToken == nil {
		return "", errNoToken()
	}
	t, err := ghAuthToken()
	if err != nil {
		return "", fmt.Errorf("GitHub token not found (set GH_TOKEN or GITHUB_TOKEN, or authenticate gh): %w", err)
	}
	t = strings.TrimSpace(t)
	if t == "" {
		return "", errNoToken()
	}
	return t, nil
}

func errNoToken() error {
	return fmt.Errorf("GitHub token not found: set GH_TOKEN or GITHUB_TOKEN, or run `gh auth login`")
}
