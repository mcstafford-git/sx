# Tasks

## 1. Establish shared consent and queue lifecycle

- [x] 1.1 Add the versioned, user-level telemetry decision to shared config and its custom serialization; verify missing/invalid/current/unknown versions, legacy true/false preferences, explicit decline, round trips, and profile changes in config tests without changing unrelated bootstrap defaults.
- [x] 1.2 Add shared telemetry policy and process-safe lifecycle locking; implement off-by-default reads, migration cleanup, disable-before-purge, and purge-before-fresh-enable. Verify isolated config/cache tests cover absent queues, cleanup errors, repeated enable, symlink safety, and `SX_CACHE_DIR` overrides.
- [x] 1.3 Gate shared queue collection and flushing on current consent; verify off/undecided/config-error cases perform zero event writes or vault sends, legacy queued data is discarded, enabled events still flush, and failed cleanup never enables transmission. Document consent and local queue retention in `docs/stats.md` alongside this change.

## 2. Enforce consent in hooks and MCP

- [x] 2.1 Gate `report-usage` before input parsing and event construction, and gate MCP skill reporting before collection and before asynchronous transmission; verify stale installed hooks no-op, normal tool calls still succeed, enabled usage is recorded, and a running MCP server observes revocation without restart.
- [x] 2.2 Update analytics hook selection in init/install and shared bootstrap handling to require new consent while preserving session/MCP option behavior; verify old analytics true and config-load failure never install analytics hooks, fresh consent allows installation, and already installed hooks remain harmless while disabled.
- [x] 2.3 Document actual hook/MCP producers and current vault delivery behavior in `docs/stats.md`, correcting the existing unconditional install and lazy-flush descriptions where needed; verify the text matches producer and vault tests for path, git, and hosted vaults.

## 3. Add explicit CLI consent controls

- [x] 3.1 Add and register `sx telemetry status|enable|disable`, using the shared service and explaining covered usage and vault destinations in command help/output; verify status distinguishes undecided/declined/enabled, enable is affirmative, disable persists before cleanup, and settings failures are actionable.
- [x] 3.2 Replace interactive init's hook-only question with default-no telemetry consent for undecided users, remembering declines; verify legacy users are asked afresh, saved new decisions are respected, canceled prompts never enable, and non-interactive/general-confirmation flows never create consent.
- [x] 3.3 Document renewed consent, commands, automation behavior, and `sx install` hook reconciliation in `docs/stats.md` and the appropriate README entry; build the CLI and verify documented status/enable/disable commands using isolated `SX_CONFIG_DIR` and `SX_CACHE_DIR` without touching real client configs.

## 4. Integrate desktop consent and extension reporting

- [x] 4.1 Expose current telemetry state and explicit updates through `app/settings.go` and `SettingsModal.tsx`, updating generated bindings as required; verify the initial control is off, explanatory text covers all producers, CLI changes are reflected on settings reload, declines persist, and errors do not show a successful enable.
- [x] 4.2 Gate `recordPluginUsage` through shared consent and synchronization without gating extension audit events; verify extension installs succeed while disabled with zero usage writes, audit append remains intact, enabled usage is recorded, and a running app observes a CLI disable.
- [x] 4.3 Document desktop opt-in/revocation and its shared scope in `docs/stats.md`; run the frontend build/type checks from `app/frontend/package.json` and relevant app settings/extension tests, verifying bindings and visible state agree.

## 5. Verify integrated consent behavior

- [x] 5.1 Add integration coverage with isolated state for migration, enabled reporting, disable, fresh re-enable, and two independent reporters racing revocation; verify no stale pending event is sent after disable completes and no pre-opt-in event survives into the next enabled period.
- [x] 5.2 Verify historical stats, mutation audit events, and explicit vault-history copy still work while consent is off; run relevant stats/app/vault-copy tests and check Linux/macOS/Windows path/locking coverage in supported platform CI without claiming unexecuted platforms passed.
- [x] 5.3 Run `openspec validate make-telemetry-optional --strict`, review implementation against every requirement/scenario, and run `make prepush` after the final relevant edit; record results and report any failure or unexecuted check before claiming implementation complete.

- [x] 5.4 Fix the verification finding where stale profile saves restore old consent: share the telemetry lock with all config saves, preserve current consent during ordinary saves, and verify stale opt-in/opt-out saves, lock contention, the original profile-confirmation CLI reproduction, and `make prepush`.

## Workflow follow-up

- Run OpenSpec verification after implementation is complete.
- Sync the telemetry-consent delta and archive only after verification and repository completion checks succeed.

## Implementation verification

- Strict OpenSpec validation passed; requirements and scenarios were reviewed against config, queue, CLI, MCP, and desktop changes.
- `make prepush` passed after the final code edit: formatting, lint (0 issues), GraphQL freshness/inline-operation checks, race-enabled tests (48 packages), and build.
- `npm run build` passed (TypeScript and Vite); existing frontend dependencies were installed with `npm ci` first.
- CLI status/enable/disable smoke checks passed using temporary config/cache roots and the built binary, with the self-updater disabled.
- Consent lifecycle tests cover legacy consent rejection, purge failures, symlinks, delayed old events, cross-process revocation, and runtime MCP/desktop setting changes. Historical usage reads and vault-history copy/audit tests passed while reporting is off.
- Telemetry tests cross-compiled for Windows amd64 and macOS arm64. Native runtime checks ran on Linux only. The existing test workflow runs on Linux; release jobs build macOS/Windows but do not provide native telemetry test execution. Desktop UI behavior was reviewed and type-checked; no graphical session was exercised.
- The git invalid-hostname test used invocation-local `NO_PROXY`/`no_proxy` entries for `invalid.localhost.invalid` to avoid the environment proxy returning HTTP 502.

## Locking finding resolved

- Moved shared telemetry lock ownership into config; ordinary saves preserve the latest persisted consent under that lock. Explicit consent updates use the private writer under the same lock, avoiding recursive acquisition. Config replacement uses the existing atomic-write helper.
- Regression tests cover stale enabled/disabled snapshots, ordinary saves being unable to grant consent, and contention with the reporting/consent lock. Targeted race-enabled config, stats, CLI, MCP, and app tests passed.
- Re-ran the exact isolated CLI reproduction with the rebuilt binary: pause `sx profile use` at confirmation, disable telemetry in a separate process, then resume. Telemetry remains disabled and the requested profile change is saved.
- `make prepush` passed after the final code edit (0 lint issues, GraphQL checks, all 48 race-enabled test packages, build); strict OpenSpec validation passed. Native macOS/Windows and graphical UI checks were not repeated for this config-only fix.
