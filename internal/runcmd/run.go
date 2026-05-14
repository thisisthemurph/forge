package runcmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/thisisthemurph/forge/internal/cli"
	"github.com/thisisthemurph/forge/internal/cursorrunner"
	"github.com/thisisthemurph/forge/internal/featureplan"
	"github.com/thisisthemurph/forge/internal/githubapi"
	"github.com/thisisthemurph/forge/internal/gitremote"
	"github.com/thisisthemurph/forge/internal/localgit"
	"github.com/thisisthemurph/forge/internal/naming"
	"github.com/thisisthemurph/forge/internal/prpublish"
)

// Run performs **State-driven dispatch** for a Feature: validates policy for mutating
// work (**Stack consistency policy**), prints **No pending work** when the scheduler
// has nothing executable, otherwise prepares branches, runs the **Cursor CLI**, pushes,
// and ensures a **Forge-managed PR**.
func Run(ctx context.Context, cfg cli.Config, cwd string, getenv func(string) string, ghAuth func() (string, error), client *githubapi.Client, stdin io.Reader, stdout, stderr io.Writer) error {
	if stdin == nil {
		stdin = os.Stdin
	}
	if stdout == nil {
		stdout = os.Stdout
	}
	if stderr == nil {
		stderr = os.Stderr
	}

	st, err := featureplan.Load(ctx, cfg, cwd, getenv, ghAuth, client)
	if err != nil {
		return err
	}
	if st.GraphErr != nil {
		return fmt.Errorf("cannot run: scheduling graph / DAG validation failed: %w", st.GraphErr)
	}
	if len(st.Plan.Warnings) > 0 {
		return fmt.Errorf("cannot run (stack consistency / Forge PR identification):\n- %s", strings.Join(st.Plan.Warnings, "\n- "))
	}
	if st.Plan.NextExecutable == nil {
		fmt.Fprintf(stdout, "Feature #%d: no pending work\n", cfg.Feature)
		return nil
	}

	sub := *st.Plan.NextExecutable
	order := st.Order
	stackPos := -1
	for i, n := range order {
		if n == sub {
			stackPos = i
			break
		}
	}
	if stackPos < 0 {
		return fmt.Errorf("internal error: Sub-issue #%d missing from Stack order", sub)
	}

	repoRoot, err := gitremote.FindRepoRoot(cwd)
	if err != nil {
		return err
	}

	featureBase, err := localgit.FeatureBaseBranchForMutation(repoRoot)
	if err != nil {
		return err
	}

	featureBranch := naming.FeatureBranch(cfg.Feature, "")
	if err := localgit.EnsureBranchFromExisting(repoRoot, featureBranch, featureBase); err != nil {
		return err
	}

	var parentBranch string
	if stackPos == 0 {
		parentBranch = featureBranch
	} else {
		prev := order[stackPos-1]
		parentBranch = naming.StackedBranch(cfg.Feature, prev, naming.SlugFromTitle(st.TitleByNumber[prev]))
	}
	headBranch := naming.StackedBranch(cfg.Feature, sub, naming.SlugFromTitle(st.TitleByNumber[sub]))
	if err := localgit.EnsureStackedFromParent(repoRoot, headBranch, parentBranch); err != nil {
		return err
	}

	bin, err := cursorrunner.ResolveBin(getenv, exec.LookPath)
	if err != nil {
		return err
	}

	title := st.TitleByNumber[sub]
	prompt := fmt.Sprintf(
		"Complete GitHub issue #%d (%q) for Feature #%d using TDD (red-green-refactor). Work only in this repository checkout.",
		sub, title, cfg.Feature,
	)
	// Headless automation flags match Cursor CLI print mode (see Cursor CLI docs).
	cursorArgs := []string{"-p", "--force", prompt}

	fmt.Fprintf(stdout, "forge: running Cursor for Sub-issue #%d\n", sub)
	if err := cursorrunner.Run(ctx, cursorrunner.Config{
		Bin:      bin,
		RepoRoot: repoRoot,
		Env:      cursorrunner.FilterEnv(os.Environ()),
		Stdin:    stdin,
		Stdout:   stdout,
		Stderr:   stderr,
	}, cursorArgs...); err != nil {
		return err
	}

	if err := localgit.PushOriginBranch(repoRoot, headBranch); err != nil {
		return err
	}

	_, err = prpublish.EnsureForgeManagedPR(ctx, client, prpublish.EnsureInput{
		Owner:         st.Owner,
		Repo:          st.Repo,
		SubIssue:      sub,
		SubIssueTitle: title,
		HeadBranch:    headBranch,
		BaseBranch:    parentBranch,
	})
	if err != nil {
		return err
	}

	return nil
}
