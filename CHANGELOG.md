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

[0.11.0] - 2025-06-06
-------------------------

Version 0.11.0 adds support to use the Kion Favorites API endpoints (requires Kion version 3.13.0 or higher). This allows Kion to be the new source of truth for your configured favorites.

### Added

- Ability to list and use favorites (aliases) that are set in Kion
- New utility command to push local favorites up to Kion (`util push-favorites`)
- Option to delete local favorites once they've been pushed up
- New `default_region` property to the config file

[0.10.0] - 2025-04-29
---------------------

Version 0.10.0 adds the ability to target/name Firefox containers when federating with favorites, improves browser support for SAML based authentication, and adds an option for printing the authentication URL as opposed to opening it in the users default browser. The SAML auth URL option can be set with `kion.saml_print_url` in the configuration file, by the `--saml-print-url` flag, or with the `KION_SAML_PRINT_URL` environment variable. Normal precedent order applies.

### Added

- Option to print the SAML authentication url [kionsoftware/kion-cli/pull/84]
- Man page to Homebrew based installs [kionsoftware/kion-cli/pull/79]
- Ability to target a specific Firefox container when federating [kionsoftware/kion-cli/pull/86]

### Changed

- Modified browser calls to be more general and use user defaults [kionsoftware/kion-cli/pull/83]

[0.9.0] - 2025.01.16
--------------------

Version 0.9.0 adds the ability to federate directly into a specific service through favorites or the `console` command. No more hopping into the console dashboard then searching for and navigating to your destination. Note that the specified service is injected into the federation URLs `Destination` parameter. So, for example, the full path for the RDS service is `/rds/home?region=us-east-1#Home`, we just need the first part of the path `rds`.

```bash
# Add a service to your favorites
favorites:
  - name: mysandbox
    account: "121212121212"
    cloud_access_role: Admin
    access_type: web
    service: rds

# Then call it
kion fav mysandbox

# Or step through the selection wizard
kion con rds
```

### Added

- Ability to federate directly to a service [kionsoftware/kion-cli/pull/77]

[0.8.0] - 2025.01.07
--------------------

