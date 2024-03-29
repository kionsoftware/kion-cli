Change Log
==========

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/). This project adheres to [Semantic Versioning](http://semver.org/) with the exception of version 0 as we find our footing. Only changes to the application should be logged here. Repository maintenance, tests, and other non application changes should be excluded.

[Unreleased] - yyyy-mm-dd
-------------------------

Notes for upgrading...

### Added
- Print metadata to stdout when federating into web consoles [https://github.com/kionsoftware/kion-cli/pull/20]
- Add flags to support headless runs [https://github.com/kionsoftware/kion-cli/pull/22]
- Fallback logic for users with restricted perms when using the `run` cmd [https://github.com/kionsoftware/kion-cli/pull/22]

### Changed

- Renamed `access_type` values for clarity [https://github.com/kionsoftware/kion-cli/pull/11]
- Improve logic around web federation [https://github.com/kionsoftware/kion-cli/pull/21]
- Dynamically output account name info if available [https://github.com/kionsoftware/kion-cli/pull/22]

### Deprecated

### Removed

### Fixed

- Fix unexpected EOF when creating Bash subshells [https://github.com/kionsoftware/kion-cli/pull/14]
- Improve CAR selection logic and usage wording [https://github.com/kionsoftware/kion-cli/pull/19]
- Fix SAML auth around private network access checks [https://github.com/kionsoftware/kion-cli/pull/23]

[0.0.2] - 2024-02-23
--------------------

### Added

- Web console access! [https://github.com/kionsoftware/kion-cli/pull/10]

### Changed

### Deprecated

### Removed

### Fixed

- Add workaround for users with `Browse Project - Minimal` permissions [https://github.com/kionsoftware/kion-cli/pull/8]
- Ensure STAK output can be eval'd [https://github.com/kionsoftware/kion-cli/pull/1]

[0.0.1] - 2024-02-02
--------------------

Initial release.
