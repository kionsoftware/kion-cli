Change Log
==========

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/). This project adheres to [Semantic Versioning](http://semver.org/) with the exception of version 0 as we find our footing. Only changes to the application should be logged here. Repository maintenance, tests, and other non application changes should be excluded.

[Unreleased] - yyyy-mm-dd
-------------------------

Notes for upgrading...

### Added

### Changed

### Deprecated

### Removed

### Fixed

[0.1.0] - 2024-05-20
--------------------

### Added

- Print metadata to stdout when federating into web consoles [kionsoftware/kion-cli/pull/20]
- Add flags to support headless runs [kionsoftware/kion-cli/pull/22]
- Fallback logic for users with restricted perms when using the `run` cmd [kionsoftware/kion-cli/pull/22]
- Logic to accommodate users with cloud access only permissions [kionsoftware/kion-cli/pull/24]
- STAK selection wizard now includes project and car IDs and account numbers [kionsoftware/kion-cli/pull/24]
- Automate AWS logout before federating into the AWS console [kionsoftware/kion-cli/pull/25]
- Support defining region on favorites or via flag [kionsoftware/kion-cli/pull/26]
- Support for old `ctkey` usage by adding compatibility commands [kionsoftware/kion-cli/pull/28]
- Ability to save short-term access keys to an AWS credentials profile [kionsoftware/kion-cli/pull/28]
- Add support for windows when printing STAKs for export [kionsoftware/kion-cli/pull/28]

### Changed

- Renamed `access_type` values for clarity [kionsoftware/kion-cli/pull/11]
- Improve logic around web federation [kionsoftware/kion-cli/pull/21]
- Dynamically output account name info if available [kionsoftware/kion-cli/pull/22]

### Fixed

- Fix unexpected EOF when creating Bash subshells [kionsoftware/kion-cli/pull/14]
- Improve CAR selection logic and usage wording [kionsoftware/kion-cli/pull/19]
- Fix SAML auth around private network access checks [kionsoftware/kion-cli/pull/23]
- Fixed automated logouts of AWS console sessions on Firefox [kionsoftware/kion-cli/pull/31]

[0.0.2] - 2024-02-23
--------------------

### Added

- Web console access! [kionsoftware/kion-cli/pull/10]

### Fixed

- Add workaround for users with `Browse Project - Minimal` permissions [kionsoftware/kion-cli/pull/8]
- Ensure STAK output can be eval'd [kionsoftware/kion-cli/pull/1]

[0.0.1] - 2024-02-02
--------------------

Initial release.
