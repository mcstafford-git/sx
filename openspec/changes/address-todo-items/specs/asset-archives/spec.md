# Spec Delta

## Purpose

Bound memory and disk consumption when downloading, inspecting, and extracting asset packages and repository discovery archives.

## ADDED Requirements

### Requirement: Compressed content has fixed byte budgets
sx SHALL limit compressed asset packages to 10 MiB and compressed repository discovery ZIPs to 50 MiB, where one MiB is 1,048,576 bytes. A complete input exactly at its limit SHALL be permitted if otherwise valid; exceeding a limit SHALL fail with an error identifying the budget.

#### Scenario: An asset exceeds its download budget
- **WHEN** an asset response exceeds 10 MiB
- **THEN** sx stops accepting content and reports the compressed asset limit
- **AND** the over-limit result is not cached or installed

#### Scenario: Repository discovery exceeds its budget
- **WHEN** a repository discovery ZIP exceeds 50 MiB
- **THEN** discovery fails with the repository limit identified
- **AND** the larger repository budget does not increase any individual asset's 10 MiB limit

#### Scenario: Transport size declarations are absent or misleading
- **WHEN** a response has no trustworthy content length or uses the progress download path
- **THEN** enforcement is based on actual consumed bytes

### Requirement: Archive expansion has fixed budgets
Each processed archive SHALL contain at most 1,000 entries, at most 10 MiB of expanded bytes in any file, and at most 50 MiB of expanded bytes in total. These limits SHALL apply to both asset and repository archives, counting directory entries toward the entry limit.

#### Scenario: An archive expands beyond a budget
- **WHEN** decompression exceeds the per-file or aggregate expanded-byte limit
- **THEN** processing stops and identifies the exceeded limit
- **AND** remaining entries are not expanded

#### Scenario: An archive has too many entries
- **WHEN** an archive declares more than 1,000 entries
- **THEN** sx rejects it before reading file bodies

#### Scenario: Archive size declarations are untrustworthy
- **WHEN** declared entry sizes do not reliably describe the produced content
- **THEN** actual expanded bytes remain bounded

### Requirement: Every archive consumption path enforces the budgets
sx SHALL enforce applicable budgets before trusting metadata or processing archive bodies, including cached packages, metadata reads, hashing, copying entries, extraction, and repository discovery assembly. Existing path-traversal protections SHALL remain effective.

#### Scenario: Metadata itself is oversized
- **WHEN** a package contains metadata that expands beyond the file limit
- **THEN** metadata inspection fails within the budget before installation begins

#### Scenario: An old cached package exceeds current limits
- **WHEN** an existing cached package is used after limits are introduced
- **THEN** the package is checked before archive content is consumed
- **AND** cache presence does not bypass the limits

#### Scenario: Discovery builds a ZIP from separate files
- **WHEN** repository discovery fetches individual files and assembles an archive
- **THEN** file count, actual downloaded file bytes, aggregate expanded bytes, and the completed compressed output stay within the repository budgets

#### Scenario: An input is exactly at a limit
- **WHEN** a valid archive or response equals an applicable limit without exceeding another
- **THEN** the resource budget alone does not reject it
