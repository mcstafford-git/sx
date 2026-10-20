# Repository Audit TODO

What this repo does: `sx` distributes AI assets across teams, vaults, and coding clients through a Go CLI and desktop app. The audit assumed one-user CLI load with occasional concurrent CLI/app runs.

## Must fix

1. **Extension policy fails open when its cache is missing** (`app/plugins.go:290`)
   - **What this is:** Organization policy can disable or allowlist desktop extensions.
   - **Problem:** A vault read failure with no valid cache returns `open`. Cached third-party code with prior consent can run despite the organization's policy.
   - **Fix:** Return `disabled` when policy retrieval fails and no valid cached policy exists.
   - **If we skip it:** A network or cache failure can bypass an organization's extension restrictions.

2. **Remote ZIP files have no size limits** (`internal/utils/zip.go:45`)
   - **What this is:** Vault assets are downloaded into memory and decompressed during installation.
   - **Problem:** Both HTTP reads and ZIP extraction are unlimited. A small compression bomb can consume all memory or disk.
   - **Fix:** Cap compressed downloads, entry count, per-entry size, and total extracted bytes using `io.LimitReader`.
   - **If we skip it:** A bad marketplace or vault asset can crash `sx` or fill the machine's disk.

3. **A failed update deletes the working installation** (`internal/handlers/dirasset/operations.go:41`)
   - **What this is:** Skills, hooks, plugins, and MCP packages are installed as directories.
   - **Problem:** The old directory is removed before the new ZIP is extracted. A corrupt later ZIP entry leaves the user with a partial or missing asset.
   - **Fix:** Extract into a sibling temporary directory, then replace the old directory only after extraction succeeds.
   - **If we skip it:** One broken update removes a previously working skill or plugin.

4. **CLI and app config writes can corrupt or lose settings** (`internal/config/profile.go:300`)
   - **What this is:** The desktop app and CLI share `config.json`, including profiles and credentials.
   - **Problem:** Writes use `os.WriteFile` without an atomic rename or cross-process lock. A crash can truncate the file; two writers can silently overwrite each other.
   - **Fix:** Use the existing `utils.WriteFileAtomic`, plus a file lock around read-modify-write operations.
   - **If we skip it:** Concurrent app and CLI changes can lose profiles or make configuration unreadable.

## Should fix

5. **The normal test suite fails behind this HTTP proxy** (`internal/git/client.go:646`)
   - **What this is:** Git errors are classified into friendly network, authentication, and missing-repository messages.
   - **Problem:** `CONNECT tunnel failed, response 502` is not recognized as a network error, causing `make test` to fail and users to receive a generic message.
   - **Fix:** Classify proxy tunnel failures as network errors and unit-test the classifier with fixed strings.
   - **If we skip it:** Proxy-based development and CI environments remain unreliable.

6. **Critical scope-editing flows are only skipped test stubs** (`internal/commands/scope_editor_flows_test.go:5`)
   - **What this is:** The interactive editor changes organization, repository, team, user, and bot access.
   - **Problem:** Sixteen documented flows call `t.Skip`, including cancel, replace, global access, and removing all scopes.
   - **Fix:** Implement the existing table of test cases using the documented non-TTY harness.
   - **If we skip it:** A regression can broaden access or erase existing scopes without CI noticing.

## Verdict

Strong test coverage and careful vault locking, but fix policy failure, ZIP limits, and destructive installation first.

## Not checked

Generated GraphQL code, every client adapter, Docker E2E tests, and live external services. `make test` ran with race detection; all reported packages passed except the proxy-sensitive Git test above. No source files were changed by the audit.
