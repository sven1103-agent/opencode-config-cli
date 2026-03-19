# AGENTS.md

## Scope

This repository contains OpenCode agent configuration assets and supporting setup materials.

Primary repo contents:
- `opencode.openai.json` and `opencode.mixed.json` are the main OpenCode config presets
- `.opencode/schemas/` contains the vendored inter-agent handoff and result schemas used by this repo's workflow
- `scripts/opencode-helper` is the iteration-1 helper CLI for bootstrapping `opencode.json` + installing/validating local schemas
- `docs/opencode-helper-cli.md` is the traceable product document for the planned helper CLI
- `README.md` explains the multi-agent configuration approach and how the repo is intended to be used

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

- If the task is about agent behavior or configuration presets, inspect `README.md`, `opencode.openai.json`, and `opencode.mixed.json`
- If the task is about local schema installation or validation, inspect `scripts/opencode-helper`
- If the task is about future helper CLI behavior, update `docs/opencode-helper-cli.md` first so product decisions remain traceable
