package naming

import (
	"strings"
	"testing"
)

func TestFeatureBranch_golden(t *testing.T) {
	t.Parallel()
	got := FeatureBranch(11, "")
	want := "forge/feature-11"
	if got != want {
		t.Fatalf("FeatureBranch(11, \"\") = %q, want %q", got, want)
	}
}

func TestStackedBranch_golden(t *testing.T) {
	t.Parallel()
	got := StackedBranch(11, 42, "")
	want := "forge/feature-11/issue-42"
	if got != want {
		t.Fatalf("StackedBranch(11, 42, \"\") = %q, want %q", got, want)
	}
}

func TestSlugFromTitle(t *testing.T) {
	t.Parallel()
	tests := []struct {
		title string
		want  string
	}{
		{"", ""},
		{"   ", ""},
		{"Add API", "add-api"},
		{"  Foo   Bar  ", "foo-bar"},
		{"café", "caf"},
		{"Issue #42: Schema!!!", "issue-42-schema"},
	}
	for _, tt := range tests {
		got := SlugFromTitle(tt.title)
		if got != tt.want {
			t.Errorf("SlugFromTitle(%q) = %q, want %q", tt.title, got, tt.want)
		}
	}
}

func TestFeatureBranch_withSlug(t *testing.T) {
	t.Parallel()
	got := FeatureBranch(11, SlugFromTitle("Invoicing v1"))
	want := "forge/feature-11-invoicing-v1"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestStackedBranch_withSlug(t *testing.T) {
	t.Parallel()
	got := StackedBranch(11, 42, SlugFromTitle("Add schema"))
	want := "forge/feature-11/issue-42-add-schema"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestSlugFromTitle_maxLength(t *testing.T) {
	t.Parallel()
	in := strings.Repeat("a", maxSlugLen+50)
	got := SlugFromTitle(in)
	want := strings.Repeat("a", maxSlugLen)
	if got != want {
		t.Fatalf("got %q (len %d), want %q (len %d)", got, len(got), want, len(want))
	}
}
