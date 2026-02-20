# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Rework object ID allocation: per-session IDs replace shared map, simplify stage entry notifications
- Better config file handling and structure
- Comprehensive production logging for save operations (warehouse, Koryo points, savedata, Hunter Navi, plate equipment)
- Disconnect type tracking (graceful, connection_lost, error) with detailed logging
- Session lifecycle logging with duration and metrics tracking
- Structured logging with timing metrics for all database save operations
- Plate data (transmog) safety net in logout flow - adds monitoring checkpoint for platedata, platebox, and platemyset persistence

### Changed

- Updated minimum Go version requirement from 1.23 to 1.25
- Improved config handling
- Refactored logout flow to save all data before cleanup (prevents data loss race conditions)
- Unified save operation into single `saveAllCharacterData()` function with proper error handling
- Removed duplicate save calls in `logoutPlayer()` function

### Fixed

- Config file handling and validation
- Fixes 3 critical race condition in handlers_stage.go.
- Fix an issue causing a crash on clans with 0 members.
- Fixed deadlock in zone change causing 60-second timeout when players change zones
- Fixed crash when sending empty packets in QueueSend/QueueSendNonBlocking
- Fixed missing stage transfer packet for empty zones
- Fixed save data corruption check rejecting valid saves due to name encoding mismatches (SJIS/UTF-8)
- Fixed incomplete saves during logout - character savedata now persisted even during ungraceful disconnects
- Fixed double-save bug in logout flow that caused unnecessary database operations
- Fixed save operation ordering - now saves data before session cleanup instead of after
- Fixed stale transmog/armor appearance shown to other players - user binary cache now invalidated when plate data is saved
- Fixed client crash when quest or scenario files are missing - now sends failure ack instead of nil data
- Fixed server crash when Discord relay receives messages with unsupported Shift-JIS characters (emoji, Lenny faces, cuneiform, etc.)
- Fixed data race in token.RNG global used concurrently across goroutines

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

## Pre-9.0.0 Development (2022-02-25 to 2022-08-03)

The period before version 9.0.0 represents the early community development phase, starting with the Community Edition reupload and continuing through multiple feature additions leading up to the first semantic versioning release.

### [Pre-release] - 2022-06-01 to 2022-08-03

Major feature implementations leading to 9.0.0:

#### Added (June-August 2022)

- **Friend System**: Friend list functionality with cross-character enumeration
- **Blacklist System**: Player blocking functionality
- **My Series System**: Basic My Series functionality with shared data and bookshelf support
- **Guild Treasure Hunts**: Complete guild treasure hunting system with cooldowns
- **House System**:
  - House interior updates and furniture loading
  - House entry handling improvements
  - Visit other players' houses with correct furniture display
- **Festa System**:
  - Initial Festa build and decoding
  - Canned Festa prizes implementation
  - Festa finale acquisition handling
  - Festa info and packet handling improvements
- **Achievement System**: Hunting career achievements concept implementation
- **Object System**:
  - Object indexing (v3, v3.1)
  - Semaphore indexes
  - Object index limits and reuse prevention
- **Transit Message**: Correct parsing of transit messages for minigames
- **World Chat**: Enabled world chat functionality
- **Rights System**: Rights command and permission updates on login
- **Customizable Login Notice**: Support for custom login notices

#### Changed (June-August 2022)

- **Stage System**: Major stage rework and improvements
- **Raviente System**: Cleanup, fixes, and announcement improvements
- **Discord Integration**: Mediated Discord handling improvements
- **Server Logging**: Improved server logging throughout
- **Configuration**: Edited default configs
- **Repository**: Extensive repository cleanup
- **Build System**: Implemented build actions and artifact generation

#### Fixed (June-August 2022)

- Critical semaphore bug fixes
- Raviente-related fixes and cleanup
- Read-locked channels issue
- Stubbed title enumeration
- Object index reuse prevention
- Crash when not in guild on logout
- Invalid schema issues
- Stage enumeration crash prevention
- Gook (book) enumeration and cleanup
- Guild SQL fixes
- Various packet parsing improvements
- Semaphore checking changes
- User insertion not broadcasting

### [Pre-release] - 2022-05-01 to 2022-06-01

Guild system enhancements and social features:

#### Added (May-June 2022)

- **Guild Features**:
  - Guild alliance support with complete implementation
  - Guild member (Pugi) management and renaming
  - Guild post SJIS (Japanese) character encoding support
  - Guild message board functionality
  - Guild meal system
  - Diva Hall adventure cat support
  - Guild adventure cat implementation
  - Alliance members included in guild member enumeration
- **Character System**:
  - Mail locking mechanism
  - Favorite quest save/load functionality
  - Title/achievement enumeration parsing
  - Character data handler rewrite
- **Game Features**:
  - Item distribution handling system
  - Road Shop weekly rotation
  - Scenario counter implementation
  - Diva adventure dispatch parsing
  - House interior query support
  - Entrance and sign server response improvements
- **Launcher**:
  - Discord bot integration with configurable channels and dev roles
  - Launcher error handling improvements
  - Launcher finalization with modal, news, menu, safety links
  - Auto character addition
  - Variable centered text support
  - Last login timestamp updates

#### Changed (May-June 2022)

- Stage and semaphore overhaul with improved casting handling
- Simplified guild handler code
- String support improvements with PascalString helpers
- Byte frame converted to local package
- Local package conversions (byteframe, pascalstring)

#### Fixed (May-June 2022)

- SJIS guild post support
- Nil guild failsafes
- SQL queries with missing counter functionality
- Enumerate airoulist parsing
- Mail item description crashes
- Ambiguous mail query
- Last character updates
- Compatibility issues
- Various packet files

### [Pre-release] - 2022-02-25 to 2022-05-01

Initial Community Edition and foundational work:

#### Added (February-May 2022)

- **Core Systems**:
  - Japanese Shift-JIS character name support
  - Character creation with automatic addition
  - Raviente system patches
  - Diva reward handling
  - Conquest quest support
  - Quest clear timer
  - Garden cat/shared account box implementation
- **Guild Features**:
  - Guild hall available on creation
  - Unlocked all street titles
  - Guild schema corrections
- **Launcher**:
  - Complete launcher implementation
  - Modal dialogs
  - News system
  - Menu and safety links
  - Button functionality
  - Caching system

#### Changed (February-May 2022)

- Save compression updates
- Migration folder moved to root
- Improved launcher code structure

#### Fixed (February-May 2022)

- Mercenary/cat handler fixes
- Error code 10054 (savedata directory creation)
- Conflicts resolution
- Various syntax corrections

---

## Historical Context

This changelog documents all known changes from the Community Edition reupload (February 25, 2022) onwards. The period before this (Einherjar Team era, ~2020-2022) has no public git history.

Earlier development by Cappuccino/Ellie42 (March 2020) focused on basic server infrastructure, multiplayer systems, and core functionality. See [AUTHORS.md](AUTHORS.md) for detailed development history.

The project began following semantic versioning with v9.0.0 (August 3, 2022) and maintains tagged releases for stable versions. Development continues on the main branch with features merged from feature branches.
