# Tasks

## 1. Enforce extension policy throughout startup and refresh

- [x] 1.1 Add failing policy regressions in the app policy/cache tests for retrieval/current-vault errors with missing, unreadable, corrupt, unsupported-mode, and wrong-profile caches; verify each fails against the current open fallback before changing production behavior.
- [x] 1.2 Make live and cache-only Go policy lookups default disabled on failure, validate cached modes, and atomically cache successful results; run policy/cache tests verifying successful no-policy and unsupported-backend lookups stay open and valid cached restrictions survive outages.
- [x] 1.3 Make frontend initial state, malformed responses, and bridge-error fallbacks disabled; add focused startup tests with cached code and prior consent using a lightweight Node/TypeScript harness without a new dependency, and verify execution is blocked until policy permits it, successful refresh enables allowed code, and profile switches do not reuse another profile's policy.
- [x] 1.4 Update `docs/app-plugins-spec.md` for startup and offline fallback behavior, including existing built-in handling; run the frontend test harness and `npm run build` in `app/frontend`, and verify documentation matches Go/frontend policy tests.

## 2. Bound archive consumption before inspection or extraction

- [ ] 2.1 Add fixed utility budgets and bounded-read primitives for 10 MiB asset input, 50 MiB repository input, 10 MiB expanded files, 50 MiB expanded totals, and 1,000 entries; verify synthetic tests allow equality, reject one-byte/count overflow, and enforce actual bytes with absent or misleading size declarations.
- [ ] 2.2 Apply archive validation and actual-byte limits to shared metadata reads, extraction, hashing, entry copying/repackaging, and relevant direct ZIP readers in vault adapters; verify compression-bomb metadata, excessive entries, corrupt later entries, path traversal, and executable permissions without allocating genuinely unbounded payloads.
- [ ] 2.3 Bound ordinary and progress HTTP downloads plus hosted-vault asset reads before accumulation; use local HTTP servers to verify known-length, chunked/unknown-length, misleading-size, exact-limit, cancellation, progress, and over-limit error propagation without live services.
- [ ] 2.4 Apply asset budgets to disk-cache loading and validation before metadata inspection; verify oversized cached input cannot bypass the limits, is rejected/evicted safely, and valid partitioned caches and cache misses still work.
- [ ] 2.5 Bound GitHub directory discovery traversal, individual downloaded files, aggregate bytes, entry count, and assembled ZIP output with the repository policy; verify fixture-based GitHub responses hit each budget and selected asset packages still use 10 MiB, without adding a whole-repository ZIP endpoint or limiting Git clone size.
- [ ] 2.6 Document fixed limits and error behavior in `docs/metadata-spec.md`, `docs/clients.md`, and the relevant README usage link; run utils/metadata/assets/GitHub/vault/app archive-related tests and verify marketplace wrappers retain the exceeded-budget diagnostic.

## 3. Preserve directory installations across failed updates

- [ ] 3.1 Add a shared staged directory-install helper using same-filesystem siblings and a per-destination lock; verify a new install, successful replacement, late extraction error, budget rejection, and simultaneous replacements leave a complete final version with correct permissions.
- [ ] 3.2 Implement backup/promotion/rollback handling and staging cleanup; verify injected promotion failure restores old files, failed restoration preserves and names the old backup, and backup-cleanup failure never deletes the working new installation.
- [ ] 3.3 Wire `dirasset.Operations.Install` and packaged MCP directory-install handlers for supported clients through the helper, without replacing scratch extraction or changing config-only MCP behavior; run handler/client regressions verifying a bad update preserves the old asset, config-only installs still work, and no partial first install is published.
- [ ] 3.4 Document replacement/recovery guarantees and their crash/access limitations in `docs/clients.md`; execute focused filesystem tests on available native platforms and verify Windows failure handling via supported platform checks, recording actual execution versus cross-compilation.

## 4. Make shared configuration updates transactional

