# Proposal

## Why

Usage telemetry currently defaults on, and the analytics-hook preference only controls hook installation rather than all collection paths. Users need explicit, revocable consent covering CLI hooks, MCP usage, and desktop extension adoption.

## What Changes

- **BREAKING**: Default telemetry off for new and existing users; require fresh explicit opt-in instead of trusting legacy analytics-hook preferences.
- Provide one user-level preference shared across profiles, CLI, and desktop, with a default-no consent prompt and explicit enable/disable/status controls.
- Gate usage collection and transmission, including previously installed hooks, MCP reporting, and desktop extension install usage.
- Delete pending local telemetry on migration and disabling, and before a fresh opt-in, so enabling starts fresh.
- Keep historical usage readable and preserve mutation audit logs.
- Update telemetry documentation and test every reporting path and the consent migration.

## Capabilities

### New Capabilities

- `telemetry-consent`: Explicit user consent, revocation, migration, and pending-event handling for usage telemetry.

### Modified Capabilities

None.

## Impact

- Shared configuration in `internal/config/profile.go`; bootstrap options and init/hook installation paths.
- `internal/stats/queue.go`, `internal/commands/report_usage.go`, `internal/mcp/server.go`, and desktop extension usage in `app/extensions_install.go`.
- CLI telemetry controls, desktop settings, and associated tests and documentation (`docs/stats.md`).
- No new service or dependency is required. Historical vault usage formats remain compatible.

## Non-Goals

- Removing historical vault data, changing audit logging, or disabling ordinary vault/API operations.
- Changing marketplace download counts supplied by hosting infrastructure.
- Adding organization-level policy, per-vault consent, telemetry dashboards, or automatic reporting prompts inside hooks and MCP calls.
