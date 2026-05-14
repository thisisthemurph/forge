package prpublish

import (
	"strings"
	"testing"
)

func TestForgeManagedPRTitle(t *testing.T) {
	t.Parallel()
	got := ForgeManagedPRTitle(7, "Git push + Forge-managed PR")
	want := "[#7] Git push + Forge-managed PR"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestForgeManagedPRBody_includesFixes(t *testing.T) {
	t.Parallel()
	got := ForgeManagedPRBody(7)
	if !strings.Contains(got, "Fixes #7") {
		t.Fatalf("expected Fixes linkage, got %q", got)
	}
}
