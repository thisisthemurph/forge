# Forge

Forge is a small Go CLI for GitHub-backed feature work. You attach sub-issues to a parent feature issue, declare dependencies in each sub-issue's `## Blocked by` section, and Forge figures out stack order, branch names, and what to do next. It can inspect that plan (`status`) or drive the next step (`run`): set up stacked git branches, invoke the Cursor CLI as the agent runner, then push and open or update Forge-managed pull requests. Humans still review and merge PRs.

Domain vocabulary and v1 behavior details live in `[CONTEXT.md](CONTEXT.md)`.

## Requirements

- Run from your git checkout; Forge walks up to the repo root as needed.
- `origin` must point at github.com so Forge can infer `owner/repo`, unless you pass `--repo owner/name` (for example when issues live upstream but you work in a fork).
- GitHub API access: set `GH_TOKEN` or `GITHUB_TOKEN`, or authenticate the GitHub CLI (`gh auth login`) so `gh auth token` works when the env vars are unset.
- For `forge run`, the `cursor` binary must be on `PATH`, or set `FORGE_CURSOR_BIN` to an absolute path to the Cursor executable.

## Install

From a clone of this repo (Go 1.23+):

```bash
go build -o ./.bin/forge ./cmd/forge
```

Optionally install onto your `PATH`:

```bash
go install ./cmd/forge
```

## Usage

```text
forge [flags] <command> <feature-issue-number>
```

Use the global flag `--repo owner/name` (or `--repo=owner/name`) when `origin` does not identify the GitHub repo where the feature issue lives.

- `<feature-issue-number>` is the GitHub issue number of the parent feature only. You do not pass individual sub-issue numbers; Forge loads every issue attached to that feature and schedules from the graph.

### `forge status <n>`

Read-only summary: repo, attached sub-issues, stack order, deterministic branch names, scheduler's “next work” if any. Scheduling or stack warnings are printed instead of failing the command.

### `forge run <n>`

Mutating loop for feature `#n`: validates the scheduling graph and stack consistency, then either prints `Feature #n: no pending work` or prepares branches, runs Cursor for the next executable sub-issue, pushes to `origin`, and ensures the PR. Graph validation failures or blocking stack-consistency warnings exit non-zero without starting the agent.

### Global flag: `--repo`

Pass `--repo owner/repo` (or `--repo=owner/repo`) when `origin` does not identify the GitHub repo where the feature issue lives. It is a persistent flag and may appear before or after the subcommand.

Examples:

```bash
forge status 42
forge run 42
forge --repo upstream/org status 42
forge status 42 --repo upstream/org
```

Avoid overlapping `forge run` invocations for the same feature or checkout; Forge does not coordinate concurrent runs.