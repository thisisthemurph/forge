package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/thisisthemurph/forge/internal/githubapi"
)

const repoFlagName = "repo"

// Deps are runtime dependencies for the forge CLI (wired from cmd/forge).
type Deps struct {
	Getenv     func(string) string
	GhAuth     func() (string, error)
	Client     *githubapi.Client
	Stdin      io.Reader
	Stdout     io.Writer
	Stderr     io.Writer
	Getwd      func() (string, error)
	ExecStatus func(ctx context.Context, cfg Config, cwd string) error
	ExecRun    func(ctx context.Context, cfg Config, cwd string) error
}

func repoOverrideFromCmd(cmd *cobra.Command) (string, error) {
	repo, err := cmd.Flags().GetString(repoFlagName)
	if err != nil {
		return "", err
	}
	repo = strings.TrimSpace(repo)
	fl := cmd.Flags().Lookup(repoFlagName)
	if fl != nil && fl.Changed && repo == "" {
		return "", fmt.Errorf("--repo requires owner/name")
	}
	return repo, nil
}

// NewRootCmd builds the Cobra command tree for forge.
func NewRootCmd(deps Deps) *cobra.Command {
	if deps.Getwd == nil {
		deps.Getwd = os.Getwd
	}
	if deps.Stdin == nil {
		deps.Stdin = os.Stdin
	}
	if deps.Stdout == nil {
		deps.Stdout = os.Stdout
	}
	if deps.Stderr == nil {
		deps.Stderr = os.Stderr
	}

	root := &cobra.Command{
		Use:           "forge",
		Short:         "Forge automates GitHub-backed feature work with Cursor.",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.ArbitraryArgs,
		RunE: func(_ *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("usage: forge [flags] (status|run) <feature-issue-number>")
			}
			return fmt.Errorf("unknown command %q (expected status or run)", args[0])
		},
	}
	root.SetOut(deps.Stdout)
	root.SetErr(deps.Stderr)
	root.CompletionOptions.DisableDefaultCmd = true

	root.PersistentFlags().String(
		repoFlagName,
		"",
		`GitHub "owner/repo" for API calls (overrides origin).`,
	)

	// Avoid registering a subcommand named "help" so `forge help` matches legacy "unknown command".
	root.SetHelpCommand(&cobra.Command{
		Use:    "__forge_internal_help_placeholder",
		Hidden: true,
		RunE:   func(*cobra.Command, []string) error { return nil },
	})

	featureArgs := func(sub string) cobra.PositionalArgs {
		return func(_ *cobra.Command, args []string) error {
			switch len(args) {
			case 0:
				return fmt.Errorf("%s requires feature issue number", sub)
			case 1:
				return nil
			default:
				return fmt.Errorf("unexpected arguments after feature issue number")
			}
		}
	}

	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Print read-only status for a Feature issue",
		Args:  featureArgs("status"),
		RunE: func(cmd *cobra.Command, args []string) error {
			n, err := strconv.Atoi(args[0])
			if err != nil || n < 1 {
				return fmt.Errorf("feature issue number must be a positive integer")
			}
			repoOverride, err := repoOverrideFromCmd(cmd)
			if err != nil {
				return err
			}
			ctx, cancel := context.WithTimeout(cmd.Context(), 60*time.Second)
			defer cancel()

			cwd, err := deps.Getwd()
			if err != nil {
				return fmt.Errorf("get working directory: %w", err)
			}
			cfg := Config{RepoOverride: repoOverride, Feature: n}
			return deps.ExecStatus(ctx, cfg, cwd)
		},
	}

	runCmd := &cobra.Command{
		Use:   "run",
		Short: "Run state-driven dispatch for a Feature issue (mutating)",
		Args:  featureArgs("run"),
		RunE: func(cmd *cobra.Command, args []string) error {
			n, err := strconv.Atoi(args[0])
			if err != nil || n < 1 {
				return fmt.Errorf("feature issue number must be a positive integer")
			}
			repoOverride, err := repoOverrideFromCmd(cmd)
			if err != nil {
				return err
			}
			ctx, cancel := context.WithTimeout(cmd.Context(), 2*time.Hour)
			defer cancel()

			cwd, err := deps.Getwd()
			if err != nil {
				return fmt.Errorf("get working directory: %w", err)
			}
			cfg := Config{RepoOverride: repoOverride, Feature: n}
			return deps.ExecRun(ctx, cfg, cwd)
		},
	}

	root.AddCommand(statusCmd, runCmd)
	return root
}
