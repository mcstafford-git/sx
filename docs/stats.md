# Usage analytics

Usage telemetry is **off by default**, including for existing users. A previous
“Yes” to installing analytics hooks is not consent to this broader reporting.

```bash
sx telemetry status     # undecided, disabled, or enabled
sx telemetry enable     # explicitly opt in
sx telemetry disable    # stop reporting and delete pending local events
sx install              # configure analytics hooks after opting in
```

Opting in records skill usage detected by client hooks, skill usage reported
through MCP, and desktop extension installs in your configured vaults. Hosted
vaults receive events over their usage endpoint; git and path vaults keep usage
under `.sx/usage/`. Events include asset name, version, type and timestamp;
vaults attribute them to your identity. Consent is shared across all profiles,
the CLI, and the desktop app on this installation.

Interactive `sx init` asks undecided users a new question, defaulting to no.
A decline is remembered. Unattended init and general confirmation flags never
opt in. You can also use **Enable usage telemetry** in desktop Settings;
opening settings does not grant consent. Restarting a current-version MCP
server or desktop app is unnecessary when changing the setting.

Without consent, old installed hooks safely no-op and usage is neither queued
nor transmitted. Migration, disabling, and a fresh opt-in discard pending local
usage queue files, so enabling starts fresh. Cleanup errors prevent enabling;
retry the settings operation after resolving the reported filesystem error.
Historical vault events, `sx stats`, explicit vault-history copies, and mutation
audit logs remain available. Ordinary downloads and vault requests still work.
Older binaries do not honor this setting: upgrade/restart them before relying
on it. Files already written into vault history are not removed.

With consent, desktop-app extension installs record the same event with
`asset_type = "app-plugin"` (appended best-effort by the app), so
extension adoption shows up in this dashboard alongside skills. That is
this vault's view; **global** marketplace install counts are a separate
mechanism — release-asset download counts aggregated into the
marketplace repo's `stats.json` — described in docs/app-plugins-spec.md.

## The `sx stats` command

```bash
sx stats                               # last 7 days
sx stats --since 30d                   # widen the window
sx stats --since all                   # lifetime totals
sx stats --assets                      # per-asset view only
sx stats --teams                       # per-team view only
sx stats --since 30d --json            # machine-readable
```

`--since` accepts `Nd` (days) or `all`. The dashboard renders four
sections: totals, top assets, per-team adoption, and top actors. The
`--assets` and `--teams` flags narrow to one section each.

### What the dashboard shows

- **Total events** — number of recorded usage events in the window.
- **Top assets** — each asset's total use count plus the number of
  unique users associated with it.
- **Team adoption** — for each team, the percentage of members who
  recorded any usage in the window (e.g. `platform 3/5 = 60%`).
- **Top actors** — users with the most usage events in the window.

## JSON output

`sx stats --json` produces:

```json
{
  "since": "2026-03-18T00:00:00Z",
  "total_events": 127,
  "assets": [
    { "AssetName": "code-reviewer", "TotalUses": 42, "UniqueActors": 17 }
  ],
  "teams": [
    { "name": "platform", "member_count": 5, "active_members": 3, "adoption_pct": 60.0 }
  ],
  "top_actors": [
    { "Actor": "alice@acme.com", "TotalUses": 9 }
  ]
}
```

When `--since all` is used the `since` field is omitted rather than
serialised as a zero time.

## Storage format

Each event is one JSON object on a line in `.sx/usage/YYYY-MM.jsonl`:

```json
{
  "ts": "2026-04-17T10:04:12.445Z",
  "actor": "alice@acme.com",
  "asset_name": "code-reviewer",
  "asset_version": "1.2.3",
  "asset_type": "skill"
}
```

## Fault tolerance

- A malformed JSONL line is logged at `warn` and skipped. One bad line
  never drops a whole batch of good ones from a flush.
- An event with an unparseable timestamp is stamped with
  `time.Unix(0, 0)` so it falls outside any `--since Nd` window rather
  than skewing "recent" totals; `--since all` still counts it.

## Delivery and local queue

With consent, hooks and MCP enqueue events under the configured cache directory's
`usage-queue/`, then attempt to flush a batch to the active vault. Hook flushing
is synchronous with a bounded timeout; MCP reporting runs in the background.
Failed deliveries leave pending queue files for a later flush while enabled.
Desktop extension usage writes directly to the current vault, best-effort.

Path vaults append events to monthly JSONL files. Git vaults append locally and
commit/push with throttling to avoid a commit for every event; management
mutations may also sweep pending vault changes into a commit. Local statistics
can see recorded vault events before a successful push. The telemetry consent
queue cleanup removes the usage spool, not files already persisted in a vault.

## Sleuth vault

The [skills.new](https://skills.new) hosted vault ingests the same
events through its usage endpoint. `sx stats` falls back to the
server's aggregation API; the web UI provides per-asset charts and
per-user drill-downs that the CLI does not render.
