package mergesnapshot

import (
	"context"

	"github.com/thisisthemurph/forge/internal/githubapi"
	"github.com/thisisthemurph/forge/internal/naming"
)

// ManagedPR is a **Forge-managed PR** candidate after **Forge PR identification** filters.
type ManagedPR struct {
	Number  int
	BaseRef string
	Merged  bool
}

// ForgeLookup is the outcome of classifying PRs for one **Sub-issue** (forge label + linkage).
type ForgeLookup struct {
	// Matches lists PRs that carry the forge label and are linked to the Sub-issue.
	Matches []*ManagedPR
}

// Snapshot answers GitHub-backed merge questions for scheduling tests and live callers.
type Snapshot interface {
	ForgeManagedPR(subIssueNumber int) ForgeLookup
}

// Memory is an in-memory **Snapshot** for tests.
type Memory map[int]ForgeLookup

// ForgeManagedPR implements Snapshot.
func (m Memory) ForgeManagedPR(n int) ForgeLookup {
	if m == nil {
		return ForgeLookup{}
	}
	lu, ok := m[n]
	if !ok {
		return ForgeLookup{}
	}
	return lu
}

// LoadFromGitHub builds a Snapshot by listing PRs for each Sub-issue’s deterministic **Stacked branch** head.
func LoadFromGitHub(ctx context.Context, c *githubapi.Client, owner, repo string, feature int, order []int, titles map[int]string) (Snapshot, error) {
	out := make(Memory, len(order))
	for _, n := range order {
		slug := naming.SlugFromTitle(titles[n])
		headBranch := naming.StackedBranch(feature, n, slug)
		pulls, err := c.ListPullRequestsByHead(ctx, owner, repo, headBranch)
		if err != nil {
			return nil, err
		}
		var matches []*ManagedPR
		for _, pr := range pulls {
			if !githubapi.HasForgeLabel(pr) || !githubapi.PullLinksSubIssue(pr, n) {
				continue
			}
			matches = append(matches, &ManagedPR{
				Number:  pr.Number,
				BaseRef: pr.BaseRef,
				Merged:  pr.Merged,
			})
		}
		out[n] = ForgeLookup{Matches: matches}
	}
	return out, nil
}
