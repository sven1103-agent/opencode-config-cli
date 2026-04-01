# CLI E2E Coverage

## Purpose

This document tracks black-box end-to-end coverage for the shipped Go CLI binary (`oc`).

It is the durable inventory for what is already covered, what is verified in CI, and what is still deferred.

## Scope

- Test the built `oc` binary, not internal package functions
- Focus on deterministic, offline scenarios first
- Track both happy paths and operator-facing failure paths
- Keep coverage aligned with supported Linux and macOS release targets

## Principles

- Run tests against the compiled binary via `os/exec`
- Isolate each test with temp directories and isolated user config state
- Keep fixtures local and deterministic
- Avoid live network dependencies in CI
- Expand coverage incrementally and update this document in the same PR

## Workflow

- Workflow file: `.github/workflows/e2e-cli.yml`
- Binary handoff env var: `OC_E2E_BINARY`
- Current CI platforms:
  - `ubuntu-latest`
  - `macos-latest`

## Test Harness

- Package: `e2e/`
- Fixture root: `e2e/testdata/`
- Execution model:
  - build `oc`
  - invoke commands via subprocesses
  - set isolated `HOME` and `XDG_CONFIG_HOME`
  - verify stdout, stderr, exit status, and written files

## Covered Scenarios

| Scenario | Command(s) | Source Type | Status | CI Verified | Notes |
|---|---|---|---|---|---|
| Print version | `oc version` | n/a | Implemented | No | Verifies built binary launches and prints version |
| Register local directory source | `oc source add <dir>` | local directory | Implemented | No | Uses fixture bundle with manifest |
| Register local archive source | `oc source add <tar.gz>` | local archive | Implemented | No | Archive generated at test runtime |
| List registered sources | `oc source list` | registry | Implemented | No | Confirms isolated registry contents |
| Apply preset from local directory source | `oc bundle apply <id> --preset <name>` | local directory | Implemented | No | Verifies written config and provenance |
| Apply preset from local archive source | `oc bundle apply <id> --preset <name>` | local archive | Implemented | No | Exercises tar extraction path |
| Show applied bundle provenance | `oc bundle status --project-root <dir>` | project provenance | Implemented | No | Reads `.opencode/bundle-provenance.json` through CLI |
| Refuse overwrite without force | `oc bundle apply <id> --preset <name>` | local directory | Implemented | No | Must fail when output already exists |
| Reject missing manifest source | `oc source add <dir>` | invalid local directory | Implemented | No | Validates manifest presence check |
| Reject invalid tarball | `oc source add <tar.gz>` and/or `oc bundle apply <id> --preset <name>` | invalid local archive | Implemented | No | Verifies archive extraction failure path |
| Reject unknown source ID | `oc bundle apply <id> --preset <name>` | registry lookup | Implemented | No | Verifies user-facing error path |

## Deferred Scenarios

| Gap | Reason | Planned Follow-up |
|---|---|---|
| GitHub release source E2E | Current implementation still has remote operations marked not implemented in this branch line | Add once remote source behavior is stable end to end |
| `httptest`-backed remote source coverage | Not needed for initial local-flow milestone | Add with remote-source implementation follow-up |
| `bundle update` E2E | Current command behavior is intentionally limited and network-dependent | Add when update behavior is production-ready |
| Windows runner coverage | Windows is out of current scope | Revisit only if product scope changes |

## Fixture Strategy

- Keep a small valid local bundle fixture in `e2e/testdata/fixture-bundle/`
- Generate `.tar.gz` archives during test execution instead of committing binary archives
- Keep invalid cases lightweight and explicit

## Environment Isolation

- Each test uses a fresh temporary project root
- Each test uses a fresh temporary config home
- Subprocess env sets:
  - `HOME`
  - `XDG_CONFIG_HOME`

## Change Log

- 2026-04-01: Document created for `US-052` initial local-flow coverage planning
- 2026-04-01: Initial local-flow scenarios implemented in `e2e/` and wired into `.github/workflows/e2e-cli.yml`
