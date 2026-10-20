# Spec Delta

## Purpose

Defines a reviewable path from a proposed repository behavior change through implementation, verification, documentation, and archival.

## ADDED Requirements

### Requirement: Behavior changes are planned in OpenSpec
Repository changes that add, modify, or remove durable behavior SHALL have an OpenSpec change whose proposal identifies the affected capabilities before implementation begins.

#### Scenario: New behavior is proposed
- **WHEN** a contributor plans a change that alters behavior relied on by users, clients, vaults, or automation
- **THEN** the contributor creates or selects an OpenSpec change before implementing it
- **AND** the proposal names every new or modified capability

#### Scenario: Work has no specification impact
- **WHEN** a change is limited to refactoring, tooling, or documentation and does not alter durable behavior
- **THEN** its OpenSpec change MAY explicitly skip specification deltas
- **AND** the proposal explains why no capability requirements change

### Requirement: Planning artifacts are traceable
Each planned change SHALL connect its motivation, requirements, design decisions, and implementation tasks so reviewers can trace every task to the behavior it supports.

#### Scenario: Change is ready for implementation
- **WHEN** implementation is about to begin
- **THEN** the change contains all artifacts required by its configured schema
- **AND** the tasks identify the concrete repository work and verification needed to satisfy the requirements

### Requirement: Implementation follows tracked tasks
Implementation SHALL be performed against one selected OpenSpec change, and completed work SHALL be reflected in that change's task state.

#### Scenario: A task is implemented
- **WHEN** implementation and task-specific verification succeed
- **THEN** the corresponding task is marked complete
- **AND** unrelated tasks remain unchanged

#### Scenario: Scope changes during implementation
- **WHEN** implementation reveals a material requirement, design, or scope change
- **THEN** the planning artifacts are updated and reconciled before the affected work is treated as complete

### Requirement: Completion requires the repository verification gate
A code change SHALL NOT be reported complete unless `make prepush` succeeds after the final relevant edit.

#### Scenario: Verification succeeds
- **WHEN** all planned implementation tasks are complete
- **AND** `make prepush` succeeds after the final relevant edit
- **THEN** the change may be reported ready for review or archival
- **AND** the verification result is summarized for reviewers

#### Scenario: Verification fails or is interrupted
- **WHEN** `make prepush` fails, cannot run, or is interrupted
- **THEN** the change is not reported complete
- **AND** the failing or unverified state is reported clearly

### Requirement: Behavior and reference documentation remain consistent
A change SHALL update affected OpenSpec requirements and repository reference documentation when their documented behavior changes.

#### Scenario: A documented contract changes
- **WHEN** implementation changes a contract described by an OpenSpec capability or a document under `docs/`
- **THEN** the affected specification or reference document is updated in the same change
- **AND** unchanged documentation is not copied into OpenSpec merely for duplication

### Requirement: Completed changes are verified before archival
An OpenSpec change SHALL be archived only after its required tasks are complete and its implementation has been checked against the planned artifacts.

#### Scenario: Change is ready to archive
- **WHEN** all required tasks are complete
- **AND** verification confirms the implementation matches the proposal, requirements, and design
- **THEN** the change may be archived into the durable capability inventory

#### Scenario: Implementation differs from the plan
- **WHEN** verification finds a material mismatch between implementation and planning artifacts
- **THEN** the change remains active
- **AND** either the implementation or the planning artifacts are corrected before archival
