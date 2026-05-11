# PRD: Forge CLI (v1)

## Problem Statement

Today you run a manual loop: point an AI at a parent **Feature** issue, pick the next **Sub-issue**, follow TDD, open a PR, review, merge to `main`, repeat. That does not encode the **Scheduling graph**, **Stack order**, stacked **PR stack target**s, or **Forge-managed PR** conventions—and nothing prevents inconsistent merges or ambiguous “which PR is Forge’s?” state.

You want a **Go CLI** that automates the mechanical parts: resolve **Feature** + attached **Sub-issues**, validate the DAG from **`## Blocked by`**, compute **Stack order**, prepare **Deterministic branch names** and checkouts, invoke **Cursor** as the **Agent runner**, then **push** and open **Forge-managed PR**s with the right base, **`forge`** label, **`Fixes #<sub-issue>`**, and title format—while **humans** still merge PRs and integrate to **default**. The CLI should fail loudly on **Repository hygiene**, **DAG validation**, **Forge PR identification**, and **Stack consistency violation**s when starting work, but stay inspectable on read-only status.

## Solution

Ship **Forge v1**: a single binary invoked from **Repository root** with **only** the parent **Feature** issue number (and optional **`--repo owner/name`** when **`origin`** is wrong). Forge uses the **GitHub API client** with **GitHub token** resolution (`GH_TOKEN` / `GITHUB_TOKEN` / `gh auth token`), discovers **Sub-issues** attached to that **Feature** on **github.com** only, parses **Blocker** edges, validates **DAG validation** rules, computes **Stack order** (topological + ascending issue-number tie-break), derives **Stack parent** / **PR stack target** per **Sub-issue**, evaluates **Blocker satisfied** and **Stack prefix merged** from GitHub merge metadata, applies **Stack consistency policy** (warn on status-style reads, error on mutating commands), performs **Pre-agent branch setup** from **Feature base branch**, runs **Cursor** via **Cursor binary resolution**, and on success performs **Git publish and PR open** to **`origin`** with **PR title**, **PR body linking**, **PR draft policy** (ready). When nothing remains, emit **No pending work**. On **Agent failure hygiene**, leave the tree as-is and exit non-zero. **Commit authorship** stays agent-only; **Concurrency** is human-operated (no locks). Output follows **CLI output** (human text only).

## User Stories

