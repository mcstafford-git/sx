# Spec Delta

## Purpose

Present actionable remote Git errors across direct and proxied connections without relying on the user's live network for classification tests.

## ADDED Requirements

### Requirement: Proxy connection failures are network errors
sx SHALL classify Git proxy tunnel connection failures, including a CONNECT tunnel failure with response 502, as network errors. The diagnostic SHALL identify the remote and preserve useful underlying failure details.

#### Scenario: HTTPS proxy rejects a tunnel
- **WHEN** Git reports CONNECT tunnel failed with response 502
- **THEN** sx reports a network error reaching the remote
- **AND** the proxy failure details remain visible

#### Scenario: Proxy error casing varies
- **WHEN** the same supported proxy failure appears with different capitalization
- **THEN** classification remains a network error

### Requirement: Remote error categories remain distinct
Authentication failures, missing repositories, known network failures, and unknown Git failures SHALL retain distinct actionable diagnostics. Proxy connection classification SHALL NOT broadly reinterpret unrelated errors merely because they contain a numeric status code.

#### Scenario: Authentication fails
- **WHEN** Git reports a recognized credential or public-key authentication failure
- **THEN** sx reports authentication guidance for the remote

#### Scenario: A repository is absent
- **WHEN** Git reports a recognized repository-not-found failure
- **THEN** sx reports the repository URL and existing URL/access guidance

#### Scenario: Failure is unknown
- **WHEN** Git fails with an unrecognized message
- **THEN** sx retains a generic Git-operation error with the original failure details
