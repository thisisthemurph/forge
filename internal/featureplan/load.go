package featureplan

import (
	"context"
	"fmt"
	"strings"

	"github.com/thisisthemurph/forge/internal/cli"
	"github.com/thisisthemurph/forge/internal/githubapi"
	"github.com/thisisthemurph/forge/internal/gitremote"
	"github.com/thisisthemurph/forge/internal/mergesnapshot"
	"github.com/thisisthemurph/forge/internal/remote"
	"github.com/thisisthemurph/forge/internal/scheduler"
	"github.com/thisisthemurph/forge/internal/scheduling"
	"github.com/thisisthemurph/forge/internal/token"
)

// State is GitHub + scheduling + merge snapshot data for one Feature issue.
type State struct {
	Feature       int
	Owner, Repo   string
	SubIssues     []githubapi.SubIssue
	Graph         *scheduling.Graph
	GraphErr      error
	Order         []int
	TitleByNumber map[int]string
	Snapshot      mergesnapshot.Snapshot
	Plan          scheduler.Plan
}

// Load resolves the repository, lists Sub-issues, validates the **Scheduling graph**,
// loads merge snapshot data, and builds the scheduler plan.
//
// When **DAG validation** fails, GraphErr is set and Graph is nil; Snapshot and Plan
// are zero values and no further GitHub merge snapshot calls are made.
func Load(ctx context.Context, cfg cli.Config, cwd string, getenv func(string) string, ghAuth func() (string, error), client *githubapi.Client) (*State, error) {
	var owner, repo string
	var err error
	if strings.TrimSpace(cfg.RepoOverride) != "" {
		owner, repo, err = remote.ParseRepoOwnerPath(cfg.RepoOverride)
		if err != nil {
			return nil, err
		}
	} else {
		repoRoot, err := gitremote.FindRepoRoot(cwd)
		if err != nil {
			return nil, err
		}
		raw, err := gitremote.GetOriginURL(repoRoot)
		if err != nil {
			return nil, fmt.Errorf("resolve origin: %w", err)
		}
		owner, repo, err = remote.OwnerRepoFromURL(raw)
		if err != nil {
			return nil, fmt.Errorf("resolve GitHub repository from origin: %w", err)
		}
	}

	tok, err := token.Resolve(getenv, ghAuth)
	if err != nil {
		return nil, err
	}
	client.Token = tok

	subs, err := client.ListAllSubIssues(ctx, owner, repo, cfg.Feature)
	if err != nil {
		return nil, err
	}

	st := &State{
		Feature:   cfg.Feature,
		Owner:     owner,
		Repo:      repo,
		SubIssues: subs,
	}

	inputs := make([]scheduling.SubIssueInput, 0, len(subs))
	for _, s := range subs {
		inputs = append(inputs, scheduling.SubIssueInput{Number: s.Number, Body: s.Body})
	}
	graph, graphErr := scheduling.AnalyzeGraph(cfg.Feature, inputs)
	if graphErr != nil {
		st.GraphErr = graphErr
		return st, nil
	}
	st.Graph = graph
	st.Order = graph.Order

	titleByNumber := make(map[int]string, len(subs))
	for _, s := range subs {
		titleByNumber[s.Number] = s.Title
	}
	st.TitleByNumber = titleByNumber

	snap, err := mergesnapshot.LoadFromGitHub(ctx, client, owner, repo, cfg.Feature, graph.Order, titleByNumber)
	if err != nil {
		return nil, err
	}
	st.Snapshot = snap
	st.Plan = scheduler.BuildPlan(cfg.Feature, graph.Order, graph.Blockers, titleByNumber, snap)
	return st, nil
}
