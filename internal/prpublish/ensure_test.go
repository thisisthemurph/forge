package prpublish

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/thisisthemurph/forge/internal/githubapi"
)

func TestEnsureForgeManagedPR_createsWhenNoneExist(t *testing.T) {
	t.Parallel()
	var createBody []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/repos/o/r/pulls":
			if r.URL.Query().Get("head") != "o:forge/feature/1/issue/10" {
				t.Fatalf("head query %q", r.URL.Query().Get("head"))
			}
			w.Write([]byte(`[]`))
		case r.Method == http.MethodPost && r.URL.Path == "/repos/o/r/pulls":
			var err error
			createBody, err = io.ReadAll(r.Body)
			if err != nil {
				t.Fatal(err)
			}
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{"number":99}`))
		case r.Method == http.MethodPost && r.URL.Path == "/repos/o/r/issues/99/labels":
			w.Write([]byte(`[]`))
		default:
			t.Fatalf("unexpected %s %s", r.Method, r.URL.Path)
		}
	}))
	t.Cleanup(srv.Close)

	c := &githubapi.Client{BaseURL: srv.URL, HTTP: srv.Client(), Token: "t"}
	n, err := EnsureForgeManagedPR(context.Background(), c, EnsureInput{
		Owner: "o", Repo: "r", SubIssue: 10, SubIssueTitle: "Alpha",
		HeadBranch: "forge/feature/1/issue/10", BaseBranch: "forge/feature/1/base",
	})
	if err != nil {
		t.Fatal(err)
	}
	if n != 99 {
		t.Fatalf("pr number %d", n)
	}
	var got map[string]any
	if err := json.Unmarshal(createBody, &got); err != nil {
		t.Fatal(err)
	}
	if got["title"] != "[#10] Alpha" || got["head"] != "forge/feature/1/issue/10" || got["base"] != "forge/feature/1/base" || got["draft"] != false {
		t.Fatalf("create payload: %v", got)
	}
}

func TestEnsureForgeManagedPR_updatesForgePRMetadata(t *testing.T) {
	t.Parallel()
	var sawPatch bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/repos/o/r/pulls":
			w.Write([]byte(`[{"number":5,"title":"[#10] Old","body":"Fixes #10","state":"open","merged":false,"base":{"ref":"wrong-base"},"labels":[{"name":"forge"}]}]`))
		case r.Method == http.MethodPatch && r.URL.Path == "/repos/o/r/pulls/5":
			sawPatch = true
			b, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatal(err)
			}
			var m map[string]any
			if err := json.Unmarshal(b, &m); err != nil {
				t.Fatal(err)
			}
			if m["title"] != "[#10] Alpha" || m["base"] != "forge/feature/1/base" || m["draft"] != false || m["body"] != "Fixes #10\n" {
				t.Fatalf("patch payload: %s", b)
			}
			w.Write([]byte(`{"number":5}`))
		case r.Method == http.MethodPost && r.URL.Path == "/repos/o/r/issues/5/labels":
			w.Write([]byte(`[]`))
		default:
			t.Fatalf("unexpected %s %s", r.Method, r.URL.Path)
		}
	}))
	t.Cleanup(srv.Close)

	c := &githubapi.Client{BaseURL: srv.URL, HTTP: srv.Client(), Token: "t"}
	n, err := EnsureForgeManagedPR(context.Background(), c, EnsureInput{
		Owner: "o", Repo: "r", SubIssue: 10, SubIssueTitle: "Alpha",
		HeadBranch: "forge/feature/1/issue/10", BaseBranch: "forge/feature/1/base",
	})
	if err != nil {
		t.Fatal(err)
	}
	if n != 5 {
		t.Fatalf("pr number %d", n)
	}
	if !sawPatch {
		t.Fatal("expected PATCH update")
	}
}

func TestEnsureForgeManagedPR_rejectsMultipleForgeCandidates(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/repos/o/r/pulls" {
			w.Write([]byte(`[
				{"number":1,"title":"[#10] A","body":"Fixes #10","state":"open","merged":false,"base":{"ref":"main"},"labels":[{"name":"forge"}]},
				{"number":2,"title":"[#10] B","body":"Fixes #10","state":"open","merged":false,"base":{"ref":"main"},"labels":[{"name":"forge"}]}
			]`))
			return
		}
		t.Fatalf("unexpected %s %s", r.Method, r.URL.Path)
	}))
	t.Cleanup(srv.Close)

	c := &githubapi.Client{BaseURL: srv.URL, HTTP: srv.Client(), Token: "t"}
	_, err := EnsureForgeManagedPR(context.Background(), c, EnsureInput{
		Owner: "o", Repo: "r", SubIssue: 10, SubIssueTitle: "Alpha",
		HeadBranch: "forge/feature/1/issue/10", BaseBranch: "forge/feature/1/base",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestEnsureForgeManagedPR_rejectsAmbiguousOpenHead(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/repos/o/r/pulls" {
			w.Write([]byte(`[
				{"number":1,"title":"Drive-by","body":"n/a","state":"open","merged":false,"base":{"ref":"main"},"labels":[]},
				{"number":2,"title":"Other","body":"n/a","state":"open","merged":false,"base":{"ref":"main"},"labels":[]}
			]`))
			return
		}
		t.Fatalf("unexpected %s %s", r.Method, r.URL.Path)
	}))
	t.Cleanup(srv.Close)

	c := &githubapi.Client{BaseURL: srv.URL, HTTP: srv.Client(), Token: "t"}
	_, err := EnsureForgeManagedPR(context.Background(), c, EnsureInput{
		Owner: "o", Repo: "r", SubIssue: 10, SubIssueTitle: "Alpha",
		HeadBranch: "forge/feature/1/issue/10", BaseBranch: "forge/feature/1/base",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}
