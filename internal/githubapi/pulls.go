package githubapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

// PullRequest is a minimal GitHub pull request payload for Forge scheduling.
type PullRequest struct {
	Number  int    `json:"number"`
	Title   string `json:"title"`
	Body    string `json:"body"`
	State   string `json:"state"`
	Merged  bool   `json:"merged"`
	BaseRef string `json:"base"`
	Labels  []struct {
		Name string `json:"name"`
	} `json:"labels"`
}

type pullJSON struct {
	Number int    `json:"number"`
	Title  string `json:"title"`
	Body   string `json:"body"`
	State  string `json:"state"`
	Merged bool   `json:"merged"`
	Base   struct {
		Ref string `json:"ref"`
	} `json:"base"`
	Labels []struct {
		Name string `json:"name"`
	} `json:"labels"`
}

func (p pullJSON) toPull() PullRequest {
	return PullRequest{
		Number:  p.Number,
		Title:   p.Title,
		Body:    p.Body,
		State:   p.State,
		Merged:  p.Merged,
		BaseRef: p.Base.Ref,
		Labels:  p.Labels,
	}
}

// ListPullRequestsByHead lists pull requests whose head matches owner:branch (state all, paginated).
func (c *Client) ListPullRequestsByHead(ctx context.Context, owner, repo, headBranch string) ([]PullRequest, error) {
	base := strings.TrimSuffix(c.BaseURL, "/")
	if base == "" {
		base = defaultBaseURL
	}
	httpClient := c.HTTP
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	head := owner + ":" + headBranch
	var all []PullRequest
	for page := 1; ; page++ {
		u, err := url.Parse(fmt.Sprintf("%s/repos/%s/%s/pulls", base, url.PathEscape(owner), url.PathEscape(repo)))
		if err != nil {
			return nil, err
		}
		q := u.Query()
		q.Set("state", "all")
		q.Set("head", head)
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
		var batch []pullJSON
		if err := json.Unmarshal(body, &batch); err != nil {
			return nil, fmt.Errorf("decode pulls: %w", err)
		}
		if len(batch) == 0 {
			break
		}
		for _, raw := range batch {
			all = append(all, raw.toPull())
		}
		if len(batch) < 100 {
			break
		}
	}
	return all, nil
}

var developmentLinkRE = regexp.MustCompile(`(?i)\b(?:fix(?:es|ed)?|close(?:s|d)?|resolve(?:s|d)?)\s*[:#]?\s*#(\d+)\b`)

var titleLinkRE = regexp.MustCompile(`^\s*\[#(\d+)\]`)

// PullLinksSubIssue reports whether a PR satisfies GitHub-style **development linkage** to a Sub-issue.
func PullLinksSubIssue(pr PullRequest, subIssue int) bool {
	if m := titleLinkRE.FindStringSubmatch(pr.Title); len(m) == 2 {
		if n, err := strconv.Atoi(m[1]); err == nil && n == subIssue {
			return true
		}
	}
	for _, m := range developmentLinkRE.FindAllStringSubmatch(pr.Body, -1) {
		if len(m) != 2 {
			continue
		}
		n, err := strconv.Atoi(m[1])
		if err == nil && n == subIssue {
			return true
		}
	}
	return false
}

// HasForgeLabel reports whether the PR carries the **`forge`** label.
func HasForgeLabel(pr PullRequest) bool {
	for _, lb := range pr.Labels {
		if strings.EqualFold(lb.Name, "forge") {
			return true
		}
	}
	return false
}
