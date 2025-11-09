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

- [Go](https://go.dev/dl/): programming language for the server logic
- [PostgreSQL](https://www.postgresql.org/download/): server database.

### Installation

1. Bring up a fresh database by using the [backup file attached with the latest release](https://github.com/ZeruLight/Erupe/releases/latest/download/SCHEMA.sql).
2. Run each script under [patch-schema](./patch-schema) as they introduce newer schema.
3. Edit [config.json](./config.json) such that the database password matches your PostgreSQL setup.
4. Launch `go run .` to run erupe directly, or use `go build` to compile Erupe.

### Note

You will need to acquire and install the client files and quest binaries separately.
See the [resources](#resources) for details.

## Resources

- [Quest and Scenario Binary Files](https://files.catbox.moe/xf0l7w.7z)
- [PewPewDojo Discord](https://discord.gg/CFnzbhQ)
- [Community FAQ Pastebin](https://pastebin.com/QqAwZSTC)
