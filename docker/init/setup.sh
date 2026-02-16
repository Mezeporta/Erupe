#!/bin/bash
set -e

echo "INIT: Restoring database schema..."
pg_restore --username="$POSTGRES_USER" --dbname="$POSTGRES_DB" --no-owner --no-acl --verbose /schemas/init.sql || {
  echo "WARN: pg_restore exited with errors (this is expected if the database already has objects)"
}

echo "Updating!"
for file in /schemas/update-schema/*; do
  echo "  Applying $file"
  psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" -1 -f "$file"
done

echo "Patching!"
for file in /schemas/patch-schema/*; do
  echo "  Applying $file"
  psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" -1 -f "$file"
done