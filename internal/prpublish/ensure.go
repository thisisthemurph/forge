package prpublish

import (
	"context"
	"fmt"
	"strings"

	"github.com/thisisthemurph/forge/internal/githubapi"
)

// EnsureInput selects repository coordinates, Sub-issue metadata, and git refs
// for **Git publish and PR open (v1)**.
//
// HeadBranch must already exist on **origin** before calling EnsureForgeManagedPR
// (push via localgit.PushOriginBranch first).
type EnsureInput struct {
	Owner, Repo string

	SubIssue      int
	SubIssueTitle string

	// HeadBranch is the short branch name pushed to origin (stacked branch).
	HeadBranch string
	// BaseBranch is the **PR stack target** for this Sub-issue.
	BaseBranch string
}

// EnsureForgeManagedPR creates or updates the **Forge-managed PR** for a Sub-issue:
// **`forge`** label, **PR stack target** as base, **PR title (v1)**, **PR body linking (v1)**,
// and **PR draft policy (v1)** (not draft). It never merges the PR.
//
// If multiple open **Forge-managed PR** candidates exist for the same head, it returns an error.
func EnsureForgeManagedPR(ctx context.Context, c *githubapi.Client, in EnsureInput) (pullNumber int, err error) {
	if strings.TrimSpace(in.Owner) == "" || strings.TrimSpace(in.Repo) == "" {
		return 0, fmt.Errorf("owner and repo are required")
	}
	if in.SubIssue < 1 {
		return 0, fmt.Errorf("sub-issue number must be positive")
	}
	if strings.TrimSpace(in.HeadBranch) == "" || strings.TrimSpace(in.BaseBranch) == "" {
		return 0, fmt.Errorf("head and base branch are required")
	}

	headKey := in.Owner + ":" + in.HeadBranch
	pulls, err := c.ListPullRequestsByHead(ctx, in.Owner, in.Repo, in.HeadBranch)
	if err != nil {
		return 0, err
	}

	var open []githubapi.PullRequest
	for _, pr := range pulls {
		if strings.EqualFold(pr.State, "open") && !pr.Merged {
			open = append(open, pr)
		}
	}

	var forgeOpen []githubapi.PullRequest
	for _, pr := range open {
		if githubapi.HasForgeLabel(pr) && githubapi.PullLinksSubIssue(pr, in.SubIssue) {
			forgeOpen = append(forgeOpen, pr)
		}
	}

	wantTitle := ForgeManagedPRTitle(in.SubIssue, in.SubIssueTitle)
	wantBody := ForgeManagedPRBody(in.SubIssue)

	switch len(forgeOpen) {
	case 0:
		if len(open) > 1 {
			return 0, fmt.Errorf("multiple open pull requests for head %q; cannot pick a Forge-managed PR", headKey)
		}
		if len(open) == 1 {
			pr := open[0]
			if err := reconcileOpenPR(ctx, c, in.Owner, in.Repo, pr.Number, wantTitle, wantBody, in.BaseBranch); err != nil {
				return 0, err
			}
			if err := c.AddIssueLabels(ctx, in.Owner, in.Repo, pr.Number, []string{"forge"}); err != nil {
				return 0, err
			}
			return pr.Number, nil
		}
		num, err := c.CreatePullRequest(ctx, in.Owner, in.Repo, wantTitle, in.HeadBranch, in.BaseBranch, wantBody, false)
		if err != nil {
			return 0, err
		}
		if err := c.AddIssueLabels(ctx, in.Owner, in.Repo, num, []string{"forge"}); err != nil {
			return 0, err
		}
		return num, nil
	case 1:
		pr := forgeOpen[0]
		if err := reconcileOpenPR(ctx, c, in.Owner, in.Repo, pr.Number, wantTitle, wantBody, in.BaseBranch); err != nil {
			return 0, err
		}
		if err := c.AddIssueLabels(ctx, in.Owner, in.Repo, pr.Number, []string{"forge"}); err != nil {
			return 0, err
		}
		return pr.Number, nil
	default:
		return 0, fmt.Errorf("multiple Forge-managed pull requests for Sub-issue #%d on head %q", in.SubIssue, headKey)
	}
}

func reconcileOpenPR(ctx context.Context, c *githubapi.Client, owner, repo string, pullNumber int, title, body, base string) error {
	draft := false
	t := title
	b := body
	br := base
	return c.UpdatePullRequest(ctx, owner, repo, pullNumber, &t, &b, &br, &draft)
}
