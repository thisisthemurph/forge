# Forge CLI — contributor guide (v1)

This page orients contributors and operators to the shipped v1 CLI.

## Agent runner (v1)

Forge’s **Agent runner** in v1 is the **Cursor CLI** only: Forge resolves a `cursor` binary on `PATH`, or an absolute path via **`FORGE_CURSOR_BIN`** when you need a nonstandard install. Other agent backends are out of scope until a later version.

TDD, tests, and code edits happen inside that agent process; Forge does not run your test suite for you.

## Human merge responsibility

Forge opens and updates **Forge-managed PR**s but **never merges** them and never merges to **default** for you. Review and merge remain human steps.

## Concurrency

Forge does not take locks across terminals or machines. Avoid overlapping `forge run` invocations against the same checkout or feature queue so pushes and branch state stay coherent.

## Security: tokens

Forge resolves a **GitHub token** for API calls and **must not** print token values in logs or normal CLI output. Treat CI secrets like any other credential.

## GitHub token (v1)

Resolution order matches **CONTEXT.md**: **`GH_TOKEN`**, then **`GITHUB_TOKEN`**. If both are unset or blank, Forge may invoke the GitHub CLI’s token helper when **`gh auth`** is configured (`gh auth login`); otherwise authentication fails with a clear message.

## Repository targeting and `--repo`

Forge derives **`owner/repo`** from **`origin`** when that remote uses **github.com**. Pass **`--repo owner/name`** when tracking issues on **upstream** while **`origin`** points at your **fork**, so the GitHub API targets the repository where the **Feature** issue lives.

## CLI output (v1)

Output is plain text for humans; there is no **`--json`** mode in v1.

## Glossary

**Forge PR identification**, **Stack consistency policy**, **Executable**, and the rest of the scheduling vocabulary are defined in [CONTEXT.md](../../CONTEXT.md). Prefer that file’s terms in issues and code so user-facing errors stay aligned with the product language.
