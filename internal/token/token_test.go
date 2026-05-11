package token

import (
	"errors"
	"testing"
)

func TestResolve(t *testing.T) {
	t.Parallel()
	t.Run("GH_TOKEN wins", func(t *testing.T) {
		t.Parallel()
		getenv := mapLookup(map[string]string{
			"GH_TOKEN":     "a",
			"GITHUB_TOKEN": "b",
		})
		tok, err := Resolve(getenv, nil)
		if err != nil {
			t.Fatal(err)
		}
		if tok != "a" {
			t.Fatalf("got %q want a", tok)
		}
	})
	t.Run("GITHUB_TOKEN when GH unset", func(t *testing.T) {
		t.Parallel()
		getenv := mapLookup(map[string]string{"GITHUB_TOKEN": "xy"})
		tok, err := Resolve(getenv, nil)
		if err != nil {
			t.Fatal(err)
		}
		if tok != "xy" {
			t.Fatalf("got %q", tok)
		}
	})
	t.Run("gh auth token when env empty", func(t *testing.T) {
		t.Parallel()
		getenv := mapLookup(map[string]string{})
		tok, err := Resolve(getenv, func() (string, error) { return "from-gh", nil })
		if err != nil {
			t.Fatal(err)
		}
		if tok != "from-gh" {
			t.Fatalf("got %q", tok)
		}
	})
	t.Run("error when nothing available", func(t *testing.T) {
		t.Parallel()
		getenv := mapLookup(map[string]string{})
		_, err := Resolve(getenv, func() (string, error) { return "", errors.New("gh not logged in") })
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func mapLookup(m map[string]string) func(string) string {
	return func(k string) string { return m[k] }
}
