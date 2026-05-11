# Forge

CLI and automation around GitHub-backed feature work: dependency-aware scheduling, stacked branches and PRs, human-in-the-loop review, and **Agent runner (v1)** integration with **Cursor** via the **Cursor CLI**.

## Language

**Feature (issue)**:
The parent GitHub issue that represents a PRD-backed feature; holds all related work as attached issues.

**Feature selector (v1)**:
The operator passes **only** the parent **Feature**’s GitHub issue number as the work-queue selector (plus shared flags such as **`--repo`**). **Sub-issue** numbers are **not** CLI arguments in v1—operators do not run Forge “against” an individual **Sub-issue** by id.

**State-driven dispatch (v1)**:
On each invocation for a **Feature**, Forge loads attached **Sub-issues**, validates the **Scheduling graph**, reads GitHub merge/PR state (and local git when relevant), and decides the next step—such as running the **Agent runner** for the next **Executable** **Sub-issue**, or continuing an in-flight **Forge-managed PR** when repository state indicates that is correct. The exact branch/PR actions follow **Deterministic branch names (v1)** and **Git publish and PR open (v1)**; operators do not choose the **Sub-issue** at the CLI.

**Concurrency (v1)**:
Forge does **not** acquire locks or coordinate multiple simultaneous runs for the same **Feature** or **Repository root**; operators avoid overlapping invocations that could corrupt a shared checkout or race pushes.

**GitHub repository scope (v1)**:
Forge resolves the **`owner/repo`** for API calls from the **`origin`** remote’s URL when it points at **github.com**; if **`origin`** is missing, not a **github.com** host, or ambiguous, Forge **errors**. A **`--repo owner/name`** (or equivalent) flag **overrides** that default—for example when the issue lives on **upstream** but work happens in a **fork**.

**GitHub product scope (v1)**:
Only the **github.com** SaaS product is in scope. **GitHub Enterprise Server** and other hosts are explicitly out of scope for v1.

**Deterministic branch names (v1)**:
The **Feature branch** and each **Sub-issue** **Stacked branch** are named by Forge from a **fixed template** using the **Feature** issue number and each **Sub-issue** number (and optional stable slug derived from issue text)—never ad hoc names chosen by the **Agent runner**. Exact path separators and prefixes are implementation details; predictability and repeatability are the requirement.

**Sub-issue**:
A GitHub issue attached to a Feature issue; one unit of implementation work with its own PR.

**Blocker**:
Another issue that must be satisfied before this **Sub-issue** may execute; declared only under the **Blocked-by section** in the sub-issue body.

**Blocked-by section**:
A markdown section headed `## Blocked by` whose bullets are same-repository issue references, for example `#15` on its own line under a bullet.

Example:

```markdown
## Blocked by
- #15
```

_Avoid_: Using "dependency" alone for people-edges—prefer **Blocker** for predecessor edges in the DAG.

## Relationships

- A **Feature** groups many **Sub-issues**.
- A **Sub-issue** names zero or more **Blockers** (predecessors); edges come only from its **Blocked-by** section.
- Forge considers **only** a single **Feature** and its attached **Sub-issues** when building the graph; nothing outside GitHub participates.
- Git work is **stacked**: **Feature branch** ← **Stack order** ← … each **Sub-issue** branching from its **stack parent** and opening a PR whose **base** is that parent branch.

Example (branch names illustrative):

```
feature/invoicing         ← from Feature base branch (see below)
 └── issue-101/schema     ← from feature/invoicing
      └── issue-102/api   ← from issue-101 branch
           └── issue-103/ui
```

**Scheduling graph**:
The DAG whose vertices are **Sub-issues** of one **Feature** and whose edges are **Blocker** relations (multiple **Blockers** per **Sub-issue** allowed). The parent **Feature** is used for discovery and validation; it is **not** a vertex and cannot appear as a **Blocker** reference.

**Stack order**:
A **total order** of all **Sub-issues** under a **Feature**, computed by Forge from the **Scheduling graph**: a **valid topological order** (every **Blocker** appears before the **Sub-issue** it blocks), with ties broken strictly by **ascending GitHub issue number** whenever more than one **Sub-issue** could legally appear next. No human override in v1. This linear order is what git stacking follows; it may serialize work that the DAG alone would allow in parallel.

DAG example (edges are **Blocker** → blocked):

```
issue-1 → issue-2
issue-1 → issue-3
issue-2 → issue-4
issue-3 → issue-4
issue-4 → issue-5
```

