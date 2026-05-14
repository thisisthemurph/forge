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
