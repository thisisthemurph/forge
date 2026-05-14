package runcmd

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/thisisthemurph/forge/internal/cli"
	"github.com/thisisthemurph/forge/internal/githubapi"
)

func TestRun_noPendingWork_emptySubIssues(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/repos/up/stream/issues/77/sub_issues" {
			w.Write([]byte(`[]`))
			return
		}
		t.Fatalf("unexpected path %s", r.URL.Path)
	}))
	t.Cleanup(srv.Close)

	cfg := cli.Config{RepoOverride: "up/stream", Feature: 77}
	getenv := func(k string) string {
		if k == "GITHUB_TOKEN" {
			return "test-token"
		}
		return ""
	}
	var out strings.Builder
	client := &githubapi.Client{BaseURL: srv.URL, HTTP: srv.Client()}
	err := Run(context.Background(), cfg, t.TempDir(), getenv, nil, client, nil, &out, &out)
	if err != nil {
		t.Fatal(err)
	}
	if got := out.String(); got != "Feature #77: no pending work\n" {
		t.Fatalf("unexpected output %q", got)
	}
}

func TestRun_graphErrorIsFatal(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/repos/up/stream/issues/42/sub_issues" {
			w.Write([]byte(`[{"number":2,"title":"Bad","state":"open","body":"## Blocked by\n- #42\n"}]`))
			return
		}
		t.Fatalf("unexpected path %s", r.URL.Path)
	}))
	t.Cleanup(srv.Close)

	cfg := cli.Config{RepoOverride: "up/stream", Feature: 42}
	getenv := func(k string) string {
		if k == "GITHUB_TOKEN" {
			return "test-token"
		}
		return ""
	}
	var out strings.Builder
	client := &githubapi.Client{BaseURL: srv.URL, HTTP: srv.Client()}
	err := Run(context.Background(), cfg, t.TempDir(), getenv, nil, client, nil, &out, &out)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "DAG validation") {
		t.Fatalf("expected DAG validation in error, got %v", err)
	}
}

func TestRun_stackWarningsAreFatal(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/up/stream/issues/1/sub_issues":
			w.Write([]byte(`[{"number":5,"title":"Do thing","state":"open","body":""}]`))
		case "/repos/up/stream/pulls":
			if r.URL.Query().Get("head") != "up:forge/feature/1/issue/5/do-thing" {
				t.Fatalf("unexpected pulls head %q", r.URL.Query().Get("head"))
			}
			payload := `[
			  {"number":10,"title":"[#5] a","body":"Fixes #5","state":"open","merged":false,"base":{"ref":"forge/feature/1/base"},"labels":[{"name":"forge"}]},
			  {"number":11,"title":"[#5] b","body":"Fixes #5","state":"open","merged":false,"base":{"ref":"forge/feature/1/base"},"labels":[{"name":"forge"}]}
			]`
			w.Write([]byte(payload))
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	t.Cleanup(srv.Close)

	cfg := cli.Config{RepoOverride: "up/stream", Feature: 1}
	getenv := func(k string) string {
		if k == "GITHUB_TOKEN" {
			return "test-token"
		}
		return ""
	}
	var out strings.Builder
	client := &githubapi.Client{BaseURL: srv.URL, HTTP: srv.Client()}
	err := Run(context.Background(), cfg, t.TempDir(), getenv, nil, client, nil, &out, &out)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "Forge PR identification") && !strings.Contains(err.Error(), "ambiguous") {
		t.Fatalf("expected Forge PR identification error, got %v", err)
	}
}
