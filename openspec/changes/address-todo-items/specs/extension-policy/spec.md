# Spec Delta

## Purpose

Preserve organizational desktop extension restrictions when policy retrieval or local policy-cache access fails.

## ADDED Requirements

### Requirement: Policy failures default to disabled without a valid cache
When extension policy cannot be retrieved, sx SHALL use a valid cached policy for the current profile or disable extensions if no such cache exists. Cached user consent SHALL NOT bypass the effective organization policy.

#### Scenario: Retrieval fails without cached policy
- **WHEN** the current vault cannot be resolved or its policy request fails
- **AND** no valid current-profile policy cache is available
- **THEN** the effective extension policy is disabled
- **AND** previously consented third-party extensions are not permitted to execute

#### Scenario: Cached restrictions survive an outage
- **WHEN** policy retrieval fails and a valid current-profile cached disabled or allowlist policy exists
- **THEN** the cached restrictions remain effective

#### Scenario: Cache is invalid
- **WHEN** policy retrieval fails and its cache is unreadable, malformed, or contains an unsupported mode
- **THEN** the effective policy is disabled

### Requirement: Successful open policies retain their meaning
A successful policy lookup with no configured restrictions, or a backend known not to support extension policy, SHALL remain open. Successful supported policy lookups SHALL refresh the current-profile cache.

#### Scenario: A vault explicitly allows extensions
- **WHEN** policy lookup succeeds with an open policy or no configured policy
- **THEN** the effective policy is open
- **AND** subsequent retrieval failures may use that valid cached result

#### Scenario: A backend has no policy support
- **WHEN** the vault is successfully resolved and does not implement extension policy
- **THEN** the effective policy remains open

#### Scenario: Profiles are switched
- **WHEN** the user switches to a profile whose policy cannot be retrieved
- **THEN** a cached policy from another profile is not used

### Requirement: Startup requires a known extension policy
Before a successful live policy lookup, extension startup SHALL use a valid current-profile cached policy or disabled policy. Missing policy information SHALL NOT temporarily allow previously consented third-party code to execute.

#### Scenario: Startup has cached code but no policy
- **WHEN** cached third-party extension code and user consent exist but no valid policy cache is available
- **THEN** the extension remains stopped until a successful lookup permits it

#### Scenario: A policy bridge call fails
- **WHEN** the frontend cannot obtain a policy from either live or cache-only policy APIs
- **THEN** the frontend uses disabled policy rather than open policy