- [ ] 4.1 Add a config mutation API that acquires the existing telemetry/config lock before reading current state and commits with the private atomic writer; verify two-process edits to distinct fields/profiles both survive, malformed existing config is not reset, and invalid mutations leave the prior file intact.
- [ ] 4.2 Migrate `SaveToProfile`, config/profile migrations, and other config-package callers to the transactional path; verify existing legacy-format, nil-versus-empty client selection, active/default-profile, compatibility-mirror, identity, and credential-preservation tests still pass.
- [ ] 4.3 Migrate CLI profile/client/init/identity-prompt writers to field-specific mutations after prompts complete; verify a delayed prompt preserves intervening unrelated edits, deleted profiles are not resurrected, and same-field updates serialize without silently resetting other fields.
- [ ] 4.4 Migrate desktop settings/repository/bridge writers and retire unrestricted production snapshot-save calls; verify CLI/app concurrent profile, credential, client-preference, and repository-tracking changes survive, and search all production config writers to confirm no stale snapshot path remains.
- [ ] 4.5 Keep explicit consent updates and automatic reporting on the same lock without recursive acquisition; run config/stats/CLI/MCP/app race-enabled regressions for disable during profile edits, fresh opt-in, queue cleanup failure, and ordinary writes that cannot grant consent.
- [ ] 4.6 Document shared config transaction behavior and conflict/error semantics in `docs/environment-variables.md`; verify callbacks contain no interactive prompts or network work, atomic replacement retains secure permissions, and the prior profile-confirmation consent regression remains fixed.

## 5. Make Git error classification proxy-independent

- [ ] 5.1 Add table-driven classifier tests for CONNECT tunnel/502 failures, casing variants, existing authentication/not-found/network errors, empty output, and unrelated numeric statuses; verify the proxy case fails before adding targeted network classification.
- [ ] 5.2 Replace live invalid-hostname classification dependence with fixed-output tests and deterministic command-wiring coverage; run `go test -race ./internal/git` with the normal proxy environment, without adding an invalid-hostname NO_PROXY exemption, and verify original diagnostic details and authentication guidance are retained.
- [ ] 5.3 Update Git troubleshooting in `docs/github-actions.md` to distinguish proxy connection failures from SSH credentials or absent repositories; verify documented messages match the classifier tests.

## 6. Replace all skipped scope-editor flows with executable coverage

- [ ] 6.1 Implement top-menu keep/global/remove/personal/cancel cases using the documented non-TTY harness and capability-based fake vaults; verify mutations are absent for keep/cancel and match current documented targets for affirmative actions.
- [ ] 6.2 Implement editor repository/path, comma-separated identity lists, remove-all, and editor-cancel cases; verify working edits are discarded on cancel, removing scopes follows current global/replace semantics, and persistence uses the supported backend capability path.
- [ ] 6.3 Implement me-prefill, me-plus-other-user, me-plus-own-email deduplication, and me-to-org replace cases; verify recorded setter targets use the actor identity and do not retain unintended identity scopes.
- [ ] 6.4 Implement append/replace, backend-capability gating, and unresolved-target-reporting cases; verify append flags, exact sent target sets, and partial-resolution messages against `docs/scoping.md`, `docs/teams.md`, `docs/users.md`, and `docs/bots.md`, correcting stale stub names/comments without changing supported scope behavior.
- [ ] 6.5 Run all 17 former scope-editor stubs and relevant scope persistence/integration tests with race detection; verify none retain `t.Skip` or are replaced by tautological intermediate-value checks, and update harness comments to describe executed coverage.

## 7. Verify the integrated audit fixes

- [ ] 7.1 Reconcile all six `TODO.md` entries with implemented behavior and verification evidence, correcting the config description and 17-flow count; verify every entry maps to completed tasks and preserve any actual unverified limitation instead of marking it resolved.
- [ ] 7.2 Run `openspec validate address-todo-items --strict`, review all five capability specs against implementation/tests/docs, and run the frontend checks; verify no regression in historical stats, audit behavior, telemetry consent, or existing scope semantics.
- [ ] 7.3 Run `make prepush` after the final relevant edit with normal proxy settings; record formatting, lint, GraphQL checks, race-enabled tests, and build results, plus native/cross-compiled platform coverage, and report failures before claiming completion.

## Workflow follow-up

- Run OpenSpec verification after implementation is complete.
- Sync the accepted capability specs and archive after verification and repository completion checks succeed.
