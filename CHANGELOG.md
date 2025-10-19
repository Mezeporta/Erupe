# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Alpelo object system backport functionality
- Better config file handling and structure

### Changed

- Improved config handling

### Fixed

- Config file handling and validation

### Security

- Bumped golang.org/x/net from 0.33.0 to 0.38.0
- Bumped golang.org/x/crypto from 0.31.0 to 0.35.0

## Removed

- Compatibility with Go 1.21 removed.

## [9.2.0] - 2023-04-01

### Added in 9.2.0

- Gacha system with box gacha and stepup gacha support
- Multiple login notices support
- Daily quest allowance configuration
- Gameplay options system
- Support for stepping stone gacha rewards
- Guild semaphore locking mechanism
- Feature weapon schema and generation system
- Gacha reward tracking and fulfillment
- Koban my mission exchange for gacha

### Changed in 9.2.0

- Reworked logging code and syntax
- Reworked broadcast functions
- Reworked netcafe course activation
- Reworked command responses for JP chat
- Refactored guild message board code
- Separated out gacha function code
- Rearranged gacha functions
- Updated golang dependencies
- Made various handlers non-fatal errors
- Moved various packet handlers
- Moved caravan event handlers
- Enhanced feature weapon RNG

### Fixed in 9.2.0

- Mail item workaround removed (replaced with proper implementation)
- Possible infinite loop in gacha rolls
- Feature weapon RNG and generation
- Feature weapon times and return expiry
- Netcafe timestamp handling
- Guild meal enumeration and timer
- Guild message board enumerating too many posts
- Gacha koban my mission exchange
- Gacha rolling and reward handling
- Gacha enumeration recommendation tag
- Login boost creating hanging connections
- Shop-db schema issues
- Scout enumeration data
- Missing primary key in schema
- Time fixes and initialization
- Concurrent stage map write issue
- Nil savedata errors on logout
- Patch schema inconsistencies
- Edge cases in rights integer handling
- Missing period in broadcast strings

### Removed in 9.2.0

- Unused database tables
- Obsolete LauncherServer code
- Unused code from gacha functionality
- Mail item workaround (replaced with proper implementation)

### Security in 9.2.0

- Escaped database connection arguments

## [9.1.1] - 2022-11-10

### Changed in 9.1.1

- Temporarily reverted versioning system
- Fixed netcafe time reset behavior

## [9.1.0] - 2022-11-04

### Added in 9.1.0

- Multi-language support system
- Support for JP strings in broadcasts
- Guild scout language support
- Screenshot sharing support
- New sign server implementation
- Multi-language string mappings
- Language-based chat command responses

### Changed in 9.1.0

- Rearranged configuration options
- Converted token to library
- Renamed sign server
- Mapped language to server instead of session

### Fixed in 9.1.0

- Various packet responses

## [9.1.0-rc3] - 2022-11-02

### Fixed in 9.1.0-rc3

- Prevented invalid bitfield issues

## [9.1.0-rc2] - 2022-10-28

### Changed in 9.1.0-rc2

- Set default featured weapons to 1

## [9.1.0-rc1] - 2022-10-24

### Removed in 9.1.0-rc1

- Migrations directory

## [9.0.1] - 2022-08-04

### Changed in 9.0.1

- Updated login notice

## [9.0.0] - 2022-08-03

### Fixed in 9.0.0

- Fixed readlocked channels issue
- Prevent rp logs being nil
- Prevent applicants from receiving message board notifications

### Added in 9.0.0

- Implement guild semaphore locking
- Support for more courses
- Option to flag corruption attempted saves as deleted
- Point limitations for currency

---

## Historical Context

This changelog documents releases from v9.0.0 onwards. For a complete history of all changes, refer to the [git repository](https://github.com/Mezeporta/Erupe).

The project follows semantic versioning and maintains tagged releases for stable versions. Development continues on the main branch with features merged from feature branches.
