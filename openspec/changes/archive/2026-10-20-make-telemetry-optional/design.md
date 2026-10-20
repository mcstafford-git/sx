# Design

## Context

See `proposal.md` for motivation and `specs/telemetry-consent/spec.md` for the contract. The existing shared `BootstrapOptions` map defaults missing values to true. `analytics_hook` controls installation, while `report-usage` and MCP reporting independently enqueue and flush events. Desktop extension installs call `RecordUsageEvents` directly and separately append audit events. The local queue lives under the configured cache directory. `docs/stats.md` contains descriptions of install recording and git flushing that must be reconciled with the actual producers and vault implementations during the documentation update.

This is a cross-cutting change with a persisted consent migration, so a design is required.

## Goals / Non-Goals

**Goals:** Keep one consent authority, fail closed, and reuse existing reporting and configuration mechanisms.

**Non-Goals:** Do not place user consent checks in generic vault history primitives: explicit history copies and statistics reads must continue to work. Do not add per-client consent, a background consent service, or a second event storage system.

## Decisions

### Store a new consent record outside legacy bootstrap preferences

Add a top-level telemetry record to `MultiProfileConfig` containing a consent version, an explicit enabled decision, and the latest opt-in time. Missing or unknown versions mean undecided and effectively disabled; the current version with false means declined. Read failures mean disabled. Keep unrelated bootstrap defaults unchanged.

The old `analytics_hook` boolean cannot prove consent to MCP and desktop reporting, so it must never be copied into the new record. Introducing the dedicated record avoids changing every bootstrap option's compatibility behavior. Preserve the record through configuration saves, migrations, and profile changes, including the custom JSON marshal/unmarshal representation in `profile.go`.

### Offer consent only through deliberate settings interactions

Provide `sx telemetry status`, `sx telemetry enable`, and `sx telemetry disable`, registered through the existing Cobra command structure. Status distinguishes undecided from declined. The explicit enable command is an affirmative action; its help and success output explain covered usage and vault destinations. No general `--yes` flag grants consent.

In interactive `sx init`, replace the hook-only analytics question with a scope-accurate telemetry question defaulting to no. Ask undecided users, including legacy users, for a fresh answer; remember declines. Existing users can use init or the new controls without being interrupted in ordinary commands. Desktop settings expose the same state and an initially off consent control with equivalent explanatory text. Read current consent when opening settings; persist only through explicit user interaction.

This avoids prompts in hooks, MCP servers, or unattended init, and provides an accessible opt-in path for desktop-only users.

### Gate producers and the shared queue

Add a small shared telemetry policy/service layer used by all three producers. Check consent before parsing hook input or constructing events, and again before enqueue or transmission. Gate `EnqueueEvent` and `FlushQueue` as backstops; route desktop automatic usage through the shared policy before its direct vault write. Keep its audit append independent.

Resolve current consent from disk for reporting attempts rather than retaining a startup snapshot in MCP or desktop processes. Late asynchronous work must recheck before committing usage. Synchronize consent changes, queue writes/flushes, and desktop automatic usage writes using a shared process-safe lock in the configured state directory; reuse the repository's cross-platform file-lock dependency/pattern. Re-read state after acquiring it. The config package owns this lock: ordinary configuration saves acquire it and preserve the latest persisted consent rather than the caller's snapshot. Explicit consent updates use the private config writer under that lock. Configuration replacement uses the existing atomic-write helper so readers cannot observe a partially written record. Bound network work with existing timeouts so settings changes cannot hang indefinitely. A write already started under enabled consent may finish before disable completes; after disable returns, no pending old write may start.

Hook installation treats analytics as enabled only with new-model consent. Do not reinstall disabled analytics hooks. Existing hooks may remain on disk and must safely no-op; removing every historic repository hook is unnecessary and unreliable. Enabling makes hooks eligible for the existing init/install hook reconciliation path, and CLI/desktop guidance tells users to run `sx install` for detected clients. Runtime enforcement, rather than hook removal, is authoritative.

### Purge pending data before allowing a fresh consent period

Under the same lock, discard the existing usage queue on first use without new-model consent and whenever telemetry is disabled. Explicit disabling saves the off decision before cleanup. Enabling from undecided or declined clears residual events before saving enabled consent. If cleanup fails, do not persist enabled consent; return an actionable error from settings operations and keep hook reporting best-effort and disabled.

Use the configured cache API and `filepath` rather than hardcoded Unix paths, respecting `SX_CACHE_DIR` and its existing alias. Remove only the usage queue, never unrelated cache files. Validate the resolved cleanup target and do not follow a queue-directory symlink into another directory. Missing queues are successful cleanup. Legacy processes cannot be forced to adopt this policy: users must restart or upgrade old binaries; stale events left by them are discarded before fresh enable.

Persist the opt-in time in the consent record and compare existing event timestamps before accepting delayed callbacks, so work created before a fresh opt-in cannot be sent afterward. Producers use nanosecond timestamps to avoid dropping events created in the same second as consent. A single lifecycle lock plus this timestamp check avoids event epochs or queue schema changes: existing JSON event and vault formats remain compatible. A redundant enable on already enabled consent is a no-op and must not discard legitimate pending events.

### Preserve historical operations

Leave vault statistics reads, audit events, and `vault copy` history migration outside the automatic telemetry policy. Do not delete `.sx/usage` or server-side history. Downloads and normal vault requests still occur as required for requested functionality; disabling telemetry gates usage reporting, not all network traffic.

## Risks / Trade-offs

- Existing installations stop contributing usage until renewed consent -> announce the compatibility change and explain how to enable it.
- A forgotten reporting path bypasses consent -> test hook, MCP, shared queue, and desktop direct-write paths independently with spies that assert zero writes/network calls while off.
- Consent changes race asynchronous reporters -> use the shared lock and test two independent reporters against disable; never cache permission for process lifetime.
- Cleanup can fail on Windows or restricted filesystems -> remain off, report settings errors, and retry safely without sending old events.
- Older sx versions ignore new consent -> document that mixed-version execution cannot provide the new guarantee; rollback requires care because legacy binaries default on.

## Migration Plan

1. Introduce the new record with missing-state semantics disabled; preserve other config fields and bootstrap defaults.
2. Wire all producers and shared queue entry points to policy before offering enable controls.
3. On migration discard old queued events. Present the new default-no question on the next interactive init or allow an explicit settings decision; never infer consent from legacy true.
4. Add CLI and desktop controls, hook installation filtering, tests, and docs. Verify Linux/macOS/Windows path and locking behavior with isolated config/cache roots and platform CI.
5. Run strict OpenSpec validation, relevant frontend checks, and `make prepush` before declaring implementation complete.

Rollback restores the earlier code but would restore its default-on behavior. Before running a legacy version, disable/remove legacy analytics hooks as appropriate and disclose that legacy MCP/desktop reporting does not honor the new record. No vault usage migration or rollback is necessary; locally purged events cannot be recovered.
