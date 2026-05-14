package cli

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/thisisthemurph/forge/internal/githubapi"
)

func TestExecute_argvErrors(t *testing.T) {
	t.Parallel()
	deps := Deps{
		Getenv: func(string) string { return "" },
		GhAuth: func() (string, error) { return "", nil },
		Client: &githubapi.Client{},
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
		Getwd:  func() (string, error) { return t.TempDir(), nil },
		ExecStatus: func(context.Context, Config, string) error {
			panic("ExecStatus must not be called in this test")
		},
		ExecRun: func(context.Context, Config, string) error {
			panic("ExecRun must not be called in this test")
		},
	}

	tests := []struct {
		name           string
		argv           []string
		wantErr        string
		wantErrSubstr  string
		wantNoExactErr bool
	}{
		{
			name:    "bare forge",
			argv:    []string{},
			wantErr: "usage: forge [flags] (status|run) <feature-issue-number>",
		},
		{
			name:    "unknown command",
			argv:    []string{"deploy", "1"},
			wantErr: `unknown command "deploy" (expected status or run)`,
		},
		{
			name:    "status missing number",
			argv:    []string{"status"},
			wantErr: "status requires feature issue number",
		},
		{
			name:    "run missing number",
			argv:    []string{"run"},
			wantErr: "run requires feature issue number",
		},
		{
			name:    "bad number",
			argv:    []string{"status", "x"},
			wantErr: "feature issue number must be a positive integer",
		},
		{
			name:    "extra args",
			argv:    []string{"status", "1", "extra"},
			wantErr: "unexpected arguments after feature issue number",
		},
		{
			name:           "repo without value",
			argv:           []string{"--repo"},
			wantNoExactErr: true,
			wantErrSubstr:  "repo",
		},
		{
			name:    "repo empty after equals",
			argv:    []string{"status", "1", "--repo="},
			wantErr: "--repo requires owner/name",
		},
		{
			name:    "repo whitespace only",
			argv:    []string{"status", "1", "--repo", "  "},
			wantErr: "--repo requires owner/name",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := NewRootCmd(deps)
			root.SetArgs(tt.argv)
			err := root.ExecuteContext(context.Background())
			if err == nil {
				t.Fatal("expected error")
			}
			if tt.wantNoExactErr {
				if tt.wantErrSubstr != "" && !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(tt.wantErrSubstr)) {
					t.Fatalf("error %q should mention %q", err.Error(), tt.wantErrSubstr)
				}
				return
			}
			if tt.wantErr != "" && err.Error() != tt.wantErr {
				t.Fatalf("error %q, want %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestExecute_statusInvokesExecStatus(t *testing.T) {
	t.Parallel()
	var saw bool
	deps := Deps{
		Getenv: func(string) string { return "" },
		GhAuth: func() (string, error) { return "", nil },
		Client: &githubapi.Client{},
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
		Getwd:  func() (string, error) { return t.TempDir(), nil },
		ExecStatus: func(_ context.Context, cfg Config, _ string) error {
			saw = true
			if cfg.RepoOverride != "o/r" || cfg.Feature != 3 {
				t.Fatalf("cfg = %+v", cfg)
			}
			return fmt.Errorf("stub-status")
		},
		ExecRun: func(context.Context, Config, string) error {
			panic("ExecRun must not be called")
		},
	}
	root := NewRootCmd(deps)
	root.SetArgs([]string{"--repo=o/r", "status", "3"})
	err := root.ExecuteContext(context.Background())
	if !saw {
		t.Fatal("expected ExecStatus to run")
	}
	if err == nil || err.Error() != "stub-status" {
		t.Fatalf("err = %v", err)
	}
}

func TestExecute_repoFlagFlexiblePlacement(t *testing.T) {
	t.Parallel()
	cases := [][]string{
		{"--repo=o/r", "status", "3"},
		{"status", "3", "--repo", "o/r"},
		{"status", "--repo", "o/r", "3"},
	}
	for _, argv := range cases {
		var got string
		deps := Deps{
			Getenv: func(string) string { return "" },
			GhAuth: func() (string, error) { return "", nil },
			Client: &githubapi.Client{},
			Stdout: &bytes.Buffer{},
			Stderr: &bytes.Buffer{},
			Getwd:  func() (string, error) { return t.TempDir(), nil },
			ExecStatus: func(_ context.Context, cfg Config, _ string) error {
				got = cfg.RepoOverride
				return fmt.Errorf("stop")
			},
			ExecRun: func(context.Context, Config, string) error {
				panic("ExecRun must not be called")
			},
		}
		root := NewRootCmd(deps)
		root.SetArgs(argv)
		_ = root.ExecuteContext(context.Background())
		if got != "o/r" {
			t.Fatalf("argv=%v repo=%q", argv, got)
		}
	}
}

func TestExecute_repoFlagLastWins(t *testing.T) {
	t.Parallel()
	var got string
	deps := Deps{
		Getenv: func(string) string { return "" },
		GhAuth: func() (string, error) { return "", nil },
		Client: &githubapi.Client{},
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
		Getwd:  func() (string, error) { return t.TempDir(), nil },
		ExecStatus: func(_ context.Context, cfg Config, _ string) error {
			got = cfg.RepoOverride
			return fmt.Errorf("stop")
		},
		ExecRun: func(context.Context, Config, string) error {
			panic("ExecRun must not be called")
		},
	}
	root := NewRootCmd(deps)
	root.SetArgs([]string{"--repo", "a/b", "--repo=c/d", "status", "1"})
	_ = root.ExecuteContext(context.Background())
	if got != "c/d" {
		t.Fatalf("repo=%q want c/d", got)
	}
}
