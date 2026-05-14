package githubapi

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreatePullRequest(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/o/r/pulls" || r.Method != http.MethodPost {
			t.Fatalf("unexpected %s %s", r.Method, r.URL.Path)
		}
		b, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		var got map[string]any
		if err := json.Unmarshal(b, &got); err != nil {
			t.Fatal(err)
		}
		if got["title"] != "[#7] Title" || got["head"] != "feature-branch" || got["base"] != "main" || got["draft"] != false || got["body"] != "Fixes #7\n" {
			t.Fatalf("unexpected payload: %v", got)
		}
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"number":44}`))
	}))
	t.Cleanup(srv.Close)

	c := &Client{BaseURL: srv.URL, HTTP: srv.Client(), Token: "t"}
	n, err := c.CreatePullRequest(context.Background(), "o", "r", "[#7] Title", "feature-branch", "main", "Fixes #7\n", false)
	if err != nil {
		t.Fatal(err)
	}
	if n != 44 {
		t.Fatalf("got number %d", n)
	}
}

func TestAddIssueLabels(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/o/r/issues/44/labels" || r.Method != http.MethodPost {
			t.Fatalf("unexpected %s %s", r.Method, r.URL.Path)
		}
		b, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		if string(b) != `{"labels":["forge"]}` {
			t.Fatalf("unexpected body %s", b)
		}
		w.Write([]byte(`[]`))
	}))
	t.Cleanup(srv.Close)

	c := &Client{BaseURL: srv.URL, HTTP: srv.Client(), Token: "t"}
	if err := c.AddIssueLabels(context.Background(), "o", "r", 44, []string{"forge"}); err != nil {
		t.Fatal(err)
	}
}

func TestUpdatePullRequest(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/o/r/pulls/44" || r.Method != http.MethodPatch {
			t.Fatalf("unexpected %s %s", r.Method, r.URL.Path)
		}
		b, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		var m map[string]any
		if err := json.Unmarshal(b, &m); err != nil {
			t.Fatal(err)
		}
		if m["title"] != "[#7] T" || m["base"] != "forge/feature/1/base" || m["draft"] != false {
			t.Fatalf("unexpected payload: %s", b)
		}
		w.Write([]byte(`{"number":44}`))
	}))
	t.Cleanup(srv.Close)

	c := &Client{BaseURL: srv.URL, HTTP: srv.Client(), Token: "t"}
	title := "[#7] T"
	baseRef := "forge/feature/1/base"
	draft := false
	if err := c.UpdatePullRequest(context.Background(), "o", "r", 44, &title, nil, &baseRef, &draft); err != nil {
		t.Fatal(err)
	}
}
