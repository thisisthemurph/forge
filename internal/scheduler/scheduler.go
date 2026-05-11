package scheduler

import (
	"fmt"

	"github.com/thisisthemurph/forge/internal/mergesnapshot"
	"github.com/thisisthemurph/forge/internal/naming"
)

// Plan is scheduler output for read-only **status** and future dispatch.
type Plan struct {
	Warnings       []string
	NextExecutable *int
}

// BuildPlan computes **Executable** detection and **Stack consistency violation** signals.
// blockers maps each Sub-issue to its **Blocker** list from the **Scheduling graph**.
func BuildPlan(feature int, order []int, blockers map[int][]int, titles map[int]string, snap mergesnapshot.Snapshot) Plan {
	var p Plan
	if len(order) == 0 {
		return p
	}
	stackIndex := make(map[int]int, len(order))
	for i, n := range order {
		stackIndex[n] = i
	}

	prStackTarget := func(stackPos int) string {
		if stackPos == 0 {
			return naming.FeatureBranch(feature, "")
		}
		prev := order[stackPos-1]
		slug := naming.SlugFromTitle(titles[prev])
		return naming.StackedBranch(feature, prev, slug)
	}

	mergedIntoTarget := func(sub int, stackPos int) bool {
		lu := snap.ForgeManagedPR(sub)
		if len(lu.Matches) != 1 {
			return false
		}
		pr := lu.Matches[0]
		if !pr.Merged {
			return false
		}
		return pr.BaseRef == prStackTarget(stackPos)
	}

	for _, n := range order {
		lu := snap.ForgeManagedPR(n)
		switch len(lu.Matches) {
		case 0:
			p.Warnings = append(p.Warnings, fmt.Sprintf(
				"Forge PR identification: Sub-issue #%d has no Forge-managed PR (expected exactly one with forge label + development linkage).",
				n,
			))
		case 1:
		default:
			p.Warnings = append(p.Warnings, fmt.Sprintf(
				"Forge PR identification: Sub-issue #%d has %d Forge-managed PRs; scheduling is ambiguous.",
				n, len(lu.Matches),
			))
		}
	}

	mergedOK := make([]bool, len(order))
	for i, sub := range order {
		mergedOK[i] = mergedIntoTarget(sub, i)
	}
	for j := range order {
		if !mergedOK[j] {
			continue
		}
		for i := 0; i < j; i++ {
			if !mergedOK[i] {
				p.Warnings = append(p.Warnings, fmt.Sprintf(
					"Stack consistency violation: Sub-issue #%d appears merged while earlier Sub-issue #%d in Stack order is not merged into its PR stack target.",
					order[j], order[i],
				))
				break
			}
		}
	}

	for i, sub := range order {
		lu := snap.ForgeManagedPR(sub)
		if len(lu.Matches) > 1 {
			continue
		}

		blocked := false
		for _, b := range blockers[sub] {
			if _, ok := stackIndex[b]; !ok {
				continue
			}
			bPos := stackIndex[b]
			if !mergedIntoTarget(b, bPos) {
				blocked = true
				break
			}
		}
		if blocked {
			continue
		}

		for j := 0; j < i; j++ {
			prev := order[j]
			if !mergedIntoTarget(prev, j) {
				blocked = true
				break
			}
		}
		if blocked {
			continue
		}

		switch len(lu.Matches) {
		case 0:
			next := sub
			p.NextExecutable = &next
			return p
		case 1:
			pr := lu.Matches[0]
			if pr.Merged && pr.BaseRef == prStackTarget(i) {
				continue
			}
			next := sub
			p.NextExecutable = &next
			return p
		}
	}
	return p
}
