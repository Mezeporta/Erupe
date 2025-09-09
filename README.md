# Erupe

## Client Compatibility

### Platforms

- PC
- PlayStation 3
- PlayStation Vita
- Wii U (Up to Z2)

### Versions (ClientMode)

- All versions after HR compression (G10-ZZ) have been tested extensively and have great functionality.
- All versions available on Wii U (G3-Z2) have been tested and should have good functionality.
- The second oldest found version is Forward.4 (FW.4), this version has basic functionality.
- The oldest found version is Season 6.0 (S6.0), however functionality is very limited.

If you have an **installed** copy of Monster Hunter Frontier on an old hard drive, **please** get in contact so we can archive it!

## Setup

If you are only looking to install Erupe, please use [a pre-compiled binary](https://github.com/ZeruLight/Erupe/releases/latest).

If you want to modify or compile Erupe yourself, please read on.

## Requirements

- [Go](https://go.dev/dl/)
- [PostgreSQL](https://www.postgresql.org/download/)

## Installation

1. Bring up a fresh database by using the [backup file attached with the latest release](https://github.com/ZeruLight/Erupe/releases/latest/download/SCHEMA.sql).
2. Run each script under [patch-schema](./schemas/patch-schema) as they introduce newer schema.
3. Edit [config.json](./config.json) such that the database password matches your PostgreSQL setup.
4. Run `go build` or `go run .` to compile Erupe.

## Docker

Please see [docker/README.md](./docker/README.md). This is intended for quick installs and development, not for production.

## Schemas

We source control the following schemas:

- Initialization Schema: This initializes the application database to a specific version (9.1.0).
- Update Schemas: These are update files that should be ran on top of the initialization schema.
- Patch Schemas: These are for development and should be run after running all initialization and update schema. These get condensed into `Update Schemas` and deleted when updated to a new release.
- Bundled Schemas: These are demo reference files to give servers standard set-ups.

Note: Patch schemas are subject to change! You should only be using them if you are following along with development.

## Resources

- [Quest and Scenario Binary Files](https://files.catbox.moe/xf0l7w.7z)
- [Mezeporta Square Discord](https://discord.gg/DnwcpXM488)
