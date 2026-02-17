# Docker for Erupe

## Quick Start

1. From the repository root, copy and edit the config:

   ```bash
   cp config.example.json docker/config.json
   ```

   Edit `docker/config.json` — set `Database.Host` to `"db"` and match the password to `docker-compose.yml` (default: `password`).

2. Place your [quest/scenario files](https://files.catbox.moe/xf0l7w.7z) in `docker/bin/`.

3. Start everything:

   ```bash
   cd docker
   docker compose up
   ```

The database is automatically initialized and patched on first start via `init/setup.sh`.

pgAdmin is available at `http://localhost:5050` (default login: `user@pgadmin.com` / `password`).

## Building Locally

By default the server service pulls the prebuilt image from GHCR. To build from source instead, edit `docker-compose.yml`: comment out the `image` line and uncomment the `build` section, then:

```bash
docker compose up --build
```

## Stopping the Server

```bash
docker compose stop     # Stop containers (preserves data)
docker compose down     # Stop and remove containers (preserves data volumes)
```

To delete all persistent data, remove these directories after stopping:

- `docker/db-data/`
- `docker/savedata/`

## Updating

After pulling new changes:

1. Check for new patch schemas in `schemas/patch-schema/` — apply them via pgAdmin or `psql` into the running database container.

2. Rebuild and restart:

   ```bash
   docker compose down
   docker compose build
   docker compose up
   ```

## Troubleshooting

**Postgres won't populate on Windows**: `init/setup.sh` must use LF line endings, not CRLF. Open it in your editor and convert.
