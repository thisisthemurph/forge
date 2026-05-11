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
		w.Write([]byte(`[{"number":1,"title":"Child","state":"open"}]`))
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
}
