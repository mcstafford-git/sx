# Proposal

## Why

The repository's development rules are split across the Makefile, CI workflows, pull request template, README, and agent instructions. A durable OpenSpec capability can make the expected path from planning through verification explicit without duplicating implementation details from those files.

## What Changes

- Define a `development-workflow` capability covering when changes need OpenSpec planning, how implementation work is tracked, and what evidence is required before work is reported complete.
- Configure OpenSpec with concise project constraints and apply/archive guidance derived from the repository's existing rules.
- Add a contributor-facing entry point that explains how OpenSpec fits alongside issues, pull requests, existing technical documentation, and `make prepush`.
- Update the pull request workflow so authors can reference an OpenSpec change when one exists, while preserving a lightweight path for changes that do not alter specified behavior.
- Keep release operations and the migration of existing technical documents into OpenSpec out of scope for this first change.

## Capabilities

### New Capabilities

- `development-workflow`: Planning, implementation tracking, verification, and completion requirements for repository changes.

### Modified Capabilities

None.

## Impact

- Adds the first durable capability under `openspec/specs/` when this change is archived.
- Updates `openspec/config.yaml` and contributor-facing workflow documentation.
- May update `.github/pull_request_template.md` to capture OpenSpec and verification evidence.
- Does not change the `sx` CLI, desktop app, public APIs, dependencies, CI commands, or release automation.
