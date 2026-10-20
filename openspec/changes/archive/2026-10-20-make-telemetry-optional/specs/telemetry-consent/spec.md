# Spec Delta

## Purpose

Give users explicit, revocable control over usage telemetry across sx CLI hooks, MCP skill usage, and desktop extension adoption.

## ADDED Requirements

### Requirement: Telemetry requires fresh explicit consent
Usage telemetry SHALL remain disabled unless the user explicitly opts in under the new consent model. Legacy analytics-hook preferences SHALL NOT constitute consent.

#### Scenario: A new user has no preference
- **WHEN** sx runs without a saved telemetry decision
- **THEN** usage telemetry remains disabled

#### Scenario: An existing user previously enabled analytics hooks
- **WHEN** sx reads a legacy analytics-hook preference without new telemetry consent
- **THEN** usage telemetry remains disabled
- **AND** an interactive consent opportunity asks for a fresh decision with no selected by default

#### Scenario: Configuration cannot establish consent
- **WHEN** consent configuration is missing, unreadable, or invalid
- **THEN** no usage telemetry is collected or transmitted
- **AND** ordinary operations remain usable where their own configuration permits

### Requirement: Users can inspect and change their telemetry decision
The CLI and desktop SHALL expose the same user-level telemetry decision across profiles. Opt-in SHALL require a deliberate affirmative action that explains the usage being recorded and its vault destination. Declining SHALL be remembered.

#### Scenario: A user opts in
- **WHEN** the user explicitly enables telemetry through the CLI or desktop
- **THEN** the decision is persisted and visible to both interfaces across profiles
- **AND** only usage occurring after that decision is eligible for recording

#### Scenario: A user declines or runs non-interactively
- **WHEN** the user declines consent or runs unattended without previously saved new-model consent
- **THEN** telemetry stays disabled
- **AND** unattended commands do not prompt or treat general confirmation flags as consent
- **AND** a saved decline is not repeatedly prompted

#### Scenario: A user inspects consent
- **WHEN** the user requests telemetry status in the CLI or opens desktop settings
- **THEN** the effective enabled or disabled state is shown
- **AND** an undecided state is distinguishable from an explicit decline

### Requirement: Consent governs every usage reporting path
All automatic usage telemetry SHALL honor the current user decision at collection and transmission time, including installed client hooks, MCP skill reporting, and desktop extension install reporting.

#### Scenario: Old hooks remain installed without consent
- **WHEN** a previously installed analytics hook invokes sx without current consent
- **THEN** it records no usage event and performs no telemetry transmission
- **AND** it completes without prompting for consent

#### Scenario: MCP or desktop reporting occurs while disabled
- **WHEN** a skill is used through MCP or an extension is installed in the desktop app while telemetry is disabled
- **THEN** the requested operation continues normally
- **AND** no usage telemetry is queued or sent

#### Scenario: A user enables telemetry
- **WHEN** new-model consent is enabled and a supported usage event occurs
- **THEN** that event may be recorded through the existing vault usage mechanism

#### Scenario: Consent is revoked during a long-running session
- **WHEN** telemetry is disabled while an MCP server or desktop app remains running
- **THEN** subsequent collection and transmission attempts honor the disabled state without a restart
- **AND** an operation already transmitted before revocation is not retracted

### Requirement: Disabled telemetry discards pending local events
Migration to the new consent model and disabling telemetry SHALL delete pending local usage events. A fresh opt-in SHALL discard any residual pending events before reporting is enabled. Cleanup failure SHALL NOT enable reporting or allow stale events to be transmitted.

#### Scenario: A legacy queue exists during migration
- **WHEN** sx first applies the new consent model to a user with pending usage events
- **THEN** those pending events are deleted without transmission
- **AND** old analytics-hook consent does not preserve the queue

#### Scenario: A user disables telemetry and later re-enables it
- **WHEN** the user disables telemetry with events pending and later opts in again
- **THEN** events from before the new opt-in are never transmitted
- **AND** telemetry collection starts fresh

#### Scenario: Queue cleanup fails
- **WHEN** pending events cannot be deleted during migration, disable, or fresh enable
- **THEN** telemetry remains effectively disabled and cleanup can be retried
- **AND** an explicit settings operation reports the cleanup failure

### Requirement: Usage consent preserves historical data and audit behavior
Changing telemetry consent SHALL preserve previously recorded vault usage and the ability to read historical statistics. Mutation audit logs and explicit vault-history copy operations SHALL retain their existing behavior.

#### Scenario: Historical statistics are requested while disabled
- **WHEN** the user reads existing usage statistics with telemetry disabled
- **THEN** historical statistics remain available without enabling collection

#### Scenario: A mutation or history copy runs while disabled
- **WHEN** an audited vault mutation or an explicit copy of existing vault history occurs
- **THEN** the existing audit or copy behavior continues
- **AND** disabling telemetry does not delete persisted vault history
