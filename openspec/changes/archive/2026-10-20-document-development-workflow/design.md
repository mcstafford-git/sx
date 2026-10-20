# Design

## Context

See `proposal.md` for motivation. The repository already has executable quality gates in the Makefile and CI, contributor guidance in the README and pull request template, detailed behavioral contracts under `docs/`, and agent-specific rules supplied outside OpenSpec. OpenSpec is newly initialized with the `spec-driven` schema and has no archived capabilities yet.

## Goals / Non-Goals

**Goals:**

- Give contributors and coding agents one documented path from behavior proposal to verified implementation.
- Keep durable requirements in OpenSpec while leaving commands and technical reference material in their current authoritative files.
- Make OpenSpec useful without requiring every typo fix, refactor, or operational note to invent a behavior delta.
- Add enough project guidance that future OpenSpec artifacts consistently respect the repository's existing completion rules.

**Non-Goals:**

- Change build, test, lint, CI, release, or deployment behavior.
- Convert the existing `docs/` specifications into OpenSpec in bulk.
- Add CI enforcement for OpenSpec validation in the first adoption change.
- Define release-management requirements; that remains a separate future capability.

## Decisions

### OpenSpec owns workflow requirements, not command definitions

The `development-workflow` capability will state durable obligations such as planning behavior changes, keeping tasks traceable, running the completion gate, and verifying before archival. The Makefile remains authoritative for what `make prepush` executes, and CI remains authoritative for automation details.

This avoids copying a list of commands that would drift whenever the Makefile changes. The alternative—describing every format, lint, test, code-generation, and build step in the spec—was rejected because it would create two sources of truth.

### Existing technical documents remain authoritative within their domains

Documents such as `docs/manifest-spec.md`, `docs/lock-spec.md`, and `docs/rbac.md` continue to describe detailed product contracts until a deliberate future change migrates a capability. The workflow will require updating affected reference documents, but it will not mirror their contents into a generic OpenSpec capability.

The alternative—importing all existing documents during initial adoption—was rejected because it would create a large, hard-to-review change and obscure whether the imported requirements are current.

### Contributor guidance uses a focused workflow document

Add `docs/development-workflow.md` as the human-readable entry point and link it from the README's development section. It will explain when to use OpenSpec, the propose/apply/verify/archive lifecycle, how existing docs fit, and the mandatory completion gate.

Putting the full workflow directly in the README was rejected because the README already serves product users, library consumers, and contributors. A short link keeps that page scannable.

### Pull requests record planning and verification without blocking exceptions

Extend `.github/pull_request_template.md` with an optional OpenSpec change reference and a concise verification result. Behavior-changing pull requests should name their change; work with no durable behavior impact can state that OpenSpec is not applicable.

Automated enforcement is deferred. A hard CI check would need reliable diff-to-capability classification and an exception policy, neither of which is necessary to establish the workflow.

### OpenSpec configuration carries only stable, non-discoverable guidance

Update `openspec/config.yaml` with brief constraints that planning must account for cross-platform behavior when relevant, keep commit subjects within 72 characters, never push without explicit permission, and run `make prepush` before reporting code work complete. Add apply and archive guidance rather than copying repository architecture or command internals into configuration.

This keeps generated artifacts consistent while respecting the config file's instruction to omit facts agents can discover from source.

## Risks / Trade-offs

- **Workflow guidance can drift from executable gates** -> Refer to `make prepush` by name and keep its step definitions only in the Makefile.
- **Contributors may overuse OpenSpec for trivial edits** -> Document the durable-behavior boundary and allow explicit no-spec changes.
- **Contributors may ignore an unenforced workflow** -> Surface it in the README and pull request template, then evaluate CI enforcement after adoption experience.
- **Existing technical documents may conflict with future OpenSpec specs** -> Migrate capabilities only through focused, reviewable changes and name the current authority during each migration.

## Migration Plan

1. Add the workflow guidance and OpenSpec configuration without changing executable repository behavior.
2. Validate this change strictly and run the existing repository completion gate.
3. Use the workflow on subsequent behavior-changing work.
4. Verify and archive this change so `development-workflow` becomes the first durable capability.

Rollback consists of reverting the contributor guidance and configuration updates before archival. After archival, a follow-up OpenSpec change should modify or remove the capability requirements rather than deleting history.
