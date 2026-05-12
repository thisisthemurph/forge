package cursorrunner

import (
	"errors"
	"strings"
	"testing"
)

func TestResolveBin_prefersForgeCursorBinWhenAbsolute(t *testing.T) {
	getenv := func(k string) string {
		if k == "FORGE_CURSOR_BIN" {
			return "/opt/cursor/bin/cursor"
		}
		return ""
	}
	lookPath := func(_ string) (string, error) {
		t.Fatal("lookPath must not be consulted when FORGE_CURSOR_BIN is set")
		return "", nil
	}
	got, err := ResolveBin(getenv, lookPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "/opt/cursor/bin/cursor" {
		t.Fatalf("expected /opt/cursor/bin/cursor, got %q", got)
	}
}

func TestResolveBin_rejectsForgeCursorBinWhenNotAbsolute(t *testing.T) {
	getenv := func(k string) string {
		if k == "FORGE_CURSOR_BIN" {
			return "cursor"
		}
		return ""
	}
	lookPath := func(_ string) (string, error) {
		t.Fatal("lookPath must not be consulted when FORGE_CURSOR_BIN is set")
		return "", nil
	}
	_, err := ResolveBin(getenv, lookPath)
	if err == nil {
		t.Fatal("expected error for non-absolute FORGE_CURSOR_BIN")
	}
	if !strings.Contains(err.Error(), "FORGE_CURSOR_BIN") || !strings.Contains(err.Error(), "absolute") {
		t.Fatalf("error should mention FORGE_CURSOR_BIN and absolute, got: %v", err)
	}
}

func TestResolveBin_fallsBackToPathLookup(t *testing.T) {
	getenv := func(_ string) string { return "" }
	lookPath := func(name string) (string, error) {
		if name != "cursor" {
			t.Fatalf("expected lookPath(\"cursor\"), got %q", name)
		}
		return "/usr/local/bin/cursor", nil
	}
	got, err := ResolveBin(getenv, lookPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "/usr/local/bin/cursor" {
		t.Fatalf("expected /usr/local/bin/cursor, got %q", got)
	}
}

func TestResolveBin_errorsWhenCursorMissingOnPath(t *testing.T) {
	getenv := func(_ string) string { return "" }
	lookPath := func(_ string) (string, error) {
		return "", errors.New("not found")
	}
	_, err := ResolveBin(getenv, lookPath)
	if err == nil {
		t.Fatal("expected error when cursor is not on PATH")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "cursor") {
		t.Fatalf("error should mention cursor, got: %v", err)
	}
}
