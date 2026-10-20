# Spec Delta

## Purpose

Preserve user profiles, credentials, preferences, and telemetry consent when CLI and desktop processes update their shared configuration.

## ADDED Requirements

### Requirement: Configuration mutations preserve unrelated changes
Successful configuration mutations SHALL apply the requested change to the latest persisted state under cross-process exclusion. Concurrent mutations of unrelated settings SHALL both survive. A stale snapshot SHALL NOT silently replace newer unrelated settings.

#### Scenario: Different profiles are edited concurrently
- **WHEN** two processes change different profiles from overlapping sessions
- **THEN** both successful changes are present in the persisted configuration

#### Scenario: A prompt remains open while another process saves
- **WHEN** a command reads configuration, waits for user input, and another process changes unrelated settings
- **THEN** confirming the original action preserves the intervening unrelated changes

#### Scenario: Same-setting mutations compete
- **WHEN** two valid actions intentionally change the same setting
- **THEN** their commits are serialized against current state
- **AND** the later commit determines that setting without resetting unrelated state

#### Scenario: A stale action is no longer valid
- **WHEN** another process deletes a profile or otherwise invalidates a pending action
- **THEN** the pending action fails clearly rather than recreating or overwriting state from its stale snapshot

### Requirement: Config replacement does not expose partial state
A configuration save SHALL publish a complete replacement with secure permissions. If lock acquisition, validation, serialization, or replacement fails before commit, sx SHALL report failure and retain the prior complete configuration.

#### Scenario: Replacement cannot complete
- **WHEN** writing the replacement fails before publication
- **THEN** the previous configuration remains readable and unchanged

#### Scenario: A reader overlaps a writer
- **WHEN** another sx process reads while configuration is being saved
- **THEN** it reads a complete old or new configuration rather than partially serialized data

### Requirement: Shared config updates preserve consent and profile invariants
Ordinary configuration mutations SHALL preserve current telemetry consent, persistent profile selection, and unrelated credentials unless their requested action explicitly changes the relevant setting. Existing runtime profile overrides SHALL NOT accidentally rewrite persistent defaults or compatibility fields.

#### Scenario: Telemetry is disabled during a profile edit
- **WHEN** telemetry is disabled while another process edits a profile
- **THEN** the completed profile edit leaves telemetry disabled

#### Scenario: Runtime profile selection is temporary
- **WHEN** a command selects a runtime profile while changing another supported configuration field
- **THEN** the saved default and compatibility mirror continue to follow the persistent profile rules

#### Scenario: Old configuration is migrated during a save
- **WHEN** a supported legacy configuration is mutated
- **THEN** its existing profiles, credentials, and explicit preferences survive migration except for the fields intentionally changed
