# Design

## Context

See `proposal.md` for motivation and the five capability deltas for observable behavior. The audit predates the telemetry work: `SaveMultiProfile` now locks, preserves latest consent, and uses `utils.WriteFileAtomic`, but callers still load and mutate other fields before that lock is acquired. `SaveToProfile`, CLI profile/client/init commands, and desktop settings/repository methods still submit whole snapshots.

Policy failure spans `GetPluginPolicy`, the cache-only `CachedPluginPolicy` startup API, and `app/frontend/src/plugins/policy.ts`, whose initial and exception states are open. Fixing only the live Go API leaves a startup execution window.

ZIP handling spans HTTP and hosted-vault downloads, cached asset metadata inspection, shared ZIP helpers, GitHub directory discovery, and direct archive readers. `metadata.ValidateZip` reads metadata before installation, so limits introduced only at extraction are too late. GitHub discovery currently downloads individual files and assembles a ZIP; no whole-repository ZIP endpoint was found in the inspected source.

`dirasset.Operations.Install` removes the old directory before extraction. Packaged MCP handlers in multiple clients extract directly into existing directories and can partially overwrite them. These installation paths share the archive-preservation problem even though the audit points to only one function.

The scope-editor file has 17 skipped flows. Its comments and `add_scope_keep_test.go` document a numbered-menu non-TTY harness, fake bulk setters, and actor resolution. Some surrounding stubs describe an older file/git identity-scope distinction; current `docs/scoping.md`, `docs/teams.md`, `docs/users.md`, `docs/bots.md`, and actual supported vault capabilities remain the behavioral authorities.

## Goals / Non-Goals

**Goals:** Close each audited failure at its shared boundary, keep tests deterministic, and preserve existing valid configuration, scope semantics, and telemetry consent.

**Non-Goals:** Do not add configurable limits, generic transaction frameworks, new remote download modes, or new scope semantics. Do not redesign extension trust or promise a power-loss-safe multi-directory transaction. This change protects directory preparation/promotion; an unrelated client-config write after a successful directory commit is not a newly introduced multi-resource transaction.

## Decisions

### 1. Use a disabled fallback consistently across the policy lifecycle

Change failed live retrieval and cache-only bootstrap to use disabled fallback. Validate cached modes against the existing supported open/allowlist/disabled values and normalize allowed lists; missing, unreadable, truncated, malformed, or unknown-mode caches are unusable. Keep profile partitioning and successful no-policy/no-policy-support behavior open. Cache successful results using the existing atomic-write helper so readers do not observe truncated cache writes.

Change the frontend initial state, bridge-error fallbacks, and malformed-result handling to disabled. Preserve existing allowlist treatment of built-ins and the existing total-off meaning of disabled. Tests must cover cached code plus prior user consent with missing policy, not just the returned Go enum.

Keeping the present open bootstrap while hardening live retrieval was rejected because third-party code can execute before the live call finishes. Requiring live policy on every startup was rejected because a valid per-profile cache already supports offline use.

### 2. Define two compressed budgets and one shared expansion policy

Use named fixed constants with binary units:

| Input or resource | Limit |
|---|---:|
| Asset package compressed bytes | 10 MiB |
| Repository discovery ZIP compressed bytes | 50 MiB |
| Expanded bytes across one archive | 50 MiB |
| Expanded bytes in one file | 10 MiB |
| Entries in one archive, directories included | 1,000 |

These are per-input limits, not whole-vault or Git-clone limits. Repository discovery assembled from files has the same file/count/aggregate expansion budgets and a 50 MiB cap on assembled compressed output; a selected package subsequently consumed as an asset uses 10 MiB. Existing raw Git cloning remains unchanged. Do not create a repository ZIP endpoint just to enforce the larger budget.

Centralize bounded reads and ZIP policy in existing utility boundaries. Use limit-plus-one reads to distinguish exactly-at-limit from overflow; content length and ZIP headers permit early rejection but are never the only enforcement. Stream expansion with a remaining aggregate counter, bound each file's actual output, and preserve checksum/read errors, executable bits, and zip-slip checks. Stop before creating excess entries. Close readers and writers promptly inside loops.

Apply the shared policy to metadata reads, hashing/copying/repackaging, archive listings, cached package consumption, and direct archive consumers such as vault adapters. Cap bytes before HTTP accumulation in both ordinary and progress paths, hosted asset reads, and GitHub file discovery. Bound GitHub discovery traversal and assembled output, rather than fetching unlimited files and checking only after allocation. Use bounded reads for local cached asset data where possible; validate cached inputs before metadata inspection and reject/evict over-limit cache entries rather than trusting them.

A larger repository budget must be an explicit internal use of the repository policy; default asset helpers must not silently inherit it. Keep error messages specific to the exceeded limit and preserve those errors through wrappers, including marketplace installation errors.

Extraction-only limits and trusting declared sizes were rejected because metadata inspection and unknown/chunked-length downloads happen earlier. New flags/environment overrides were rejected in favor of the agreed fixed policy.

### 3. Stage directory updates and retain the old version until promotion succeeds

Add a focused shared directory-install helper alongside the existing archive/install utilities. Extract a bounded, validated ZIP into a temporary sibling directory on the destination filesystem. Keep the original directory untouched during preparation. For replacement, move the old directory to a unique sibling backup, promote the prepared directory, then remove the backup. If promotion fails, restore the backup. If restoration fails, preserve and name the backup in the returned error; cleanup must never delete the only surviving working version.

