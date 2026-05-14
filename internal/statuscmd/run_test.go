package statuscmd

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/thisisthemurph/forge/internal/cli"
	"github.com/thisisthemurph/forge/internal/githubapi"
)

func TestRun_statusWithRepoOverride(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/up/stream/issues/99/sub_issues":
			w.Write([]byte(`[{"number":1,"title":"Child","state":"open","body":""}]`))
		case "/repos/up/stream/pulls":
			if r.URL.Query().Get("head") != "up:forge/feature/99/issue/1/child" {
				t.Fatalf("unexpected pulls head %q", r.URL.Query().Get("head"))
			}
			w.Write([]byte(`[]`))
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	t.Cleanup(srv.Close)

	cfg := cli.Config{RepoOverride: "up/stream", Feature: 99}
	getenv := func(k string) string {
		if k == "GITHUB_TOKEN" {
			return "test-token"
		}
		return ""
	}
	var out strings.Builder
	client := &githubapi.Client{BaseURL: srv.URL, HTTP: srv.Client()}
	err := Run(context.Background(), cfg, t.TempDir(), getenv, nil, client, &out)
	if err != nil {
		t.Fatal(err)
	}
	got := out.String()
	if !strings.Contains(got, "Feature #99") || !strings.Contains(got, "#1") || !strings.Contains(got, "Child") {
		t.Fatalf("output:\n%s", got)
	}
	if !strings.Contains(got, "Stack order") {
		t.Fatalf("expected Stack order section:\n%s", got)
	}
	if !strings.Contains(got, "Feature branch:") || !strings.Contains(got, "forge/feature/99/base") {
		t.Fatalf("expected Feature branch line:\n%s", got)
	}
	if !strings.Contains(got, "forge/feature/99/issue/1/child") {
		t.Fatalf("expected Stacked branch for #1:\n%s", got)
	}
	if !strings.Contains(got, "Scheduler") || !strings.Contains(got, "Next planned work: Sub-issue #1") {
		t.Fatalf("expected scheduler section:\n%s", got)
	}
}

func TestRun_statusStackOrderWithBlockers(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/up/stream/issues/50/sub_issues":
			payload := `[
  {"number":10,"title":"Alpha","state":"open","body":""},
  {"number":20,"title":"Beta","state":"open","body":"## Blocked by\n- #10\n"},
  {"number":30,"title":"Gamma","state":"open","body":"## Blocked by\n- #10\n"}
]`
			w.Write([]byte(payload))
		case "/repos/up/stream/pulls":
			w.Write([]byte(`[]`))
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	t.Cleanup(srv.Close)

	cfg := cli.Config{RepoOverride: "up/stream", Feature: 50}
	getenv := func(k string) string {
		if k == "GITHUB_TOKEN" {
			return "test-token"
		}
		return ""
	}
	var out strings.Builder
	client := &githubapi.Client{BaseURL: srv.URL, HTTP: srv.Client()}
	err := Run(context.Background(), cfg, t.TempDir(), getenv, nil, client, &out)
	if err != nil {
		t.Fatal(err)
	}
	got := out.String()
	// After #10, both #20 and #30 are ready; ascending tie-break => 20 before 30
	wantSeq := []string{"#10", "#20", "#30"}
	idx := strings.Index(got, "Stack order")
	if idx < 0 {
		t.Fatalf("missing stack order:\n%s", got)
	}
	tail := got[idx:]
	for _, w := range wantSeq {
		if !strings.Contains(tail, w) {
			t.Fatalf("expected %q in stack section:\n%s", w, got)
		}
	}
	// Ensure numeric order: #20 appears before #30 in the stack section
	pos20 := strings.Index(tail, "#20")
	pos30 := strings.Index(tail, "#30")
	if pos20 < 0 || pos30 < 0 || pos20 >= pos30 {
		t.Fatalf("expected #20 before #30 in stack section:\n%s", got)
	}
	branchLines := []string{
		"forge/feature/50/base",
		"forge/feature/50/issue/10/alpha",
		"forge/feature/50/issue/20/beta",
		"forge/feature/50/issue/30/gamma",
	}
	for _, line := range branchLines {
		if !strings.Contains(got, line) {
			t.Fatalf("expected branch name %q in output:\n%s", line, got)
		}
	}
	// Stacked branch lines follow stack order: alpha before beta before gamma
	posAlpha := strings.Index(got, "issue/10/alpha")
	posBeta := strings.Index(got, "issue/20/beta")
	posGamma := strings.Index(got, "issue/30/gamma")
	if posAlpha < 0 || posBeta < 0 || posGamma < 0 || !(posAlpha < posBeta && posBeta < posGamma) {
		t.Fatalf("expected stacked branches in stack order:\n%s", got)
	}
	if !strings.Contains(got, "Next planned work: Sub-issue #10") {
		t.Fatalf("expected next work on #10:\n%s", got)
	}
}