1. As a maintainer, I want to run Forge from my repo root with a **Feature** issue number, so that I do not pass **Sub-issue** ids at the CLI.
2. As a maintainer, I want Forge to resolve **`owner/repo`** from **`origin`** on **github.com**, so that I do not duplicate repo coordinates.
3. As a maintainer, I want **`--repo owner/name`** when my **`origin`** is a fork but issues live upstream, so that API calls hit the correct GitHub repo.
4. As a maintainer, I want Forge to error if **`origin`** is missing or not **github.com**, so that I fix remotes before wasting agent time.
5. As a maintainer, I want **GitHub token** resolution in the documented order, so that CI and laptops both authenticate predictably.
6. As a maintainer, I want Forge to load all **Sub-issues** attached to the parent **Feature**, so that the **Scheduling graph** matches how I organize PRDs on GitHub.
7. As a maintainer, I want **`## Blocked by`** with lines like `- #15` parsed into **Blocker** edges, so that dependencies are explicit in issue bodies.
8. As a maintainer, I want **DAG validation** to reject blockers that are not sibling **Sub-issues** of the same **Feature**, so that the graph never references parents or external issues by mistake.
9. As a maintainer, I want **Stack order** to be a valid topological order with **ascending GitHub issue number** tie-break, so that parallel-ready issues get a deterministic git stack.
10. As a maintainer, I want **Stack parent** / **PR stack target** derived from **Stack order**, so that each PR targets the previous stacked branch—not **default**.
11. As a maintainer, I want **Feature branch** created from my **Feature base branch** (current checkout), so that the stack starts from the line I intentionally have checked out.
12. As a maintainer, I want **Repository hygiene** enforced before mutating work (branch + agent), so that I never cut a **Feature branch** from dirty or detached **HEAD** by accident.
13. As a maintainer, I want **Deterministic branch names** for **Feature branch** and **Stacked branch** lines, so that I can reason about branches across machines without the agent naming them.
14. As a maintainer, I want **Pre-agent branch setup** to create/check out the correct branch before **Cursor** runs, so that the agent always works on the intended ref.
15. As a maintainer, I want **Cursor binary resolution** (`cursor` on **PATH**, optional **`FORGE_CURSOR_BIN`**), so that nonstandard installs still work.
16. As a maintainer, I want **Cursor environment** to omit **`FORGE_*`** variables, so that the agent relies on checkout and project config only.
17. As a maintainer, I want Forge to invoke **Cursor CLI** and treat exit status as success/failure, so that TDD and validation stay entirely agent-side per **TDD and validation**.
18. As a maintainer, I want **Agent failure hygiene** (no automatic hard reset) on agent failure, so that I can inspect partial work.
19. As a maintainer, I want **Commit authorship** such that Forge never creates content commits—only the agent does—so that authorship and review stay honest.
20. As a maintainer, I want **Git publish and PR open** after a successful agent run (**push** to **`origin`**, open/update PR via API), so that GitHub state matches the stack without me clicking “Open PR.”
21. As a maintainer, I want **Forge-managed PR**s to carry the **`forge`** label and be linked to the **Sub-issue** per **Forge PR identification**, so that scheduling can find exactly one PR per **Sub-issue**.
22. As a maintainer, I want **PR title** `[\#<sub-issue>] <issue title>` and **PR body linking** with **`Fixes #<sub-issue>`**, so that notifications and auto-close behave consistently.
23. As a maintainer, I want **PR draft policy** to open PRs **ready for review**, so that my review queue matches today’s habits.
24. As a maintainer, I want **Blocker satisfied** derived from merged **Forge-managed PR** into the correct **PR stack target**, so that merge metadata drives scheduling—not issue-close labels alone.
25. As a maintainer, I want **Stack prefix merged** enforced for **Executable** work together with DAG **Blockers**, so that stacked git reality matches **Stack order**.
26. As a maintainer, I want **Stack consistency policy**: warnings on read-only **status**, hard errors on commands that start agents or mutate git/GitHub, so that I can inspect bad states without being blocked from `status`.
27. As a maintainer, I want **Human merge responsibility** unchanged—Forge never merges PRs or **default**—so that review stays human-gated.
28. As a maintainer, I want **No pending work** to print `Feature <N>: no pending work` and exit 0, so that scripts and humans see completion clearly.
29. As a maintainer, I want **State-driven dispatch** to choose the next **Executable** **Sub-issue** or continue an in-flight **Forge-managed PR** line based on GitHub + local state, so that I only ever pass the parent **Feature** number.
30. As a maintainer, I want **Concurrency** left to humans (no Forge locks), so that v1 stays simple even if two terminals could conflict.
31. As a maintainer, I want **GitHub product scope** limited to **github.com** SaaS in v1, so that scope stays shippable.
32. As a maintainer, I want **CLI output** to be human-readable text (no **`--json`** yet), so that day-one usage is simple terminal UX.
33. As a maintainer, I want sub-issue discovery and PR/issue mutations to use a **GitHub API client** (not **`gh`** subprocesses) for normal operations, so that the binary is self-contained aside from token bootstrap.
34. As a contributor, I want clear error messages for **DAG validation**, **Forge PR identification**, and **Stack consistency violation**, so that I can fix issue bodies or labels without reading source.
35. As a contributor, I want a **`status`** (or equivalent) read-only command for a **Feature**, so that I can see stack position, merge state, and warnings before running the agent.
36. As a contributor, I want the **Agent runner** integration documented as Cursor-only in v1, so that I do not expect other agents to work yet.
37. As a security-conscious user, I want tokens never printed in logs, so that accidental leakage is avoided.
38. As a repo owner, I want the **`forge`** label applied consistently when opening PRs, so that identification rules stay machine-checkable.
39. As a reviewer, I want non–**Forge-managed PR**s ignored for scheduling, so that drive-by PRs do not confuse **Blocker satisfied**.
40. As a power user, I want deterministic behavior documented in **CONTEXT.md** glossary terms, so that product language matches the implementation’s user-facing errors.

## Implementation Decisions

