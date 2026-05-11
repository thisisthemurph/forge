package mergesnapshot

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/thisisthemurph/forge/internal/githubapi"
)

func TestLoadFromGitHub_findsForgeLinkedPR(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/o/r/pulls" {
			t.Fatalf("path %s", r.URL.Path)
		}
		head := r.URL.Query().Get("head")
		switch head {
		case "o:forge/feature-50/issue-10-alpha":
			w.Write([]byte(`[{"number":9,"title":"[#10] Alpha","body":"Fixes #10","state":"open","merged":false,"base":{"ref":"forge/feature-50"},"labels":[{"name":"forge"}]}]`))
		default:
			w.Write([]byte(`[]`))
		}
	}))
	t.Cleanup(srv.Close)

	c := &githubapi.Client{BaseURL: srv.URL, HTTP: srv.Client(), Token: "t"}
	snap, err := LoadFromGitHub(context.Background(), c, "o", "r", 50, []int{10}, map[int]string{10: "Alpha"})
	if err != nil {
		t.Fatal(err)
	}
	lu := snap.ForgeManagedPR(10)
	if len(lu.Matches) != 1 || lu.Matches[0].Number != 9 {
		t.Fatalf("lookup %#v", lu)
	}
}

func TestLoadFromGitHub_ignoresUnlinkedForgePR(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`[{"number":1,"title":"Other","body":"","state":"open","merged":false,"base":{"ref":"main"},"labels":[{"name":"forge"}]}]`))
	}))
	t.Cleanup(srv.Close)

	c := &githubapi.Client{BaseURL: srv.URL, HTTP: srv.Client(), Token: "t"}
	snap, err := LoadFromGitHub(context.Background(), c, "o", "r", 1, []int{10}, map[int]string{10: "A"})
	if err != nil {
		t.Fatal(err)
	}
	if len(snap.ForgeManagedPR(10).Matches) != 0 {
		t.Fatal("expected no linked forge PR")
	}
}
