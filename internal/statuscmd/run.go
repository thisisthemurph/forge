package statuscmd

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/thisisthemurph/forge/internal/cli"
	"github.com/thisisthemurph/forge/internal/githubapi"
	"github.com/thisisthemurph/forge/internal/gitremote"
	"github.com/thisisthemurph/forge/internal/remote"
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
	return nil
}