Serialize competing replacements for the same destination with the existing cross-platform file-lock dependency, using a stable lock path outside the renamed directories. Do not hold a vault lock or perform network fetches while holding the install lock. Clean failed staging directories; distinguish successful promotion with backup-cleanup failure from failed installation. A first install failure leaves no final partial asset.

Wire `dirasset.Operations.Install` and packaged MCP directory extraction paths through this helper; leave temporary scratch extraction and file-based handlers on their existing paths with the new archive limits. Validate metadata before promotion and preserve file permissions. Use controlled filesystem-operation failures in helper tests to cover both promotion and rollback errors without relying on Unix permissions alone.

Replacing a nonempty directory with one rename is not portable. In-place extraction or deleting the old tree first cannot meet the preservation contract. A brief rename gap is acceptable; uninterrupted access and crash-recovery journaling are explicitly outside scope. Windows file-in-use failures must leave the old installation or its recoverable backup intact.

### 4. Perform configuration mutations against current state under one lock

Add a focused config mutation API that acquires the existing shared telemetry/config lock, loads/migrates the current configuration, applies a short field-specific mutation callback, validates resulting profile invariants, and atomically saves through the private writer. Missing configuration is allowed only for explicit creation; malformed existing configuration is an error, never a reason to start from an empty map.

Migrate production read-modify-write callers: `SaveToProfile`, CLI profile/client/init updates and identity prompts, desktop settings/repository changes, and profile migration helpers. Gather prompts and network results before entering the transaction; revalidate the affected profile and requested action against current state inside it. Update only requested fields: saving a new credential or client preference must not restore another profile's old identity or credentials. An explicitly requested full profile replacement may replace that profile, while retaining unrelated profiles and global state.

Retire the unrestricted public snapshot-save path from production callers instead of maintaining two independently safe-looking APIs. Keep the low-level full-file writer private and require lock ownership at its call sites. Adapt tests/fixtures to the new mutation or initialization API. Preserve runtime override behavior, active/default-profile invariants, legacy serialization, nil-versus-empty client-list semantics, and the existing dedicated consent writer.

Reuse the lock already shared by reporting and consent. Provide an internal locked mutation path for consent so it does not recursively acquire the same lock. Ordinary mutations cannot change telemetry consent; only explicit telemetry actions do. Commits on the same setting are serialized with the later valid action winning; unrelated field edits both survive. A deleted profile is not resurrected by a delayed prompt.

A lock around only `SaveMultiProfile` was rejected because the stale read precedes it. Generic whole-file last-writer-wins and automatic merging of full snapshots were rejected because they conceal intent and can recreate deleted profiles. Transaction callbacks express the actual requested fields with less ambiguity.

### 5. Test error classification with fixed Git output

Extend `classifyRemoteError` with case-insensitive proxy-tunnel failure matching, including the observed CONNECT/502 output. Keep existing auth, repository-not-found, known-network, and generic categories intact; do not match an arbitrary `502` token. Add table-driven tests with supplied output strings and ordinary synthetic errors. Retain small command-wiring coverage without relying on an intentionally invalid public hostname to select a particular error.

Do not change proxy environment variables or require NO_PROXY exceptions to pass the normal suite. The recorded proxy workaround from earlier changes is historical evidence, not part of the new fix.

### 6. Turn scope stubs into behavioral tests

Replace all 17 skips with executable cases using the existing non-TTY harness and fake vault capabilities. Cover top-menu keep/global/remove/personal/cancel, editor add/remove/cancel, comma-separated targets, me-resolution/deduplication, append/replace, and unresolved-target reporting. Assert persisted target sets and mutation counts, not only menu labels or intermediate structs.

Use the actual capability interfaces and current authoritative scope behavior. Where an old test name/comment assumes file/git vaults cannot support identity scopes, update the expectation to capability-based behavior; do not change working production scope logic to satisfy stale comments. Preserve relevant backend coverage for git/path and hosted capabilities. This is test work and does not justify a new scope capability delta.

## Risks / Trade-offs

- Valid large assets will newly fail -> document exact fixed limits and test equality and one-byte-over boundaries using synthetic inputs.
- Shared ZIP helpers have many callers -> migrate callers explicitly and test metadata/cache/direct-reader paths, not only extraction.
- Startup may temporarily disable extensions without a policy cache -> successful lookup restores allowed execution; document the deliberate failure-safe change.
- Cross-process locking can introduce deadlocks -> acquire one config lock, keep prompts/network outside callbacks, and test reporting, consent, and config mutations together.
- Directory rollback can fail due to filesystem state -> retain the backup and report recovery paths; run native Windows/macOS/Linux focused tests where supported and state any unexecuted platforms.
- Scope stub descriptions can be stale -> anchor tests to documented current contracts and capability interfaces; surface any actual contract conflict before implementation changes behavior.

## Migration Plan

1. Add regression tests demonstrating each current failure and fill scope cases incrementally.
2. Land archive budgets before staged installation; harden policy and config independently against their tests.
3. Preserve existing config/vault formats and cached policy format. Old cached packages are checked when consumed; no bulk cache purge or destructive vault rewrite is required.
4. Update reference docs and reconcile each TODO entry only after its verification succeeds.
5. Validate OpenSpec strictly, run relevant frontend checks and platform checks, and run `make prepush` with the normal inherited proxy environment rather than the earlier invalid-hostname exemption.

Rollback restores the previous binaries/code and reintroduces the audited weaknesses; it requires no data migration. Do not delete retained installation backups as part of rollback, and never weaken or reset existing telemetry decisions.
