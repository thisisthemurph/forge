package statuscmd

import (
	"context"
	"fmt"
	"io"

	"github.com/thisisthemurph/forge/internal/cli"
	"github.com/thisisthemurph/forge/internal/featureplan"
	"github.com/thisisthemurph/forge/internal/githubapi"
	"github.com/thisisthemurph/forge/internal/naming"
)

// Run executes the read-only status flow for a Feature issue.
func Run(ctx context.Context, cfg cli.Config, cwd string, getenv func(string) string, ghAuth func() (string, error), client *githubapi.Client, stdout io.Writer) error {
	if cfg.Subcommand != "status" {
		return fmt.Errorf("internal error: unsupported command %q", cfg.Subcommand)
	}

	st, err := featureplan.Load(ctx, cfg, cwd, getenv, ghAuth, client)
	if err != nil {
		return err
	}

	fmt.Fprintf(stdout, "Feature #%d in %s/%s\n", cfg.Feature, st.Owner, st.Repo)
	if len(st.SubIssues) == 0 {
		fmt.Fprintf(stdout, "No sub-issues attached.\n")
		return nil
	}
	fmt.Fprintf(stdout, "Sub-issues (%d):\n", len(st.SubIssues))
	for _, s := range st.SubIssues {
		fmt.Fprintf(stdout, "  #%d\t%s\t%s\n", s.Number, s.State, s.Title)
	}

	if st.GraphErr != nil {
		fmt.Fprintf(stdout, "\nWarnings (Stack consistency policy on status):\n")
		fmt.Fprintf(stdout, "  - Scheduling graph / DAG validation: %v\n", st.GraphErr)
		return nil
	}
	order := st.Order
	fmt.Fprintf(stdout, "\nStack order (Scheduling graph):\n")
	for _, n := range order {
		fmt.Fprintf(stdout, "  #%d\n", n)
	}

	fmt.Fprintf(stdout, "\nBranch names (Deterministic branch names):\n")
	fmt.Fprintf(stdout, "  Feature branch: %s\n", naming.FeatureBranch(cfg.Feature, ""))
	for _, n := range order {
		slug := naming.SlugFromTitle(st.TitleByNumber[n])
		fmt.Fprintf(stdout, "  #%d (stacked): %s\n", n, naming.StackedBranch(cfg.Feature, n, slug))
	}

	if len(st.Plan.Warnings) > 0 {
		fmt.Fprintf(stdout, "\nWarnings (Stack consistency policy on status):\n")
		for _, w := range st.Plan.Warnings {
			fmt.Fprintf(stdout, "  - %s\n", w)
		}
	}

	fmt.Fprintf(stdout, "\nScheduler (merge snapshot + Stack order):\n")
	if st.Plan.NextExecutable != nil {
		fmt.Fprintf(stdout, "  Next planned work: Sub-issue #%d\n", *st.Plan.NextExecutable)
	} else {
		fmt.Fprintf(stdout, "  Next planned work: none (no open Stack position matched **Executable** rules)\n")
	}
	return nil
}
