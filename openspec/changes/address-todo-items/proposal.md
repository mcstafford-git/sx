# Proposal

## Why

The repository audit in `TODO.md` identifies failure paths that can bypass extension policy, exhaust resources, destroy working installations, or lose shared settings. Close those gaps and replace unreliable or skipped tests with repeatable checks before users encounter them during routine CLI/app use.

## What Changes

- **BREAKING:** On extension-policy retrieval failure, use a valid per-profile cached policy or disable extensions; never assume open access from a missing or invalid cache.
- **BREAKING:** Enforce fixed archive limits: 10 MiB compressed per asset package, 50 MiB compressed per repository discovery ZIP, 50 MiB total expanded bytes, 10 MiB per expanded file, and 1,000 archive entries. Exact limits are allowed; exceeding them fails with an actionable error.
- Bound downloads and archive reads before metadata inspection, hashing, or installation, including progress downloads and cached assets.
- Preserve working directory installations when validation, extraction, or replacement fails by staging replacements and recovering the previous version on promotion failure.
- Extend shared-config protection from telemetry preservation to complete read-modify-write transactions across CLI and desktop callers; retain atomic replacement and consent semantics.
- Classify Git proxy tunnel failures as network errors and make classifier coverage independent of live DNS/proxy behavior.
- Implement all 17 currently skipped scope-editor flows without changing the documented scope semantics.
- Reconcile `TODO.md` with verified outcomes during implementation, including its outdated config-write description and flow count.

## Capabilities

### New Capabilities

- `extension-policy`: Failure-safe retrieval and caching of desktop extension policy.
- `asset-archives`: Fixed resource budgets for asset downloads, repository discovery ZIPs, and archive processing.
- `asset-installation`: Preservation of existing directory assets across failed updates.
- `shared-configuration`: Serialized configuration mutations that preserve unrelated settings and consent.
- `git-remote-errors`: Actionable classification of remote Git failures, including proxy connection failures.

### Modified Capabilities

None. The existing `telemetry-consent` contract is preserved and regression-tested. Scope-editor work adds tests for existing behavior in `docs/scoping.md`; it does not introduce a specification delta solely for test coverage.

## Impact

- Desktop policy lookup/cache in `app/plugins.go` and associated frontend policy consumption tests.
- Shared ZIP utilities and metadata readers, asset fetch/cache paths, HTTP/hosted-vault downloads, GitHub and marketplace repository discovery.
- Directory installation handlers, shared config APIs, CLI config/profile/client/init callers, and desktop settings/repository callers.
- Git error classification tests and the non-TTY scope-editor test harness.
- Reference documentation for policy fallback, archive budgets, update failures, and shared state; audit TODO reconciliation.

## Non-Goals

- Configurable archive limits, new dependencies, policy redesign, or changing explicit successful open policies.
- Limiting the size of a Git clone or placing a single budget on all assets in a vault.
- Changing scope meanings, adding access-control features, or migrating existing vault/config formats unnecessarily.
- Promising uninterrupted access or power-loss recovery during a multi-step directory replacement; recovery from ordinary reported failures is in scope.
