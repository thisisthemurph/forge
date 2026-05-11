package scheduling

import (
	"reflect"
	"strings"
	"testing"
)

func TestAnalyze_parentFeatureIsInvalidBlocker(t *testing.T) {
	t.Parallel()
	_, err := Analyze(100, []SubIssueInput{
		{Number: 50, Body: "## Blocked by\n- #100\n"},
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "DAG validation") || !strings.Contains(err.Error(), "Feature") {
		t.Fatalf("error should mention DAG validation and Feature, got: %v", err)
	}
}

func TestAnalyze_nonSiblingBlocker(t *testing.T) {
	t.Parallel()
	_, err := Analyze(1, []SubIssueInput{
		{Number: 10, Body: "## Blocked by\n- #999\n"},
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "DAG validation") || !strings.Contains(err.Error(), "Sub-issue") {
		t.Fatalf("error: %v", err)
	}
}

func TestAnalyze_cycleRejected(t *testing.T) {
	t.Parallel()
	_, err := Analyze(99, []SubIssueInput{
		{Number: 1, Body: "## Blocked by\n- #2\n"},
		{Number: 2, Body: "## Blocked by\n- #1\n"},
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "cycle") {
		t.Fatalf("error: %v", err)
	}
}

func TestAnalyze_contextDAGStackOrder(t *testing.T) {
	t.Parallel()
	// CONTEXT.md example: 1→2, 1→3, 2→4, 3→4, 4→5  =>  1,2,3,4,5
	subs := []SubIssueInput{
		{Number: 1, Body: ""},
		{Number: 2, Body: "## Blocked by\n- #1\n"},
		{Number: 3, Body: "## Blocked by\n- #1\n"},
		{Number: 4, Body: "## Blocked by\n- #2\n- #3\n"},
		{Number: 5, Body: "## Blocked by\n- #4\n"},
	}
	got, err := Analyze(99, subs)
	if err != nil {
		t.Fatal(err)
	}
	want := []int{1, 2, 3, 4, 5}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Stack order = %v, want %v", got, want)
	}
}

func TestAnalyze_numericTieBreakAmongReady(t *testing.T) {
	t.Parallel()
	// 10 before both 20 and 30; then 20 before 30; 40 blocked by both 20 and 30
	subs := []SubIssueInput{
		{Number: 10, Body: ""},
		{Number: 20, Body: "## Blocked by\n- #10\n"},
		{Number: 30, Body: "## Blocked by\n- #10\n"},
		{Number: 40, Body: "## Blocked by\n- #20\n- #30\n"},
	}
	got, err := Analyze(1, subs)
	if err != nil {
		t.Fatal(err)
	}
	want := []int{10, 20, 30, 40}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Stack order = %v, want %v", got, want)
	}
}

func TestAnalyze_emptySubs(t *testing.T) {
	t.Parallel()
	got, err := Analyze(1, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("got %v", got)
	}
}

func TestAnalyze_joinsMultipleValidationErrors(t *testing.T) {
	t.Parallel()
	_, err := Analyze(100, []SubIssueInput{
		{Number: 10, Body: "## Blocked by\n- #100\n"},
		{Number: 11, Body: "## Blocked by\n- #200\n"},
	})
	if err == nil {
		t.Fatal("expected error")
	}
	msg := err.Error()
	if !strings.Contains(msg, "#10") || !strings.Contains(msg, "#11") {
		t.Fatalf("expected errors for both Sub-issues, got: %s", msg)
	}
}