- **Module: scheduling graph** — Encapsulates parsing **`## Blocked by`**, building the **Scheduling graph**, **DAG validation**, and computing **Stack order** (Kahn-style topological walk with **ascending GitHub issue number** tie-break). Exposes a small API: validated graph or structured errors (invalid reference, cycle if ever introduced by bad data, missing section treated as no blockers). Pure over parsed issue payloads plus the set of **Sub-issue** numbers under the **Feature**.
- **Module: naming** — Encapsulates **Deterministic branch names** for **Feature branch** and each **Stacked branch** from **Feature** + **Sub-issue** identifiers (and optional stable slug rules). Hides template strings and sanitization policy behind one function per branch “role.”
- **Module: merge snapshot** — Abstract interface that answers GitHub-backed questions: for each **Sub-issue**, is there exactly one **Forge-managed PR** (by **`forge`** label + development linkage), merge status into **PR stack target**, open vs merged. A live implementation calls the **GitHub API client**; tests use an in-memory fake implementing the same interface so **Blocker satisfied** / **Stack prefix merged** / **Stack consistency violation** logic stays testable without network.
- **Module: scheduler / state** — Composes **scheduling graph**, **merge snapshot**, and **Stack order** to compute **Executable** next steps, detect **Stack consistency violation**, and decide **State-driven dispatch** (continue in-flight vs start next). Emits a typed “plan” object consumed by the command layer (no side effects).
- **Module: local git** — **Repository root** operations: parse **`origin`** for **github.com** `owner/repo`, read **Feature base branch**, enforce **Repository hygiene**, create/check out branches (**Pre-agent branch setup**), and **push** to **`origin`**. Depends on running `git` as subprocess or go-git later; v1 can subprocess `git` for familiarity.
- **Module: GitHub API facade** — Token resolution per **GitHub token**, REST/GraphQL calls for listing child issues, reading issue bodies/titles, listing PRs, creating/updating PRs, applying labels, linking issues. Not a thin dump of all of go-github—only operations Forge needs, so the surface stays stable when swapping HTTP details.
- **Module: cursor runner** — **Cursor binary resolution**, executes **Cursor CLI** with working directory **Repository root**, inherits stdio, maps exit codes to **Agent failure hygiene** outcomes. No **`FORGE_*`** injection per **Cursor environment**.
- **Module: PR publisher** — Builds **PR title**, **PR body linking**, applies **`forge`** label, sets base to **PR stack target**, ensures development linkage as supported by the API. Invoked only after successful agent exit, before or after push depending on safest ordering (implementation chooses consistent order documented in code comments).
- **Command layer** — Parses global flags (**`--repo`**), **Feature selector** positional argument, subcommands (`run` / `status` or a single entry command—pick one UX and document it). Maps scheduler errors to user-facing messages per **Stack consistency policy** (warn vs error by command class).
- **Deep-module emphasis** — Keep **scheduling graph** and **scheduler** narrow and testable; keep **GitHub API facade** behind interfaces used by **merge snapshot** so policy tests do not need real GitHub.

**Maintainer check (post-triage):** Confirm this module decomposition matches how you want the codebase layered. Reply on the tracking issue with any renames or merges (e.g. fold **naming** into **local git** if too granular).

## Testing Decisions

- **Good tests** assert externally observable behavior: given issue bodies and a fake **merge snapshot**, the **scheduler** reports the correct next action and the correct warn/error classification; given markdown bodies, **DAG validation** and **Stack order** match golden expectations; **naming** returns stable strings for the same inputs.
- **Modules to test with automated tests:** **scheduling graph** (highest priority), **scheduler** with fake **merge snapshot**, **naming**, and pure helpers for **`## Blocked by`** parsing edge cases (empty section, multiple bullets, malformed lines).
- **Modules to test lightly or via integration later:** **GitHub API facade** (httptest for a few critical JSON shapes), **local git** (integration test in CI with a temp repo), **cursor runner** (smoke test skipped by default if `cursor` not installed).
- **Prior art:** Repository is greenfield; follow standard Go table-driven tests and keep policy logic free of CLI flag parsing.

**Maintainer check (post-triage):** Confirm which modules above must ship with unit tests in v1 vs follow-up PRs.

## Out of Scope

- **GitHub Enterprise Server** and non-**github.com** hosts (**GitHub product scope**).
- Forge merging PRs or merging to **default** (**Human merge responsibility**).
- Advisory or mandatory multi-process **Concurrency** locks.
- Alternate **Push remote** (only **`origin`** in v1).
- **`FORGE_*`** environment injection for **Cursor** (**Cursor environment**).
- Machine-readable **`--json`** output (**CLI output**).
- Forge-run test/lint commands (**TDD and validation** stays agent-only).
- Non-**Cursor** **Agent runner** implementations.
- Human override of **Stack order** (tie-break is issue number only).
- Local-only files as source of truth for **Forge-managed PR** identity (GitHub labels + linkage only).
- Empty or **Forge-authored** commits (**Commit authorship**).

## Further Notes

- Domain vocabulary and v1 policies are captured in repository **`CONTEXT.md`**; implementation and user-facing errors should reuse those terms.
- **Publishing this PRD:** The workspace had no **`git remote`** configured at authoring time, so the issue was not auto-created. After adding a GitHub remote (or choosing `owner/repo`), create the issue with triage:

  ```bash
  gh issue create \
    --repo OWNER/REPO \
    --title "PRD: Forge CLI (v1)" \
    --body-file docs/prd/forge-cli-v1.md \
    --label "needs-triage"
  ```

  Ensure the **`needs-triage`** label exists on the target repository (see `.cursor/skills/setup-matt-pocock-skills/triage-labels.md`).

- If the body is too large for one issue, split **User Stories** into a tracking comment and keep the issue body to Problem/Solution/Decisions/Testing/Out of Scope.
