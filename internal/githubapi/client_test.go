package githubapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestClient_ListAllSubIssues_success(t *testing.T) {
	t.Parallel()
	var pages int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("method %s", r.Method)
		}
		if auth := r.Header.Get("Authorization"); auth != "Bearer tok" {
			t.Fatalf("Authorization header: %q", auth)
		}
		if r.URL.Path != "/repos/acme/r/issues/42/sub_issues" {
			t.Fatalf("path %s", r.URL.Path)
		}
		p := r.URL.Query().Get("page")
		if p == "" {
			p = "1"
		}
		pages++
		switch p {
		case "1":
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`[{"number":10,"title":"First","state":"open"},{"number":11,"title":"Second","state":"closed"}]`))
		default:
			w.Write([]byte(`[]`))
		}
	}))
	t.Cleanup(srv.Close)

	c := &Client{
		BaseURL: srv.URL,
		HTTP:    srv.Client(),
		Token:   "tok",
	}
	got, err := c.ListAllSubIssues(context.Background(), "acme", "r", 42)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].Number != 10 || got[1].Number != 11 {
		t.Fatalf("unexpected %+v", got)
	}
	if pages != 1 {
		t.Fatalf("expected single page for short result, got %d", pages)
	}
}

func TestClient_ListAllSubIssues_fullPageFetchesNext(t *testing.T) {
	t.Parallel()
	batch := make([]map[string]any, 100)
	for i := range batch {
		batch[i] = map[string]any{"number": i + 1, "title": "x", "state": "open"}
	}
	page1, err := json.Marshal(batch)
	if err != nil {
		t.Fatal(err)
	}
	var pages int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := r.URL.Query().Get("page")
		if p == "" {
			p = "1"
		}
		pages++
		switch p {
		case "1":
			w.Write(page1)
		default:
			w.Write([]byte(`[]`))
		}
	}))
	t.Cleanup(srv.Close)

	c := &Client{
		BaseURL: srv.URL,
		HTTP:    srv.Client(),
		Token:   "tok",
	}
	got, err := c.ListAllSubIssues(context.Background(), "o", "r", 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 100 {
		t.Fatalf("len %d", len(got))
	}
	if pages != 2 {
		t.Fatalf("expected two HTTP pages, got %d", pages)
	}
}

func TestClient_ListAllSubIssues_apiError(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"message":"Bad credentials"}`))
	}))
	t.Cleanup(srv.Close)

	c := &Client{
		BaseURL: srv.URL,
		HTTP:    srv.Client(),
		Token:   "wrong",
	}
	_, err := c.ListAllSubIssues(context.Background(), "o", "r", 1)
	if err == nil {
		t.Fatal("expected error")
	}
	if want := "Bad credentials"; !strings.Contains(err.Error(), want) {
		t.Fatalf("error %q should mention %q", err.Error(), want)
	}
}
