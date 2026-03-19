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

- Keep changes aligned with the repo's planning-first, schema-validated workflow
- Preserve the role of `.opencode/schemas/handoff.schema.json` and `.opencode/schemas/result.schema.json` as the local canonical artifact contracts for this repo
- Prefer additive, traceable documentation changes over informal notes scattered across files
- Treat traceability as a first-class requirement: when a PR implements a user story, start the title with `US-###`, include `Implements: US-###` in the PR body, and update the source story in `docs/opencode-helper-cli.md` with `Status: Done` plus the PR link or number
- For multi-agent work, keep handoffs and results traceable through the session artifact workflow under `.opencode/sessions/` and validate those artifacts against `.opencode/schemas/handoff.schema.json` and `.opencode/schemas/result.schema.json`
- Treat helper CLI planning in `docs/opencode-helper-cli.md` as the source of truth for requirements, features, and later user stories
- Avoid changing unrelated session artifacts under `.opencode/sessions/`

## Git Worktrees (Branch + Worktree Per PR)

- Default: every task/change is done in its own branch and its own git worktree, and results in its own PR (even single-file changes)
- Never do feature work directly on `main`
- Worktree location convention: `.worktrees/<branch>/` (repo-local)
- Workflow:
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
