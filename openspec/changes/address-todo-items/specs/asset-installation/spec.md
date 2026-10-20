# Spec Delta

## Purpose

Keep directory-based AI assets usable when an attempted installation or update fails before the new version can be safely published.

## ADDED Requirements

### Requirement: Failed preparation preserves the working installation
A directory asset update SHALL preserve the previous installation and its contents if metadata validation, resource-budget checks, or extraction fails. A failed first installation SHALL NOT leave a partial asset in the final installation location.

#### Scenario: A later archive entry is corrupt
- **WHEN** an update has readable metadata but extraction fails on a later entry
- **THEN** the previously installed files remain unchanged at the installation location
- **AND** sx reports the failed update

#### Scenario: An update exceeds archive limits
- **WHEN** a replacement package fails an archive budget check
- **THEN** the working version remains installed

#### Scenario: A first installation fails
- **WHEN** preparation fails and no prior version is installed
- **THEN** the final installation location contains no partial new asset

### Requirement: Replacement failures retain a recoverable working version
sx SHALL prepare the complete replacement before displacing a working asset. If final replacement fails, sx SHALL restore the old installation when restoration is possible. If restoration also fails, sx SHALL preserve the old version's backup and report its recovery location rather than deleting it.

#### Scenario: Final replacement fails and restoration succeeds
- **WHEN** the prepared replacement cannot be promoted after the old version has been displaced
- **THEN** the old version is restored to the installation location
- **AND** the operation reports failure

#### Scenario: Restoration itself fails
- **WHEN** replacement and restoration both fail
- **THEN** the old version remains available in its backup location
- **AND** the error identifies that location and both failures

#### Scenario: An update succeeds
- **WHEN** preparation and replacement succeed
- **THEN** the final location contains the complete replacement including required file permissions
- **AND** temporary staging and redundant backup data are removed or cleanup failure is reported without discarding the working replacement
