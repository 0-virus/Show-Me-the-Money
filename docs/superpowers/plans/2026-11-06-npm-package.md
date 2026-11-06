# npm Package Implementation Plan

> **For agentic workers:** Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Prepare and locally verify @halmasy1/moneycopy@1.0.3 without publishing.
**Architecture:** Existing Go binaries bundled with a dependency-free Node CLI. Python prepares the files from the versioned dist directory before npm pack.
**Tech Stack:** Go native executables, Node CommonJS, npm, Python standard library.
**Spec:** ../specs/2026-11-06-npm-package-design.md

## Global Constraints
- Name @halmasy1/moneycopy; command moneycopy; version 1.0.3.
- Windows x64, macOS x64/arm64. No runtime downloads or install hooks.
- Reuse current dist; never publish or change credentials.

## Review Focus
- Paths containing spaces and shell punctuation in arguments: spawn without a shell.
- Unsupported OS/architecture and missing binaries: actionable error, nonzero exit.
- Child failures/signals: preserve exit semantics.
- Mac executable permissions in Windows-created tarballs: check and repair at launch.
- Stale binaries or unexpected files in tarball: verify versions, headers, allowlist, installed --version.

## Task 1: CLI and package preparation
Files: package.json, bin/moneycopy.cjs, scripts/prepare-npm.py, tests/npm-launcher.test.cjs, .gitignore.
Interfaces: resolveBinary(platform, arch, root), run(args, options) -> exit code; prepack generates native/ and THIRD-PARTY-LICENSES.txt from dist.
- [ ] Write failing tests for platform routing, argument forwarding, errors and child status.
- [ ] Implement launcher and manifest; implement preparation and version checks.
- [ ] Run node tests and npm pack; inspect actual tar contents and native bytes.
- [ ] Commit launcher and packaging configuration.

## Task 2: Distribution verification and instructions
Files: README.md, docs/verification.md, scripts/verify-npm.py.
- [ ] Install the actual tarball into artifacts/npm-smoke using --ignore-scripts and a local prefix; execute installed shim --version and play in a terminal.
- [ ] Validate tar allowlist, metadata, native hashes and Mac headers/permissions.
- [ ] Document preparation, packing, login, publish, npx and global-install commands; label publication as pending.
- [ ] Perform one fresh code review, resolve material findings, repack final docs and commit.
