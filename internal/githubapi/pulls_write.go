package githubapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// CreatePullRequest opens a pull request via the GitHub REST API.
// draft should be false for **PR draft policy (v1)** (ready for review).
func (c *Client) CreatePullRequest(ctx context.Context, owner, repo, title, head, base, body string, draft bool) (int, error) {
	baseURL := strings.TrimSuffix(c.BaseURL, "/")
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	httpClient := c.HTTP
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	u := fmt.Sprintf("%s/repos/%s/%s/pulls", baseURL, url.PathEscape(owner), url.PathEscape(repo))

	payload := map[string]any{
		"title": title,
		"head":  head,
		"base":  base,
		"body":  body,
		"draft": draft,
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return 0, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(raw))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("User-Agent", "forge-cli/0")
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	bodyBytes, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return 0, err
	}
	if resp.StatusCode != http.StatusCreated {
		msg := extractMessage(bodyBytes)
		return 0, fmt.Errorf("GitHub API create pull returned %s: %s", resp.Status, msg)
	}
	var out struct {
		Number int `json:"number"`
	}
	if err := json.Unmarshal(bodyBytes, &out); err != nil {
		return 0, fmt.Errorf("decode create pull response: %w", err)
	}
	if out.Number == 0 {
		return 0, fmt.Errorf("create pull response missing number")
	}
	return out.Number, nil
}

// UpdatePullRequest updates fields on an open pull request.
// Only non-nil pointer fields are sent in the JSON body.
func (c *Client) UpdatePullRequest(ctx context.Context, owner, repo string, pullNumber int, title, body, base *string, draft *bool) error {
	baseURL := strings.TrimSuffix(c.BaseURL, "/")
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	httpClient := c.HTTP
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	u := fmt.Sprintf("%s/repos/%s/%s/pulls/%d", baseURL, url.PathEscape(owner), url.PathEscape(repo), pullNumber)

	payload := map[string]any{}
	if title != nil {
		payload["title"] = *title
	}
	if body != nil {
		payload["body"] = *body
	}
	if base != nil {
		payload["base"] = *base
	}
	if draft != nil {
		payload["draft"] = *draft
	}
	if len(payload) == 0 {
		return fmt.Errorf("UpdatePullRequest: no fields to update")
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, u, bytes.NewReader(raw))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("User-Agent", "forge-cli/0")
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	bodyBytes, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		msg := extractMessage(bodyBytes)
		return fmt.Errorf("GitHub API update pull returned %s: %s", resp.Status, msg)
	}
	return nil
}

// AddIssueLabels adds labels to an issue or pull request (PRs use the issues endpoint).
func (c *Client) AddIssueLabels(ctx context.Context, owner, repo string, issueOrPRNumber int, labels []string) error {
	if len(labels) == 0 {
		return nil
	}
	baseURL := strings.TrimSuffix(c.BaseURL, "/")
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	httpClient := c.HTTP
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	u := fmt.Sprintf("%s/repos/%s/%s/issues/%d/labels", baseURL, url.PathEscape(owner), url.PathEscape(repo), issueOrPRNumber)

	payload := map[string]any{"labels": labels}
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(raw))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("User-Agent", "forge-cli/0")
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	bodyBytes, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		msg := extractMessage(bodyBytes)
		return fmt.Errorf("GitHub API add labels returned %s: %s", resp.Status, msg)
	}
	return nil
}
