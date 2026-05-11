package remote

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

var (
	ErrNotGitHub     = errors.New("remote does not point at github.com")
	ErrInvalidRemote = errors.New("could not parse owner/repo from remote URL")
)

// ParseRepoOwnerPath parses a --repo value "owner/name" (optional .git on name).
func ParseRepoOwnerPath(repoFlag string) (owner, repo string, err error) {
	repoFlag = strings.TrimSpace(repoFlag)
	if repoFlag == "" {
		return "", "", fmt.Errorf("--repo must be owner/name")
	}
	parts := strings.Split(repoFlag, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" || strings.Contains(parts[1], "/") {
		return "", "", fmt.Errorf("--repo must be owner/name")
	}
	name := strings.TrimSuffix(strings.TrimSpace(parts[1]), ".git")
	owner = strings.TrimSpace(parts[0])
	if owner == "" || name == "" {
		return "", "", fmt.Errorf("--repo must be owner/name")
	}
	return strings.ToLower(owner), strings.ToLower(name), nil
}

// OwnerRepoFromURL extracts GitHub owner and repository from common git remote URL forms.
// Only github.com SaaS hosts are accepted (v1 product scope).
func OwnerRepoFromURL(raw string) (owner, repo string, err error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", "", ErrInvalidRemote
	}
	if o, r, ok := parseGitHubSCP(raw); ok {
		return o, r, nil
	}
	if o, r, ok := parseGitHubSSHURL(raw); ok {
		return o, r, nil
	}
	return parseGitHubHTTPS(raw)
}

func parseGitHubSCP(raw string) (owner, repo string, ok bool) {
	const prefix = "git@github.com:"
	lower := strings.ToLower(raw)
	idx := strings.Index(lower, prefix)
	if idx < 0 {
		return "", "", false
	}
	rest := raw[idx+len(prefix):]
	rest = strings.TrimSuffix(rest, ".git")
	segs := strings.SplitN(rest, "/", 2)
	if len(segs) != 2 || segs[0] == "" || segs[1] == "" {
		return "", "", false
	}
	return strings.ToLower(segs[0]), strings.ToLower(segs[1]), true
}

func parseGitHubSSHURL(raw string) (owner, repo string, ok bool) {
	if !strings.HasPrefix(strings.ToLower(raw), "ssh://") {
		return "", "", false
	}
	u, err := url.Parse(raw)
	if err != nil || !strings.EqualFold(u.Host, "github.com") {
		return "", "", false
	}
	path := strings.Trim(strings.TrimPrefix(u.Path, "/"), "/")
	path = strings.TrimSuffix(path, ".git")
	segs := strings.SplitN(path, "/", 2)
	if len(segs) != 2 || segs[0] == "" || segs[1] == "" {
		return "", "", false
	}
	return strings.ToLower(segs[0]), strings.ToLower(segs[1]), true
}

func parseGitHubHTTPS(raw string) (owner, repo string, err error) {
	u, err := url.Parse(raw)
	if err != nil {
		return "", "", fmt.Errorf("%w: %w", ErrInvalidRemote, err)
	}
	if u.Hostname() == "" {
		return "", "", ErrInvalidRemote
	}
	if !strings.EqualFold(u.Hostname(), "github.com") {
		return "", "", ErrNotGitHub
	}
	path := strings.Trim(strings.TrimPrefix(u.Path, "/"), "/")
	path = strings.TrimSuffix(path, ".git")
	segs := strings.SplitN(path, "/", 2)
	if len(segs) != 2 || segs[0] == "" || segs[1] == "" {
		return "", "", ErrInvalidRemote
	}
	return strings.ToLower(segs[0]), strings.ToLower(segs[1]), nil
}
