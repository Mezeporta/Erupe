# Database Configuration

PostgreSQL database configuration and setup for Erupe.

## Configuration

```json
{
  "Database": {
    "Host": "localhost",
    "Port": 5432,
    "User": "postgres",
    "Password": "",
    "Database": "erupe"
  }
}
```

## Settings Reference

| Setting | Type | Default | Description |
|---------|------|---------|-------------|
| `Host` | string | `"localhost"` | Database host address |
| `Port` | number | `5432` | PostgreSQL port |
| `User` | string | `"postgres"` | Database user |
| `Password` | string | *required* | Database password (must not be empty) |
| `Database` | string | `"erupe"` | Database name |

**Important:** The `Password` field must not be empty. The server will refuse to start if the password is blank.

## Initial Setup

### 1. Install PostgreSQL

**Ubuntu/Debian:**

```bash
sudo apt update
sudo apt install postgresql postgresql-contrib
```

**macOS (Homebrew):**

```bash
brew install postgresql
brew services start postgresql
```

**Windows:**
Download and install from [postgresql.org](https://www.postgresql.org/download/windows/)

### 2. Create Database User

```bash
# Switch to postgres user
sudo -u postgres psql

# Create user with password
CREATE USER erupe WITH PASSWORD 'your_secure_password';

# Grant privileges
ALTER USER erupe CREATEDB;

# Exit psql
\q
```

### 3. Create Database

```bash
# Create database
createdb -U erupe erupe

# Or via psql
psql -U postgres
CREATE DATABASE erupe OWNER erupe;
\q
```

### 4. Apply Schema

From the Erupe root directory:

```bash
# Apply initial schema (bootstraps to version 9.1.0)
psql -U erupe -d erupe -f schemas/schema.sql
```

### 5. Apply Patches (Development Branch)

If using the development branch, apply patch schemas in order:

```bash
# Apply patches sequentially
psql -U erupe -d erupe -f patch-schema/01_patch.sql
psql -U erupe -d erupe -f patch-schema/02_patch.sql
# ... continue in order
```

**Note:** Patch schemas are development updates and may change. They get consolidated into update schemas on release.

### 6. (Optional) Load Bundled Data

Load demo data for shops, events, and gacha:

```bash
psql -U erupe -d erupe -f schemas/bundled-schema/shops.sql
psql -U erupe -d erupe -f schemas/bundled-schema/events.sql
psql -U erupe -d erupe -f schemas/bundled-schema/gacha.sql
```

## Schema Management

Erupe uses a multi-tiered schema system:

### Schema Types

1. **Initialization Schema** (`schemas/schema.sql`)
   - Bootstraps database to version 9.1.0
   - Creates all tables, indexes, and base data
   - Use for fresh installations

2. **Patch Schemas** (`patch-schema/*.sql`)
   - Development-time updates
   - Numbered sequentially (01, 02, 03...)
   - Applied in order during active development
   - **May change during development cycle**

3. **Update Schemas** (for releases)
   - Production-ready consolidated updates
   - Stable and tested
   - Created when patches are finalized for release

4. **Bundled Schemas** (`schemas/bundled-schema/*.sql`)
   - Optional demo/template data
   - Shops, events, gacha pools
   - Not required but helpful for testing

### Applying Schemas

**Fresh Installation:**

```bash
psql -U erupe -d erupe -f schemas/schema.sql
```

**Development (with patches):**

```bash
# First apply base schema
psql -U erupe -d erupe -f schemas/schema.sql

# Then apply patches in order
for patch in patch-schema/*.sql; do
    psql -U erupe -d erupe -f "$patch"
done
```

**With Password:**

```bash
PGPASSWORD='your_password' psql -U erupe -d erupe -f schema.sql
```

## Configuration Examples

### Local Development

```json
{
  "Database": {
    "Host": "localhost",
    "Port": 5432,
    "User": "postgres",
    "Password": "dev_password",
    "Database": "erupe_dev"
  }
}
```

### Production (Dedicated Database Server)

```json
{
  "Database": {
    "Host": "db.example.com",
    "Port": 5432,
    "User": "erupe",
    "Password": "very_secure_password_here",
    "Database": "erupe_production"
  }
}
```

### Docker Container

```json
{
  "Database": {
    "Host": "db",
    "Port": 5432,
    "User": "postgres",
    "Password": "docker_password",
    "Database": "erupe"
  }
}
```

The `Host: "db"` refers to the Docker Compose service name.

## Docker Setup

Using Docker Compose (see `docker-compose.yml`):

```yaml
services:
  db:
    image: postgres:13
    environment:
      POSTGRES_PASSWORD: test
      POSTGRES_DB: erupe
    ports:
      - "5432:5432"
    volumes:
      - pgdata:/var/lib/postgresql/data
```

Start database:

```bash
docker compose up db
```

Apply schema to Docker database:

```bash
docker compose exec db psql -U postgres -d erupe -f /path/to/schema.sql
```

## Database Maintenance

### Backup Database

```bash
# Full backup
pg_dump -U erupe erupe > erupe_backup.sql

# Compressed backup
pg_dump -U erupe erupe | gzip > erupe_backup.sql.gz

# With password
PGPASSWORD='password' pg_dump -U erupe erupe > backup.sql
```

### Restore Database

```bash
# Drop and recreate database
dropdb -U erupe erupe
createdb -U erupe erupe

# Restore from backup
psql -U erupe -d erupe -f erupe_backup.sql

# From compressed backup
gunzip -c erupe_backup.sql.gz | psql -U erupe -d erupe
```

### Clean Database (Development)

```bash
# Connect to database
psql -U erupe -d erupe

-- Delete all user data
DELETE FROM guild_characters;
DELETE FROM guilds;
DELETE FROM characters;
DELETE FROM sign_sessions;
DELETE FROM users;

-- Exit
\q
```

Or use `CleanDB: true` in [Development Mode](development-mode.md) (⚠️ destructive!).

### Check Database Size

```bash
psql -U erupe -d erupe -c "SELECT pg_size_pretty(pg_database_size('erupe'));"
```

### Vacuum Database

Reclaim space and optimize:

```bash
psql -U erupe -d erupe -c "VACUUM ANALYZE;"
```

## Troubleshooting

### Connection Refused

**Error:** `could not connect to server: Connection refused`

**Solutions:**

- Verify PostgreSQL is running: `sudo systemctl status postgresql`
- Check port is correct: `5432` (default)
- Verify host is accessible
- Check firewall rules

### Authentication Failed

**Error:** `password authentication failed for user "erupe"`

**Solutions:**

- Verify password is correct in config
- Check `pg_hba.conf` authentication method
- Ensure user exists: `psql -U postgres -c "\du"`

### Database Does Not Exist

**Error:** `database "erupe" does not exist`

**Solutions:**

- Create database: `createdb -U erupe erupe`
- Verify database name matches config

### Permission Denied

**Error:** `permission denied for table users`

**Solutions:**

```sql
-- Grant all privileges on database
GRANT ALL PRIVILEGES ON DATABASE erupe TO erupe;

-- Grant all privileges on all tables
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO erupe;
```

### Schema Version Mismatch

**Error:** Server starts but data doesn't load properly

**Solutions:**

- Check if all patches were applied in order
- Verify schema version in database
- Re-apply missing patches

## Security Best Practices

1. **Use Strong Passwords**: Never use default or weak passwords
2. **Limit Network Access**: Use firewall rules to restrict database access
3. **Don't Expose PostgreSQL Publicly**: Only allow connections from Erupe server
4. **Use SSL/TLS**: Enable SSL for production databases
5. **Regular Backups**: Automate daily backups
6. **Separate Users**: Don't use `postgres` superuser for Erupe

## Performance Tuning

For larger servers, optimize PostgreSQL:

### Connection Pooling

Consider using PgBouncer for connection pooling:

```bash
sudo apt install pgbouncer
```

### PostgreSQL Configuration

Edit `/etc/postgresql/13/main/postgresql.conf`:

```conf
# Increase shared buffers (25% of RAM)
shared_buffers = 2GB

# Increase work memory
work_mem = 16MB

# Increase maintenance work memory
maintenance_work_mem = 512MB

# Enable query logging (development)
log_statement = 'all'
log_duration = on
```

Restart PostgreSQL:

```bash
sudo systemctl restart postgresql
```

## Related Documentation

- [Basic Settings](basic-settings.md) - Basic server configuration
- [Development Mode](development-mode.md) - CleanDB option
- [Server Configuration](server-configuration.md) - Server setup
- [CLAUDE.md](../CLAUDE.md#database-operations) - Database operations guide
