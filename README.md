# Erupe Community Edition

Erupe is a community-built server for Monster Hunter Frontier - All versions.
It is a complete reverse engineered solution to self-host a Monter Hunter Frontier server.

> ![IMPORTANT]
> The purpose of this branch is to have a clean transition from a functional 9.2.0 release, to a future 9.3.0 version.
> Over the last 2 years after the release of 9.2.0, many commits introduced broken features.

## Setup

If you are only looking to install Erupe, please use a [pre-compiled binary](https://github.com/ZeruLight/Erupe/releases/latest).

If you want to modify or compile Erupe yourself, please read on.

### Requirements

Install is simple, you need:

- [Go 1.25+](https://go.dev/dl/): programming language for the server logic
- [PostgreSQL](https://www.postgresql.org/download/): server database

### Installation

#### First-time Setup

1. Clone the repository:

   ```bash
   git clone https://github.com/ZeruLight/Erupe.git
   cd Erupe
   ```

2. Create a new PostgreSQL database and install the schema:

   ```bash
   # Download and apply the base schema
   wget https://github.com/ZeruLight/Erupe/releases/latest/download/SCHEMA.sql
   psql -U your_user -d your_database -f SCHEMA.sql
   ```

3. Run each script under [patch-schema](./patch-schema) to apply schema updates:

   ```bash
   psql -U your_user -d your_database -f patch-schema/01_patch.sql
   # Repeat for each patch file in order
   ```

4. Copy [config.example.json](./config.example.json) to `config.json` and edit it:

   ```bash
   cp config.example.json config.json
   # Edit config.json with your database credentials
   ```

5. Install dependencies and run:

   ```bash
   go mod download
   go run .
   ```

   Or build a binary:

   ```bash
   go build
   ./erupe-ce
   ```

#### Updating an Existing Installation

1. Pull the latest changes:

   ```bash
   git pull origin main
   ```

2. Update dependencies:

   ```bash
   go mod tidy
   ```

3. Apply any new schema patches from [patch-schema](./patch-schema)

4. Rebuild and restart:

   ```bash
   go build
   ./erupe-ce
   ```

### Note

You will need to acquire and install the client files and quest binaries separately.
See the [resources](#resources) for details.

## Resources

- [Quest and Scenario Binary Files](https://files.catbox.moe/xf0l7w.7z)
- [PewPewDojo Discord](https://discord.gg/CFnzbhQ): community for discussion.
- [Mogapedia's Discord](https://discord.gg/f77VwBX5w7): Discord community responsible for this branch.
- [Community FAQ Pastebin](https://pastebin.com/QqAwZSTC)
