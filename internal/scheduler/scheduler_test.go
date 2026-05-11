package scheduler

import (
	"strings"
	"testing"

	"github.com/thisisthemurph/forge/internal/mergesnapshot"
	"github.com/thisisthemurph/forge/internal/naming"
)

func TestNextExecutable_singleOpenStackNoBlockers(t *testing.T) {
	t.Parallel()
	feature := 50
	order := []int{10}
	titles := map[int]string{10: "Alpha"}
	snap := mergesnapshot.Memory{
		10: {Matches: nil},
	}
	plan := BuildPlan(feature, order, map[int][]int{10: nil}, titles, snap)
	if plan.NextExecutable == nil || *plan.NextExecutable != 10 {
		t.Fatalf("NextExecutable = %v, want 10", plan.NextExecutable)
	}
	if len(plan.Warnings) == 0 || !strings.Contains(plan.Warnings[0], "Forge PR identification") {
		t.Fatalf("expected identification warning, got %#v", plan.Warnings)
	}
}

func TestNextExecutable_waitsForStackPrefix(t *testing.T) {
	t.Parallel()
	feature := 50
	order := []int{10, 20}
	titles := map[int]string{10: "Alpha", 20: "Beta"}
	featureBranch := naming.FeatureBranch(feature, "")
	snap := mergesnapshot.Memory{
		10: {Matches: nil},
		20: {Matches: nil},
	}
	plan := BuildPlan(feature, order, map[int][]int{10: nil, 20: nil}, titles, snap)
	if plan.NextExecutable == nil || *plan.NextExecutable != 10 {
		t.Fatalf("NextExecutable = %v, want 10", plan.NextExecutable)
	}
	snap2 := mergesnapshot.Memory{
		10: {Matches: []*mergesnapshot.ManagedPR{{Number: 1, BaseRef: featureBranch, Merged: true}}},
		20: {Matches: nil},
	}
	plan2 := BuildPlan(feature, order, map[int][]int{10: nil, 20: nil}, titles, snap2)
	if plan2.NextExecutable == nil || *plan2.NextExecutable != 20 {
		t.Fatalf("NextExecutable = %v, want 20", plan2.NextExecutable)
	}
}

func TestNextExecutable_respectsDAGBlockers(t *testing.T) {
	t.Parallel()
	feature := 1
	order := []int{10, 20}
	titles := map[int]string{10: "A", 20: "B"}
	featureBranch := naming.FeatureBranch(feature, "")
	snap := mergesnapshot.Memory{
		10: {Matches: []*mergesnapshot.ManagedPR{{Number: 99, BaseRef: featureBranch, Merged: true}}},
		20: {Matches: nil},
	}
	plan := BuildPlan(feature, order, map[int][]int{10: nil, 20: {10}}, titles, snap)
	if plan.NextExecutable == nil || *plan.NextExecutable != 20 {
		t.Fatalf("NextExecutable = %v, want 20", plan.NextExecutable)
	}

	snapBlocked := mergesnapshot.Memory{
		10: {Matches: nil},
		20: {Matches: nil},
	}
	planBlocked := BuildPlan(feature, order, map[int][]int{10: nil, 20: {10}}, titles, snapBlocked)
	if planBlocked.NextExecutable == nil || *planBlocked.NextExecutable != 10 {
		t.Fatalf("NextExecutable = %v, want 10 while blocker open", planBlocked.NextExecutable)
	}
}

func TestAmbiguousForgePRNotExecutable(t *testing.T) {
	t.Parallel()
	feature := 1
	order := []int{10}
	titles := map[int]string{10: "A"}
	snap := mergesnapshot.Memory{
		10: {Matches: []*mergesnapshot.ManagedPR{{Number: 1}, {Number: 2}}},
	}
	plan := BuildPlan(feature, order, map[int][]int{10: nil}, titles, snap)
	if plan.NextExecutable != nil {
		t.Fatalf("NextExecutable = %v, want nil", *plan.NextExecutable)
	}
	var found bool
	for _, w := range plan.Warnings {
		if strings.Contains(w, "ambiguous") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected ambiguous warning, got %#v", plan.Warnings)
	}
}

func TestStackConsistencyViolationWarning(t *testing.T) {
	t.Parallel()
	feature := 1
	order := []int{10, 20, 30}
	titles := map[int]string{10: "A", 20: "B", 30: "C"}
	b10 := naming.StackedBranch(feature, 10, naming.SlugFromTitle(titles[10]))
	snap := mergesnapshot.Memory{
		10: {Matches: nil},
		20: {Matches: []*mergesnapshot.ManagedPR{{Number: 2, BaseRef: b10, Merged: true}}},
		30: {Matches: nil},
	}
	plan := BuildPlan(feature, order, map[int][]int{10: nil, 20: {10}, 30: {20}}, titles, snap)
	var found bool
	for _, w := range plan.Warnings {
		if strings.Contains(w, "Stack consistency violation") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected stack consistency warning, got %#v", plan.Warnings)
	}
}
