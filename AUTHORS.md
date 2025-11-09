# List of authors who contributed to Erupe

## Point of current development

The project is currently developed under <https://github.com/ZeruLight/Erupe>

## History of development

Development of this project dates back to 2019, and was developed under various umbrellas over time:

### Cappuccino (Fist/Ando/Ellie42) - "The Erupe Developers" (2019-2020)

<https://github.com/Ellie42/Erupe> / <https://github.com/ricochhet/Erupe-Legacy>

**Initial proof-of-concept** and foundational work:

* Basic server infrastructure (Sign, Entrance, Channel servers)
* Account registration and character creation systems
* Initial multiplayer lobby functionality
* Core network communication layer
* Save data compression using delta/diff encoding
* Stage management and reservation systems for multiplayer quests
* Party system supporting up to 4 players
* Chat system (local, party, private messaging)
* Hunter Navi NPC interactions
* Diva Defense feature
* Quest selection and basic quest support
* PostgreSQL database integration with migration support

**Technical Details:**

* Repository created: March 6, 2020
* Public commits: March 4-12, 2020 (9 days of visible development)
* Total commits: 142
* Status: Still active closed source

The original developers created this as an educational project to learn server emulation. This version established the fundamental architecture that all subsequent versions built upon.

### Einherjar Team (~2020-2022 Feb)

**Major expansion period** (estimated March 2020 - February 2022):

Unfortunately, **no public git history exists** for this critical development period. The Einherjar Team's work was used as the foundation for all subsequent community repositories. Based on features present in the Community Edition fork (February 2022) that weren't in the original Cappuccino version, the Einherjar Team likely implemented:

* Extensive quest system improvements
* Guild system foundations
* Economy and item distribution systems
* Additional game mechanics and features
* Stability improvements and bug fixes
* Database schema expansions

This ~2-year period represents the largest gap in documented history. If anyone has information about this team's contributions, please contact the project maintainers.

### Community Edition (2022)

<https://github.com/xl3lackout/Erupe>

**Community-driven consolidation** (February 6 - August 7, 2022):

* Guild system enhancements:
  * Guild alliances support
  * Guild member management (Pugi renaming)
  * SJIS support for guild posts (Japanese characters)
  * Guild message boards
* Character and account improvements:
  * Mail system with locking mechanism
  * Favorite quest persistence
  * Title/achievement enumeration
  * Character data handler rewrites
* Game economy features:
  * Item distribution handling
  * Road Shop rotation system
  * Scenario counter tracking
* Technical improvements:
  * Stage and semaphore overhaul
  * Discord bot integration with chat broadcasting
  * Error handling enhancements in launcher
  * Configuration improvements

**Technical Details:**

* Repository created: February 6, 2022
* Active development: May 11 - August 7, 2022 (3 months)
* Total commits: 69
* Contributors: Ando, Fists Team, the French Team, Mai's Team, and the MHFZ community

This version focused on making the server accessible to the broader community and implementing social/multiplayer features.

### ZeruLight / Mezeporta (2022-present)

<https://github.com/ZeruLight/Erupe> (now <https://github.com/Mezeporta/Erupe>)

**Major feature expansion and maturation** (March 24, 2022 - Present):

**Version 9.0.0 (August 2022)** - Major systems implementation:

* MezFes festival gameplay (singleplayer minigames)
* Friends lists and block lists (blacklists)
* Guild systems:
  * Guild Treasure Hunts
  * Guild Cooking system
  * Guild semaphore locking
* Series Quests playability
* My Series visits customization
* Raviente rework (multiple simultaneous instances)
* Stage system improvements
* Currency point limitations

**Version 9.1.0 (November 2022)** - Internationalization:

* Multi-language support system (Japanese initially)
* JP string support in broadcasts
* Guild scout language support
* Screenshot sharing support
* New sign server implementation
* Language-based chat command responses
* Configuration restructuring

**Version 9.2.0 (April 2023)** - Gacha and advanced systems:

* Complete gacha system (box gacha, stepup gacha)
* Multiple login notices
* Daily quest allowance configuration
* Gameplay options system
* Feature weapon schema and generation
* Gacha reward tracking and fulfillment
* Koban my mission exchange
* NetCafe course activation improvements
* Guild meal enumeration and timers
* Mail system improvements
* Logging and broadcast function overhauls

**Unreleased/Current (2023-2025)** - Stability and quality improvements:

* Comprehensive production logging for all save operations
* Session lifecycle tracking with metrics
* Disconnect type tracking (graceful, connection_lost, error)
* Critical race condition fixes in stage handlers
* Deadlock fixes in zone changes
* Save data corruption fixes
* Transmog/plate data persistence fixes
* Logout flow improvements preventing data loss
* Config file handling improvements
* Alpelo object system backport
* Security updates (golang dependencies)

**Technical Details:**

* Repository created: March 24, 2022
* Latest activity: January 2025 (actively maintained)
* Total commits: 1,295+
* Contributors: 20+
* Releases: 9 major releases
* Multi-version support: Season 6.0 to ZZ
* Multi-platform: PC, PS3, PS Vita, Wii U (up to Z2)

This version transformed Erupe from a proof-of-concept into a feature-complete, stable server emulator with extensive game system implementations and ongoing maintenance.

### sekaiwish Fork (2024)

<https://github.com/sekaiwish/Erupe>

**Recent fork** (November 10, 2024):

* Fork of Mezeporta/Erupe
* Total commits: 1,260
* Purpose and specific contributions: Unknown (recently created)

This is a recent fork and its specific goals or contributions are not yet documented.

## Authorship of the code

Authorship is assigned for each commit within the git history, which is stored in these git repos:

* <https://github.com/ZeruLight/Erupe>
* <https://github.com/Ellie42/Erupe>
* <https://github.com/ricochhet/Erupe-Legacy>
* <https://github.com/xl3lackout/Erupe>

Note the divergence between Ellie42's branch and xl3lackout's where history has been lost.

Unfortunately, we have no detailed information on the history of Erupe before 2022.
If somebody can provide information, please contact us, so that we can make this history available.

## Exceptions with third-party libraries

The third-party libraries have their own way of addressing authorship and the authorship of commits importing/updating
a third-party library reflects who did the importing instead of who wrote the code within the commit.

The authors of third-party libraries are not explicitly mentioned, and usually is possible to obtain from the files belonging to the third-party libraries.
