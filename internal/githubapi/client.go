package githubapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const defaultBaseURL = "https://api.github.com"

// Client calls the GitHub REST API for Forge v1 needs.
type Client struct {
	BaseURL string
	HTTP    *http.Client
	Token   string
}

// SubIssue is a GitHub issue attached as a sub-issue of a Feature issue.
type SubIssue struct {
	Number int    `json:"number"`
	Title  string `json:"title"`
	State  string `json:"state"`
	Body   string `json:"body"`
}

// ListAllSubIssues returns every sub-issue for parentIssue in owner/repo (paginated).
func (c *Client) ListAllSubIssues(ctx context.Context, owner, repo string, parentIssue int) ([]SubIssue, error) {
	base := strings.TrimSuffix(c.BaseURL, "/")
	if base == "" {
		base = defaultBaseURL
	}
	httpClient := c.HTTP
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	var all []SubIssue
	for page := 1; ; page++ {
		u, err := url.Parse(fmt.Sprintf("%s/repos/%s/%s/issues/%d/sub_issues", base, url.PathEscape(owner), url.PathEscape(repo), parentIssue))
		if err != nil {
			return nil, err
		}
		q := u.Query()
		q.Set("per_page", "100")
		q.Set("page", fmt.Sprintf("%d", page))
		u.RawQuery = q.Encode()

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+c.Token)
		req.Header.Set("Accept", "application/vnd.github+json")
		req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
		req.Header.Set("User-Agent", "forge-cli/0")

		resp, err := httpClient.Do(req)
		if err != nil {
			return nil, err
		}
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != http.StatusOK {
			msg := extractMessage(body)
			return nil, fmt.Errorf("GitHub API %s returned %s: %s", u.Path, resp.Status, msg)
		}
		var batch []SubIssue
		if err := json.Unmarshal(body, &batch); err != nil {
			return nil, fmt.Errorf("decode sub_issues: %w", err)
		}
		if len(batch) == 0 {
			break
		}
		all = append(all, batch...)
		if len(batch) < 100 {
			break
		}
	}
	return all, nil
}

func extractMessage(body []byte) string {
	var wrap struct {
		Message string `json:"message"`
	}
	if err := json.Unmarshal(body, &wrap); err == nil && wrap.Message != "" {
		return wrap.Message
	}
	return strings.TrimSpace(string(body))
}