Same graph, one valid **Stack order** Forge produces with numeric tie-break:

```
issue-1 → issue-2 → issue-3 → issue-4 → issue-5
```

**Stack parent** (for git):
The branch a **Sub-issue**’s work branches from: the **Feature branch** for the first **Sub-issue** in **Stack order**; for every later **Sub-issue**, the branch for the **immediately preceding** **Sub-issue** in **Stack order** (not necessarily its only **Blocker** in the DAG).

**Stack prefix merged**:
Every **Sub-issue** that appears **before** a given one in **Stack order** has its **Forge-managed PR** merged into the stacked branch line. Forge may treat this together with **Blocker satisfied** when deciding what work is allowed to proceed without breaking the DAG or the stack.

**DAG validation**:
Every `#n` in **Blocked by** must resolve to another **Sub-issue** of the **same** parent **Feature**. References to the parent issue, another **Feature**, issues outside that parent’s sub-issue list, or non-GitHub targets are invalid; Forge must reject the graph (or that edge) rather than silently ignoring it.

**Repository root**:
The directory the user runs Forge from in v1—the root of their Git checkout. Forge reads local Git state here (e.g. **current branch**).

**Repository hygiene (v1)**:
Before Forge creates a **Feature branch** or invokes the **Cursor CLI** for work that mutates the checkout, Forge **errors** unless (1) `HEAD` resolves to a **branch** (detached `HEAD` is forbidden), and (2) the index and worktree are **clean**—no staged or unstaged changes to **tracked** files. A future `--force` (or similar) escape hatch may relax this; not required for v1.

**Feature base branch**:
The branch checked out at **Repository root** when Forge creates the **Feature branch** for a **Feature**. In v1 the **Feature branch** is cut from this ref (often `main`, but whatever the user has checked out).

**Feature branch**:
One branch per **Feature**, created from the **Feature base branch** (not implicitly from `origin/main` unless that is what is checked out). It anchors the whole PRD for that parent issue; **Sub-issue** work reaches **Feature base branch** only through the **Feature branch** and stacked merges, not by branching straight from **Feature base branch** for stacked tiers.

**Stacked branch**:
A **Sub-issue**’s implementation branch that is created from its **stack parent** (the branch immediately above it in the stack), never directly from **Feature base branch**. Dependent work targets predecessor branches so autonomous runs are not blocked on unrelated merges upstream.

**PR stack target**:
The **base branch** of a **Sub-issue**’s **Forge-managed PR**: the **stack parent** branch—the previous **Sub-issue**’s line in **Stack order**—so diffs stay small and reviewable. The **Feature branch** is the **PR stack target** for the first **Sub-issue** in **Stack order**.

**Blocker satisfied**:
For a **Blocker** edge *A* → *B*, *A* is satisfied for *B* when *A*’s **Forge-managed PR** has been merged into *A*’s **PR stack target** (per GitHub merge metadata), so *B*’s DAG constraints are met. Non–Forge-managed PRs do not affect scheduling.

**Executable (Sub-issue)**:
A **Sub-issue** that is **Stack-eligible** in the sense Forge uses to start or resume work: every **Blocker** in the DAG is **Blocker satisfied**, and the **Stack prefix merged** condition holds through the predecessor in **Stack order** (so git ancestry matches the flattened sequence).

**Forge-managed PR**:
The pull request GitHub treats as Forge’s for a **Sub-issue**, identified there—not in local-only files. Forge applies the **`forge`** label to that PR and links the PR and **Sub-issue** in GitHub’s development linkage so humans can trace work.

**Forge PR identification**:
Among PRs in the repo, the **Forge-managed PR** for a **Sub-issue** is the one that both carries the **`forge`** label and is linked to that **Sub-issue** via GitHub’s development tracking. There must be **exactly one** such PR (open or merged history per your checks); zero or more than one is an invalid state—Forge must surface an error and refuse ambiguous scheduling. Other PRs are ignored for **Blocker satisfied** and execution.

**Human merge responsibility**:
Reviewers merge **Forge-managed PR**s along the stack (each into its **PR stack target**) and, when they choose, integrate the completed **Feature** line to **default** (or their usual integration branch). Forge does **not** merge stacked branches or **default** in v1; it only observes GitHub’s merge metadata for scheduling.

**Agent runner (v1)**:
Only **Cursor** is supported. Forge starts work by invoking the **Cursor CLI** from **Repository root**, using the same checkout the human is using.

