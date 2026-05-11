package scheduling

import (
	"errors"
	"fmt"
)

// SubIssueInput is scheduling input for one Sub-issue (number + body markdown).
type SubIssueInput struct {
	Number int
	Body   string
}

// Graph is a validated **Scheduling graph** plus **Stack order** for one Feature.
type Graph struct {
	// Order is **Stack order**: a topological order with ascending issue-number tie-break.
	Order []int
	// Blockers maps each Sub-issue number to its declared **Blocker** predecessors.
	Blockers map[int][]int
}

// AnalyzeGraph validates Blocker edges and returns Stack order plus Blocker edges.
func AnalyzeGraph(feature int, subs []SubIssueInput) (*Graph, error) {
	if len(subs) == 0 {
		return &Graph{}, nil
	}
	subSet := make(map[int]struct{}, len(subs))
	for _, s := range subs {
		subSet[s.Number] = struct{}{}
	}

	blockersBy := make(map[int][]int, len(subs))
	var valErrs []error
	for _, s := range subs {
		raw := ParseBlockedBySection(s.Body)
		var dedup []int
		seen := map[int]struct{}{}
		for _, b := range raw {
			if _, ok := seen[b]; ok {
				continue
			}
			seen[b] = struct{}{}
			dedup = append(dedup, b)
		}
		for _, b := range dedup {
			switch {
			case b == feature:
				valErrs = append(valErrs, fmt.Errorf(
					"DAG validation: Sub-issue #%d references parent Feature #%d as a Blocker; only sibling Sub-issues are allowed",
					s.Number, feature,
				))
			case b == s.Number:
				valErrs = append(valErrs, fmt.Errorf(
					"DAG validation: Sub-issue #%d lists itself under Blocked by",
					s.Number,
				))
			default:
				if _, ok := subSet[b]; !ok {
					valErrs = append(valErrs, fmt.Errorf(
						"DAG validation: Sub-issue #%d references #%d, which is not a Sub-issue under this Feature",
						s.Number, b,
					))
				}
			}
		}
		blockersBy[s.Number] = dedup
	}
	if len(valErrs) > 0 {
		return nil, errors.Join(valErrs...)
	}

	inDeg := make(map[int]int, len(subSet))
	adj := make(map[int][]int)
	for n := range subSet {
		inDeg[n] = 0
	}
	for blocked, blockers := range blockersBy {
		for _, b := range blockers {
			inDeg[blocked]++
			adj[b] = append(adj[b], blocked)
		}
	}

	order := make([]int, 0, len(subSet))
	inOrder := make(map[int]struct{}, len(subSet))
	for len(order) < len(subSet) {
		var ready []int
		for n := range subSet {
			if inDeg[n] != 0 {
				continue
			}
			if _, done := inOrder[n]; done {
				continue
			}
			ready = append(ready, n)
		}
		if len(ready) == 0 {
			return nil, fmt.Errorf("DAG validation: Scheduling graph contains a cycle among Sub-issues")
		}
		pick := ready[0]
		for _, x := range ready[1:] {
			if x < pick {
				pick = x
			}
		}
		order = append(order, pick)
		inOrder[pick] = struct{}{}
		for _, w := range adj[pick] {
			inDeg[w]--
		}
	}
	return &Graph{Order: order, Blockers: blockersBy}, nil
}

// Analyze validates Blocker edges against the Feature and sibling Sub-issues,
// then returns Stack order: a topological order with ascending issue-number
// tie-break among nodes whose Blockers are already placed.
func Analyze(feature int, subs []SubIssueInput) ([]int, error) {
	g, err := AnalyzeGraph(feature, subs)
	if err != nil {
		return nil, err
	}
	if g == nil {
		return nil, nil
	}
	return g.Order, nil
}
