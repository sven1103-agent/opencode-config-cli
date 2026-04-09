# Bundle Contract

This document defines the contract that configuration bundles must comply with to work with the `oc` CLI.

## Ownership

The bundle contract is **owned by the opencode-config-cli repository** where the CLI lives. All schema definitions, validation rules, and contract changes are managed here.

Configuration bundles (like [qbicsoftware/opencode-config-bundle](https://github.com/qbicsoftware/opencode-config-bundle)) are **consumers** of this contract.

---

## Bundle Manifest Schema

### File Location

```
<bundle-root>/
  opencode-bundle.manifest.json  <- required at bundle root
  <preset-entrypoint>.json       <- preset files referenced by manifest
  .opencode/schemas/
    handoff.schema.json          <- canonical handoff contract
    result.schema.json           <- canonical result contract
```

### Schema (JSON Schema Draft 2020-12)

The schema is **embedded in the CLI binary** for offline validation.

The single source of truth is `internal/bundle/1.0.0.schema.json` in the repository.
This file is embedded into the `oc` CLI binary at build time.

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "type": "object",
  "required": ["manifest_version", "bundle_name", "bundle_version", "presets"],
  "properties": {
    "manifest_version": {
      "type": "string",
      "pattern": "^[0-9]+\\.[0-9]+\\.[0-9]+$",
      "description": "Semantic version of the bundle manifest schema (e.g., 1.0.0). CLI supports this version and N-1."
    },
    "bundle_name": {
      "type": "string",
      "description": "Stable identifier for the bundle"
    },
    "bundle_version": {
      "type": "string",
      "description": "Release tag or version string (e.g., v1.0.0)"
    },
    "source_repo": {
      "type": "string",
      "description": "URL of source repository"
    },
    "source_commit": {
      "type": "string",
      "description": "Git commit SHA"
    },
    "release_tag": {
      "type": "string",
      "description": "Release tag"
    },
    "presets": {
      "type": "array",
      "items": {
        "type": "object",
        "required": ["name", "description", "entrypoint"],
        "properties": {
          "name": {
            "type": "string",
            "description": "Stable preset identifier (lowercase, hyphens allowed)"
          },
          "description": {
            "type": "string",
            "description": "Short human-readable description"
          },
          "entrypoint": {
            "type": "string",
            "description": "Path to preset JSON file relative to bundle root"
          },
          "prompt_files": {
            "type": "array",
            "items": { "type": "string" },
            "description": "Array of prompt file paths to materialize (can be empty)"
          }
        }
      }
    }
  }
}
```

### Required Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `manifest_version` | string (semver) | Yes | Must be `1.0.0` or compatible. CLI supports this version and N-1. |
| `bundle_name` | string | Yes | Stable identifier for the bundle |
| `bundle_version` | string | Yes | Release tag or version string |
| `presets` | array | Yes | Array of preset descriptor objects |

### Preset Descriptor Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Stable identifier (lowercase, hyphens allowed) |
| `description` | string | Yes | Short human-readable description |
| `entrypoint` | string | Yes | Path to preset JSON file relative to bundle root |
| `prompt_files` | array | Yes | Array of prompt file paths (can be empty) |

### Optional Fields

| Field | Type | Description |
|-------|------|-------------|
| `source_repo` | string | URL of source repository |
| `source_commit` | string | Git commit SHA |
| `release_tag` | string | Release tag (often matches `bundle_version`) |

---

## CLI Validation Rules

The CLI rejects a bundle when:

- `opencode-bundle.manifest.json` is missing from bundle root
- Manifest is not valid JSON
- `manifest_version` is unsupported (not semver-compatible with CLI's supported versions)
- Any required top-level field is missing
- Any preset is missing required fields
- An `entrypoint` path does not exist in the bundle
- A `prompt_files` entry does not exist in the bundle

---

## GitHub Release Distribution

This section defines how a bundle must be published when it is distributed as a GitHub-release source for the `oc` CLI.

### Supported Release Types

- Stable releases are valid bundle sources.
- Prereleases are also valid bundle sources and are recommended for smoke testing.
- A bundle repository may publish only prereleases, only stable releases, or both.

The CLI should not treat a prerelease-only repository as invalid by default.

### Version Selection Semantics

For GitHub-release sources, the CLI is expected to support these selection modes:

- explicit version tag selection via `--version <tag>`
- `latest` selection for the latest stable release
- interactive version selection when multiple usable versions exist and the command is running in a TTY

Interactive selection should list versions newest first and clearly label prereleases.

If a repository has no stable releases but does have prereleases, the CLI should surface that clearly instead of returning a generic GitHub API `404` error.

### Required Release Assets

For a GitHub-release bundle source, the release must publish these assets:

- a bundle archive, typically `opencode-config-bundle-<tag>.tar.gz`
- a checksum file for the published bundle archive, typically `opencode-config-bundle-<tag>-checksums.txt`

The bundle archive must unpack into a normalized bundle root that contains exactly one `opencode-bundle.manifest.json` and all files referenced by that manifest.

The checksum file must:

- use SHA-256
- contain an entry for the published bundle archive
- match the uploaded archive byte-for-byte

### Archive Layout

The uploaded archive must contain either:

- the bundle files directly at the archive root, or
- a single top-level directory that contains the bundle root

In both cases, after extraction the CLI must be able to resolve a normalized bundle root containing:

```text
<bundle-root>/
  opencode-bundle.manifest.json
  <preset-entrypoint>.json
  .opencode/schemas/
    handoff.schema.json
    result.schema.json
```

### GitHub Auto-Generated Source Archives

GitHub's auto-generated source archives (`tarball_url` / `zipball_url`) should not be relied on as the primary distribution contract for bundle consumers.

Reasons:

- their naming is not bundle-contract-specific
- they do not provide contract-specific checksum metadata by default
- they make the intended consumable artifact less obvious to maintainers and users

For predictable distribution and verification, maintainers should publish an explicit bundle archive asset and matching checksum asset with each release.

---

## Example GitHub Actions Release Workflow

This example shows a minimal release workflow for a bundle repository. It stages the bundle contents, creates the release archive, generates a SHA-256 checksum file, and uploads both assets to the GitHub release.

```yaml
name: Release Bundle

on:
  push:
    tags:
      - "v*"
      - "*-alpha.*"
      - "*-beta.*"
      - "*-rc.*"

permissions:
  contents: write

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - name: Check out repository
        uses: actions/checkout@v4

      - name: Stage bundle contents
        run: |
          set -eu
          mkdir -p dist/bundle/.opencode/schemas
          cp opencode-bundle.manifest.json dist/bundle/
          cp opencode.*.json dist/bundle/
          cp .opencode/schemas/handoff.schema.json dist/bundle/.opencode/schemas/
          cp .opencode/schemas/result.schema.json dist/bundle/.opencode/schemas/

      - name: Create bundle archive
        run: |
          set -eu
          tar -czf "opencode-config-bundle-${GITHUB_REF_NAME}.tar.gz" -C dist bundle

      - name: Generate SHA-256 checksums
        run: |
          set -eu
          shasum -a 256 "opencode-config-bundle-${GITHUB_REF_NAME}.tar.gz" > "opencode-config-bundle-${GITHUB_REF_NAME}-checksums.txt"

      - name: Upload release assets
        uses: softprops/action-gh-release@v2
        with:
          files: |
            opencode-config-bundle-${{ github.ref_name }}.tar.gz
            opencode-config-bundle-${{ github.ref_name }}-checksums.txt
```

Maintainers may generate the manifest file ahead of time or as part of the release workflow, but the uploaded archive must always satisfy the bundle-root contract above.

---

## Future Extensions

### US-EXT-001 - Support Custom Tools in Bundles

**Status**: Backlog  
**Milestone**: v1.1.0

Add optional `tools` field to support custom OpenCode tools in bundles:

```json
{
  "manifest_version": 1,
  "bundle_version": "v1.1.0",
  "tools": [
    {
      "name": "database",
      "entrypoint": ".opencode/tools/database.ts",
      "description": "Query the project database"
    }
  ]
}
```

See [Custom Tools Docs](https://opencode.ai/docs/custom-tools) for background on how OpenCode tools work.

---

*Contract version: 1.0.0*
*Last updated: 2026-04-01*
