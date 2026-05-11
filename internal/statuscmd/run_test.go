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
		if r.URL.Path != "/repos/up/stream/issues/99/sub_issues" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		w.Write([]byte(`[{"number":1,"title":"Child","state":"open","body":""}]`))
	}))
	t.Cleanup(srv.Close)

	cfg := cli.Config{RepoOverride: "up/stream", Subcommand: "status", Feature: 99}
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
}

func TestRun_statusStackOrderWithBlockers(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/up/stream/issues/50/sub_issues" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		payload := `[
  {"number":10,"title":"A","state":"open","body":""},
  {"number":20,"title":"B","state":"open","body":"## Blocked by\n- #10\n"},
  {"number":30,"title":"C","state":"open","body":"## Blocked by\n- #10\n"}
]`
		w.Write([]byte(payload))
	}))
	t.Cleanup(srv.Close)

	cfg := cli.Config{RepoOverride: "up/stream", Subcommand: "status", Feature: 50}
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
}
