# Docker for erupe

## Prerequisites

Before using Docker with erupe, you need to obtain the database schemas:

1. **Option A**: Copy the `schemas/` directory from the main branch
2. **Option B**: Download from GitHub releases and set up the directory structure

The `schemas/` directory should be placed at the same level as the erupe project root, or you can modify the docker-compose.yml to point to wherever you place it.

Expected structure:

```text
schemas/
├── init.sql                    # Base schema
├── update-schema/              # Version update schemas
│   └── 9.2-update.sql
├── patch-schema/               # Development patches
│   └── *.sql
└── bundled-schema/             # Optional demo data
    └── *.sql
```

## Building the container

Run the following from the root of the source folder. In this example we give it the tag of dev to separate it from any other container versions.

```bash
docker build . -t erupe:dev
```

## Running the container in isolation

This is just running the container. You can do volume mounts into the container for the `config.json` to tell it to communicate to a database. You will need to do this also for other folders such as `bin` and `savedata`

```bash
docker run erupe:dev
```

## Docker compose

Docker compose allows you to run multiple containers at once. The docker compose in this folder has 4 services set up:

- **postgres** - PostgreSQL database server
- **pgadmin** - Admin interface to make database changes
- **erupe** - The game server
- **web** - Apache web server for hosting static files

We automatically populate the database to the latest version on start. If you are updating, you will need to apply the new schemas manually.

### Configuration

Before we get started, you should make sure the database info matches what's in the docker-compose file for the environment variables `POSTGRES_PASSWORD`, `POSTGRES_USER` and `POSTGRES_DB`. You can set the host to be the service name `db`.

Here is an example of what you would put in the config.json if you were to leave the defaults. **It is strongly recommended to change the password.**

```json
"Database": {
    "Host": "db",
    "Port": 5432,
    "User": "postgres",
    "Password": "password",
    "Database": "erupe"
}
```

Place this file at `./config.json` (in the project root)

You will need to do the same for your bins - place these in `./bin`

### Setting up the web hosted materials

Clone the Servers repo into `./docker/Servers`

Make sure your hosts are pointing to where this is hosted.

### Starting the server

Navigate to the `docker/` directory and run:

```bash
cd docker
docker-compose up -d
```

This boots the database, pgadmin, and the server in a detached state.

If you want all the logs and you want it to be in an attached state:

```bash
docker-compose up
```

### Accessing services

- **erupe server**: Configured ports (53310, 53312, 54001-54008, etc.)
- **pgAdmin**: <http://localhost:5050>
  - Email: <user@pgadmin.com>
  - Password: password
- **Web server**: <http://localhost:80>
- **PostgreSQL**: <localhost:5432>

### Turning off the server safely

```bash
docker-compose stop
```

### Turning off the server destructively

```bash
docker-compose down
```

Make sure if you want to delete your data, you delete the folders that persisted:

- `./docker/savedata`
- `./docker/db-data`

## Testing with Docker

For running integration tests with a database, use the test configuration:

```bash
# Start test database
docker-compose -f docker/docker-compose.test.yml up -d

# Run tests (from project root)
go test ./...

# Stop test database
docker-compose -f docker/docker-compose.test.yml down
```

The test database:

- Runs on port 5433 (to avoid conflicts with development database)
- Uses in-memory storage (tmpfs) for faster tests
- Is ephemeral - data is lost when the container stops

## Troubleshooting

### Q: My Postgres will not populate

**A:** Your `setup.sh` is maybe saved as CRLF, it needs to be saved as LF.

On Linux/Mac:

```bash
dos2unix docker/init/setup.sh
```

Or manually convert line endings in your editor.

### Q: I get "schemas not found" errors

**A:** Make sure you have the `schemas/` directory set up correctly. Check that:

1. The directory exists at `../schemas/` relative to the docker directory
2. It contains `init.sql` and the subdirectories
3. The volume mount in docker-compose.yml points to the correct location

### Q: The server starts but can't connect to the database

**A:** Ensure your `config.json` has the correct database settings:

- Host should be `db` (the service name in docker-compose)
- Credentials must match the environment variables in docker-compose.yml

### Q: Container builds but `go run` fails

**A:** Check the logs with:

```bash
docker-compose logs server
```

Common issues:

- Missing or incorrect `config.json`
- Missing `bin/` directory
- Database not ready (health check should prevent this)

## Development Workflow

1. Make code changes on your host machine
2. Restart the server container to pick up changes:

   ```bash
   docker-compose restart server
   ```

3. View logs:

   ```bash
   docker-compose logs -f server
   ```

For faster development, you can run the server outside Docker and just use Docker for the database:

```bash
# Start only the database
docker-compose up -d db

# Run server locally
go run .
```
