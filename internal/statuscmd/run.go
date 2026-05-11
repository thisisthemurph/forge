package statuscmd

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/thisisthemurph/forge/internal/cli"
	"github.com/thisisthemurph/forge/internal/githubapi"
	"github.com/thisisthemurph/forge/internal/gitremote"
	"github.com/thisisthemurph/forge/internal/mergesnapshot"
	"github.com/thisisthemurph/forge/internal/naming"
	"github.com/thisisthemurph/forge/internal/remote"
	"github.com/thisisthemurph/forge/internal/scheduler"
	"github.com/thisisthemurph/forge/internal/scheduling"
	"github.com/thisisthemurph/forge/internal/token"
)

// Run executes the read-only status flow for a Feature issue.
func Run(ctx context.Context, cfg cli.Config, cwd string, getenv func(string) string, ghAuth func() (string, error), client *githubapi.Client, stdout io.Writer) error {
	if cfg.Subcommand != "status" {
		return fmt.Errorf("internal error: unsupported command %q", cfg.Subcommand)
	}

	var owner, repo string
	var err error
	if strings.TrimSpace(cfg.RepoOverride) != "" {
		owner, repo, err = remote.ParseRepoOwnerPath(cfg.RepoOverride)
		if err != nil {
			return err
		}
	} else {
		repoRoot, err := gitremote.FindRepoRoot(cwd)
		if err != nil {
			return err
		}
		raw, err := gitremote.GetOriginURL(repoRoot)
		if err != nil {
			return fmt.Errorf("resolve origin: %w", err)
		}
		owner, repo, err = remote.OwnerRepoFromURL(raw)
		if err != nil {
			return fmt.Errorf("resolve GitHub repository from origin: %w", err)
		}
	}

	tok, err := token.Resolve(getenv, ghAuth)
	if err != nil {
		return err
	}
	client.Token = tok

	subs, err := client.ListAllSubIssues(ctx, owner, repo, cfg.Feature)
	if err != nil {
		return err
	}

	fmt.Fprintf(stdout, "Feature #%d in %s/%s\n", cfg.Feature, owner, repo)
	if len(subs) == 0 {
		fmt.Fprintf(stdout, "No sub-issues attached.\n")
		return nil
	}
	fmt.Fprintf(stdout, "Sub-issues (%d):\n", len(subs))
	for _, s := range subs {
		fmt.Fprintf(stdout, "  #%d\t%s\t%s\n", s.Number, s.State, s.Title)
	}

	inputs := make([]scheduling.SubIssueInput, 0, len(subs))
	for _, s := range subs {
		inputs = append(inputs, scheduling.SubIssueInput{Number: s.Number, Body: s.Body})
	}
	graph, err := scheduling.AnalyzeGraph(cfg.Feature, inputs)
	if err != nil {
		fmt.Fprintf(stdout, "\nScheduling graph:\n  error: %v\n", err)
		return nil
	}
	order := graph.Order
	fmt.Fprintf(stdout, "\nStack order (Scheduling graph):\n")
	for _, n := range order {
		fmt.Fprintf(stdout, "  #%d\n", n)
	}

	titleByNumber := make(map[int]string, len(subs))
	for _, s := range subs {
		titleByNumber[s.Number] = s.Title
	}
	fmt.Fprintf(stdout, "\nBranch names (Deterministic branch names):\n")
	fmt.Fprintf(stdout, "  Feature branch: %s\n", naming.FeatureBranch(cfg.Feature, ""))
	for _, n := range order {
		slug := naming.SlugFromTitle(titleByNumber[n])
		fmt.Fprintf(stdout, "  #%d (stacked): %s\n", n, naming.StackedBranch(cfg.Feature, n, slug))
	}

	snap, err := mergesnapshot.LoadFromGitHub(ctx, client, owner, repo, cfg.Feature, order, titleByNumber)
	if err != nil {
		return err
	}
	plan := scheduler.BuildPlan(cfg.Feature, order, graph.Blockers, titleByNumber, snap)

	if len(plan.Warnings) > 0 {
		fmt.Fprintf(stdout, "\nWarnings (Stack consistency policy on status):\n")
		for _, w := range plan.Warnings {
			fmt.Fprintf(stdout, "  - %s\n", w)
		}
	}

	fmt.Fprintf(stdout, "\nScheduler (merge snapshot + Stack order):\n")
	if plan.NextExecutable != nil {
		fmt.Fprintf(stdout, "  Next planned work: Sub-issue #%d\n", *plan.NextExecutable)
	} else {
		fmt.Fprintf(stdout, "  Next planned work: none (no open Stack position matched **Executable** rules)\n")
	}
	return nil
}
