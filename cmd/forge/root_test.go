package main

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/thisisthemurph/forge/internal/githubapi"
)

func TestExecute_argvErrors(t *testing.T) {
	t.Parallel()
	deps := &forgeDeps{
		Getenv: func(string) string { return "" },
		GhAuth: func() (string, error) { return "", nil },
		Client: &githubapi.Client{},
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
		Getwd:  func() (string, error) { return t.TempDir(), nil },
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
			root := newRootCmd(deps)
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

func TestExecute_statusWithHTTPServer(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/o/r/issues/3/sub_issues" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		w.Write([]byte(`[]`))
	}))
	t.Cleanup(srv.Close)

	var out strings.Builder
	deps := &forgeDeps{
		Getenv: func(k string) string {
			if k == "GITHUB_TOKEN" {
				return "test-token"
			}
			return ""
		},
		GhAuth: nil,
		Client: &githubapi.Client{BaseURL: srv.URL, HTTP: srv.Client()},
		Stdout: &out,
		Stderr: &out,
		Getwd:  func() (string, error) { return t.TempDir(), nil },
	}

	root := newRootCmd(deps)
	root.SetArgs([]string{"--repo=o/r", "status", "3"})
	err := root.ExecuteContext(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	got := out.String()
	if !strings.Contains(got, "Feature #3") || !strings.Contains(got, "No sub-issues attached") {
		t.Fatalf("output:\n%s", got)
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
		argv := argv
		t.Run(strings.Join(argv, " "), func(t *testing.T) {
			t.Parallel()
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/repos/o/r/issues/3/sub_issues" {
					t.Errorf("unexpected path %s", r.URL.Path)
					return
				}
				w.Write([]byte(`[]`))
			}))
			t.Cleanup(srv.Close)

			var out strings.Builder
			deps := &forgeDeps{
				Getenv: func(k string) string {
					if k == "GITHUB_TOKEN" {
						return "test-token"
					}
					return ""
				},
				GhAuth: nil,
				Client: &githubapi.Client{BaseURL: srv.URL, HTTP: srv.Client()},
				Stdout: &out,
				Stderr: &out,
				Getwd:  func() (string, error) { return t.TempDir(), nil },
			}
			root := newRootCmd(deps)
			root.SetArgs(argv)
			err := root.ExecuteContext(context.Background())
			if err != nil {
				t.Fatalf("err=%v", err)
			}
			if !strings.Contains(out.String(), "Feature #3") {
				t.Fatalf("output:\n%s", out.String())
			}
		})
	}
}

func TestExecute_repoFlagLastWins(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/c/d/issues/1/sub_issues" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		w.Write([]byte(`[]`))
	}))
	t.Cleanup(srv.Close)

	var out strings.Builder
	deps := &forgeDeps{
		Getenv: func(k string) string {
			if k == "GITHUB_TOKEN" {
				return "test-token"
			}
			return ""
		},
		GhAuth: nil,
		Client: &githubapi.Client{BaseURL: srv.URL, HTTP: srv.Client()},
		Stdout: &out,
		Stderr: &out,
		Getwd:  func() (string, error) { return t.TempDir(), nil },
	}

	root := newRootCmd(deps)
	root.SetArgs([]string{"--repo", "a/b", "--repo=c/d", "status", "1"})
	err := root.ExecuteContext(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "Feature #1") {
		t.Fatalf("output:\n%s", out.String())
	}
}
