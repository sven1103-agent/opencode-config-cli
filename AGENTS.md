# AGENTS.md

## Scope

This repository contains the OpenCode Helper CLI - a tool for managing OpenCode configuration bundles.

Primary repo contents:
- `cmd/` contains the Go CLI implementation
- `internal/` contains internal packages
- `scripts/opencode-helper` is the shell-based helper for bootstrapping
- `docs/opencode-helper-cli.md` is the traceable product document for the helper CLI
- `README.md` explains the helper CLI and how to use configuration bundles

Note: OpenCode configuration presets have been moved to [qbicsoftware/opencode-config-bundle](https://github.com/qbicsoftware/opencode-config-bundle)

## Working Expectations

## Preflight (Required)

Before doing any substantive work for a new user story or implementation task in this repo:
- If the request is read-only (analysis, planning, review, explanation), stay in the current workspace and do not create a worktree.
- Otherwise, treat the request as a new unit of work and create a dedicated branch + worktree immediately, even if the current branch is not `main`.
- Never implement a new user story in an existing branch/worktree that was created for a different task.
- First propose: branch name, worktree path, and base branch (`origin/main`), then wait for explicit user confirmation.
- Run: `git fetch` and `git worktree list` (avoid duplicate worktrees).
- Create the worktree: `git worktree add ".worktrees/<branch>" -b "<branch>" origin/main`.
- Run all subsequent commands with `workdir=.worktrees/<branch>`.

- Keep changes aligned with the repo's planning-first, schema-validated workflow
- Preserve the role of `.opencode/schemas/handoff.schema.json` and `.opencode/schemas/result.schema.json` as the local canonical artifact contracts for this repo
- Prefer additive, traceable documentation changes over informal notes scattered across files
- Treat traceability as a first-class requirement: every implementation PR must declare a single primary user story ID (`US-###`) and follow the traceability rules in `docs/opencode-helper-cli.md` (primary story selection, multi-story handling, and when it is allowed to mark `Status: Done`).
- For multi-agent work, keep handoffs and results traceable through the session artifact workflow under `.opencode/sessions/` and validate those artifacts against `.opencode/schemas/handoff.schema.json` and `.opencode/schemas/result.schema.json`
- Treat helper CLI planning in `docs/opencode-helper-cli.md` as the source of truth for requirements, features, and later user stories
- Avoid changing unrelated session artifacts under `.opencode/sessions/`

## Git Worktrees (Branch + Worktree Per PR)

- Default: every task/change is done in its own branch and its own git worktree, and results in its own PR (even single-file changes)
- Trigger: create the branch/worktree as soon as the request becomes a new implementation unit of work, not only when the current branch happens to be `main`
- Never do feature work directly on `main`
- Never reuse an existing branch/worktree for a different user story or unrelated implementation task
- Worktree location convention: `.worktrees/<branch>/` (repo-local)
- Workflow:
  - For read-only requests, stay in the current workspace
  - `git fetch`
  - `git worktree list` (avoid duplicates)
  - Create a new branch worktree from `origin/main`: `git worktree add ".worktrees/<branch>" -b "<branch>" origin/main`
  - Run subsequent commands using the worktree directory as the working directory (avoid `cd ... &&`)
- Confirmation gate: before creating a new branch/worktree or pushing/creating a PR, state the planned branch name + worktree path + base branch and wait for explicit user confirmation
- Cleanup: only remove worktrees when the PR is merged/closed (or the user asks): `git worktree remove ".worktrees/<branch>"`

## Quick Orientation

- If the task is about the helper CLI, inspect `README.md` and `cmd/`
- If the task is about configuration bundles, see [qbicsoftware/opencode-config-bundle](https://github.com/qbicsoftware/opencode-config-bundle)
- If the task is about future helper CLI behavior, update `docs/opencode-helper-cli.md` first so product decisions remain traceable
