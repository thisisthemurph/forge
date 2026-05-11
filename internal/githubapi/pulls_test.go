package githubapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPullLinksSubIssue_titleAndBody(t *testing.T) {
	t.Parallel()
	pr := PullRequest{Title: "[#42] Do the thing", Body: "See also"}
	if !PullLinksSubIssue(pr, 42) {
		t.Fatal("expected title linkage")
	}
	pr2 := PullRequest{Title: "PR", Body: "Fixes #7\n\nDetails"}
	if !PullLinksSubIssue(pr2, 7) {
		t.Fatal("expected body linkage")
	}
	pr3 := PullRequest{Title: "x", Body: "unrelated"}
	if PullLinksSubIssue(pr3, 7) {
		t.Fatal("expected no linkage")
	}
}

func TestListPullRequestsByHead(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/o/r/pulls" {
			t.Fatalf("path %s", r.URL.Path)
		}
		if r.URL.Query().Get("head") != "o:forge/feature-1/issue-10-alpha" {
			t.Fatalf("head %q", r.URL.Query().Get("head"))
		}
		w.Write([]byte(`[{"number":9,"title":"[#10] x","body":"Fixes #10","state":"open","merged":false,"base":{"ref":"forge/feature-1"},"labels":[{"name":"forge"}]}]`))
	}))
	t.Cleanup(srv.Close)

	c := &Client{BaseURL: srv.URL, HTTP: srv.Client(), Token: "t"}
	got, err := c.ListPullRequestsByHead(context.Background(), "o", "r", "forge/feature-1/issue-10-alpha")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Number != 9 || got[0].BaseRef != "forge/feature-1" {
		t.Fatalf("got %+v", got)
	}
}

func TestHasForgeLabel(t *testing.T) {
	t.Parallel()
	pr := PullRequest{Labels: []struct {
		Name string `json:"name"`
	}{{Name: "Forge"}}}
	if !HasForgeLabel(pr) {
		t.Fatal("expected forge label match case-insensitive")
	}
}