**Pre-agent branch setup (v1)**:
Before each **Cursor CLI** invocation, Forge creates (if missing) and checks out the correct **Feature branch** or **Stacked branch** using **Deterministic branch names (v1)** from the right **stack parent** per **Stack order** and **State-driven dispatch (v1)**. The **Agent runner** does not choose branch names or parents.

**Cursor environment (v1)**:
Forge does **not** inject **`FORGE_*`** (or other Forge-specific) environment variables into the **Cursor CLI** process in v1; the agent relies on the checkout, repository files, and its own configuration.

**TDD and validation (v1)**:
Test-driven development and pre-commit validation are entirely the **Agent runner**’s responsibility. Forge does **not** configure, require, or run a repo test command in v1.

**Agent failure hygiene (v1)**:
If the **Cursor CLI** exits non-success, Forge **does not** rewrite the working tree (no automatic **`git reset --hard`** or similar); the checkout is left **as-is** on the current branch for inspection. Forge exits with a non-zero status and surfaces what it knows about the failure.

**No pending work (v1)**:
When an invocation for **Feature** `<N>` finds no further automated actions, Forge prints **`Feature <N>: no pending work`** to standard output and exits with status **0**.

**Commit authorship (v1)**:
Every **commit** on **Feature branch** and **Stacked branch** lines that Forge manages must come from the **Agent runner** only. Forge does **not** author its own empty or scaffold commits in v1; it relies on git operations (branch pointers, push) and the agent’s commits for content.

**Git publish and PR open (v1)**:
After a successful **Cursor CLI** run for a **Sub-issue**, Forge **pushes** the resulting commits to the appropriate remote and **opens or updates** the **Forge-managed PR** via the **GitHub API client (v1)**—including **base branch**, **`forge`** label, and **Sub-issue** linkage. Humans still **merge** those PRs under **Human merge responsibility**.

**PR body linking (v1)**:
The **Forge-managed PR** body Forge creates includes **`Fixes #<sub-issue>`** for that **Sub-issue** so GitHub can close it when the PR merges (subject to repository automation settings).

**PR draft policy (v1)**:
**Forge-managed PR**s are opened **ready for review** (not **draft**) by default in v1.

**PR title (v1)**:
The **Forge-managed PR** title is **`[#<sub-issue>] <issue title>`**, using the **Sub-issue**’s number and title as returned from GitHub at PR creation time.

**Push remote (v1)**:
Forge pushes to the **`origin`** remote in v1 (no alternate remote name unless added later).

**Cursor CLI**:
The command-line interface of the **Cursor** product; Forge shells out to it (exact flags and subcommands are implementation details outside this glossary).

**Cursor binary resolution (v1)**:
Forge invokes the **`cursor`** command from **`PATH`** by default. If **`FORGE_CURSOR_BIN`** is set to an absolute path, Forge uses that executable instead (for nonstandard installs).

**CLI output (v1)**:
Forge prints **human-oriented text** only in v1; there is **no** **`--json`** (or similar) machine-readable output mode yet.

**GitHub API client (v1)**:
Forge calls GitHub’s **HTTP APIs** (REST and, when needed, GraphQL) from the Go binary—**not** by invoking the **`gh`** CLI for ordinary issue/PR/label operations.

**GitHub token (v1)**:
Credential used with the **GitHub API client (v1)**. Forge resolves it in this order: **`GH_TOKEN`**, then **`GITHUB_TOKEN`**; if neither is set, Forge may run **`gh auth token`** when the GitHub CLI is installed and authenticated—otherwise Forge errors with clear setup guidance.

**Stack consistency violation**:
GitHub’s merge or branch state disagrees with **Stack order** / **Stack prefix merged** expectations—for example a later **Sub-issue**’s **Forge-managed PR** is merged while an earlier **Sub-issue** in **Stack order** is not, or the stack line is otherwise incoherent.

**Stack consistency policy (v1)**:
On read-only inspection (e.g. `forge status`), Forge **warns** and continues so humans can see the problem. On commands that **start or schedule** autonomous work or mutate repo state, Forge **errors** with actionable remediation; it does **not** silently relax **Stack prefix merged** or pretend the stack is valid.

_A future optional label `forge:<feature-issue-number>` may appear if multiple Features in one repo need disambiguation; not required for v1._

## Example dialogue

> **Dev:** "Can a **Sub-issue** block work outside its **Feature**?"
> **Domain expert:** "No—only sibling **Sub-issues**, validated against the parent."

## Flagged ambiguities

- None yet.