Version 0.8.0 adds support for Firefox containers, support for Windows Command Prompt and PowerShell, and improved shell history support when dropping into sub-shells (the `kion stak` command). Firefox container support adds the ability to federate into multiple AWS accounts at the same time for more advanced workflows. Note that Firefox container support requires the [Open external links in a container](https://addons.mozilla.org/en-US/firefox/addon/open-url-in-container) add-on as well as an update to your `~/.kion.yml` configuration file. See the repo `README.md` for more details. A big thank you to @joraff and @mjburling for help with development and testing!

### Added

- Added support for Firefox containers [kionsoftware/kion-cli/pull/72]
- Added support for Windows Command Prompt and PowerShell [kionsoftware/kion-cli/pull/72]
- Configured subshells to pull and set `HISTFILE` in zsh [kionsoftware/kion-cli/pull/70]

### Fixed

- Exit non-zero if an error is encountered [kionsoftware/kion-cli/pull/65]

[0.7.0] - 2024.09.18
--------------------

Version 0.7.0 adds flag support to console federation, addresses a bug that presented when using paths with AWS IAM roles, and adds a method for keeping your Kion password in encrypted storage (eg the system keyring).

### Added

- Added support for flags (account alias/number and car) to console federation [kionsoftware/kion-cli/pull/61]
- Added support for storing login password in encrypted storage [kionsoftware/kion-cli/pull/53]

### Fixed

- Console federation bug when using AWS IAM roles containing a path [kionsoftware/kion-cli/pull/60]

[0.6.0] - 2024-08-01
--------------------

Version 0.6.0 adds support for account aliases coming in Kion 3.9.9 and 3.10.2. Account aliases are globally unique user defined identifiers for accounts stored in Kion. Aliases can be used with the `stak` and `run` commands instead of specifying account numbers.

### Added

- Added support for account aliases [kionsoftware/kion-cli/pull/51]


[0.5.0] - 2024-06-24
--------------------

This release changes how caching is handled for Gnome users. After upgrading a new empty cache in the default `login` keyring will be used. The old `kion-cli` keyring can be safely removed.

### Changed

- Updated keyring config for Gnome Wallet (libsecret) to use the default `login` keyring [kionsoftware/kion-cli/pull/49]


[0.4.1] - 2024-06-24
--------------------

### Fixed

- Patched the package `github.com/dvsekhvalnov/jose2go` to version 1.6.0 to address Dependabot security findings [kionsoftware/kion-cli/pull/48]


[0.4.0] - 2024-06-18
--------------------

SAML Authentication is now supported for Kion versions `< 3.8.0`. No additional configuration is required for use, see `README.md` for details on SAML authentication with the CLI.

### Added

- A new version constraint will switch between SAML authentication behaviors based on the target Kion version. [kionsoftware/kion-cli/pull/46]

[0.3.0] - 2024-06-03
--------------------

You can now use Kion CLI with multiple instances of Kion through the use of configuration profiles or by pointing to alternate configuration files. Here are some usage examples:

```bash
# point to another configuration file
KION_CONFIG=~/.kion.development.yml kion stak

# use a 'development' profile within your ~/.kion.yml configuration file
kion --profile development fav sandbox
```

A configuration file for the profile usage example above would look something like this:

```yaml
# default profile if none specified
kion:
  url: https://kion.mycompany.com
  api_key: "app_123"
favorites:
  - name: production
    account: "232323232323"
    cloud_access_role: ReadOnly

# alternate profiles called with the global `--profile [name]` flag
profiles:
  development:
    kion:
      url: https://dev.kion.mycompany.com
      api_key: "app_abc"
    favorites:
      - name: sandbox
        account: "121212121212"
        cloud_access_role: Admin
```

### Added

- Users can now set a custom config file with the `KION_CONFIG` environment variable [kionsoftware/kion-cli/pull/42]
- Users can define profiles to use Kion CLI with multiple Kion instances [kionsoftware/kion-cli/pull/42]
- Created a `util` command and `flush-cache` subcommand to flush the cache [kionsoftware/kion-cli/pull/42]

### Fixed

- Corrected an issue where the Kion CLI configuration file was not actually optional [kionsoftware/kion-cli/pull/42]

[0.2.1] - 2024-05-30
--------------------

### Changed

- Federating into the web console is now handled without iframes or javascript [kionsoftware/kion-cli/pull/40]

[0.2.0] - 2024-05-24
--------------------

Caching and AWS `credential_process` support has been added to the Kion CLI! See the AWS docs [HERE](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-sourcing-external.html) for more information as well as the `README.md` document in this repo for examples on how to use Kion CLI as a credential provider.

Kion CLI will now use cached STAKs by default to improve performance and reduce the number of calls to Kion. STAKs will be considered as valid for 15 minutes unless Kion reports back a longer STAK duration. Note that Kion is expected to start returning the duration of a STAK along with the STAK itself starting on versions 3.6.29, 3.7.19, 3.8.13, and 3.9.5.

The cache will be stored in the system's keychain, and depending on your operating system, you may be prompted to allow Kion CLI to access the cache entry on your first run.

Cached STAKs will be used by default unless:
- Caching is disabled via the `--disable-cache` global flag
- Caching is disabled in the `~/.kion.yml` configuration file by setting `kion.disable_cache: true`
- The credential has less than 5 seconds left and Kion CLI is being used as an AWS credential provider
- The credential has less than 5 seconds left and Kion CLI is being used to run an ad hoc command
- The credential has less than 5 minutes left and Kion CLI is being used to print keys
- The credential has less than 5 minutes left and Kion CLI is being used to create an authenticated subshell
- The credential has less than 10 minutes left and Kion CLI is being used to create an AWS configuration profile

Lastly, the following environment variables will no longer be set when using the `run` command to execute ad hoc commands:

  ```bash
  KION_ACCOUNT_NUM
  KION_ACCOUNT_ALIAS
  KION_CAR
  ```

### Added

- Support to use Kion CLI as a credential process subsystem for AWS profiles [kionsoftware/kion-cli/pull/38]
- Add caching for faster operations [kionsoftware/kion-cli/pull/38]
- SAML tokens are now cached for 9.5 minutes [kionsoftware/kion-cli/pull/39]

### Changed

- Kion session data has moved from the `~/.kion.yml` configuration file to the cache [kionsoftware/kion-cli/pull/39]

### Removed

- `KION_*` env variables removed from subshell environments when using the `run` command [kionsoftware/kion-cli/pull/38]

[0.1.1] - 2024-05-20
--------------------

### Fixed

- Corrected version number when running `--version` flag [kionsoftware/kion-cli/pull/36]

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
